"""
Authentication endpoint tests for the API-Direct backend
Tests user registration, login, token management, and security
"""

import pytest
from fastapi.testclient import TestClient
from unittest.mock import Mock, patch
import jwt
from datetime import datetime, timedelta
import json

# Import the main app
import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from main import app

client = TestClient(app)


class TestUserRegistration:
    """Test user registration endpoint"""
    
    def test_register_valid_user(self):
        """Test successful user registration"""
        response = client.post("/auth/register", json={
            "email": "newuser@example.com",
            "password": "SecurePassword123!",
            "username": "newuser",
            "full_name": "New User"
        })
        
        assert response.status_code == 201
        data = response.json()
        assert "user_id" in data
        assert data["email"] == "newuser@example.com"
        assert "password" not in data  # Password should not be returned
        
    def test_register_duplicate_email(self):
        """Test registration with existing email"""
        # First registration
        client.post("/auth/register", json={
            "email": "duplicate@example.com",
            "password": "SecurePassword123!",
            "username": "user1"
        })
        
        # Attempt duplicate
        response = client.post("/auth/register", json={
            "email": "duplicate@example.com",
            "password": "SecurePassword123!",
            "username": "user2"
        })
        
        assert response.status_code == 409
        assert "already exists" in response.json()["detail"].lower()
        
    def test_register_invalid_email(self):
        """Test registration with invalid email formats"""
        invalid_emails = [
            "notanemail",
            "@example.com",
            "user@",
            "user..name@example.com",
            "user@example",
            ""
        ]
        
        for email in invalid_emails:
            response = client.post("/auth/register", json={
                "email": email,
                "password": "SecurePassword123!",
                "username": "testuser"
            })
            
            assert response.status_code == 422
            assert "email" in str(response.json()["detail"]).lower()
            
    def test_register_weak_password(self):
        """Test registration with weak passwords"""
        weak_passwords = [
            "short",  # Too short
            "alllowercase",  # No uppercase
            "ALLUPPERCASE",  # No lowercase
            "NoNumbers!",  # No numbers
            "NoSpecial123",  # No special characters
            "password123!",  # Common password
        ]
        
        for password in weak_passwords:
            response = client.post("/auth/register", json={
                "email": "test@example.com",
                "password": password,
                "username": "testuser"
            })
            
            assert response.status_code == 400
            assert "password" in str(response.json()["detail"]).lower()
            
    def test_register_sql_injection_attempt(self):
        """Test SQL injection prevention in registration"""
        response = client.post("/auth/register", json={
            "email": "test@example.com",
            "password": "SecurePassword123!",
            "username": "admin'; DROP TABLE users;--"
        })
        
        # Should either sanitize or reject, but not cause error
        assert response.status_code in [201, 400]
        
        # Verify database is intact by attempting another registration
        verify_response = client.post("/auth/register", json={
            "email": "verify@example.com",
            "password": "SecurePassword123!",
            "username": "verifyuser"
        })
        assert verify_response.status_code in [201, 409]


class TestUserLogin:
    """Test user login endpoint"""
    
    @pytest.fixture(autouse=True)
    def setup_user(self):
        """Create a test user before each test"""
        client.post("/auth/register", json={
            "email": "testlogin@example.com",
            "password": "TestPassword123!",
            "username": "testlogin"
        })
        
    def test_login_valid_credentials(self):
        """Test successful login"""
        response = client.post("/auth/login", json={
            "email": "testlogin@example.com",
            "password": "TestPassword123!"
        })
        
        assert response.status_code == 200
        data = response.json()
        assert "access_token" in data
        assert "token_type" in data
        assert data["token_type"] == "bearer"
        assert "refresh_token" in data
        
    def test_login_invalid_password(self):
        """Test login with wrong password"""
        response = client.post("/auth/login", json={
            "email": "testlogin@example.com",
            "password": "WrongPassword123!"
        })
        
        assert response.status_code == 401
        assert "Invalid credentials" in response.json()["detail"]
        
    def test_login_nonexistent_user(self):
        """Test login with non-existent email"""
        response = client.post("/auth/login", json={
            "email": "nonexistent@example.com",
            "password": "TestPassword123!"
        })
        
        assert response.status_code == 401
        assert "Invalid credentials" in response.json()["detail"]
        
    def test_login_rate_limiting(self):
        """Test rate limiting on login attempts"""
        # Make multiple failed login attempts
        for i in range(6):
            response = client.post("/auth/login", json={
                "email": "testlogin@example.com",
                "password": f"WrongPassword{i}!"
            })
        
        # The 6th attempt should be rate limited
        assert response.status_code == 429
        assert "Too many login attempts" in response.json()["detail"]
        
    def test_login_xss_prevention(self):
        """Test XSS prevention in login error messages"""
        response = client.post("/auth/login", json={
            "email": "<script>alert('xss')</script>@example.com",
            "password": "TestPassword123!"
        })
        
        assert response.status_code in [401, 422]
        # Ensure script tags are not reflected in response
        assert "<script>" not in json.dumps(response.json())


class TestTokenManagement:
    """Test JWT token functionality"""
    
    @pytest.fixture
    def auth_headers(self):
        """Get auth headers with valid token"""
        # Register and login
        client.post("/auth/register", json={
            "email": "tokentest@example.com",
            "password": "TestPassword123!",
            "username": "tokentest"
        })
        
        response = client.post("/auth/login", json={
            "email": "tokentest@example.com",
            "password": "TestPassword123!"
        })
        
        token = response.json()["access_token"]
        return {"Authorization": f"Bearer {token}"}
        
    def test_protected_endpoint_with_valid_token(self, auth_headers):
        """Test accessing protected endpoint with valid token"""
        response = client.get("/auth/me", headers=auth_headers)
        
        assert response.status_code == 200
        data = response.json()
        assert data["email"] == "tokentest@example.com"
        
    def test_protected_endpoint_without_token(self):
        """Test accessing protected endpoint without token"""
        response = client.get("/auth/me")
        
        assert response.status_code == 401
        assert "Not authenticated" in response.json()["detail"]
        
    def test_protected_endpoint_with_invalid_token(self):
        """Test accessing protected endpoint with invalid token"""
        headers = {"Authorization": "Bearer invalid.token.here"}
        response = client.get("/auth/me", headers=headers)
        
        assert response.status_code == 401
        assert "Could not validate credentials" in response.json()["detail"]
        
    def test_token_expiration(self):
        """Test that expired tokens are rejected"""
        # Create an expired token
        expired_token = jwt.encode(
            {
                "sub": "tokentest@example.com",
                "exp": datetime.utcnow() - timedelta(hours=1)
            },
            "secret_key",  # This should match your app's secret
            algorithm="HS256"
        )
        
        headers = {"Authorization": f"Bearer {expired_token}"}
        response = client.get("/auth/me", headers=headers)
        
        assert response.status_code == 401
        
    def test_refresh_token(self):
        """Test token refresh functionality"""
        # Login to get tokens
        login_response = client.post("/auth/login", json={
            "email": "tokentest@example.com",
            "password": "TestPassword123!"
        })
        
        refresh_token = login_response.json()["refresh_token"]
        
        # Use refresh token to get new access token
        response = client.post("/auth/refresh", json={
            "refresh_token": refresh_token
        })
        
        assert response.status_code == 200
        data = response.json()
        assert "access_token" in data
        assert data["access_token"] != login_response.json()["access_token"]


class TestPasswordManagement:
    """Test password reset and change functionality"""
    
    def test_password_reset_request(self):
        """Test password reset request"""
        # Create user
        client.post("/auth/register", json={
            "email": "resettest@example.com",
            "password": "OldPassword123!",
            "username": "resettest"
        })
        
        # Request password reset
        response = client.post("/auth/password-reset", json={
            "email": "resettest@example.com"
        })
        
        assert response.status_code == 200
        assert "Password reset email sent" in response.json()["message"]
        
    def test_password_reset_nonexistent_user(self):
        """Test password reset for non-existent user"""
        response = client.post("/auth/password-reset", json={
            "email": "nonexistent@example.com"
        })
        
        # Should return success to prevent email enumeration
        assert response.status_code == 200
        assert "Password reset email sent" in response.json()["message"]
        
    @patch('backend.api.main.send_email')
    def test_password_change_authenticated(self, mock_send_email, auth_headers):
        """Test changing password while authenticated"""
        response = client.post("/auth/change-password", 
            headers=auth_headers,
            json={
                "current_password": "TestPassword123!",
                "new_password": "NewPassword123!"
            }
        )
        
        assert response.status_code == 200
        
        # Verify old password no longer works
        login_response = client.post("/auth/login", json={
            "email": "tokentest@example.com",
            "password": "TestPassword123!"
        })
        assert login_response.status_code == 401
        
        # Verify new password works
        login_response = client.post("/auth/login", json={
            "email": "tokentest@example.com",
            "password": "NewPassword123!"
        })
        assert login_response.status_code == 200


class TestEmailVerification:
    """Test email verification flow"""
    
    @patch('backend.api.main.send_email')
    def test_email_verification_flow(self, mock_send_email):
        """Test complete email verification flow"""
        # Register user
        register_response = client.post("/auth/register", json={
            "email": "verifytest@example.com",
            "password": "TestPassword123!",
            "username": "verifytest"
        })
        
        assert mock_send_email.called
        
        # Extract verification token from mock call
        verification_token = mock_send_email.call_args[0][2]  # Assuming token is 3rd argument
        
        # Verify email
        response = client.get(f"/auth/verify-email?token={verification_token}")
        
        assert response.status_code == 200
        assert "Email verified successfully" in response.json()["message"]
        
    def test_invalid_verification_token(self):
        """Test email verification with invalid token"""
        response = client.get("/auth/verify-email?token=invalid-token")
        
        assert response.status_code == 400
        assert "Invalid or expired token" in response.json()["detail"]


class TestSecurityHeaders:
    """Test security headers on auth endpoints"""
    
    def test_security_headers_present(self):
        """Test that security headers are set correctly"""
        response = client.post("/auth/login", json={
            "email": "test@example.com",
            "password": "TestPassword123!"
        })
        
        # Check security headers
        assert response.headers.get("X-Content-Type-Options") == "nosniff"
        assert response.headers.get("X-Frame-Options") == "DENY"
        assert response.headers.get("X-XSS-Protection") == "1; mode=block"
        assert "Content-Security-Policy" in response.headers


if __name__ == "__main__":
    pytest.main([__file__, "-v"])