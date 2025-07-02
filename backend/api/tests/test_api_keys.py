"""
Test API key management functionality
"""

import pytest
from unittest.mock import Mock, AsyncMock, patch, MagicMock
from fastapi.testclient import TestClient
from fastapi import HTTPException
import sys
import os
import uuid
import bcrypt
from datetime import datetime

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

# Mock docker before importing
mock_docker = Mock()
sys.modules['docker'] = mock_docker

# Create a proper async context manager class
class AsyncContextManager:
    def __init__(self):
        self.conn = AsyncMock()
        
    async def __aenter__(self):
        return self.conn
        
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        return None

# Mock the database before importing main
with patch('asyncpg.create_pool', new_callable=AsyncMock) as mock_pool:
    with patch('redis.from_url') as mock_redis:
        mock_pool.return_value = AsyncMock()
        mock_redis.return_value = Mock()
        
        import main
        from main import app, create_access_token
        from auth.api_keys import APIKeyManager
        
        # Set the global db_pool variable
        main.db_pool = Mock()
        main.db_pool.acquire = Mock(return_value=AsyncContextManager())
        main.api_key_manager = Mock()

client = TestClient(app)


class TestAPIKeyManagement:
    """Test API key creation and management"""
    
    def setup_method(self):
        """Setup test user"""
        self.test_user_id = str(uuid.uuid4())
        self.test_email = "test@example.com"
        self.access_token = create_access_token(self.test_user_id, self.test_email)
        self.auth_headers = {"Authorization": f"Bearer {self.access_token}"}
        
        # Mock database for authentication
        mock_conn = main.db_pool.acquire.return_value.conn
        mock_conn.fetchrow.return_value = {
            "id": self.test_user_id,
            "email": self.test_email,
            "name": "Test User",
            "company": None,
            "profile_image": None,
            "created_at": datetime.utcnow()
        }
    
    def test_create_api_key(self):
        """Test creating a new API key"""
        # Initialize API key manager if needed
        if not hasattr(main, 'api_key_manager') or main.api_key_manager is None:
            from auth.api_keys import APIKeyManager
            main.api_key_manager = APIKeyManager(main.db_pool)
        
        # Mock API key manager
        main.api_key_manager.create_api_key = AsyncMock(return_value={
            "id": str(uuid.uuid4()),
            "name": "Test CLI Key",
            "key": "ak_test123456789",
            "key_prefix": "ak_test1",
            "scopes": ["read", "write"],
            "expires_at": None,
            "created_at": datetime.utcnow().isoformat()
        })
        
        response = client.post(
            "/api/keys",
            json={
                "name": "Test CLI Key",
                "scopes": ["read", "write"],
                "expires_in_days": 30
            },
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        assert "key" in data  # Actual key only returned on creation
        assert data["name"] == "Test CLI Key"
        assert data["scopes"] == ["read", "write"]
        assert "key_prefix" in data
        assert "id" in data
    
    def test_list_api_keys(self):
        """Test listing API keys"""
        # Mock database responses
        mock_conn = main.db_pool.acquire.return_value.conn
        
        key_id = str(uuid.uuid4())
        mock_conn.fetch.return_value = [
            {
                'id': key_id,
                'name': 'Production Key',
                'key_prefix': 'prod_xxx',
                'scopes': ['read', 'write', 'deploy'],
                'last_used_at': None,
                'expires_at': None,
                'created_at': Mock(isoformat=lambda: '2024-01-01T00:00:00')
            }
        ]
        
        # Initialize API key manager if needed
        if not hasattr(main, 'api_key_manager') or main.api_key_manager is None:
            from auth.api_keys import APIKeyManager
            main.api_key_manager = APIKeyManager(main.db_pool)
        
        # Mock list_api_keys method
        main.api_key_manager.list_api_keys = AsyncMock(return_value=[
            {
                'id': str(key_id),
                'name': 'Production Key',
                'key': None,  # Key is not returned in list
                'key_prefix': 'prod_xxx',
                'scopes': ['read', 'write', 'deploy'],
                'last_used_at': None,
                'expires_at': None,
                'created_at': '2024-01-01T00:00:00',
                'is_expired': False
            }
        ])
        
        response = client.get("/api/keys", headers=self.auth_headers)
        
        assert response.status_code == 200
        data = response.json()
        assert len(data) == 1
        assert data[0]["name"] == "Production Key"
        # Key field exists but is None in list response
        assert data[0]["key"] is None
    
    def test_revoke_api_key(self):
        """Test revoking an API key"""
        # Mock database responses
        mock_conn = main.db_pool.acquire.return_value.conn
        mock_conn.execute.return_value = "DELETE 1"
        
        # Initialize API key manager if needed
        if not hasattr(main, 'api_key_manager') or main.api_key_manager is None:
            from auth.api_keys import APIKeyManager
            main.api_key_manager = APIKeyManager(main.db_pool)
        
        # Mock revoke_api_key method
        main.api_key_manager.revoke_api_key = AsyncMock(return_value=True)
        
        key_id = str(uuid.uuid4())
        response = client.delete(f"/api/keys/{key_id}", headers=self.auth_headers)
        
        assert response.status_code == 200
        assert response.json()["message"] == "API key revoked successfully"
    
    def test_api_key_authentication(self):
        """Test authenticating with API key"""
        # Create a test API key
        test_key = "test_api_key_12345"
        key_hash = bcrypt.hashpw(test_key.encode('utf-8'), bcrypt.gensalt()).decode('utf-8')
        
        # Mock database responses
        mock_conn = main.db_pool.acquire.return_value.conn
        
        # Mock key validation
        mock_conn.fetch.return_value = [
            {
                'id': str(uuid.uuid4()),
                'user_id': self.test_user_id,
                'key_hash': key_hash,
                'scopes': ['read', 'write'],
                'expires_at': None,
                'email': self.test_email,
                'name': 'Test User',
                'is_active': True
            }
        ]
        
        # Initialize API key manager if needed
        if not hasattr(main, 'api_key_manager') or main.api_key_manager is None:
            from auth.api_keys import APIKeyManager
            main.api_key_manager = APIKeyManager(main.db_pool)
        
        # Test the API key
        response = client.get(
            "/api/keys/test",
            headers={"X-API-Key": test_key}
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["message"] == "API key is valid"
        assert data["email"] == self.test_email
    
    def test_max_api_keys_limit(self):
        """Test maximum API keys limit"""
        # Mock database responses
        mock_conn = main.db_pool.acquire.return_value.conn
        
        # Mock that user already has 10 keys
        mock_conn.fetchval.return_value = 10
        
        # Initialize API key manager if needed
        if not hasattr(main, 'api_key_manager') or main.api_key_manager is None:
            from auth.api_keys import APIKeyManager
            main.api_key_manager = APIKeyManager(main.db_pool)
        
        # Mock the create_api_key to raise the exception
        main.api_key_manager.create_api_key = AsyncMock(side_effect=HTTPException(
            status_code=400,
            detail="Maximum number of API keys (10) reached"
        ))
        
        response = client.post(
            "/api/keys",
            json={"name": "One Too Many"},
            headers=self.auth_headers
        )
        
        assert response.status_code == 400
        assert "Maximum number of API keys" in response.json()["detail"]
    
    def test_api_key_required(self):
        """Test that API key is required for protected endpoints"""
        response = client.get("/api/keys/test")
        
        assert response.status_code == 401
        assert "API key required" in response.json()["detail"]
    
    def test_invalid_api_key(self):
        """Test invalid API key"""
        # Initialize API key manager if needed
        if not hasattr(main, 'api_key_manager') or main.api_key_manager is None:
            from auth.api_keys import APIKeyManager
            main.api_key_manager = APIKeyManager(main.db_pool)
        
        # Mock that no keys match
        mock_conn = main.db_pool.acquire.return_value.conn
        mock_conn.fetch.return_value = []
        
        response = client.get(
            "/api/keys/test",
            headers={"X-API-Key": "invalid_key"}
        )
        
        assert response.status_code == 401
        assert "Invalid API key" in response.json()["detail"]


class TestAPIKeyScopes:
    """Test API key scope permissions"""
    
    def test_scope_validation(self):
        """Test scope-based access control"""
        # Use the existing db_pool mock
        manager = APIKeyManager(main.db_pool)
        
        # Test user with limited scopes
        user_info = {
            "user_id": "123",
            "scopes": ["read", "write"]
        }
        
        # Should have read scope
        assert manager.check_scope(user_info, "read") is True
        
        # Should not have deploy scope
        assert manager.check_scope(user_info, "deploy") is False
        
        # Admin scope overrides all
        admin_info = {
            "user_id": "456",
            "scopes": ["admin"]
        }
        assert manager.check_scope(admin_info, "deploy") is True


if __name__ == "__main__":
    pytest.main([__file__, "-v"])