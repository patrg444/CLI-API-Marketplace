"""
Basic database test to verify setup
"""

import pytest
import asyncio
import sys
import os

# Add parent directory to path
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from database import DatabaseManager, User


@pytest.mark.asyncio
async def test_basic_database_setup():
    """Test basic database setup works"""
    # Create manager
    manager = DatabaseManager(
        connection_string="sqlite+aiosqlite:///:memory:",
        pool_size=5,
        max_overflow=10
    )
    
    # Initialize
    await manager.initialize()
    # Create tables
    await manager.create_tables()
    
    try:
        # Test we can get a session
        async with manager.get_session() as session:
            # Create a user
            user = User(
                email="test@example.com",
                username="testuser",
                password_hash="hash123"
            )
            session.add(user)
            await session.commit()
            
            # Verify it worked
            assert user.id is not None
            print(f"Created user with ID: {user.id}")
            
    finally:
        await manager.close()


@pytest.mark.asyncio
async def test_pool_metrics():
    """Test we can get pool metrics"""
    manager = DatabaseManager(
        connection_string="sqlite+aiosqlite:///:memory:"
    )
    await manager.initialize()
    
    try:
        metrics = manager.get_pool_metrics()
        assert 'size' in metrics
        assert 'total' in metrics
        print(f"Pool metrics: {metrics}")
    finally:
        await manager.close()


if __name__ == "__main__":
    # Run tests directly
    asyncio.run(test_basic_database_setup())
    asyncio.run(test_pool_metrics())
    print("All basic tests passed!")