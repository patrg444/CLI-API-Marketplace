"""
Test suite for API Version Management System
"""

import pytest
import asyncio
import asyncpg
from unittest.mock import AsyncMock, Mock, patch
import sys
from datetime import datetime, timedelta
from uuid import UUID, uuid4

# Mock docker module before imports
sys.modules['docker'] = Mock()

# Import after mocking
from backend.api.version_manager import VersionManager, VersionType, VersionStatus


class AsyncContextManager:
    def __init__(self, conn):
        self.conn = conn
    
    async def __aenter__(self):
        return self.conn
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        return None


@pytest.fixture
def mock_db_pool():
    pool = AsyncMock()
    conn = AsyncMock()
    
    # Create a proper async context manager mock
    async def acquire():
        return AsyncContextManager(conn)
    
    pool.acquire = acquire
    
    return pool, conn


@pytest.fixture
def version_manager(mock_db_pool):
    pool, _ = mock_db_pool
    return VersionManager(pool)


class TestVersionManager:
    
    @pytest.mark.asyncio
    async def test_create_version_table(self, version_manager, mock_db_pool):
        """Test creating version table"""
        pool, conn = mock_db_pool
        
        await version_manager.create_version_table()
        
        # Verify table creation SQL was executed
        conn.execute.assert_called_once()
        call_args = conn.execute.call_args[0][0]
        assert "CREATE TABLE IF NOT EXISTS api_versions" in call_args
        assert "version VARCHAR(20) NOT NULL" in call_args
        assert "major INTEGER NOT NULL" in call_args
    
    def test_parse_version_valid(self, version_manager):
        """Test parsing valid semantic versions"""
        # Test standard version
        major, minor, patch, prerelease = version_manager.parse_version("1.2.3")
        assert major == 1
        assert minor == 2
        assert patch == 3
        assert prerelease is None
        
        # Test with prerelease
        major, minor, patch, prerelease = version_manager.parse_version("2.0.0-beta.1")
        assert major == 2
        assert minor == 0
        assert patch == 0
        assert prerelease == "beta.1"
    
    def test_parse_version_invalid(self, version_manager):
        """Test parsing invalid versions"""
        with pytest.raises(ValueError):
            version_manager.parse_version("1.2")
        
        with pytest.raises(ValueError):
            version_manager.parse_version("invalid")
        
        with pytest.raises(ValueError):
            version_manager.parse_version("1.2.3.4")
    
    def test_compare_versions(self, version_manager):
        """Test version comparison"""
        # Test major version difference
        assert version_manager.compare_versions("2.0.0", "1.0.0") == 1
        assert version_manager.compare_versions("1.0.0", "2.0.0") == -1
        
        # Test minor version difference
        assert version_manager.compare_versions("1.2.0", "1.1.0") == 1
        assert version_manager.compare_versions("1.1.0", "1.2.0") == -1
        
        # Test patch version difference
        assert version_manager.compare_versions("1.0.2", "1.0.1") == 1
        assert version_manager.compare_versions("1.0.1", "1.0.2") == -1
        
        # Test equal versions
        assert version_manager.compare_versions("1.0.0", "1.0.0") == 0
        
        # Test prerelease versions
        assert version_manager.compare_versions("1.0.0", "1.0.0-beta") == 1
        assert version_manager.compare_versions("1.0.0-beta", "1.0.0") == -1
        assert version_manager.compare_versions("1.0.0-beta.2", "1.0.0-beta.1") == 1
    
    @pytest.mark.asyncio
    async def test_create_version_first_version(self, version_manager, mock_db_pool):
        """Test creating first version of an API"""
        pool, conn = mock_db_pool
        api_id = str(uuid4())
        user_id = str(uuid4())
        
        # Mock no existing version
        conn.fetchrow.side_effect = [None, {"id": uuid4(), "version": "1.0.0"}]
        
        result = await version_manager.create_version(
            api_id=api_id,
            version_type=VersionType.MAJOR,
            user_id=user_id,
            release_notes="Initial release"
        )
        
        # Verify version was created
        assert conn.fetchrow.call_count == 2
        insert_call = conn.fetchrow.call_args_list[1][0][0]
        assert "INSERT INTO api_versions" in insert_call
        assert "1.0.0" in str(conn.fetchrow.call_args_list[1][0])
    
    @pytest.mark.asyncio
    async def test_create_version_increment(self, version_manager, mock_db_pool):
        """Test incrementing versions correctly"""
        pool, conn = mock_db_pool
        api_id = str(uuid4())
        user_id = str(uuid4())
        
        # Test major version increment
        conn.fetchrow.side_effect = [
            {"version": "1.2.3", "major": 1, "minor": 2, "patch": 3},
            {"id": uuid4(), "version": "2.0.0"}
        ]
        
        result = await version_manager.create_version(
            api_id=api_id,
            version_type=VersionType.MAJOR,
            user_id=user_id
        )
        
        insert_args = conn.fetchrow.call_args_list[1][0]
        assert "2.0.0" in str(insert_args)
        
        # Test minor version increment
        conn.fetchrow.side_effect = [
            {"version": "1.2.3", "major": 1, "minor": 2, "patch": 3},
            {"id": uuid4(), "version": "1.3.0"}
        ]
        
        result = await version_manager.create_version(
            api_id=api_id,
            version_type=VersionType.MINOR,
            user_id=user_id
        )
        
        insert_args = conn.fetchrow.call_args_list[3][0]
        assert "1.3.0" in str(insert_args)
        
        # Test patch version increment
        conn.fetchrow.side_effect = [
            {"version": "1.2.3", "major": 1, "minor": 2, "patch": 3},
            {"id": uuid4(), "version": "1.2.4"}
        ]
        
        result = await version_manager.create_version(
            api_id=api_id,
            version_type=VersionType.PATCH,
            user_id=user_id
        )
        
        insert_args = conn.fetchrow.call_args_list[5][0]
        assert "1.2.4" in str(insert_args)
    
    @pytest.mark.asyncio
    async def test_publish_version(self, version_manager, mock_db_pool):
        """Test publishing a draft version"""
        pool, conn = mock_db_pool
        api_id = str(uuid4())
        
        # Set up transaction mock
        conn.transaction = AsyncMock()
        conn.transaction.return_value.__aenter__ = AsyncMock()
        conn.transaction.return_value.__aexit__ = AsyncMock()
        
        conn.fetchrow.return_value = {"id": uuid4(), "version": "2.0.0", "status": "active"}
        
        result = await version_manager.publish_version(
            api_id=api_id,
            version="2.0.0",
            deprecation_period_days=90
        )
        
        # Verify deprecation of old versions
        assert conn.execute.call_count >= 2
        deprecate_call = conn.execute.call_args_list[0][0][0]
        assert "UPDATE api_versions" in deprecate_call
        assert "SET status" in deprecate_call
        
        # Verify new version published
        publish_call = conn.fetchrow.call_args[0][0]
        assert "UPDATE api_versions" in publish_call
        assert "SET status" in publish_call
        assert "published_at = CURRENT_TIMESTAMP" in publish_call
    
    @pytest.mark.asyncio
    async def test_get_versions(self, version_manager, mock_db_pool):
        """Test retrieving API versions"""
        pool, conn = mock_db_pool
        api_id = str(uuid4())
        
        conn.fetch.return_value = [
            {
                "id": uuid4(),
                "version": "2.0.0",
                "status": "active",
                "created_by_username": "testuser"
            },
            {
                "id": uuid4(),
                "version": "1.0.0",
                "status": "deprecated",
                "created_by_username": "testuser"
            }
        ]
        
        # Test without drafts
        versions = await version_manager.get_versions(api_id)
        
        assert len(versions) == 2
        assert versions[0]["version"] == "2.0.0"
        assert versions[1]["version"] == "1.0.0"
        
        # Verify query
        query = conn.fetch.call_args[0][0]
        assert "FROM api_versions v" in query
        assert "ORDER BY v.major DESC" in query
    
    @pytest.mark.asyncio
    async def test_check_version_compatibility(self, version_manager, mock_db_pool):
        """Test version compatibility checking"""
        pool, conn = mock_db_pool
        api_id = str(uuid4())
        
        # Test compatible version
        conn.fetchrow.side_effect = [
            {"version": "2.0.0", "minimum_supported_version": "1.0.0"},
            None
        ]
        
        result = await version_manager.check_version_compatibility(
            api_id=api_id,
            client_version="1.5.0"
        )
        
        assert result["compatible"] is True
        assert result["current_version"] == "2.0.0"
        
        # Test version too old
        conn.fetchrow.side_effect = [
            {"version": "3.0.0", "minimum_supported_version": "2.0.0"},
            None
        ]
        
        result = await version_manager.check_version_compatibility(
            api_id=api_id,
            client_version="1.0.0"
        )
        
        assert result["compatible"] is False
        assert "below minimum supported version" in result["reason"]
        assert result["upgrade_required"] is True
        
        # Test deprecated version
        conn.fetchrow.side_effect = [
            {"version": "2.0.0", "minimum_supported_version": "1.0.0"},
            {
                "status": "deprecated",
                "deprecation_date": datetime.now(),
                "retirement_date": datetime.now() + timedelta(days=90)
            }
        ]
        
        result = await version_manager.check_version_compatibility(
            api_id=api_id,
            client_version="1.5.0"
        )
        
        assert result["compatible"] is True
        assert "warning" in result
        assert result["warning"] == "This version is deprecated"
    
    @pytest.mark.asyncio
    async def test_generate_version_diff(self, version_manager, mock_db_pool):
        """Test generating diff between versions"""
        pool, conn = mock_db_pool
        api_id = str(uuid4())
        
        # Mock version data
        v1_data = {
            "version": "1.0.0",
            "major": 1,
            "minor": 0,
            "patch": 0,
            "breaking_changes": []
        }
        
        v2_data = {
            "version": "2.0.0",
            "major": 2,
            "minor": 0,
            "patch": 0,
            "breaking_changes": ["API endpoints restructured"]
        }
        
        conn.fetchrow.side_effect = [v1_data, v2_data]
        conn.fetch.return_value = []  # No versions between
        
        # Mock get_version_changelog
        with patch.object(version_manager, 'get_version_changelog') as mock_changelog:
            mock_changelog.return_value = {
                "versions": [],
                "all_breaking_changes": ["API endpoints restructured"],
                "all_deprecated_features": []
            }
            
            diff = await version_manager.generate_version_diff(
                api_id=api_id,
                version1="1.0.0",
                version2="2.0.0"
            )
            
            assert diff["old_version"] == "1.0.0"
            assert diff["new_version"] == "2.0.0"
            assert diff["version_jump"]["major"] == 1
            assert diff["is_breaking_change"] is True


if __name__ == "__main__":
    pytest.main([__file__, "-v"])