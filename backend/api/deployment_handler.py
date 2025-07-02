"""
Real deployment handler for API Direct
Deploys user APIs as AWS Lambda functions
"""
import os
import json
import boto3
import zipfile
import tempfile
import subprocess
from typing import Dict, Any
from datetime import datetime
import uuid
from fastapi import HTTPException

class DeploymentHandler:
    def __init__(self):
        region = os.getenv('AWS_REGION', 'us-east-1')
        self.s3_client = boto3.client('s3', region_name=region)
        self.lambda_client = boto3.client('lambda', region_name=region)
        self.api_gateway = boto3.client('apigatewayv2', region_name=region)
        self.iam_client = boto3.client('iam', region_name=region)
        
        # Buckets from environment
        self.code_bucket = os.getenv('CODE_STORAGE_BUCKET', 'apidirect-code-storage-e6dce744')
        self.lambda_role_arn = None
        self._ensure_lambda_role()
    
    def _ensure_lambda_role(self):
        """Create or get IAM role for Lambda execution"""
        role_name = 'apidirect-lambda-execution-role'
        
        try:
            # Check if role exists
            response = self.iam_client.get_role(RoleName=role_name)
            self.lambda_role_arn = response['Role']['Arn']
        except self.iam_client.exceptions.NoSuchEntityException:
            # Create role
            trust_policy = {
                "Version": "2012-10-17",
                "Statement": [{
                    "Effect": "Allow",
                    "Principal": {"Service": "lambda.amazonaws.com"},
                    "Action": "sts:AssumeRole"
                }]
            }
            
            response = self.iam_client.create_role(
                RoleName=role_name,
                AssumeRolePolicyDocument=json.dumps(trust_policy),
                Description='Execution role for API Direct Lambda functions'
            )
            
            # Attach basic execution policy
            self.iam_client.attach_role_policy(
                RoleName=role_name,
                PolicyArn='arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole'
            )
            
            self.lambda_role_arn = response['Role']['Arn']
            
            # Wait for role to be ready
            import time
            time.sleep(10)
    
    def deploy_api(self, user_id: str, api_name: str, code_path: str, runtime: str = "python3.9") -> Dict[str, Any]:
        """Deploy an API to AWS Lambda"""
        
        # Generate unique function name
        function_name = f"apidirect-{user_id}-{api_name}-{uuid.uuid4().hex[:8]}"
        
        try:
            # 1. Package the code
            zip_path = self._package_code(code_path)
            
            # 2. Upload to S3
            s3_key = f"deployments/{user_id}/{api_name}/{datetime.utcnow().isoformat()}.zip"
            self.s3_client.upload_file(zip_path, self.code_bucket, s3_key)
            
            # 3. Create Lambda function
            lambda_response = self.lambda_client.create_function(
                FunctionName=function_name,
                Runtime=runtime,
                Role=self.lambda_role_arn,
                Handler='handler.lambda_handler',  # We'll wrap their code
                Code={
                    'S3Bucket': self.code_bucket,
                    'S3Key': s3_key
                },
                Description=f'API Direct: {api_name}',
                Timeout=30,
                MemorySize=512,
                Environment={
                    'Variables': {
                        'API_NAME': api_name,
                        'USER_ID': user_id
                    }
                }
            )
            
            # 4. Create API Gateway
            api_response = self.api_gateway.create_api(
                Name=f"{api_name}-{user_id}",
                ProtocolType='HTTP',
                Version='1.0',
                Description=f'API Direct gateway for {api_name}'
            )
            
            api_id = api_response['ApiId']
            api_endpoint = api_response['ApiEndpoint']
            
            # 5. Create Lambda integration
            region = os.getenv('AWS_REGION', 'us-east-1')
            integration_response = self.api_gateway.create_integration(
                ApiId=api_id,
                IntegrationType='AWS_PROXY',
                IntegrationUri=f"arn:aws:lambda:{region}:{lambda_response['FunctionArn'].split(':')[4]}:function:{function_name}",
                PayloadFormatVersion='2.0'
            )
            
            # 6. Create route
            self.api_gateway.create_route(
                ApiId=api_id,
                RouteKey='$default',
                Target=f"integrations/{integration_response['IntegrationId']}"
            )
            
            # 7. Create deployment stage
            self.api_gateway.create_stage(
                ApiId=api_id,
                StageName='prod',
                AutoDeploy=True
            )
            
            # 8. Add Lambda permission for API Gateway
            self.lambda_client.add_permission(
                FunctionName=function_name,
                StatementId=f'apigateway-{api_id}',
                Action='lambda:InvokeFunction',
                Principal='apigateway.amazonaws.com',
                SourceArn=f"arn:aws:execute-api:{region}:*:{api_id}/*/*"
            )
            
            # Clean up temp file
            os.remove(zip_path)
            
            return {
                'status': 'deployed',
                'function_name': function_name,
                'api_endpoint': f"{api_endpoint}/prod",
                'api_id': api_id,
                'deployment_time': datetime.utcnow().isoformat()
            }
            
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Deployment failed: {str(e)}")
    
    def _package_code(self, code_path: str) -> str:
        """Package code into Lambda-compatible zip"""
        
        # Create temp zip file
        temp_zip = tempfile.mktemp(suffix='.zip')
        
        with zipfile.ZipFile(temp_zip, 'w', zipfile.ZIP_DEFLATED) as zf:
            # Add user's code
            for root, dirs, files in os.walk(code_path):
                for file in files:
                    file_path = os.path.join(root, file)
                    arcname = os.path.relpath(file_path, code_path)
                    zf.write(file_path, arcname)
            
            # Add Lambda handler wrapper
            handler_code = '''
import json
import sys
import os

# Add current directory to path
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

# Import user's main module
try:
    from main import app  # FastAPI app
    from mangum import Mangum
    lambda_handler = Mangum(app)
except ImportError:
    # Fallback for simple functions
    from main import handler as user_handler
    
    def lambda_handler(event, context):
        return {
            'statusCode': 200,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps(user_handler(event, context))
        }
'''
            zf.writestr('handler.py', handler_code)
            
            # Add requirements if needed
            if os.path.exists(os.path.join(code_path, 'requirements.txt')):
                # Install dependencies to a temp directory
                temp_deps = tempfile.mkdtemp()
                subprocess.run([
                    'pip', 'install', '-r', 
                    os.path.join(code_path, 'requirements.txt'),
                    '-t', temp_deps,
                    '--platform', 'manylinux2014_x86_64',
                    '--only-binary=:all:'
                ], check=True)
                
                # Add dependencies to zip
                for root, dirs, files in os.walk(temp_deps):
                    for file in files:
                        file_path = os.path.join(root, file)
                        arcname = os.path.relpath(file_path, temp_deps)
                        zf.write(file_path, arcname)
        
        return temp_zip
    
    def get_deployment_status(self, function_name: str) -> Dict[str, Any]:
        """Get status of a deployed function"""
        try:
            response = self.lambda_client.get_function(FunctionName=function_name)
            return {
                'status': response['Configuration']['State'],
                'last_modified': response['Configuration']['LastModified'],
                'runtime': response['Configuration']['Runtime'],
                'memory': response['Configuration']['MemorySize'],
                'timeout': response['Configuration']['Timeout']
            }
        except self.lambda_client.exceptions.ResourceNotFoundException:
            return {'status': 'not_found'}
    
    def delete_deployment(self, function_name: str, api_id: str) -> bool:
        """Delete a deployment"""
        try:
            # Delete Lambda function
            self.lambda_client.delete_function(FunctionName=function_name)
            
            # Delete API Gateway
            self.api_gateway.delete_api(ApiId=api_id)
            
            return True
        except Exception as e:
            print(f"Error deleting deployment: {e}")
            return False