"""
Test trial and sandbox system functionality
"""

import pytest
from unittest.mock import Mock, AsyncMock, patch
from fastapi.testclient import TestClient
from fastapi import HTTPException
import sys
import os
import uuid
from datetime import datetime, timedelta
import json

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
        from trial_manager import TrialManager
        
        # Set the global db_pool variable
        main.db_pool = Mock()
        main.db_pool.acquire = Mock(return_value=AsyncContextManager())
        main.api_key_manager = Mock()
        main.trial_manager = Mock()

client = TestClient(app)


class TestTrialSystem:
    """Test trial functionality"""
    
    def setup_method(self):
        """Setup test user and API"""
        self.test_user_id = str(uuid.uuid4())
        self.test_email = "test@example.com"
        self.access_token = create_access_token(self.test_user_id, self.test_email)
        self.auth_headers = {"Authorization": f"Bearer {self.access_token}"}
        self.test_api_id = str(uuid.uuid4())
        
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
    
    def test_start_trial(self):
        """Test starting a new trial"""
        # Initialize trial manager if needed
        if not hasattr(main, 'trial_manager') or main.trial_manager is None:
            main.trial_manager = TrialManager(main.db_pool)
        
        # Mock trial creation
        main.trial_manager.start_trial = AsyncMock(return_value={
            "trial_id": str(uuid.uuid4()),
            "api_id": self.test_api_id,
            "api_name": "Test API",
            "status": "active",
            "requests_limit": 1000,
            "requests_used": 0,
            "rate_limit": 10,
            "expires_at": (datetime.utcnow() + timedelta(days=7)).isoformat(),
            "started_at": datetime.utcnow().isoformat()
        })
        
        response = client.post(
            f"/api/trials/start?api_id={self.test_api_id}",
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "active"
        assert data["requests_limit"] == 1000
        assert "trial_id" in data
    
    def test_check_trial_status(self):
        """Test checking trial status"""
        # Initialize trial manager
        if not hasattr(main, 'trial_manager') or main.trial_manager is None:
            main.trial_manager = TrialManager(main.db_pool)
        
        # Mock trial check - active trial
        main.trial_manager.check_trial_access = AsyncMock(return_value={
            "trial_id": str(uuid.uuid4()),
            "requests_remaining": 950,
            "rate_limit": 10,
            "expires_at": (datetime.utcnow() + timedelta(days=5)).isoformat()
        })
        
        response = client.get(
            f"/api/trials/{self.test_api_id}/status",
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["has_trial"] is True
        assert data["trial"]["requests_remaining"] == 950
    
    def test_no_trial_status(self):
        """Test status when no trial exists"""
        # Initialize trial manager
        if not hasattr(main, 'trial_manager') or main.trial_manager is None:
            main.trial_manager = TrialManager(main.db_pool)
        
        # Mock no trial
        main.trial_manager.check_trial_access = AsyncMock(return_value=None)
        
        response = client.get(
            f"/api/trials/{self.test_api_id}/status",
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["has_trial"] is False
    
    def test_sandbox_request_with_mock(self):
        """Test making sandbox request with mock response"""
        # Initialize trial manager
        if not hasattr(main, 'trial_manager') or main.trial_manager is None:
            main.trial_manager = TrialManager(main.db_pool)
        
        # Mock trial access check
        main.trial_manager.check_trial_access = AsyncMock(return_value={
            "trial_id": str(uuid.uuid4()),
            "requests_remaining": 100,
            "rate_limit": 10
        })
        
        # Mock sandbox request
        main.trial_manager.create_sandbox_request = AsyncMock(return_value={
            "status": 200,
            "headers": {"Content-Type": "application/json"},
            "data": {"message": "Hello from mock!"},
            "is_sandbox": True,
            "is_mocked": True
        })
        
        # Mock record usage
        main.trial_manager.record_trial_usage = AsyncMock(return_value=True)
        
        response = client.post(
            f"/api/sandbox/{self.test_api_id}/request",
            params={
                "endpoint": "/test",
                "method": "GET",
                "use_mock": True
            },
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == 200
        assert data["is_mocked"] is True
        assert data["data"]["message"] == "Hello from mock!"
    
    def test_sandbox_request_without_trial(self):
        """Test sandbox request without active trial"""
        # Initialize trial manager
        if not hasattr(main, 'trial_manager') or main.trial_manager is None:
            main.trial_manager = TrialManager(main.db_pool)
        
        # Mock no trial access
        main.trial_manager.check_trial_access = AsyncMock(return_value=None)
        
        response = client.post(
            f"/api/sandbox/{self.test_api_id}/request",
            params={
                "endpoint": "/test",
                "method": "GET"
            },
            headers=self.auth_headers
        )
        
        assert response.status_code == 403
        assert "No active trial" in response.json()["detail"]
    
    def test_create_mock_response(self):
        """Test creating mock response (API owner only)"""
        # Initialize trial manager
        if not hasattr(main, 'trial_manager') or main.trial_manager is None:
            main.trial_manager = TrialManager(main.db_pool)
        
        # Mock API ownership check
        mock_conn = main.db_pool.acquire.return_value.conn
        mock_conn.fetchval.return_value = self.test_user_id  # User owns the API
        
        # Mock create mock response
        main.trial_manager.create_mock_response = AsyncMock(return_value={
            "mock_id": str(uuid.uuid4()),
            "message": "Mock response created successfully"
        })
        
        response = client.post(
            f"/api/sandbox/{self.test_api_id}/mock",
            params={
                "endpoint_pattern": "^/users/\\d+$",
                "method": "GET",
                "status": 200
            },
            json={
                "id": 123,
                "name": "Test User",
                "email": "test@example.com"
            },
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        assert "mock_id" in data
        assert data["message"] == "Mock response created successfully"
    
    def test_create_mock_response_unauthorized(self):
        """Test creating mock response by non-owner"""
        # Initialize trial manager
        if not hasattr(main, 'trial_manager') or main.trial_manager is None:
            main.trial_manager = TrialManager(main.db_pool)
        
        # Mock API ownership check - different owner
        mock_conn = main.db_pool.acquire.return_value.conn
        mock_conn.fetchval.return_value = str(uuid.uuid4())  # Different user owns the API
        
        response = client.post(
            f"/api/sandbox/{self.test_api_id}/mock",
            params={
                "endpoint_pattern": "^/users/\\d+$",
                "method": "GET"
            },
            json={"test": "data"},
            headers=self.auth_headers
        )
        
        assert response.status_code == 403
        assert "Only API owners can create mock responses" in response.json()["detail"]
    
    def test_get_trial_analytics(self):
        """Test getting trial analytics (API owner only)"""
        # Initialize trial manager
        if not hasattr(main, 'trial_manager') or main.trial_manager is None:
            main.trial_manager = TrialManager(main.db_pool)
        
        # Mock API ownership check
        mock_conn = main.db_pool.acquire.return_value.conn
        mock_conn.fetchval.return_value = self.test_user_id  # User owns the API
        
        # Mock analytics
        main.trial_manager.get_trial_analytics = AsyncMock(return_value={
            "trial_stats": {
                "total_trials": 150,
                "active_trials": 25,
                "converted_trials": 45,
                "expired_trials": 80,
                "avg_requests_used": 450.5,
                "conversion_rate": 30.0
            },
            "sandbox_stats": {
                "total_sandbox_requests": 5000,
                "mocked_requests": 3500,
                "real_requests": 1500,
                "avg_response_time": 125.5
            }
        })
        
        response = client.get(
            f"/api/trials/{self.test_api_id}/analytics",
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["trial_stats"]["total_trials"] == 150
        assert data["trial_stats"]["conversion_rate"] == 30.0
        assert data["sandbox_stats"]["total_sandbox_requests"] == 5000


class TestTrialManager:
    """Test trial manager directly"""
    
    def test_trial_expiration(self):
        """Test trial expiration logic"""
        # Use the existing db_pool mock
        manager = TrialManager(main.db_pool)
        
        # Test expired by date
        mock_conn = main.db_pool.acquire.return_value.conn
        mock_conn.fetchrow.return_value = {
            'id': uuid.uuid4(),
            'user_id': uuid.uuid4(),
            'api_id': uuid.uuid4(),
            'status': 'active',
            'requests_used': 50,
            'requests_limit': 1000,
            'expires_at': datetime.utcnow() - timedelta(days=1),  # Expired yesterday
            'trial_rate_limit': 10
        }
        
        # The check_trial_access should return None for expired trial
        async def test_expired():
            result = await manager.check_trial_access("user123", "api123")
            assert result is None
        
        # Run the async test
        import asyncio
        asyncio.run(test_expired())
    
    def test_trial_request_limit(self):
        """Test trial request limit enforcement"""
        manager = TrialManager(main.db_pool)
        
        # Test exhausted requests
        mock_conn = main.db_pool.acquire.return_value.conn
        mock_conn.fetchrow.return_value = {
            'id': uuid.uuid4(),
            'user_id': uuid.uuid4(),
            'api_id': uuid.uuid4(),
            'status': 'active',
            'requests_used': 1000,
            'requests_limit': 1000,  # All requests used
            'expires_at': datetime.utcnow() + timedelta(days=7),
            'trial_rate_limit': 10
        }
        
        async def test_limit():
            result = await manager.check_trial_access("user123", "api123")
            assert result is None
        
        import asyncio
        asyncio.run(test_limit())


if __name__ == "__main__":
    pytest.main([__file__, "-v"])