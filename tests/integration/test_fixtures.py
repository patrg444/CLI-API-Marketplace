"""
Shared test fixtures and utilities for integration tests
"""

import pytest
import asyncio
import aiohttp
import os
from datetime import datetime, timedelta
import jwt
import json
from typing import Dict, Optional, AsyncGenerator
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
import redis.asyncio as redis


# Test configuration
TEST_CONFIG = {
    "api_base_url": os.getenv("TEST_API_URL", "http://localhost:8000"),
    "database_url": os.getenv("TEST_DB_URL", "postgresql+asyncpg://test:test@localhost/testdb"),
    "redis_url": os.getenv("TEST_REDIS_URL", "redis://localhost:6379/1"),
    "jwt_secret": os.getenv("TEST_JWT_SECRET", "test-secret-key"),
}


@pytest.fixture(scope="session")
def event_loop():
    """Create an instance of the default event loop for the test session."""
    loop = asyncio.get_event_loop_policy().new_event_loop()
    yield loop
    loop.close()


@pytest.fixture(scope="session")
async def db_engine():
    """Create database engine for tests"""
    engine = create_async_engine(
        TEST_CONFIG["database_url"],
        echo=False,
        pool_pre_ping=True,
        pool_size=5
    )
    
    yield engine
    
    await engine.dispose()


@pytest.fixture
async def db_session(db_engine):
    """Create a database session for tests"""
    async_session = sessionmaker(
        db_engine,
        class_=AsyncSession,
        expire_on_commit=False
    )
    
    async with async_session() as session:
        yield session
        await session.rollback()


@pytest.fixture
async def redis_client():
    """Create Redis client for tests"""
    client = await redis.from_url(
        TEST_CONFIG["redis_url"],
        encoding="utf-8",
        decode_responses=True
    )
    
    yield client
    
    await client.flushdb()
    await client.close()


@pytest.fixture
async def api_client() -> AsyncGenerator[aiohttp.ClientSession, None]:
    """Create HTTP client for API tests"""
    timeout = aiohttp.ClientTimeout(total=30)
    connector = aiohttp.TCPConnector(limit=100)
    
    async with aiohttp.ClientSession(
        base_url=TEST_CONFIG["api_base_url"],
        timeout=timeout,
        connector=connector,
        headers={"Content-Type": "application/json"}
    ) as session:
        yield session


@pytest.fixture
def test_user() -> Dict[str, any]:
    """Create test user data"""
    return {
        "id": "test-user-123",
        "email": "test@example.com",
        "username": "testuser",
        "password": "TestPass123!",
        "created_at": datetime.utcnow().isoformat()
    }


@pytest.fixture
def test_api() -> Dict[str, any]:
    """Create test API data"""
    return {
        "id": "test-api-456",
        "name": "Test Weather API",
        "description": "A test weather API for integration testing",
        "base_url": "https://api.test-weather.com",
        "owner_id": "test-user-123",
        "categories": ["weather", "data"],
        "status": "active",
        "created_at": datetime.utcnow().isoformat()
    }


@pytest.fixture
def test_subscription() -> Dict[str, any]:
    """Create test subscription data"""
    return {
        "id": "test-sub-789",
        "user_id": "test-user-123",
        "api_id": "test-api-456",
        "plan": "starter",
        "status": "active",
        "api_key": "sk_test_abcdef123456",
        "created_at": datetime.utcnow().isoformat(),
        "current_period_end": (datetime.utcnow() + timedelta(days=30)).isoformat()
    }


@pytest.fixture
def auth_headers(test_user) -> Dict[str, str]:
    """Create authentication headers with JWT token"""
    token_payload = {
        "sub": test_user["id"],
        "email": test_user["email"],
        "username": test_user["username"],
        "exp": datetime.utcnow() + timedelta(hours=1),
        "iat": datetime.utcnow()
    }
    
    token = jwt.encode(
        token_payload,
        TEST_CONFIG["jwt_secret"],
        algorithm="HS256"
    )
    
    return {"Authorization": f"Bearer {token}"}


@pytest.fixture
def admin_headers() -> Dict[str, str]:
    """Create admin authentication headers"""
    token_payload = {
        "sub": "admin-user-id",
        "email": "admin@apidirect.dev",
        "role": "admin",
        "exp": datetime.utcnow() + timedelta(hours=1),
        "iat": datetime.utcnow()
    }
    
    token = jwt.encode(
        token_payload,
        TEST_CONFIG["jwt_secret"],
        algorithm="HS256"
    )
    
    return {"Authorization": f"Bearer {token}"}


class TestDataBuilder:
    """Helper class to build test data"""
    
    @staticmethod
    def create_user(**kwargs) -> Dict:
        """Create user with custom attributes"""
        base_user = {
            "id": f"user-{datetime.utcnow().timestamp()}",
            "email": f"user{datetime.utcnow().timestamp()}@example.com",
            "username": f"user{int(datetime.utcnow().timestamp())}",
            "password": "SecurePass123!",
            "email_verified": False,
            "created_at": datetime.utcnow().isoformat()
        }
        base_user.update(kwargs)
        return base_user
    
    @staticmethod
    def create_api(**kwargs) -> Dict:
        """Create API with custom attributes"""
        base_api = {
            "id": f"api-{datetime.utcnow().timestamp()}",
            "name": f"Test API {datetime.utcnow().timestamp()}",
            "description": "Test API description",
            "base_url": f"https://api-{int(datetime.utcnow().timestamp())}.example.com",
            "documentation_url": "https://docs.example.com",
            "categories": ["test"],
            "status": "active",
            "created_at": datetime.utcnow().isoformat()
        }
        base_api.update(kwargs)
        return base_api
    
    @staticmethod
    def create_subscription(**kwargs) -> Dict:
        """Create subscription with custom attributes"""
        base_subscription = {
            "id": f"sub-{datetime.utcnow().timestamp()}",
            "plan": "free",
            "status": "active",
            "api_key": f"sk_test_{int(datetime.utcnow().timestamp())}",
            "created_at": datetime.utcnow().isoformat(),
            "current_period_end": (datetime.utcnow() + timedelta(days=30)).isoformat()
        }
        base_subscription.update(kwargs)
        return base_subscription


class APITestHelper:
    """Helper methods for API testing"""
    
    @staticmethod
    async def create_test_user(api_client: aiohttp.ClientSession, **kwargs) -> Dict:
        """Create a test user via API"""
        user_data = TestDataBuilder.create_user(**kwargs)
        
        async with api_client.post("/auth/register", json={
            "email": user_data["email"],
            "password": user_data["password"],
            "username": user_data["username"]
        }) as resp:
            result = await resp.json()
            return result
    
    @staticmethod
    async def create_test_api(
        api_client: aiohttp.ClientSession,
        auth_headers: Dict[str, str],
        **kwargs
    ) -> Dict:
        """Create a test API via API"""
        api_data = TestDataBuilder.create_api(**kwargs)
        
        async with api_client.post(
            "/apis",
            json=api_data,
            headers=auth_headers
        ) as resp:
            result = await resp.json()
            return result
    
    @staticmethod
    async def subscribe_to_api(
        api_client: aiohttp.ClientSession,
        auth_headers: Dict[str, str],
        api_id: str,
        plan: str = "starter"
    ) -> Dict:
        """Subscribe to an API"""
        async with api_client.post(
            "/subscriptions",
            json={"api_id": api_id, "plan": plan},
            headers=auth_headers
        ) as resp:
            result = await resp.json()
            return result
    
    @staticmethod
    async def make_api_call(
        api_client: aiohttp.ClientSession,
        api_id: str,
        api_key: str,
        endpoint: str,
        **params
    ) -> Dict:
        """Make a call to an API through the gateway"""
        headers = {"X-API-Key": api_key}
        
        async with api_client.get(
            f"/gateway/{api_id}{endpoint}",
            headers=headers,
            params=params
        ) as resp:
            result = await resp.json()
            return result


class WebSocketTestHelper:
    """Helper methods for WebSocket testing"""
    
    @staticmethod
    async def connect_websocket(token: str) -> any:
        """Connect to WebSocket with authentication"""
        import websockets
        
        ws_url = TEST_CONFIG["api_base_url"].replace("http", "ws")
        ws = await websockets.connect(f"{ws_url}/ws?token={token}")
        return ws
    
    @staticmethod
    async def wait_for_message(websocket, message_type: str, timeout: float = 5.0) -> Dict:
        """Wait for specific message type from WebSocket"""
        start_time = asyncio.get_event_loop().time()
        
        while asyncio.get_event_loop().time() - start_time < timeout:
            try:
                message = await asyncio.wait_for(websocket.recv(), timeout=1.0)
                data = json.loads(message)
                if data.get("type") == message_type:
                    return data
            except asyncio.TimeoutError:
                continue
        
        raise TimeoutError(f"Did not receive message of type {message_type} within {timeout} seconds")


class DatabaseTestHelper:
    """Helper methods for database testing"""
    
    @staticmethod
    async def seed_test_data(db_session: AsyncSession):
        """Seed database with test data"""
        # This would contain actual ORM operations
        # For now, it's a placeholder
        pass
    
    @staticmethod
    async def cleanup_test_data(db_session: AsyncSession, prefix: str = "test-"):
        """Clean up test data from database"""
        # This would contain actual cleanup operations
        # For now, it's a placeholder
        pass


# Pytest markers for different test types
def pytest_configure(config):
    """Configure pytest with custom markers"""
    config.addinivalue_line("markers", "slow: marks tests as slow")
    config.addinivalue_line("markers", "integration: marks tests as integration tests")
    config.addinivalue_line("markers", "requires_redis: marks tests that require Redis")
    config.addinivalue_line("markers", "requires_db: marks tests that require database")
    config.addinivalue_line("markers", "requires_websocket: marks tests that require WebSocket")