"""
Simplified database tests that work with our implementation
"""

import pytest
import asyncio
from datetime import datetime, timedelta
from sqlalchemy import text, select
from sqlalchemy.exc import IntegrityError

import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from database import (
    DatabaseManager,
    User,
    API,
    Subscription,
    Transaction,
    Base
)


@pytest.fixture
async def db_manager():
    """Create a test database manager"""
    # Use in-memory SQLite for tests
    manager = DatabaseManager(
        connection_string="sqlite+aiosqlite:///:memory:",
        pool_size=5,
        max_overflow=10
    )
    await manager.initialize()
    try:
        yield manager
    finally:
        await manager.close()


@pytest.fixture
async def db_session(db_manager):
    """Get a database session for testing"""
    async with db_manager.get_session() as session:
        yield session


class TestDatabaseConnection:
    """Test database connection management"""
    
    @pytest.mark.asyncio
    async def test_connection_initialization(self, db_manager):
        """Test database connection is properly initialized"""
        assert db_manager.engine is not None
        assert db_manager.async_session_maker is not None
    
    @pytest.mark.asyncio
    async def test_simple_query(self, db_manager):
        """Test executing a simple query"""
        async with db_manager.get_session() as session:
            result = await session.execute(text("SELECT 1 as value"))
            row = result.first()
            assert row.value == 1
    
    @pytest.mark.asyncio
    async def test_health_check(self, db_manager):
        """Test database health check"""
        health = await db_manager.health_check()
        
        assert health['status'] == 'healthy'
        assert health['response_time_ms'] >= 0
        assert 'connection_count' in health
        assert 'pool_size' in health


class TestUserOperations:
    """Test user-related database operations"""
    
    @pytest.mark.asyncio
    async def test_create_user(self, db_manager):
        """Test creating a new user"""
        async with db_manager.get_session() as session:
            user = User(
                email="test@example.com",
                username="testuser",
                password_hash="hashed_password",
                name="Test User"
            )
            session.add(user)
            await session.commit()
            
            # Verify user was created
            result = await session.execute(
                select(User).where(User.email == "test@example.com")
            )
            saved_user = result.scalar_one()
            assert saved_user.email == "test@example.com"
            assert saved_user.username == "testuser"
            assert saved_user.name == "Test User"
    
    @pytest.mark.asyncio
    async def test_unique_email_constraint(self, db_manager):
        """Test unique email constraint"""
        async with db_manager.get_session() as session:
            # Create first user
            user1 = User(
                email="duplicate@example.com",
                username="user1",
                password_hash="hash1"
            )
            session.add(user1)
            await session.commit()
        
        async with db_manager.get_session() as session:
            # Try to create duplicate
            user2 = User(
                email="duplicate@example.com",
                username="user2",
                password_hash="hash2"
            )
            session.add(user2)
            
            with pytest.raises(IntegrityError):
                await session.commit()
    
    @pytest.mark.asyncio
    async def test_update_user(self, db_manager):
        """Test updating user information"""
        async with db_manager.get_session() as session:
            # Create user
            user = User(
                email="update@example.com",
                username="updateuser",
                password_hash="hash",
                name="Original Name"
            )
            session.add(user)
            await session.commit()
            user_id = user.id
        
        async with db_manager.get_session() as session:
            # Update user
            result = await session.execute(
                select(User).where(User.id == user_id)
            )
            user = result.scalar_one()
            user.name = "Updated Name"
            user.company = "Test Company"
            await session.commit()
        
        async with db_manager.get_session() as session:
            # Verify update
            result = await session.execute(
                select(User).where(User.id == user_id)
            )
            updated_user = result.scalar_one()
            assert updated_user.name == "Updated Name"
            assert updated_user.company == "Test Company"


class TestAPIOperations:
    """Test API-related database operations"""
    
    @pytest.mark.asyncio
    async def test_create_api(self, db_manager):
        """Test creating a new API"""
        async with db_manager.get_session() as session:
            # Create user first
            user = User(
                email="api_owner@example.com",
                username="apiowner",
                password_hash="hash"
            )
            session.add(user)
            await session.flush()
            
            # Create API
            api = API(
                name="test-api",
                user_id=user.id,
                owner_id=user.id,
                description="Test API Description",
                deployment_type="hosted",
                pricing_type="freemium"
            )
            session.add(api)
            await session.commit()
            
            # Verify API was created
            result = await session.execute(
                select(API).where(API.name == "test-api")
            )
            saved_api = result.scalar_one()
            assert saved_api.name == "test-api"
            assert saved_api.description == "Test API Description"
            assert saved_api.deployment_type == "hosted"
    
    @pytest.mark.asyncio
    async def test_user_api_relationship(self, db_manager):
        """Test relationship between users and APIs"""
        async with db_manager.get_session() as session:
            # Create user
            user = User(
                email="relation@example.com",
                username="relationuser",
                password_hash="hash"
            )
            session.add(user)
            await session.flush()
            
            # Create multiple APIs
            for i in range(3):
                api = API(
                    name=f"api-{i}",
                    user_id=user.id,
                    owner_id=user.id,
                    description=f"API {i}"
                )
                session.add(api)
            
            await session.commit()
            user_id = user.id
        
        async with db_manager.get_session() as session:
            # Query user with APIs
            result = await session.execute(
                select(User).where(User.id == user_id)
            )
            user = result.scalar_one()
            
            # Check APIs count
            api_result = await session.execute(
                select(API).where(API.owner_id == user_id)
            )
            apis = api_result.scalars().all()
            assert len(apis) == 3
            assert all(api.owner_id == user_id for api in apis)


class TestSubscriptionOperations:
    """Test subscription-related operations"""
    
    @pytest.mark.asyncio
    async def test_create_subscription(self, db_manager):
        """Test creating a subscription"""
        async with db_manager.get_session() as session:
            # Create user and API
            user = User(
                email="subscriber@example.com",
                username="subscriber",
                password_hash="hash"
            )
            session.add(user)
            await session.flush()
            
            api = API(
                name="premium-api",
                user_id=user.id,
                owner_id=user.id,
                description="Premium API"
            )
            session.add(api)
            await session.flush()
            
            # Create subscription
            subscription = Subscription(
                user_id=user.id,
                api_id=api.id,
                plan="pro",
                status="active",
                amount=99.99,
                currency="USD"
            )
            session.add(subscription)
            await session.commit()
            
            # Verify subscription
            result = await session.execute(
                select(Subscription).where(
                    Subscription.user_id == user.id,
                    Subscription.api_id == api.id
                )
            )
            saved_sub = result.scalar_one()
            assert saved_sub.plan == "pro"
            assert saved_sub.status == "active"
            assert float(saved_sub.amount) == 99.99


class TestTransactionOperations:
    """Test transaction/billing operations"""
    
    @pytest.mark.asyncio
    async def test_database_transaction_rollback(self, db_manager):
        """Test transaction rollback on error"""
        async with db_manager.get_session() as session:
            try:
                # Start transaction
                user = User(
                    email="rollback@example.com",
                    username="rollbackuser",
                    password_hash="hash"
                )
                session.add(user)
                await session.flush()
                
                # This should cause an error (duplicate username)
                duplicate = User(
                    email="different@example.com",
                    username="rollbackuser",  # Duplicate
                    password_hash="hash"
                )
                session.add(duplicate)
                await session.commit()
            except IntegrityError:
                # Transaction should be rolled back
                pass
        
        # Verify no users were created
        async with db_manager.get_session() as session:
            result = await session.execute(
                select(User).where(User.email == "rollback@example.com")
            )
            assert result.scalar_one_or_none() is None


class TestQueryPerformance:
    """Test query performance and optimization"""
    
    @pytest.mark.asyncio
    async def test_bulk_insert(self, db_manager):
        """Test bulk insert performance"""
        async with db_manager.get_session() as session:
            # Create multiple users at once
            users = []
            for i in range(100):
                users.append(User(
                    email=f"bulk{i}@example.com",
                    username=f"bulk{i}",
                    password_hash=f"hash{i}"
                ))
            
            session.add_all(users)
            await session.commit()
            
            # Verify all were created
            result = await session.execute(
                select(User).where(User.email.like("bulk%@example.com"))
            )
            saved_users = result.scalars().all()
            assert len(saved_users) == 100
    
    @pytest.mark.asyncio
    async def test_pagination(self, db_manager):
        """Test query pagination"""
        async with db_manager.get_session() as session:
            # Create test data
            for i in range(25):
                api = API(
                    name=f"page-api-{i:02d}",
                    user_id="00000000-0000-0000-0000-000000000000",
                    owner_id="00000000-0000-0000-0000-000000000000",
                    description=f"API {i}"
                )
                session.add(api)
            await session.commit()
            
            # Test pagination
            page_size = 10
            
            # Page 1
            result = await session.execute(
                select(API)
                .where(API.name.like("page-api-%"))
                .order_by(API.name)
                .limit(page_size)
                .offset(0)
            )
            page1 = result.scalars().all()
            assert len(page1) == 10
            assert page1[0].name == "page-api-00"
            
            # Page 2
            result = await session.execute(
                select(API)
                .where(API.name.like("page-api-%"))
                .order_by(API.name)
                .limit(page_size)
                .offset(10)
            )
            page2 = result.scalars().all()
            assert len(page2) == 10
            assert page2[0].name == "page-api-10"
            
            # Page 3 (partial)
            result = await session.execute(
                select(API)
                .where(API.name.like("page-api-%"))
                .order_by(API.name)
                .limit(page_size)
                .offset(20)
            )
            page3 = result.scalars().all()
            assert len(page3) == 5
            assert page3[0].name == "page-api-20"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])