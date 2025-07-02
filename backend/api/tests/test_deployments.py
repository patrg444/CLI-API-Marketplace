"""
Test deployment functionality
"""

import pytest
from unittest.mock import Mock, AsyncMock, patch, MagicMock
from fastapi.testclient import TestClient
import sys
import os
import uuid
import json
import base64

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

# Mock docker before importing
mock_docker = Mock()
sys.modules['docker'] = mock_docker

# Mock the database before importing main
with patch('asyncpg.create_pool', new_callable=AsyncMock) as mock_pool:
    with patch('redis.from_url') as mock_redis:
        mock_pool.return_value = AsyncMock()
        mock_redis.return_value = Mock()
        
        from main import app, db_pool
        from deployments import DeploymentManager
        from auth.api_keys import APIKeyManager
        
        # Set up the mock pool
        app.state.db_pool = mock_pool.return_value

client = TestClient(app)


class TestDeploymentAPI:
    """Test deployment API endpoints"""
    
    def setup_method(self):
        """Setup test API key for authentication"""
        self.test_user_id = str(uuid.uuid4())
        self.test_api_key = "test_api_key_deployment"
        self.auth_headers = {"X-API-Key": self.test_api_key}
        
        # Mock docker client
        mock_docker.from_env.return_value = Mock()
    
    @patch('main.deployment_manager')
    @patch('main.db_pool')
    def test_deploy_api_hosted(self, mock_db, mock_deployment_manager):
        """Test deploying a hosted API"""
        # Mock API key validation
        mock_conn = AsyncMock()
        mock_db.acquire.return_value.__aenter__.return_value = mock_conn
        
        # Mock user lookup for API key
        mock_conn.fetch.return_value = [{
            'id': uuid.uuid4(),
            'user_id': self.test_user_id,
            'key_hash': 'mock_hash',
            'scopes': ['read', 'write', 'deploy'],
            'expires_at': None,
            'email': 'test@example.com',
            'name': 'Test User',
            'is_active': True
        }]
        
        # Mock deployment creation
        deployment_id = str(uuid.uuid4())
        api_id = str(uuid.uuid4())
        mock_deployment_manager.create_deployment.return_value = {
            'deployment_id': deployment_id,
            'api_id': api_id,
            'status': 'pending',
            'message': 'Deployment queued for processing'
        }
        
        # Test deployment
        source_code = base64.b64encode(b"print('Hello API')").decode()
        response = client.post(
            "/api/deploy",
            json={
                "api_name": "test-api",
                "source_code": source_code,
                "runtime": "python:3.9",
                "env_vars": {"API_KEY": "secret"},
                "description": "Test API",
                "version": "1.0.0",
                "deployment_type": "hosted"
            },
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data['deployment_id'] == deployment_id
        assert data['api_id'] == api_id
        assert data['status'] == 'pending'
    
    @patch('main.deployment_manager')
    @patch('main.db_pool')
    def test_deploy_api_byoa_requires_premium(self, mock_db, mock_deployment_manager):
        """Test BYOA deployment requires premium subscription"""
        # Mock API key validation
        mock_conn = AsyncMock()
        mock_db.acquire.return_value.__aenter__.return_value = mock_conn
        
        # Mock non-premium user
        mock_conn.fetch.return_value = [{
            'id': uuid.uuid4(),
            'user_id': self.test_user_id,
            'key_hash': 'mock_hash',
            'scopes': ['read', 'write', 'deploy'],
            'expires_at': None,
            'email': 'test@example.com',
            'name': 'Test User',
            'is_active': True,
            'is_premium': False  # Not premium
        }]
        
        response = client.post(
            "/api/deploy",
            json={
                "api_name": "test-api",
                "source_code": "https://github.com/user/repo",
                "deployment_type": "byoa",
                "custom_endpoint": "https://my-domain.com/api"
            },
            headers=self.auth_headers
        )
        
        assert response.status_code == 403
        assert "premium subscription" in response.json()['detail'].lower()
    
    @patch('main.deployment_manager')
    @patch('main.db_pool')
    def test_get_deployment_status(self, mock_db, mock_deployment_manager):
        """Test getting deployment status"""
        # Mock API key validation
        mock_conn = AsyncMock()
        mock_db.acquire.return_value.__aenter__.return_value = mock_conn
        mock_conn.fetch.return_value = [{
            'id': uuid.uuid4(),
            'user_id': self.test_user_id,
            'key_hash': 'mock_hash',
            'scopes': ['read'],
            'expires_at': None,
            'email': 'test@example.com',
            'name': 'Test User',
            'is_active': True
        }]
        
        # Mock deployment status
        deployment_id = str(uuid.uuid4())
        mock_deployment_manager.get_deployment_status.return_value = {
            'deployment_id': deployment_id,
            'api_name': 'test-api',
            'status': 'building',
            'version': '1.0.0',
            'endpoint_url': None,
            'started_at': '2024-01-01T00:00:00',
            'completed_at': None,
            'build_logs': 'Building Docker image...',
            'config': {'runtime': 'python:3.9'}
        }
        
        response = client.get(
            f"/api/deployments/{deployment_id}",
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data['deployment_id'] == deployment_id
        assert data['status'] == 'building'
        assert 'build_logs' in data
    
    @patch('main.deployment_manager')
    @patch('main.db_pool')
    def test_list_deployments(self, mock_db, mock_deployment_manager):
        """Test listing deployments"""
        # Mock API key validation
        mock_conn = AsyncMock()
        mock_db.acquire.return_value.__aenter__.return_value = mock_conn
        mock_conn.fetch.return_value = [{
            'id': uuid.uuid4(),
            'user_id': self.test_user_id,
            'key_hash': 'mock_hash',
            'scopes': ['read'],
            'expires_at': None,
            'email': 'test@example.com',
            'name': 'Test User',
            'is_active': True
        }]
        
        # Mock deployments list
        mock_deployment_manager.list_deployments.return_value = [
            {
                'deployment_id': str(uuid.uuid4()),
                'api_id': str(uuid.uuid4()),
                'api_name': 'api-1',
                'version': '1.0.0',
                'status': 'success',
                'method': 'cli',
                'started_at': '2024-01-01T00:00:00',
                'completed_at': '2024-01-01T00:05:00',
                'duration_seconds': 300
            },
            {
                'deployment_id': str(uuid.uuid4()),
                'api_id': str(uuid.uuid4()),
                'api_name': 'api-2',
                'version': '2.0.0',
                'status': 'failed',
                'method': 'web',
                'started_at': '2024-01-02T00:00:00',
                'completed_at': '2024-01-02T00:02:00',
                'duration_seconds': 120
            }
        ]
        
        response = client.get(
            "/api/deployments?limit=10",
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        assert 'deployments' in data
        assert len(data['deployments']) == 2
        assert data['deployments'][0]['status'] == 'success'
        assert data['deployments'][1]['status'] == 'failed'


class TestDeploymentManager:
    """Test DeploymentManager functionality"""
    
    @pytest.fixture
    def deployment_manager(self):
        mock_pool = AsyncMock()
        config = {
            'docker_registry': 'registry.test.com',
            'deployment_namespace': 'test',
            'max_concurrent_builds': 5
        }
        return DeploymentManager(mock_pool, config)
    
    @pytest.mark.asyncio
    async def test_create_deployment_new_api(self, deployment_manager):
        """Test creating deployment for new API"""
        # Mock database interactions
        mock_conn = AsyncMock()
        mock_acquire = MagicMock()
        mock_acquire.__aenter__ = AsyncMock(return_value=mock_conn)
        mock_acquire.__aexit__ = AsyncMock(return_value=None)
        deployment_manager.db_pool.acquire = Mock(return_value=mock_acquire)
        
        # API doesn't exist
        mock_conn.fetchrow.return_value = None
        
        # Mock ID generation
        api_id = str(uuid.uuid4())
        deployment_id = str(uuid.uuid4())
        mock_conn.fetchval.side_effect = [api_id, deployment_id]
        
        # Create deployment
        result = await deployment_manager.create_deployment(
            user_id='user-123',
            api_name='new-api',
            source_code='print("hello")',
            config={'runtime': 'python:3.9'},
            deployment_type='hosted'
        )
        
        assert result['deployment_id'] == deployment_id
        assert result['api_id'] == api_id
        assert result['status'] == 'pending'
        
        # Verify API was created - check that fetchval was called twice
        assert mock_conn.fetchval.call_count == 2
        
        # Check first call was for API creation
        first_call = mock_conn.fetchval.call_args_list[0]
        assert 'INSERT INTO apis' in first_call[0][0]
        assert first_call[0][1] == 'user-123'  # user_id
        assert first_call[0][2] == 'new-api'   # api_name
    
    @pytest.mark.asyncio
    async def test_rollback_deployment(self, deployment_manager):
        """Test rolling back to previous deployment"""
        # Mock database
        mock_conn = AsyncMock()
        mock_acquire = MagicMock()
        mock_acquire.__aenter__ = AsyncMock(return_value=mock_conn)
        mock_acquire.__aexit__ = AsyncMock(return_value=None)
        deployment_manager.db_pool.acquire = Mock(return_value=mock_acquire)
        
        # Mock target deployment
        target_deployment_id = str(uuid.uuid4())
        api_id = str(uuid.uuid4())
        mock_conn.fetchrow.return_value = {
            'id': target_deployment_id,
            'api_id': api_id,
            'api_name': 'test-api',
            'version': '1.0.0',
            'status': 'success',
            'config_snapshot': json.dumps({'runtime': 'python:3.9'})
        }
        
        # Mock new deployment creation
        new_deployment_id = str(uuid.uuid4())
        mock_conn.fetchval.return_value = new_deployment_id
        
        # Perform rollback
        result = await deployment_manager.rollback_deployment(
            user_id='user-123',
            api_id=api_id,
            target_deployment_id=target_deployment_id
        )
        
        assert result['deployment_id'] == new_deployment_id
        assert '1.0.0' in result['message']
        assert 'initiated' in result['message']
    
    def test_generate_dockerfile_python(self, deployment_manager):
        """Test Dockerfile generation for Python"""
        config = {'runtime': 'python:3.9'}
        dockerfile = deployment_manager._generate_dockerfile(config)
        
        assert 'FROM python:3.9' in dockerfile
        assert 'pip install -r requirements.txt' in dockerfile
        assert 'CMD ["python", "main.py"]' in dockerfile
    
    def test_generate_dockerfile_node(self, deployment_manager):
        """Test Dockerfile generation for Node.js"""
        config = {'runtime': 'node:16'}
        dockerfile = deployment_manager._generate_dockerfile(config)
        
        assert 'FROM node:16' in dockerfile
        assert 'npm install' in dockerfile
        assert 'CMD ["node", "index.js"]' in dockerfile
    
    def test_generate_dockerfile_unsupported(self, deployment_manager):
        """Test Dockerfile generation for unsupported runtime"""
        config = {'runtime': 'ruby:3.0'}
        
        with pytest.raises(ValueError, match="Unsupported runtime"):
            deployment_manager._generate_dockerfile(config)


if __name__ == "__main__":
    pytest.main([__file__, "-v"])