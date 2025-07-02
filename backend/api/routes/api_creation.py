"""
API creation routes for the console
"""
from fastapi import APIRouter, HTTPException, Depends, Request, Query, Body
from typing import Dict, Any, List, Optional
from datetime import datetime
import uuid
import json

from ..database import get_db
from ..auth import get_current_user
from ..github_integration import github_integration
from ..models import User

router = APIRouter(prefix="/api/creation", tags=["api_creation"])

# In-memory storage for demo (replace with database)
draft_apis = {}
api_templates = {
    "gpt-wrapper": {
        "id": "gpt-wrapper",
        "name": "GPT Wrapper API",
        "description": "Production-ready OpenAI GPT wrapper with caching and rate limiting",
        "category": "AI/ML",
        "language": "Python",
        "runtime": "python3.11",
        "framework": "FastAPI",
        "features": ["Response caching", "Rate limiting", "Cost optimization", "Error handling", "Usage analytics"],
        "main_file": "main.py",
        "start_command": "uvicorn main:app --host 0.0.0.0 --port 8080",
        "endpoints": [
            {"method": "POST", "path": "/chat/completions"},
            {"method": "GET", "path": "/models"},
            {"method": "GET", "path": "/usage"},
            {"method": "GET", "path": "/health"}
        ],
        "environment": {
            "required": ["OPENAI_API_KEY"],
            "optional": {
                "CACHE_TTL": "3600",
                "MAX_TOKENS": "2048",
                "DEFAULT_MODEL": "gpt-3.5-turbo"
            }
        }
    },
    "image-classifier": {
        "id": "image-classifier",
        "name": "Image Classification API",
        "description": "Computer vision API using pre-trained Vision Transformer models",
        "category": "AI/ML",
        "language": "Python",
        "runtime": "python3.11",
        "framework": "FastAPI",
        "features": ["Vision Transformer models", "Multi-format support", "Batch processing", "GPU optimization", "Confidence scoring"],
        "main_file": "main.py",
        "start_command": "uvicorn main:app --host 0.0.0.0 --port 8080",
        "endpoints": [
            {"method": "POST", "path": "/classify"},
            {"method": "POST", "path": "/batch/classify"},
            {"method": "GET", "path": "/models"},
            {"method": "GET", "path": "/health"}
        ],
        "environment": {
            "required": ["MODEL_PATH"],
            "optional": {
                "USE_GPU": "false",
                "BATCH_SIZE": "32",
                "IMAGE_SIZE": "224"
            }
        }
    },
    "basic-rest": {
        "id": "basic-rest",
        "name": "Basic REST API",
        "description": "Simple REST API with CRUD operations",
        "category": "Web API",
        "language": "Python",
        "runtime": "python3.11",
        "framework": "FastAPI",
        "features": ["REST endpoints", "JSON responses", "Basic validation"],
        "main_file": "main.py",
        "start_command": "uvicorn main:app --host 0.0.0.0 --port 8080",
        "endpoints": [
            {"method": "GET", "path": "/items"},
            {"method": "POST", "path": "/items"},
            {"method": "GET", "path": "/items/{id}"},
            {"method": "PUT", "path": "/items/{id}"},
            {"method": "DELETE", "path": "/items/{id}"},
            {"method": "GET", "path": "/health"}
        ],
        "environment": {
            "required": [],
            "optional": {
                "API_VERSION": "v1",
                "LOG_LEVEL": "info"
            }
        }
    }
}

@router.post("/github/oauth/callback")
async def github_oauth_callback(
    code: str = Query(..., description="OAuth authorization code"),
    state: str = Query(..., description="OAuth state parameter"),
    current_user: User = Depends(get_current_user)
):
    """Handle GitHub OAuth callback"""
    try:
        # Exchange code for access token
        token_data = await github_integration.exchange_code_for_token(code)
        access_token = token_data.get("access_token")
        
        if not access_token:
            raise HTTPException(status_code=400, detail="Failed to get access token")
        
        # Get user info
        github_user = await github_integration.get_user_info(access_token)
        
        # Store the GitHub connection (in production, store in database)
        # For now, return the token to the frontend
        return {
            "access_token": access_token,
            "github_user": {
                "login": github_user.get("login"),
                "name": github_user.get("name"),
                "avatar_url": github_user.get("avatar_url")
            }
        }
        
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@router.get("/github/repos")
async def list_github_repos(
    access_token: str = Query(..., description="GitHub access token"),
    page: int = Query(1, ge=1),
    per_page: int = Query(30, ge=1, le=100),
    current_user: User = Depends(get_current_user)
):
    """List user's GitHub repositories"""
    try:
        repos = await github_integration.list_user_repos(access_token, page, per_page)
        
        # Transform repo data for frontend
        return {
            "repos": [
                {
                    "id": repo.get("id"),
                    "name": repo.get("name"),
                    "full_name": repo.get("full_name"),
                    "description": repo.get("description"),
                    "private": repo.get("private"),
                    "language": repo.get("language"),
                    "clone_url": repo.get("clone_url"),
                    "html_url": repo.get("html_url"),
                    "default_branch": repo.get("default_branch"),
                    "updated_at": repo.get("updated_at"),
                    "size": repo.get("size"),
                    "stargazers_count": repo.get("stargazers_count")
                }
                for repo in repos
            ],
            "page": page,
            "per_page": per_page
        }
        
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@router.post("/github/analyze")
async def analyze_github_repo(
    data: Dict[str, Any] = Body(...),
    current_user: User = Depends(get_current_user)
):
    """Analyze a GitHub repository"""
    try:
        access_token = data.get("access_token")
        repo_full_name = data.get("repo_full_name")
        clone_url = data.get("clone_url")
        
        if not all([access_token, repo_full_name, clone_url]):
            raise HTTPException(status_code=400, detail="Missing required fields")
        
        # Get repository info
        owner, repo = repo_full_name.split("/")
        repo_info = await github_integration.get_repo_info(access_token, owner, repo)
        
        # Clone and analyze the repository
        analysis = await github_integration.clone_and_analyze_repo(clone_url, access_token)
        
        # Create draft API configuration
        api_config = await github_integration.create_api_from_analysis(
            str(current_user.id),
            repo_info,
            analysis
        )
        
        # Store draft (in production, store in database)
        draft_id = str(uuid.uuid4())
        draft_apis[draft_id] = {
            "id": draft_id,
            "user_id": str(current_user.id),
            "config": api_config,
            "analysis": analysis,
            "repo_info": repo_info,
            "created_at": datetime.utcnow().isoformat(),
            "status": "draft"
        }
        
        return {
            "draft_id": draft_id,
            "analysis": analysis,
            "config": api_config
        }
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/templates")
async def list_templates(
    category: Optional[str] = Query(None, description="Filter by category"),
    current_user: User = Depends(get_current_user)
):
    """List available API templates"""
    templates = list(api_templates.values())
    
    if category:
        templates = [t for t in templates if t.get("category") == category]
    
    return {
        "templates": templates,
        "categories": ["AI/ML", "Web API", "Database API", "Integration", "Authentication", "GraphQL", "Microservice"]
    }

@router.get("/templates/{template_id}")
async def get_template(
    template_id: str,
    current_user: User = Depends(get_current_user)
):
    """Get a specific template"""
    template = api_templates.get(template_id)
    
    if not template:
        raise HTTPException(status_code=404, detail="Template not found")
    
    return template

@router.post("/templates/{template_id}/scaffold")
async def scaffold_from_template(
    template_id: str,
    data: Dict[str, Any] = Body(...),
    current_user: User = Depends(get_current_user)
):
    """Create API from template"""
    template = api_templates.get(template_id)
    
    if not template:
        raise HTTPException(status_code=404, detail="Template not found")
    
    # Create API configuration from template
    api_config = {
        "name": data.get("name", template["name"]),
        "description": data.get("description", template["description"]),
        "user_id": str(current_user.id),
        "source": {
            "type": "template",
            "template_id": template_id
        },
        "runtime": {
            "language": template["language"],
            "version": template["runtime"],
            "framework": template["framework"]
        },
        "build": {
            "main_file": template["main_file"],
            "start_command": template["start_command"],
            "port": data.get("port", 8080),
            "health_check": "/health"
        },
        "environment": template["environment"],
        "endpoints": template["endpoints"],
        "created_at": datetime.utcnow().isoformat(),
        "status": "draft"
    }
    
    # Store draft
    draft_id = str(uuid.uuid4())
    draft_apis[draft_id] = {
        "id": draft_id,
        "user_id": str(current_user.id),
        "config": api_config,
        "template": template,
        "created_at": datetime.utcnow().isoformat(),
        "status": "draft"
    }
    
    return {
        "draft_id": draft_id,
        "config": api_config
    }

@router.get("/drafts/{draft_id}")
async def get_draft(
    draft_id: str,
    current_user: User = Depends(get_current_user)
):
    """Get a draft API configuration"""
    draft = draft_apis.get(draft_id)
    
    if not draft:
        raise HTTPException(status_code=404, detail="Draft not found")
    
    if draft["user_id"] != str(current_user.id):
        raise HTTPException(status_code=403, detail="Access denied")
    
    return draft

@router.put("/drafts/{draft_id}")
async def update_draft(
    draft_id: str,
    updates: Dict[str, Any] = Body(...),
    current_user: User = Depends(get_current_user)
):
    """Update a draft API configuration"""
    draft = draft_apis.get(draft_id)
    
    if not draft:
        raise HTTPException(status_code=404, detail="Draft not found")
    
    if draft["user_id"] != str(current_user.id):
        raise HTTPException(status_code=403, detail="Access denied")
    
    # Update the configuration
    if "config" in updates:
        draft["config"].update(updates["config"])
    
    draft["updated_at"] = datetime.utcnow().isoformat()
    
    return draft

@router.post("/drafts/{draft_id}/deploy")
async def deploy_draft(
    draft_id: str,
    current_user: User = Depends(get_current_user),
    db = Depends(get_db)
):
    """Deploy a draft API"""
    draft = draft_apis.get(draft_id)
    
    if not draft:
        raise HTTPException(status_code=404, detail="Draft not found")
    
    if draft["user_id"] != str(current_user.id):
        raise HTTPException(status_code=403, detail="Access denied")
    
    try:
        # In production, this would trigger the actual deployment process
        # For now, we'll simulate it
        api_id = str(uuid.uuid4())
        
        # Create API record
        api_data = {
            "id": api_id,
            "user_id": draft["user_id"],
            "name": draft["config"]["name"],
            "description": draft["config"]["description"],
            "config": draft["config"],
            "status": "deploying",
            "created_at": datetime.utcnow().isoformat()
        }
        
        # Remove draft
        del draft_apis[draft_id]
        
        return {
            "api_id": api_id,
            "status": "deploying",
            "message": "API deployment started"
        }
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/upload/analyze")
async def analyze_uploaded_file(
    request: Request,
    current_user: User = Depends(get_current_user)
):
    """Analyze an uploaded ZIP file"""
    # This would handle file upload and analysis
    # For now, return a mock response
    return {
        "draft_id": str(uuid.uuid4()),
        "analysis": {
            "language": "Python",
            "runtime": "python3.11",
            "framework": "FastAPI",
            "main_file": "main.py",
            "endpoints": [
                {"method": "GET", "path": "/"},
                {"method": "GET", "path": "/health"}
            ],
            "environment": {
                "required": ["API_KEY"],
                "optional": {"PORT": "8080"}
            }
        }
    }

@router.get("/github/oauth/url")
async def get_github_oauth_url(
    redirect_uri: str = Query(..., description="OAuth redirect URI"),
    current_user: User = Depends(get_current_user)
):
    """Get GitHub OAuth URL"""
    state = str(uuid.uuid4())
    oauth_url = github_integration.get_oauth_url(redirect_uri, state)
    
    return {
        "oauth_url": oauth_url,
        "state": state
    }