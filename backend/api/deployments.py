"""
Deployment management for API-Direct
Handles API deployments, builds, and lifecycle management
"""

import os
import json
import asyncio
import secrets
import hashlib
from datetime import datetime, timedelta
from typing import Dict, Any, Optional, List
import asyncpg
import logging
from fastapi import HTTPException
import docker

logger = logging.getLogger(__name__)


class DeploymentManager:
    """Manages API deployments and lifecycle"""
    
    def __init__(self, db_pool: asyncpg.Pool, config: Dict[str, Any]):
        self.db_pool = db_pool
        self.config = config
        self.docker_client = docker.from_env()
        self.deployment_queue = asyncio.Queue()
        
    async def create_deployment(
        self,
        user_id: str,
        api_name: str,
        source_code: str,
        config: Dict[str, Any],
        deployment_type: str = "hosted"
    ) -> Dict[str, Any]:
        """
        Create a new API deployment
        
        Args:
            user_id: Owner's user ID
            api_name: Name of the API
            source_code: Base64 encoded source code or Git URL
            config: Deployment configuration
            deployment_type: 'hosted' or 'byoa'
            
        Returns:
            Deployment details including ID and status
        """
        async with self.db_pool.acquire() as conn:
            # Check if API exists
            api = await conn.fetchrow("""
                SELECT id, status, deployment_type 
                FROM apis 
                WHERE user_id = $1 AND name = $2
            """, user_id, api_name)
            
            if not api:
                # Create new API entry
                api_id = await conn.fetchval("""
                    INSERT INTO apis (
                        user_id, name, description, deployment_type, status
                    ) VALUES ($1, $2, $3, $4, $5)
                    RETURNING id
                """, user_id, api_name, config.get('description', ''), 
                    deployment_type, 'building')
            else:
                api_id = api['id']
                
                # Update API status
                await conn.execute("""
                    UPDATE apis 
                    SET status = 'building', updated_at = NOW()
                    WHERE id = $1
                """, api_id)
            
            # Create deployment record
            deployment_id = await conn.fetchval("""
                INSERT INTO deployments (
                    api_id, version, status, deployment_method,
                    config_snapshot
                ) VALUES ($1, $2, $3, $4, $5)
                RETURNING id
            """, api_id, config.get('version', '1.0.0'), 'pending', 
                'cli', json.dumps(config))
            
            # Queue deployment for processing
            await self.deployment_queue.put({
                'deployment_id': str(deployment_id),
                'api_id': str(api_id),
                'user_id': user_id,
                'api_name': api_name,
                'source_code': source_code,
                'config': config,
                'deployment_type': deployment_type
            })
            
            logger.info(f"Created deployment {deployment_id} for API {api_name}")
            
            return {
                'deployment_id': str(deployment_id),
                'api_id': str(api_id),
                'status': 'pending',
                'message': 'Deployment queued for processing'
            }
    
    async def process_deployment(self, deployment_data: Dict[str, Any]):
        """Process a queued deployment"""
        deployment_id = deployment_data['deployment_id']
        
        try:
            # Update deployment status
            await self._update_deployment_status(
                deployment_id, 'building', 'Starting build process'
            )
            
            # Build the API
            if deployment_data['deployment_type'] == 'hosted':
                endpoint_url = await self._build_hosted_api(deployment_data)
            else:
                endpoint_url = await self._build_byoa_api(deployment_data)
            
            # Update API with endpoint
            async with self.db_pool.acquire() as conn:
                await conn.execute("""
                    UPDATE apis 
                    SET status = 'running', 
                        endpoint_url = $1,
                        deployed_at = NOW(),
                        updated_at = NOW()
                    WHERE id = $2
                """, endpoint_url, deployment_data['api_id'])
                
                # Update deployment
                await conn.execute("""
                    UPDATE deployments
                    SET status = 'success',
                        completed_at = NOW()
                    WHERE id = $1
                """, deployment_id)
            
            logger.info(f"Deployment {deployment_id} completed successfully")
            
            # Notify via WebSocket
            await self._notify_deployment_complete(
                deployment_data['user_id'],
                deployment_data['api_id'],
                endpoint_url
            )
            
        except Exception as e:
            logger.error(f"Deployment {deployment_id} failed: {e}")
            await self._update_deployment_status(
                deployment_id, 'failed', str(e)
            )
            
            # Update API status
            async with self.db_pool.acquire() as conn:
                await conn.execute("""
                    UPDATE apis 
                    SET status = 'error'
                    WHERE id = $1
                """, deployment_data['api_id'])
    
    async def _build_hosted_api(self, deployment_data: Dict[str, Any]) -> str:
        """Build and deploy API on our infrastructure"""
        api_id = deployment_data['api_id']
        config = deployment_data['config']
        
        # Generate unique subdomain
        subdomain = f"{deployment_data['api_name']}-{api_id[:8]}".lower()
        endpoint_url = f"https://{subdomain}.api-direct.io"
        
        # Create Docker container
        container_name = f"api-{api_id}"
        
        # Build Docker image
        dockerfile_content = self._generate_dockerfile(config)
        image_tag = f"apidirect/{api_id}:latest"
        
        # In production, this would build the actual Docker image
        # For now, we'll simulate the process
        await asyncio.sleep(2)  # Simulate build time
        
        # Deploy container (simulated)
        container_config = {
            'name': container_name,
            'image': image_tag,
            'environment': config.get('env_vars', {}),
            'labels': {
                'api_id': api_id,
                'user_id': deployment_data['user_id']
            },
            'restart_policy': {'Name': 'unless-stopped'},
            'ports': {'8080/tcp': None}  # Dynamic port allocation
        }
        
        # In production, start the container
        # container = self.docker_client.containers.run(**container_config)
        
        return endpoint_url
    
    async def _build_byoa_api(self, deployment_data: Dict[str, Any]) -> str:
        """Build API for bring-your-own-account deployment"""
        config = deployment_data['config']
        
        # Generate deployment package
        package_id = secrets.token_urlsafe(16)
        package_url = f"https://packages.api-direct.io/{package_id}.zip"
        
        # Create deployment instructions
        instructions = {
            'package_url': package_url,
            'deploy_commands': self._generate_deploy_commands(config),
            'environment_variables': config.get('env_vars', {}),
            'required_services': config.get('services', [])
        }
        
        # Store deployment package info
        async with self.db_pool.acquire() as conn:
            await conn.execute("""
                UPDATE apis
                SET runtime_config = runtime_config || $1
                WHERE id = $2
            """, json.dumps({'deployment_package': instructions}), 
                deployment_data['api_id'])
        
        # Return user's custom endpoint
        return config.get('custom_endpoint', 'https://your-domain.com/api')
    
    def _generate_dockerfile(self, config: Dict[str, Any]) -> str:
        """Generate Dockerfile based on config"""
        runtime = config.get('runtime', 'python:3.9')
        
        if 'python' in runtime:
            return f"""
FROM {runtime}

WORKDIR /app

COPY requirements.txt .
RUN pip install -r requirements.txt

COPY . .

EXPOSE 8080
CMD ["python", "main.py"]
"""
        elif 'node' in runtime:
            return f"""
FROM {runtime}

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

EXPOSE 8080
CMD ["node", "index.js"]
"""
        else:
            raise ValueError(f"Unsupported runtime: {runtime}")
    
    def _generate_deploy_commands(self, config: Dict[str, Any]) -> List[str]:
        """Generate deployment commands for BYOA"""
        runtime = config.get('runtime', 'python:3.9')
        
        if 'python' in runtime:
            return [
                "pip install -r requirements.txt",
                "python main.py"
            ]
        elif 'node' in runtime:
            return [
                "npm install",
                "node index.js"
            ]
        else:
            return ["echo 'Please configure deployment commands'"]
    
    async def _update_deployment_status(
        self, 
        deployment_id: str, 
        status: str, 
        message: str = None
    ):
        """Update deployment status and logs"""
        async with self.db_pool.acquire() as conn:
            if message:
                await conn.execute("""
                    UPDATE deployments
                    SET status = $1,
                        build_logs = COALESCE(build_logs, '') || $2 || E'\n'
                    WHERE id = $3
                """, status, f"[{datetime.utcnow().isoformat()}] {message}", 
                    deployment_id)
            else:
                await conn.execute("""
                    UPDATE deployments
                    SET status = $1
                    WHERE id = $2
                """, status, deployment_id)
    
    async def _notify_deployment_complete(
        self, 
        user_id: str, 
        api_id: str,
        endpoint_url: str
    ):
        """Notify user via WebSocket about deployment completion"""
        try:
            from main import websocket_manager
            
            await websocket_manager.notify_api_status_change(
                user_id=user_id,
                api_id=api_id,
                api_name="API",
                status="deployed",
                endpoint_url=endpoint_url
            )
        except Exception as e:
            logger.error(f"Failed to send WebSocket notification: {e}")
    
    async def get_deployment_status(
        self, 
        user_id: str, 
        deployment_id: str
    ) -> Dict[str, Any]:
        """Get deployment status and logs"""
        async with self.db_pool.acquire() as conn:
            deployment = await conn.fetchrow("""
                SELECT d.*, a.name as api_name, a.endpoint_url
                FROM deployments d
                JOIN apis a ON d.api_id = a.id
                WHERE d.id = $1 AND a.user_id = $2
            """, deployment_id, user_id)
            
            if not deployment:
                raise HTTPException(status_code=404, detail="Deployment not found")
            
            return {
                'deployment_id': str(deployment['id']),
                'api_name': deployment['api_name'],
                'status': deployment['status'],
                'version': deployment['version'],
                'endpoint_url': deployment['endpoint_url'],
                'started_at': deployment['started_at'].isoformat(),
                'completed_at': deployment['completed_at'].isoformat() if deployment['completed_at'] else None,
                'build_logs': deployment['build_logs'],
                'config': deployment['config_snapshot']
            }
    
    async def list_deployments(
        self, 
        user_id: str, 
        api_id: Optional[str] = None,
        limit: int = 20
    ) -> List[Dict[str, Any]]:
        """List deployments for a user or specific API"""
        async with self.db_pool.acquire() as conn:
            query = """
                SELECT d.*, a.name as api_name
                FROM deployments d
                JOIN apis a ON d.api_id = a.id
                WHERE a.user_id = $1
            """
            params = [user_id]
            
            if api_id:
                query += " AND a.id = $2"
                params.append(api_id)
            
            query += " ORDER BY d.created_at DESC LIMIT $" + str(len(params) + 1)
            params.append(limit)
            
            deployments = await conn.fetch(query, *params)
            
            return [
                {
                    'deployment_id': str(d['id']),
                    'api_id': str(d['api_id']),
                    'api_name': d['api_name'],
                    'version': d['version'],
                    'status': d['status'],
                    'method': d['deployment_method'],
                    'started_at': d['started_at'].isoformat(),
                    'completed_at': d['completed_at'].isoformat() if d['completed_at'] else None,
                    'duration_seconds': d['build_duration_seconds']
                }
                for d in deployments
            ]
    
    async def rollback_deployment(
        self, 
        user_id: str, 
        api_id: str,
        target_deployment_id: str
    ) -> Dict[str, Any]:
        """Rollback to a previous deployment"""
        async with self.db_pool.acquire() as conn:
            # Verify ownership and get target deployment
            target = await conn.fetchrow("""
                SELECT d.*, a.name as api_name
                FROM deployments d
                JOIN apis a ON d.api_id = a.id
                WHERE d.id = $1 AND a.id = $2 AND a.user_id = $3
                AND d.status = 'success'
            """, target_deployment_id, api_id, user_id)
            
            if not target:
                raise HTTPException(
                    status_code=404, 
                    detail="Target deployment not found or not successful"
                )
            
            # Create new deployment as rollback
            new_deployment_id = await conn.fetchval("""
                INSERT INTO deployments (
                    api_id, version, status, deployment_method,
                    config_snapshot
                ) VALUES ($1, $2, $3, $4, $5)
                RETURNING id
            """, api_id, f"{target['version']}-rollback", 'pending',
                'rollback', target['config_snapshot'])
            
            # Queue the rollback
            await self.deployment_queue.put({
                'deployment_id': str(new_deployment_id),
                'api_id': api_id,
                'user_id': user_id,
                'api_name': target['api_name'],
                'source_code': None,  # Use existing image
                'config': json.loads(target['config_snapshot']),
                'deployment_type': 'rollback',
                'target_deployment': str(target_deployment_id)
            })
            
            return {
                'deployment_id': str(new_deployment_id),
                'message': f'Rollback to version {target["version"]} initiated'
            }
    
    async def delete_api(self, user_id: str, api_id: str) -> bool:
        """Delete an API and clean up resources"""
        async with self.db_pool.acquire() as conn:
            # Verify ownership
            api = await conn.fetchrow("""
                SELECT * FROM apis
                WHERE id = $1 AND user_id = $2
            """, api_id, user_id)
            
            if not api:
                raise HTTPException(status_code=404, detail="API not found")
            
            # Stop running containers (in production)
            container_name = f"api-{api_id}"
            try:
                # container = self.docker_client.containers.get(container_name)
                # container.stop()
                # container.remove()
                pass
            except Exception as e:
                logger.error(f"Failed to stop container: {e}")
            
            # Delete from database (cascades to deployments)
            await conn.execute("""
                DELETE FROM apis WHERE id = $1
            """, api_id)
            
            logger.info(f"Deleted API {api_id}")
            return True
    
    async def start_deployment_worker(self):
        """Start background worker to process deployments"""
        while True:
            try:
                deployment_data = await self.deployment_queue.get()
                asyncio.create_task(self.process_deployment(deployment_data))
            except Exception as e:
                logger.error(f"Deployment worker error: {e}")
                await asyncio.sleep(5)