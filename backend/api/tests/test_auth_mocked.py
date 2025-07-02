"""
Auth tests with mocked database
"""

import pytest
from unittest.mock import Mock, AsyncMock, patch
from fastapi.testclient import TestClient
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

# Mock the database before importing main
with patch('asyncpg.create_pool', new_callable=AsyncMock) as mock_pool:
    with patch('redis.from_url') as mock_redis:
        mock_pool.return_value = AsyncMock()
        mock_redis.return_value = Mock()
        
        from main import app, db_pool
        
        # Set up the mock pool
        app.state.db_pool = mock_pool.return_value

client = TestClient(app)


class TestRegistration:
    """Test user registration with mocked database"""
    
    @patch('main.db_pool')
    def test_register_new_user(self, mock_db):
        """Test successful registration"""
        # Mock database responses
        mock_conn = AsyncMock()
        mock_db.acquire.return_value.__aenter__.return_value = mock_conn
        mock_conn.fetchrow.side_effect = [
            None,  # No existing user
            {  # Created user
                'id': '12345',
                'name': 'Test User',
                'email': 'test@example.com',
                'company': 'Test Company',
                'profile_image': None,
                'created_at': Mock(isoformat=lambda: '2024-01-01T00:00:00')
            }
        ]
        mock_conn.fetchval.return_value = '12345'  # New user ID
        
        response = client.post("/auth/register", json={
            "email": "test@example.com",
            "password": "SecurePassword123!",
            "name": "Test User",
            "company": "Test Company"
        })
        
        assert response.status_code == 200
        data = response.json()
        assert 'access_token' in data
        assert data['token_type'] == 'bearer'
        assert 'user' in data
        assert data['user']['email'] == 'test@example.com'
    
    @patch('main.db_pool')
    def test_register_duplicate_email(self, mock_db):
        """Test registration with existing email"""
        # Mock database responses
        mock_conn = AsyncMock()
        mock_db.acquire.return_value.__aenter__.return_value = mock_conn
        mock_conn.fetchrow.return_value = {'id': '12345'}  # Existing user
        
        response = client.post("/auth/register", json={
            "email": "existing@example.com",
            "password": "SecurePassword123!",
            "name": "Test User",
            "company": "Test Company"
        })
        
        assert response.status_code == 400
        assert "already registered" in response.json()['detail']
    
    def test_register_invalid_data(self):
        """Test registration with invalid data"""
        # Missing required fields
        response = client.post("/auth/register", json={
            "email": "test@example.com"
        })
        assert response.status_code == 422
        
        # Invalid email
        response = client.post("/auth/register", json={
            "email": "not-an-email",
            "password": "SecurePassword123!",
            "name": "Test User"
        })
        assert response.status_code == 422


class TestLogin:
    """Test user login"""
    
    @patch('main.db_pool')
    def test_login_valid_credentials(self, mock_db):
        """Test login with valid credentials"""
        # Mock database responses
        mock_conn = AsyncMock()
        mock_db.acquire.return_value.__aenter__.return_value = mock_conn
        
        # Mock password hash that will match
        from main import hash_password
        password_hash = hash_password("SecurePassword123!")
        
        mock_conn.fetchrow.return_value = {
            'id': '12345',
            'name': 'Test User',
            'email': 'test@example.com',
            'company': 'Test Company',
            'profile_image': None,
            'password_hash': password_hash,
            'created_at': Mock(isoformat=lambda: '2024-01-01T00:00:00')
        }
        
        response = client.post("/auth/login", json={
            "email": "test@example.com",
            "password": "SecurePassword123!"
        })
        
        assert response.status_code == 200
        data = response.json()
        assert 'access_token' in data
        assert data['token_type'] == 'bearer'
        assert 'user' in data
    
    @patch('main.db_pool')
    def test_login_invalid_credentials(self, mock_db):
        """Test login with invalid credentials"""
        # Mock database responses
        mock_conn = AsyncMock()
        mock_db.acquire.return_value.__aenter__.return_value = mock_conn
        mock_conn.fetchrow.return_value = None  # User not found
        
        response = client.post("/auth/login", json={
            "email": "nonexistent@example.com",
            "password": "WrongPassword123!"
        })
        
        assert response.status_code == 401
        assert "Invalid email or password" in response.json()['detail']


class TestAuthentication:
    """Test authentication middleware"""
    
    def test_protected_endpoint_without_token(self):
        """Test accessing protected endpoint without token"""
        response = client.get("/auth/me")
        assert response.status_code == 403
        assert "Not authenticated" in response.json()['detail']
    
    @patch('main.db_pool')
    def test_protected_endpoint_with_valid_token(self, mock_db):
        """Test accessing protected endpoint with valid token"""
        # First login to get token
        mock_conn = AsyncMock()
        mock_db.acquire.return_value.__aenter__.return_value = mock_conn
        
        from main import hash_password
        password_hash = hash_password("SecurePassword123!")
        
        mock_conn.fetchrow.return_value = {
            'id': '12345',
            'name': 'Test User',
            'email': 'test@example.com',
            'company': 'Test Company',
            'profile_image': None,
            'password_hash': password_hash,
            'created_at': Mock(isoformat=lambda: '2024-01-01T00:00:00')
        }
        
        # Login
        login_response = client.post("/auth/login", json={
            "email": "test@example.com",
            "password": "SecurePassword123!"
        })
        token = login_response.json()['access_token']
        
        # Use token to access protected endpoint
        response = client.get("/auth/me", headers={
            "Authorization": f"Bearer {token}"
        })
        
        assert response.status_code == 200
        assert response.json()['email'] == 'test@example.com'


if __name__ == "__main__":
    pytest.main([__file__, "-v"])