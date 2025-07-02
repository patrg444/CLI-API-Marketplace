"""
GitHub integration for API creation workflow
"""
import os
import json
import base64
import asyncio
from typing import Dict, List, Optional, Any
from datetime import datetime, timedelta
import aiohttp
from fastapi import HTTPException
import subprocess
import tempfile
import shutil
from pathlib import Path

class GitHubIntegration:
    def __init__(self):
        self.github_api_url = "https://api.github.com"
        self.oauth_client_id = os.environ.get("GITHUB_OAUTH_CLIENT_ID")
        self.oauth_client_secret = os.environ.get("GITHUB_OAUTH_CLIENT_SECRET")
        
    async def exchange_code_for_token(self, code: str) -> Dict[str, Any]:
        """Exchange OAuth code for access token"""
        async with aiohttp.ClientSession() as session:
            data = {
                "client_id": self.oauth_client_id,
                "client_secret": self.oauth_client_secret,
                "code": code
            }
            
            async with session.post(
                "https://github.com/login/oauth/access_token",
                json=data,
                headers={"Accept": "application/json"}
            ) as response:
                if response.status != 200:
                    raise HTTPException(status_code=400, detail="Failed to exchange code for token")
                
                result = await response.json()
                if "error" in result:
                    raise HTTPException(status_code=400, detail=result.get("error_description", "OAuth error"))
                
                return result
    
    async def get_user_info(self, access_token: str) -> Dict[str, Any]:
        """Get authenticated user info"""
        async with aiohttp.ClientSession() as session:
            headers = {
                "Authorization": f"Bearer {access_token}",
                "Accept": "application/vnd.github.v3+json"
            }
            
            async with session.get(f"{self.github_api_url}/user", headers=headers) as response:
                if response.status != 200:
                    raise HTTPException(status_code=401, detail="Failed to get user info")
                
                return await response.json()
    
    async def list_user_repos(self, access_token: str, page: int = 1, per_page: int = 30) -> List[Dict[str, Any]]:
        """List user's repositories"""
        async with aiohttp.ClientSession() as session:
            headers = {
                "Authorization": f"Bearer {access_token}",
                "Accept": "application/vnd.github.v3+json"
            }
            
            params = {
                "page": page,
                "per_page": per_page,
                "sort": "updated",
                "direction": "desc"
            }
            
            async with session.get(
                f"{self.github_api_url}/user/repos",
                headers=headers,
                params=params
            ) as response:
                if response.status != 200:
                    raise HTTPException(status_code=401, detail="Failed to list repositories")
                
                return await response.json()
    
    async def get_repo_info(self, access_token: str, owner: str, repo: str) -> Dict[str, Any]:
        """Get detailed repository information"""
        async with aiohttp.ClientSession() as session:
            headers = {
                "Authorization": f"Bearer {access_token}",
                "Accept": "application/vnd.github.v3+json"
            }
            
            async with session.get(
                f"{self.github_api_url}/repos/{owner}/{repo}",
                headers=headers
            ) as response:
                if response.status != 200:
                    raise HTTPException(status_code=404, detail="Repository not found")
                
                return await response.json()
    
    async def get_repo_languages(self, access_token: str, owner: str, repo: str) -> Dict[str, int]:
        """Get repository languages"""
        async with aiohttp.ClientSession() as session:
            headers = {
                "Authorization": f"Bearer {access_token}",
                "Accept": "application/vnd.github.v3+json"
            }
            
            async with session.get(
                f"{self.github_api_url}/repos/{owner}/{repo}/languages",
                headers=headers
            ) as response:
                if response.status != 200:
                    return {}
                
                return await response.json()
    
    async def get_repo_structure(self, access_token: str, owner: str, repo: str, path: str = "") -> List[Dict[str, Any]]:
        """Get repository file structure"""
        async with aiohttp.ClientSession() as session:
            headers = {
                "Authorization": f"Bearer {access_token}",
                "Accept": "application/vnd.github.v3+json"
            }
            
            async with session.get(
                f"{self.github_api_url}/repos/{owner}/{repo}/contents/{path}",
                headers=headers
            ) as response:
                if response.status != 200:
                    return []
                
                return await response.json()
    
    async def clone_and_analyze_repo(self, clone_url: str, access_token: Optional[str] = None) -> Dict[str, Any]:
        """Clone repository and analyze it using the CLI detector"""
        temp_dir = None
        
        try:
            # Create temporary directory
            temp_dir = tempfile.mkdtemp(prefix="api-direct-")
            repo_path = os.path.join(temp_dir, "repo")
            
            # Clone the repository
            clone_command = ["git", "clone", "--depth", "1"]
            
            # If private repo, use token in URL
            if access_token and clone_url.startswith("https://"):
                clone_url = clone_url.replace("https://", f"https://{access_token}@")
            
            clone_command.extend([clone_url, repo_path])
            
            # Run git clone
            result = subprocess.run(
                clone_command,
                capture_output=True,
                text=True,
                timeout=60  # 60 second timeout
            )
            
            if result.returncode != 0:
                raise HTTPException(
                    status_code=400,
                    detail=f"Failed to clone repository: {result.stderr}"
                )
            
            # Run the CLI detector on the cloned repo
            analysis = await self.run_detector_analysis(repo_path)
            
            return analysis
            
        except subprocess.TimeoutExpired:
            raise HTTPException(status_code=408, detail="Repository clone timed out")
        except Exception as e:
            raise HTTPException(status_code=500, detail=str(e))
        finally:
            # Clean up temporary directory
            if temp_dir and os.path.exists(temp_dir):
                shutil.rmtree(temp_dir, ignore_errors=True)
    
    async def run_detector_analysis(self, repo_path: str) -> Dict[str, Any]:
        """Run the CLI detector on a repository path"""
        try:
            # Import the detector module
            import sys
            cli_path = str(Path(__file__).parent.parent.parent / "cli")
            if cli_path not in sys.path:
                sys.path.insert(0, cli_path)
            
            from pkg.detector.detector import AnalyzeProject
            
            # Run the detector
            detection = AnalyzeProject(repo_path)
            
            # Convert to dictionary format
            analysis = {
                "language": detection.Language,
                "runtime": detection.Runtime,
                "framework": detection.Framework,
                "main_file": detection.MainFile,
                "start_command": detection.StartCommand,
                "port": detection.Port,
                "health_check": detection.HealthCheck,
                "requirements_file": detection.RequirementsFile,
                "env_file": detection.EnvFile,
                "endpoints": [
                    {"method": ep.Method, "path": ep.Path}
                    for ep in detection.Endpoints
                ],
                "environment": {
                    "required": detection.Environment.Required,
                    "optional": detection.Environment.Optional
                }
            }
            
            # Add some metadata
            analysis["detected_at"] = datetime.utcnow().isoformat()
            analysis["detection_method"] = "cli_detector"
            
            return analysis
            
        except ImportError:
            # Fallback to basic detection if CLI detector not available
            return await self.basic_repo_analysis(repo_path)
        except Exception as e:
            # Log error but return basic analysis
            print(f"Detector analysis failed: {e}")
            return await self.basic_repo_analysis(repo_path)
    
    async def basic_repo_analysis(self, repo_path: str) -> Dict[str, Any]:
        """Basic repository analysis fallback"""
        analysis = {
            "language": "Unknown",
            "runtime": "unknown",
            "framework": "Unknown",
            "main_file": "",
            "start_command": "",
            "port": 8080,
            "health_check": "/health",
            "requirements_file": "",
            "env_file": "",
            "endpoints": [],
            "environment": {
                "required": [],
                "optional": {}
            },
            "detected_at": datetime.utcnow().isoformat(),
            "detection_method": "basic"
        }
        
        # Check for common files
        files_to_check = {
            "package.json": ("Node.js", "node18"),
            "requirements.txt": ("Python", "python3.11"),
            "go.mod": ("Go", "go1.21"),
            "Gemfile": ("Ruby", "ruby3.0"),
            "pom.xml": ("Java", "java17"),
            "Cargo.toml": ("Rust", "rust1.70")
        }
        
        for file, (lang, runtime) in files_to_check.items():
            if os.path.exists(os.path.join(repo_path, file)):
                analysis["language"] = lang
                analysis["runtime"] = runtime
                analysis["requirements_file"] = file
                break
        
        # Check for main files
        main_files = ["main.py", "app.py", "server.js", "index.js", "main.go", "app.rb"]
        for file in main_files:
            if os.path.exists(os.path.join(repo_path, file)):
                analysis["main_file"] = file
                break
        
        # Check for .env files
        env_files = [".env.example", ".env.sample", ".env.template"]
        for file in env_files:
            if os.path.exists(os.path.join(repo_path, file)):
                analysis["env_file"] = file
                break
        
        return analysis
    
    async def create_api_from_analysis(
        self,
        user_id: str,
        repo_info: Dict[str, Any],
        analysis: Dict[str, Any],
        config_overrides: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Create API configuration from analysis"""
        # Merge analysis with any user overrides
        if config_overrides:
            for key, value in config_overrides.items():
                if value is not None:
                    analysis[key] = value
        
        # Create API configuration
        api_config = {
            "name": config_overrides.get("name") or repo_info.get("name", "new-api"),
            "description": config_overrides.get("description") or repo_info.get("description", ""),
            "user_id": user_id,
            "source": {
                "type": "github",
                "repo_url": repo_info.get("clone_url"),
                "repo_name": repo_info.get("full_name"),
                "branch": repo_info.get("default_branch", "main")
            },
            "runtime": {
                "language": analysis["language"],
                "version": analysis["runtime"],
                "framework": analysis["framework"]
            },
            "build": {
                "main_file": analysis["main_file"],
                "start_command": analysis["start_command"],
                "port": analysis["port"],
                "health_check": analysis["health_check"],
                "requirements_file": analysis["requirements_file"]
            },
            "environment": analysis["environment"],
            "endpoints": analysis["endpoints"],
            "created_at": datetime.utcnow().isoformat(),
            "status": "draft"
        }
        
        return api_config
    
    def get_oauth_url(self, redirect_uri: str, state: str) -> str:
        """Generate GitHub OAuth URL"""
        params = {
            "client_id": self.oauth_client_id,
            "redirect_uri": redirect_uri,
            "scope": "repo read:user",
            "state": state
        }
        
        query_string = "&".join([f"{k}={v}" for k, v in params.items()])
        return f"https://github.com/login/oauth/authorize?{query_string}"

# Export singleton instance
github_integration = GitHubIntegration()