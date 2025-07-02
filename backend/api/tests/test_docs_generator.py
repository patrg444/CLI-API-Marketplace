"""
Test API documentation generator functionality
"""

import pytest
from unittest.mock import Mock, AsyncMock, patch
from fastapi.testclient import TestClient
from fastapi import HTTPException
import sys
import os
import uuid
from datetime import datetime
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
        from docs_generator import DocsGenerator, EndpointDoc
        
        # Set the global db_pool variable
        main.db_pool = Mock()
        main.db_pool.acquire = Mock(return_value=AsyncContextManager())
        main.api_key_manager = Mock()
        main.trial_manager = Mock()
        main.docs_generator = Mock()

client = TestClient(app)


class TestDocsGenerator:
    """Test documentation generation functionality"""
    
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
    
    def test_generate_openapi_spec(self):
        """Test OpenAPI spec generation"""
        # Initialize docs generator
        if not hasattr(main, 'docs_generator') or main.docs_generator is None:
            main.docs_generator = DocsGenerator(main.db_pool)
        
        # Mock OpenAPI spec generation
        main.docs_generator.generate_openapi_spec = AsyncMock(return_value={
            "openapi": "3.0.0",
            "info": {
                "title": "Test API",
                "version": "1.0.0",
                "description": "A test API for marketplace"
            },
            "servers": [
                {"url": "https://api.example.com", "description": "Production"}
            ],
            "paths": {
                "/users": {
                    "get": {
                        "summary": "List users",
                        "responses": {
                            "200": {"description": "Success"}
                        }
                    }
                }
            }
        })
        
        response = client.get(f"/api/docs/{self.test_api_id}/openapi")
        
        assert response.status_code == 200
        data = response.json()
        assert data["openapi"] == "3.0.0"
        assert data["info"]["title"] == "Test API"
        assert "/users" in data["paths"]
    
    def test_generate_code_examples(self):
        """Test code example generation"""
        # Initialize docs generator
        if not hasattr(main, 'docs_generator') or main.docs_generator is None:
            main.docs_generator = DocsGenerator(main.db_pool)
        
        # Mock code example generation
        main.docs_generator.generate_code_examples = AsyncMock(return_value={
            "curl": 'curl -X GET "https://api.example.com/users" -H "X-API-Key: YOUR_API_KEY"',
            "python": '''import requests

url = "https://api.example.com/users"
headers = {
    "Accept": "application/json",
    "X-API-Key": "YOUR_API_KEY"
}

response = requests.get(url, headers=headers)
print(response.json())''',
            "javascript": '''const url = "https://api.example.com/users";

const options = {
  method: "GET",
  headers: {
    "Accept": "application/json",
    "X-API-Key": "YOUR_API_KEY"
  }
};

fetch(url, options)
  .then(response => response.json())
  .then(data => console.log(data))
  .catch(error => console.error('Error:', error));'''
        })
        
        response = client.get(
            f"/api/docs/{self.test_api_id}/code-examples",
            params={
                "endpoint": "/users",
                "method": "GET",
                "languages": "curl,python,javascript"
            }
        )
        
        assert response.status_code == 200
        data = response.json()
        assert "curl" in data
        assert "python" in data
        assert "javascript" in data
        assert "X-API-Key" in data["curl"]
    
    def test_export_postman_collection(self):
        """Test Postman collection export"""
        # Initialize docs generator
        if not hasattr(main, 'docs_generator') or main.docs_generator is None:
            main.docs_generator = DocsGenerator(main.db_pool)
        
        # Mock Postman export
        main.docs_generator.export_postman_collection = AsyncMock(return_value={
            "info": {
                "name": "Test API",
                "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
            },
            "auth": {
                "type": "apikey",
                "apikey": [{
                    "key": "key",
                    "value": "{{api_key}}",
                    "type": "string"
                }]
            },
            "item": [{
                "name": "Users",
                "item": [{
                    "name": "List users",
                    "request": {
                        "method": "GET",
                        "url": {
                            "raw": "{{base_url}}/users",
                            "host": ["{{base_url}}"],
                            "path": ["users"]
                        }
                    }
                }]
            }]
        })
        
        response = client.get(f"/api/docs/{self.test_api_id}/postman")
        
        assert response.status_code == 200
        data = response.json()
        assert data["info"]["name"] == "Test API"
        assert len(data["item"]) > 0
        assert data["auth"]["type"] == "apikey"
    
    def test_generate_sdk(self):
        """Test SDK generation"""
        # Initialize docs generator
        if not hasattr(main, 'docs_generator') or main.docs_generator is None:
            main.docs_generator = DocsGenerator(main.db_pool)
        
        # Mock SDK generation
        main.docs_generator.generate_sdk = AsyncMock(return_value='''"""
Test API Python SDK
"""

import requests
from typing import Dict, Any, Optional


class TestAPIClient:
    """Client for Test API"""
    
    def __init__(self, api_key: str, base_url: str = "https://api.example.com"):
        self.api_key = api_key
        self.base_url = base_url.rstrip("/")
        self.session = requests.Session()
        self.session.headers.update({
            "X-API-Key": api_key,
            "Accept": "application/json"
        })
    
    def list_users(self, **kwargs) -> Dict[str, Any]:
        """List users"""
        url = f"{self.base_url}/users"
        response = self.session.get(url, params=kwargs)
        response.raise_for_status()
        return response.json()
''')
        
        response = client.get(f"/api/docs/{self.test_api_id}/sdk/python")
        
        assert response.status_code == 200
        assert response.headers["content-type"] == "text/plain"
        assert "content-disposition" in response.headers
        assert "TestAPIClient" in response.text
    
    def test_add_endpoint_documentation(self):
        """Test adding endpoint documentation (API owner only)"""
        # Mock API ownership check
        mock_conn = main.db_pool.acquire.return_value.conn
        mock_conn.fetchval.side_effect = [
            self.test_user_id,  # User owns the API
            str(uuid.uuid4())   # New endpoint ID
        ]
        
        endpoint_doc = {
            "path": "/users/{id}",
            "method": "GET",
            "summary": "Get user by ID",
            "description": "Retrieve a specific user by their ID",
            "parameters": [{
                "name": "id",
                "in": "path",
                "description": "User ID",
                "required": True,
                "schema": {"type": "string", "format": "uuid"}
            }],
            "responses": {
                "200": {
                    "description": "User found",
                    "content": {
                        "application/json": {
                            "schema": {
                                "type": "object",
                                "properties": {
                                    "id": {"type": "string"},
                                    "name": {"type": "string"},
                                    "email": {"type": "string"}
                                }
                            }
                        }
                    }
                },
                "404": {"description": "User not found"}
            },
            "tags": ["Users"],
            "deprecated": False,
            "security": [{"apiKey": []}]
        }
        
        response = client.post(
            f"/api/docs/{self.test_api_id}/endpoints",
            json=endpoint_doc,
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        assert response.json()["message"] == "Endpoint documentation added successfully"
    
    def test_add_endpoint_documentation_unauthorized(self):
        """Test adding documentation by non-owner"""
        # Mock API ownership check - different owner
        mock_conn = main.db_pool.acquire.return_value.conn
        mock_conn.fetchval.return_value = str(uuid.uuid4())  # Different user owns the API
        
        endpoint_doc = {
            "path": "/test",
            "method": "GET",
            "summary": "Test endpoint",
            "responses": {"200": {"description": "Success"}}
        }
        
        response = client.post(
            f"/api/docs/{self.test_api_id}/endpoints",
            json=endpoint_doc,
            headers=self.auth_headers
        )
        
        assert response.status_code == 403
        assert "Only API owners can add documentation" in response.json()["detail"]


class TestCodeGeneration:
    """Test code generation methods"""
    
    def test_curl_generation(self):
        """Test cURL code generation"""
        generator = DocsGenerator(main.db_pool)
        
        api = {
            "base_url": "https://api.example.com",
            "auth_type": "apiKey",
            "auth_header": "X-API-Key"
        }
        
        endpoint = {
            "path": "/users",
            "method": "GET",
            "parameters": [{
                "name": "limit",
                "location": "query",
                "example": "10"
            }]
        }
        
        code = generator._generate_curl(api, endpoint, None)
        
        assert "curl -X GET" in code
        assert "https://api.example.com/users" in code
        assert 'X-API-Key: YOUR_API_KEY' in code
        assert "-H \"Accept: application/json\"" in code
    
    def test_python_generation(self):
        """Test Python code generation"""
        generator = DocsGenerator(main.db_pool)
        
        api = {
            "base_url": "https://api.example.com",
            "auth_type": "bearer"
        }
        
        endpoint = {
            "path": "/users",
            "method": "POST",
            "parameters": []
        }
        
        request_body = {
            "name": "John Doe",
            "email": "john@example.com"
        }
        
        code = generator._generate_python(api, endpoint, request_body)
        
        assert "import requests" in code
        assert 'url = "https://api.example.com/users"' in code
        assert '"Authorization": "Bearer YOUR_TOKEN"' in code
        assert "response = requests.post" in code
        assert "json=data" in code
    
    def test_javascript_generation(self):
        """Test JavaScript code generation"""
        generator = DocsGenerator(main.db_pool)
        
        api = {"base_url": "https://api.example.com"}
        endpoint = {"path": "/users", "method": "GET", "parameters": []}
        
        code = generator._generate_javascript(api, endpoint, None)
        
        assert 'const url = "https://api.example.com/users"' in code
        assert 'method: "GET"' in code
        assert "fetch(url, options)" in code
        assert ".then(response => response.json())" in code


if __name__ == "__main__":
    pytest.main([__file__, "-v"])