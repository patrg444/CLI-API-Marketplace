"""
Real deployment routes for API Direct
Handles actual Lambda deployments
"""
from fastapi import APIRouter, HTTPException, Depends, UploadFile, File, BackgroundTasks
from fastapi.security import HTTPAuthorizationCredentials
from typing import Dict, Any, Optional
import os
import tempfile
import shutil
import zipfile
import json
from datetime import datetime
import boto3
from pydantic import BaseModel

from ..auth.mock_auth import MockAuthService, validate_api_key
from ..deployment_handler import DeploymentHandler
from ..database import get_db, Deployment, get_db_manager
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select

router = APIRouter()

# Initialize deployment handler
deployment_handler = DeploymentHandler()

# S3 client for code storage
s3_client = boto3.client('s3')
CODE_BUCKET = os.getenv('CODE_STORAGE_BUCKET', 'apidirect-code-storage-e6dce744')

class DeploymentRequest(BaseModel):
    api_name: str
    runtime: str = "python3.9"
    environment_variables: Dict[str, str] = {}

class DeploymentResponse(BaseModel):
    deployment_id: str
    api_id: str
    status: str
    endpoint: str
    created_at: str

@router.post("/deploy/upload")
async def upload_code(
    file: UploadFile = File(...),
    api_key: str = Depends(validate_api_key)
):
    """Upload code package for deployment"""
    try:
        # Create temp directory
        temp_dir = tempfile.mkdtemp()
        zip_path = os.path.join(temp_dir, "upload.zip")
        
        # Save uploaded file
        with open(zip_path, "wb") as f:
            content = await file.read()
            f.write(content)
        
        # Extract to verify it's a valid zip
        extract_dir = os.path.join(temp_dir, "extracted")
        os.makedirs(extract_dir)
        
        with zipfile.ZipFile(zip_path, 'r') as zip_ref:
            zip_ref.extractall(extract_dir)
        
        # Upload to S3
        s3_key = f"uploads/{api_key['user_id']}/{datetime.utcnow().isoformat()}-{file.filename}"
        s3_client.upload_file(zip_path, CODE_BUCKET, s3_key)
        
        # Clean up
        shutil.rmtree(temp_dir)
        
        return {
            "status": "uploaded",
            "s3_key": s3_key,
            "size": len(content),
            "filename": file.filename
        }
        
    except Exception as e:
        raise HTTPException(status_code=400, detail=f"Upload failed: {str(e)}")

@router.post("/deploy", response_model=DeploymentResponse)
async def deploy_api(
    deployment: DeploymentRequest,
    background_tasks: BackgroundTasks,
    api_key: str = Depends(validate_api_key),
    db: AsyncSession = Depends(get_db)
):
    """Deploy an API to AWS Lambda"""
    try:
        user_id = api_key['user_id']
        
        # Create deployment record
        db_deployment = Deployment(
            api_id=deployment.api_name,
            user_id=user_id,
            status="deploying",
            environment="production"
        )
        db.add(db_deployment)
        await db.commit()
        await db.refresh(db_deployment)
        
        # Deploy to Lambda (in background)
        background_tasks.add_task(
            deploy_to_lambda,
            str(db_deployment.id),
            user_id,
            deployment.api_name,
            deployment.runtime
        )
        
        return DeploymentResponse(
            deployment_id=str(db_deployment.id),
            api_id=deployment.api_name,
            status="deploying",
            endpoint="https://api.apidirect.dev/gateway/pending",
            created_at=datetime.utcnow().isoformat()
        )
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Deployment failed: {str(e)}")

async def deploy_to_lambda(
    deployment_id: str,
    user_id: str,
    api_name: str,
    runtime: str
):
    """Background task to deploy to Lambda"""
    db_manager = get_db_manager()
    async with db_manager.get_session() as db:
        try:
            # Get latest uploaded code
            response = s3_client.list_objects_v2(
                Bucket=CODE_BUCKET,
                Prefix=f"uploads/{user_id}/",
                MaxKeys=1
            )
            
            if 'Contents' not in response:
                raise Exception("No code uploaded")
            
            latest_upload = response['Contents'][0]['Key']
            
            # Download code
            temp_dir = tempfile.mkdtemp()
            zip_path = os.path.join(temp_dir, "code.zip")
            s3_client.download_file(CODE_BUCKET, latest_upload, zip_path)
            
            # Extract code
            extract_dir = os.path.join(temp_dir, "code")
            os.makedirs(extract_dir)
            with zipfile.ZipFile(zip_path, 'r') as zip_ref:
                zip_ref.extractall(extract_dir)
            
            # Deploy using handler
            result = deployment_handler.deploy_api(
                user_id=user_id,
                api_name=api_name,
                code_path=extract_dir,
                runtime=runtime
            )
            
            # Update deployment record
            stmt = select(Deployment).where(Deployment.id == deployment_id)
            result_db = await db.execute(stmt)
            deployment = result_db.scalar_one_or_none()
            if deployment:
                deployment.status = "deployed"
                deployment.endpoint = result['api_endpoint']
                deployment.deployment_metadata = result
                await db.commit()
            
            # Clean up
            shutil.rmtree(temp_dir)
            
        except Exception as e:
            # Update deployment as failed
            stmt = select(Deployment).where(Deployment.id == deployment_id)
            result_db = await db.execute(stmt)
            deployment = result_db.scalar_one_or_none()
            if deployment:
                deployment.status = "failed"
                deployment.error = str(e)
                await db.commit()

@router.get("/deployments")
async def list_deployments(
    api_key: str = Depends(validate_api_key),
    db: AsyncSession = Depends(get_db)
):
    """List all deployments for the authenticated user"""
    user_id = api_key['user_id']
    
    stmt = select(Deployment).where(
        Deployment.user_id == user_id
    ).order_by(Deployment.created_at.desc())
    result = await db.execute(stmt)
    deployments = result.scalars().all()
    
    return {
        "deployments": [
            {
                "id": str(d.id),
                "api_id": d.api_id,
                "status": d.status,
                "endpoint": d.endpoint,
                "environment": d.environment,
                "created_at": d.created_at.isoformat()
            }
            for d in deployments
        ]
    }

@router.get("/deployments/{deployment_id}")
async def get_deployment(
    deployment_id: str,
    api_key: str = Depends(validate_api_key),
    db: AsyncSession = Depends(get_db)
):
    """Get deployment details"""
    user_id = api_key['user_id']
    
    stmt = select(Deployment).where(
        Deployment.id == deployment_id,
        Deployment.user_id == user_id
    )
    result = await db.execute(stmt)
    deployment = result.scalar_one_or_none()
    
    if not deployment:
        raise HTTPException(status_code=404, detail="Deployment not found")
    
    # Get Lambda status if deployed
    lambda_status = {}
    if deployment.status == "deployed" and deployment.deployment_metadata:
        metadata = deployment.deployment_metadata if isinstance(deployment.deployment_metadata, dict) else json.loads(deployment.deployment_metadata or '{}')
        if 'function_name' in metadata:
            lambda_status = deployment_handler.get_deployment_status(
                metadata['function_name']
            )
    
    return {
        "id": str(deployment.id),
        "api_id": deployment.api_id,
        "status": deployment.status,
        "endpoint": deployment.endpoint,
        "environment": deployment.environment,
        "created_at": deployment.created_at.isoformat(),
        "lambda_status": lambda_status,
        "error": deployment.error
    }

@router.delete("/deployments/{deployment_id}")
async def delete_deployment(
    deployment_id: str,
    api_key: str = Depends(validate_api_key),
    db: AsyncSession = Depends(get_db)
):
    """Delete a deployment"""
    user_id = api_key['user_id']
    
    stmt = select(Deployment).where(
        Deployment.id == deployment_id,
        Deployment.user_id == user_id
    )
    result = await db.execute(stmt)
    deployment = result.scalar_one_or_none()
    
    if not deployment:
        raise HTTPException(status_code=404, detail="Deployment not found")
    
    # Delete from Lambda if deployed
    if deployment.status == "deployed" and deployment.deployment_metadata:
        metadata = deployment.deployment_metadata if isinstance(deployment.deployment_metadata, dict) else json.loads(deployment.deployment_metadata or '{}')
        success = deployment_handler.delete_deployment(
            function_name=metadata.get('function_name'),
            api_id=metadata.get('api_id')
        )
        
        if not success:
            raise HTTPException(status_code=500, detail="Failed to delete Lambda function")
    
    # Delete record
    await db.delete(deployment)
    await db.commit()
    
    return {"status": "deleted"}