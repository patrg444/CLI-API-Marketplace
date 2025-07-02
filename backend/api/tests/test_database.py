"""
Database operation tests for the API-Direct backend
Tests connection management, transactions, queries, and migrations
"""

import pytest
import asyncio
from datetime import datetime, timedelta
from unittest.mock import Mock, patch, AsyncMock
import asyncpg
from sqlalchemy import create_engine, text
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from sqlalchemy.pool import NullPool

try:
    from ..database import (
        DatabaseManager,
        get_db,
        User,
        API,
        Subscription,
        Transaction,
        DeadlockError,
        IntegrityError
    )
except ImportError:
    from database import (
        DatabaseManager,
        get_db,
        User,
        API,
        Subscription,
        Transaction,
        DeadlockError,
        IntegrityError
    )


class TestDatabaseConnection:
    """Test database connection management"""
    
    @pytest.fixture
    async def db_manager(self):
        """Create a test database manager"""
        # Use in-memory SQLite for tests
        manager = DatabaseManager(
            connection_string="sqlite+aiosqlite:///:memory:",
            pool_size=5,
            max_overflow=10
        )
        await manager.initialize()
        yield manager
        await manager.close()
    
    @pytest.mark.asyncio
    async def test_connection_pool_initialization(self, db_manager):
        """Test connection pool is properly initialized"""
        assert db_manager.engine is not None
        assert db_manager.async_session_maker is not None
        
        # Verify pool settings
        pool = db_manager.engine.pool
        assert pool.size() <= 5  # Pool size
        assert pool.overflow() <= 10  # Max overflow
    
    @pytest.mark.asyncio
    async def test_get_session(self, db_manager):
        """Test getting a database session"""
        async with db_manager.get_session() as session:
            assert isinstance(session, AsyncSession)
            assert session.is_active
            
            # Test simple query
            result = await session.execute(text("SELECT 1"))
            assert result.scalar() == 1
    
    @pytest.mark.asyncio
    async def test_connection_pool_exhaustion(self, db_manager):
        """Test behavior when connection pool is exhausted"""
        sessions = []
        
        # Acquire all connections
        for _ in range(15):  # pool_size + max_overflow
            session = await db_manager.get_session().__aenter__()
            sessions.append(session)
        
        # Try to get one more connection (should timeout or raise)
        with pytest.raises(asyncio.TimeoutError):
            async with asyncio.timeout(1):
                async with db_manager.get_session() as session:
                    pass
        
        # Clean up
        for session in sessions:
            await session.close()
    
    @pytest.mark.asyncio
    async def test_connection_retry_on_failure(self):
        """Test connection retry logic"""
        retry_count = 0
        
        async def mock_connect():
            nonlocal retry_count
            retry_count += 1
            if retry_count < 3:
                raise asyncpg.PostgresConnectionError("Connection failed")
            return Mock()
        
        manager = DatabaseManager("postgresql://test")
        with patch.object(manager, '_create_connection', side_effect=mock_connect):
            connection = await manager.connect_with_retry()
            
            assert retry_count == 3
            assert connection is not None
    
    @pytest.mark.asyncio
    async def test_connection_health_check(self, db_manager):
        """Test database health check"""
        health = await db_manager.health_check()
        
        assert health['status'] == 'healthy'
        assert 'connection_count' in health
        assert 'pool_size' in health
        assert health['response_time_ms'] >= 0


class TestTransactionManagement:
    """Test database transaction handling"""
    
    @pytest.fixture
    async def db_session(self, db_manager):
        """Get a database session for testing"""
        async with db_manager.get_session() as session:
            yield session
    
    @pytest.mark.asyncio
    async def test_successful_transaction_commit(self, db_session):
        """Test successful transaction commit"""
        async with db_session.begin():
            # Create test user
            user = User(
                email="test@example.com",
                username="testuser",
                password_hash="hashed_password"
            )
            db_session.add(user)
        
        # Verify user was saved
        result = await db_session.execute(
            text("SELECT * FROM users WHERE email = :email"),
            {"email": "test@example.com"}
        )
        saved_user = result.first()
        assert saved_user is not None
        assert saved_user.email == "test@example.com"
    
    @pytest.mark.asyncio
    async def test_transaction_rollback_on_error(self, db_session):
        """Test transaction rollback on error"""
        try:
            async with db_session.begin():
                # Create test user
                user = User(
                    email="rollback@example.com",
                    username="rollbackuser",
                    password_hash="hashed_password"
                )
                db_session.add(user)
                
                # Simulate error
                raise Exception("Simulated error")
        except Exception:
            pass
        
        # Verify user was not saved
        result = await db_session.execute(
            text("SELECT * FROM users WHERE email = :email"),
            {"email": "rollback@example.com"}
        )
        assert result.first() is None
    
    @pytest.mark.asyncio
    async def test_nested_transactions(self, db_session):
        """Test nested transaction handling with savepoints"""
        async with db_session.begin():
            # Outer transaction
            user = User(
                email="outer@example.com",
                username="outeruser",
                password_hash="hash1"
            )
            db_session.add(user)
            
            # Nested transaction (savepoint)
            async with db_session.begin_nested():
                api = API(
                    name="test-api",
                    owner_id=user.id,
                    description="Test API"
                )
                db_session.add(api)
                
                # Rollback nested transaction
                raise Exception("Rollback savepoint")
        
        # User should be saved, API should not
        users = await db_session.execute(
            text("SELECT COUNT(*) FROM users WHERE email = :email"),
            {"email": "outer@example.com"}
        )
        assert users.scalar() == 1
        
        apis = await db_session.execute(
            text("SELECT COUNT(*) FROM apis WHERE name = :name"),
            {"name": "test-api"}
        )
        assert apis.scalar() == 0
    
    @pytest.mark.asyncio
    async def test_transaction_isolation(self, db_manager):
        """Test transaction isolation levels"""
        async with db_manager.get_session() as session1:
            async with db_manager.get_session() as session2:
                # Start transaction in session1
                async with session1.begin():
                    user = User(
                        email="isolated@example.com",
                        username="isolated",
                        password_hash="hash"
                    )
                    session1.add(user)
                    await session1.flush()
                    
                    # Try to read from session2 (should not see uncommitted data)
                    result = await session2.execute(
                        text("SELECT * FROM users WHERE email = :email"),
                        {"email": "isolated@example.com"}
                    )
                    assert result.first() is None
                
                # After commit, session2 should see the data
                result = await session2.execute(
                    text("SELECT * FROM users WHERE email = :email"),
                    {"email": "isolated@example.com"}
                )
                assert result.first() is not None


class TestQueryOperations:
    """Test database query operations"""
    
    @pytest.mark.asyncio
    async def test_bulk_insert_performance(self, db_session):
        """Test bulk insert operations"""
        users = []
        for i in range(1000):
            users.append({
                'email': f'user{i}@example.com',
                'username': f'user{i}',
                'password_hash': f'hash{i}',
                'created_at': datetime.utcnow()
            })
        
        start_time = asyncio.get_event_loop().time()
        
        # Bulk insert
        await db_session.execute(
            text("""
                INSERT INTO users (email, username, password_hash, created_at)
                VALUES (:email, :username, :password_hash, :created_at)
            """),
            users
        )
        await db_session.commit()
        
        end_time = asyncio.get_event_loop().time()
        
        # Verify all inserted
        count = await db_session.execute(text("SELECT COUNT(*) FROM users"))
        assert count.scalar() == 1000
        
        # Performance check (should be fast)
        assert end_time - start_time < 1.0  # Less than 1 second for 1000 records
    
    @pytest.mark.asyncio
    async def test_complex_join_query(self, db_session):
        """Test complex join queries"""
        # Setup test data
        user = User(
            email="creator@example.com",
            username="creator",
            password_hash="hash"
        )
        db_session.add(user)
        await db_session.flush()
        
        api = API(
            name="test-api",
            owner_id=user.id,
            description="Test API",
            pricing_type="freemium"
        )
        db_session.add(api)
        await db_session.flush()
        
        subscription = Subscription(
            user_id=user.id,
            api_id=api.id,
            plan="pro",
            status="active"
        )
        db_session.add(subscription)
        await db_session.commit()
        
        # Complex join query
        result = await db_session.execute(
            text("""
                SELECT 
                    u.username,
                    a.name as api_name,
                    s.plan,
                    s.status
                FROM users u
                JOIN apis a ON a.owner_id = u.id
                JOIN subscriptions s ON s.api_id = a.id
                WHERE u.email = :email
            """),
            {"email": "creator@example.com"}
        )
        
        row = result.first()
        assert row.username == "creator"
        assert row.api_name == "test-api"
        assert row.plan == "pro"
        assert row.status == "active"
    
    @pytest.mark.asyncio
    async def test_query_with_pagination(self, db_session):
        """Test paginated query results"""
        # Create test data
        for i in range(50):
            api = API(
                name=f"api-{i:03d}",
                owner_id=1,
                description=f"API {i}",
                created_at=datetime.utcnow() - timedelta(days=i)
            )
            db_session.add(api)
        await db_session.commit()
        
        # Test pagination
        page_size = 10
        for page in range(5):
            offset = page * page_size
            
            result = await db_session.execute(
                text("""
                    SELECT name, description
                    FROM apis
                    ORDER BY created_at DESC
                    LIMIT :limit OFFSET :offset
                """),
                {"limit": page_size, "offset": offset}
            )
            
            rows = result.fetchall()
            assert len(rows) == page_size
            
            # Verify correct items
            for i, row in enumerate(rows):
                expected_index = offset + i
                assert row.name == f"api-{expected_index:03d}"
    
    @pytest.mark.asyncio
    async def test_query_timeout(self, db_session):
        """Test query timeout handling"""
        # Simulate slow query
        with pytest.raises(asyncio.TimeoutError):
            async with asyncio.timeout(0.1):  # 100ms timeout
                await db_session.execute(
                    text("SELECT pg_sleep(1)")  # Sleep for 1 second
                )


class TestDatabaseMigrations:
    """Test database migration operations"""
    
    @pytest.mark.asyncio
    async def test_migration_execution(self, db_manager):
        """Test database migration execution"""
        migrations = [
            """
            CREATE TABLE IF NOT EXISTS test_migration (
                id SERIAL PRIMARY KEY,
                version VARCHAR(50) NOT NULL,
                applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
            """,
            """
            ALTER TABLE test_migration 
            ADD COLUMN description TEXT
            """,
            """
            CREATE INDEX idx_test_migration_version 
            ON test_migration(version)
            """
        ]
        
        async with db_manager.get_session() as session:
            for migration in migrations:
                await session.execute(text(migration))
            await session.commit()
            
            # Verify migrations applied
            result = await session.execute(
                text("""
                    SELECT column_name 
                    FROM information_schema.columns 
                    WHERE table_name = 'test_migration'
                """)
            )
            columns = [row[0] for row in result]
            assert 'id' in columns
            assert 'version' in columns
            assert 'description' in columns
    
    @pytest.mark.asyncio
    async def test_migration_rollback(self, db_manager):
        """Test migration rollback on failure"""
        async with db_manager.get_session() as session:
            try:
                async with session.begin():
                    # First migration succeeds
                    await session.execute(
                        text("CREATE TABLE test_table (id INT PRIMARY KEY)")
                    )
                    
                    # Second migration fails
                    await session.execute(
                        text("CREATE TABLE test_table (id INT PRIMARY KEY)")  # Duplicate
                    )
            except Exception:
                pass
            
            # Verify table was not created due to rollback
            result = await session.execute(
                text("""
                    SELECT COUNT(*) 
                    FROM information_schema.tables 
                    WHERE table_name = 'test_table'
                """)
            )
            assert result.scalar() == 0


class TestDatabaseOptimization:
    """Test database optimization and performance"""
    
    @pytest.mark.asyncio
    async def test_index_usage(self, db_session):
        """Test that queries use appropriate indexes"""
        # Create table with index
        await db_session.execute(
            text("""
                CREATE TABLE indexed_table (
                    id SERIAL PRIMARY KEY,
                    email VARCHAR(255) UNIQUE,
                    status VARCHAR(50),
                    created_at TIMESTAMP
                );
                CREATE INDEX idx_status_created 
                ON indexed_table(status, created_at DESC);
            """)
        )
        
        # Insert test data
        for i in range(1000):
            await db_session.execute(
                text("""
                    INSERT INTO indexed_table (email, status, created_at)
                    VALUES (:email, :status, :created_at)
                """),
                {
                    'email': f'user{i}@example.com',
                    'status': 'active' if i % 2 == 0 else 'inactive',
                    'created_at': datetime.utcnow() - timedelta(hours=i)
                }
            )
        await db_session.commit()
        
        # Query that should use index
        explain_result = await db_session.execute(
            text("""
                EXPLAIN SELECT * FROM indexed_table
                WHERE status = 'active'
                ORDER BY created_at DESC
                LIMIT 10
            """)
        )
        
        explain_plan = str(explain_result.fetchall())
        # Verify index is being used (exact output varies by database)
        assert 'index' in explain_plan.lower() or 'idx_status_created' in explain_plan
    
    @pytest.mark.asyncio
    async def test_connection_pool_metrics(self, db_manager):
        """Test connection pool monitoring"""
        metrics = await db_manager.get_pool_metrics()
        
        assert 'size' in metrics
        assert 'checked_in' in metrics
        assert 'checked_out' in metrics
        assert 'overflow' in metrics
        assert 'total' in metrics
        
        # Verify metrics are reasonable
        assert metrics['total'] == metrics['checked_in'] + metrics['checked_out']
        assert metrics['overflow'] >= 0
        assert metrics['size'] > 0


class TestDatabaseErrorHandling:
    """Test database error handling scenarios"""
    
    @pytest.mark.asyncio
    async def test_constraint_violation(self, db_session):
        """Test handling of constraint violations"""
        # Create user
        user = User(
            email="unique@example.com",
            username="unique",
            password_hash="hash"
        )
        db_session.add(user)
        await db_session.commit()
        
        # Try to create duplicate
        duplicate = User(
            email="unique@example.com",  # Duplicate email
            username="different",
            password_hash="hash"
        )
        db_session.add(duplicate)
        
        with pytest.raises(IntegrityError) as exc_info:
            await db_session.commit()
        
        assert "unique" in str(exc_info.value).lower()
    
    @pytest.mark.asyncio
    async def test_deadlock_retry(self, db_manager):
        """Test deadlock detection and retry"""
        async def create_deadlock():
            async with db_manager.get_session() as session1:
                async with db_manager.get_session() as session2:
                    # Session 1 locks resource A
                    await session1.execute(
                        text("SELECT * FROM users WHERE id = 1 FOR UPDATE")
                    )
                    
                    # Session 2 locks resource B
                    await session2.execute(
                        text("SELECT * FROM apis WHERE id = 1 FOR UPDATE")
                    )
                    
                    # Session 1 tries to lock resource B (deadlock)
                    with pytest.raises(DeadlockError):
                        await session1.execute(
                            text("SELECT * FROM apis WHERE id = 1 FOR UPDATE")
                        )
        
        # Test retry mechanism
        retry_count = 0
        max_retries = 3
        
        while retry_count < max_retries:
            try:
                await create_deadlock()
                break
            except DeadlockError:
                retry_count += 1
                await asyncio.sleep(0.1 * retry_count)  # Exponential backoff
        
        assert retry_count > 0  # Should have experienced at least one deadlock


if __name__ == "__main__":
    pytest.main([__file__, "-v"])