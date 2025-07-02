"""
Test suite for Rate Limiting and Quota Management
"""

import pytest
import asyncio
from unittest.mock import AsyncMock, Mock, patch
import sys
from datetime import datetime, timedelta
from uuid import uuid4
import redis.asyncio as aioredis

# Mock docker module before imports
sys.modules['docker'] = Mock()

# Import after mocking
from backend.api.rate_limiter import (
    RateLimiter, QuotaManager, RateLimitWindow, 
    RateLimitTier, QuotaType, DEFAULT_RATE_LIMITS
)


class MockRedis:
    def __init__(self):
        self.data = {}
        self.sorted_sets = {}
        self.expires = {}
    
    def pipeline(self):
        return self
    
    async def zremrangebyscore(self, key, min_score, max_score):
        if key not in self.sorted_sets:
            return 0
        
        # Remove items within score range
        original_len = len(self.sorted_sets[key])
        self.sorted_sets[key] = {
            k: v for k, v in self.sorted_sets[key].items()
            if not (min_score <= v <= max_score)
        }
        return original_len - len(self.sorted_sets[key])
    
    async def zcard(self, key):
        return len(self.sorted_sets.get(key, {}))
    
    async def zadd(self, key, mapping):
        if key not in self.sorted_sets:
            self.sorted_sets[key] = {}
        self.sorted_sets[key].update(mapping)
        return len(mapping)
    
    async def zrem(self, key, member):
        if key in self.sorted_sets and member in self.sorted_sets[key]:
            del self.sorted_sets[key][member]
            return 1
        return 0
    
    async def zrange(self, key, start, stop, withscores=False):
        if key not in self.sorted_sets:
            return []
        
        items = sorted(self.sorted_sets[key].items(), key=lambda x: x[1])
        if withscores:
            return [(k, v) for k, v in items[start:stop+1]]
        return [k for k, v in items[start:stop+1]]
    
    async def expire(self, key, seconds):
        self.expires[key] = datetime.utcnow() + timedelta(seconds=seconds)
        return 1
    
    async def execute(self):
        # Return dummy results for pipeline
        return [0, 0, 1, 1]
    
    async def hincrby(self, key, field, increment):
        if key not in self.data:
            self.data[key] = {}
        if field not in self.data[key]:
            self.data[key][field] = 0
        self.data[key][field] += increment
        return self.data[key][field]


@pytest.fixture
def mock_redis():
    return MockRedis()


@pytest.fixture
def mock_db_pool():
    pool = AsyncMock()
    conn = AsyncMock()
    
    # Mock connection context manager
    pool.acquire.return_value.__aenter__.return_value = conn
    pool.acquire.return_value.__aexit__.return_value = None
    
    return pool, conn


@pytest.fixture
def rate_limiter(mock_redis, mock_db_pool):
    pool, _ = mock_db_pool
    return RateLimiter(mock_redis, pool)


@pytest.fixture
def quota_manager(mock_db_pool, mock_redis):
    pool, _ = mock_db_pool
    return QuotaManager(pool, mock_redis)


class TestRateLimiter:
    
    @pytest.mark.asyncio
    async def test_check_rate_limit_allowed(self, rate_limiter, mock_redis):
        """Test rate limit check when within limits"""
        key = "test_user"
        window = RateLimitWindow.MINUTE
        limit = 10
        
        # First request should be allowed
        allowed, metadata = await rate_limiter.check_rate_limit(key, window, limit)
        
        assert allowed is True
        assert metadata['limit'] == limit
        assert metadata['remaining'] == limit - 1
        assert metadata['window'] == 'minute'
    
    @pytest.mark.asyncio
    async def test_check_rate_limit_exceeded(self, rate_limiter, mock_redis):
        """Test rate limit check when limit exceeded"""
        key = "test_user"
        window = RateLimitWindow.MINUTE
        limit = 3
        
        # Make requests up to limit
        for i in range(limit):
            allowed, _ = await rate_limiter.check_rate_limit(key, window, limit)
            assert allowed is True
        
        # Next request should be denied
        allowed, metadata = await rate_limiter.check_rate_limit(key, window, limit)
        
        assert allowed is False
        assert metadata['limit'] == limit
        assert metadata['remaining'] == 0
        assert 'retry_after' in metadata
        assert metadata['retry_after'] > 0
    
    @pytest.mark.asyncio
    async def test_get_user_limits_default(self, rate_limiter, mock_db_pool):
        """Test getting default user limits"""
        pool, conn = mock_db_pool
        user_id = str(uuid4())
        
        # Mock no user data (defaults to free tier)
        conn.fetchrow.return_value = None
        
        limits = await rate_limiter.get_user_limits(user_id)
        
        # Should get free tier limits
        assert limits == DEFAULT_RATE_LIMITS[RateLimitTier.FREE]
    
    @pytest.mark.asyncio
    async def test_get_user_limits_with_tier(self, rate_limiter, mock_db_pool):
        """Test getting user limits based on tier"""
        pool, conn = mock_db_pool
        user_id = str(uuid4())
        
        # Mock pro tier user
        conn.fetchrow.return_value = {
            'tier': 'pro',
            'custom_rate_limits': None,
            'rate_limit_multiplier': 1.0
        }
        
        limits = await rate_limiter.get_user_limits(user_id)
        
        # Should get pro tier limits
        assert limits == DEFAULT_RATE_LIMITS[RateLimitTier.PRO]
    
    @pytest.mark.asyncio
    async def test_get_user_limits_with_multiplier(self, rate_limiter, mock_db_pool):
        """Test user limits with multiplier"""
        pool, conn = mock_db_pool
        user_id = str(uuid4())
        
        # Mock user with 2x multiplier
        conn.fetchrow.return_value = {
            'tier': 'basic',
            'custom_rate_limits': None,
            'rate_limit_multiplier': 2.0
        }
        
        limits = await rate_limiter.get_user_limits(user_id)
        
        # Limits should be doubled
        for window, limit in limits.items():
            expected = DEFAULT_RATE_LIMITS[RateLimitTier.BASIC][window] * 2
            assert limit == expected
    
    @pytest.mark.asyncio
    async def test_get_user_limits_with_custom(self, rate_limiter, mock_db_pool):
        """Test user limits with custom overrides"""
        pool, conn = mock_db_pool
        user_id = str(uuid4())
        
        # Mock user with custom limits
        conn.fetchrow.return_value = {
            'tier': 'basic',
            'custom_rate_limits': {
                'minute': 100,
                'hour': 5000
            },
            'rate_limit_multiplier': 1.0
        }
        
        limits = await rate_limiter.get_user_limits(user_id)
        
        # Custom limits should override defaults
        assert limits[RateLimitWindow.MINUTE] == 100
        assert limits[RateLimitWindow.HOUR] == 5000
        # Other windows should use defaults
        assert limits[RateLimitWindow.DAY] == DEFAULT_RATE_LIMITS[RateLimitTier.BASIC][RateLimitWindow.DAY]
    
    @pytest.mark.asyncio
    async def test_check_api_rate_limit(self, rate_limiter, mock_db_pool):
        """Test API-specific rate limiting"""
        pool, conn = mock_db_pool
        api_id = str(uuid4())
        user_id = str(uuid4())
        
        # Mock user limits
        conn.fetchrow.side_effect = [
            None,  # User data (defaults to free)
            {  # API data
                'rate_limits': {'minute': 20},
                'tier': None
            }
        ]
        
        # Check should pass for API limits
        allowed, metadata = await rate_limiter.check_api_rate_limit(
            api_id, user_id, cost=1
        )
        
        assert allowed is True
    
    @pytest.mark.asyncio
    async def test_global_api_limit(self, rate_limiter, mock_db_pool):
        """Test global API rate limiting"""
        pool, conn = mock_db_pool
        api_id = str(uuid4())
        
        # Mock API with limits
        conn.fetchrow.return_value = {
            'rate_limits': {'minute': 10},
            'tier': None
        }
        
        # Global limit should be 10x individual
        allowed, metadata = await rate_limiter.check_global_api_limit(api_id)
        
        assert allowed is True
        
        # Simulate hitting global limit
        # Would need to make 100+ requests to test denial


class TestQuotaManager:
    
    @pytest.mark.asyncio
    async def test_get_user_quota_free_tier(self, quota_manager, mock_db_pool):
        """Test getting quota for free tier user"""
        pool, conn = mock_db_pool
        user_id = str(uuid4())
        
        # Mock no subscription data (free tier)
        conn.fetchrow.return_value = None
        
        quota_info = await quota_manager.get_user_quota(user_id)
        
        assert quota_info['tier'] == 'free'
        assert quota_info['quotas'][QuotaType.REQUESTS.value]['limit'] == 10000
        assert quota_info['quotas'][QuotaType.BANDWIDTH.value]['limit'] == 1073741824  # 1GB
        assert quota_info['period'] == 'month'
    
    @pytest.mark.asyncio
    async def test_get_user_quota_with_usage(self, quota_manager, mock_db_pool):
        """Test getting quota with current usage"""
        pool, conn = mock_db_pool
        user_id = str(uuid4())
        
        # Mock user with some usage
        conn.fetchrow.return_value = {
            'tier': 'basic',
            'quota_limits': {
                'requests': 100000,
                'bandwidth': 10737418240,  # 10GB
                'compute_time': 36000000   # 10 hours
            },
            'quota_period': 'month',
            'period_start': datetime.utcnow() - timedelta(days=15),
            'period_end': datetime.utcnow() + timedelta(days=15),
            'used_requests': 25000,
            'used_bandwidth': 2147483648,  # 2GB
            'used_compute': 7200000  # 2 hours
        }
        
        quota_info = await quota_manager.get_user_quota(user_id)
        
        assert quota_info['tier'] == 'basic'
        assert quota_info['quotas'][QuotaType.REQUESTS.value]['used'] == 25000
        assert quota_info['quotas'][QuotaType.REQUESTS.value]['limit'] == 100000
        assert quota_info['quotas'][QuotaType.BANDWIDTH.value]['used'] == 2147483648
    
    @pytest.mark.asyncio
    async def test_check_quota_allowed(self, quota_manager, mock_db_pool):
        """Test quota check when within limits"""
        pool, conn = mock_db_pool
        user_id = str(uuid4())
        
        # Mock user with plenty of quota
        conn.fetchrow.return_value = {
            'tier': 'pro',
            'quota_limits': {'requests': 1000000},
            'quota_period': 'month',
            'period_start': datetime.utcnow() - timedelta(days=1),
            'period_end': datetime.utcnow() + timedelta(days=29),
            'used_requests': 100000,
            'used_bandwidth': 0,
            'used_compute': 0
        }
        
        allowed, metadata = await quota_manager.check_quota(
            user_id, QuotaType.REQUESTS, cost=1
        )
        
        assert allowed is True
        assert metadata['remaining'] == 899999
    
    @pytest.mark.asyncio
    async def test_check_quota_exceeded(self, quota_manager, mock_db_pool):
        """Test quota check when limit exceeded"""
        pool, conn = mock_db_pool
        user_id = str(uuid4())
        
        # Mock user at quota limit
        conn.fetchrow.return_value = {
            'tier': 'free',
            'quota_limits': {'requests': 10000},
            'quota_period': 'month',
            'period_start': datetime.utcnow() - timedelta(days=15),
            'period_end': datetime.utcnow() + timedelta(days=15),
            'used_requests': 9999,
            'used_bandwidth': 0,
            'used_compute': 0
        }
        
        # Should allow exactly 1 more request
        allowed, metadata = await quota_manager.check_quota(
            user_id, QuotaType.REQUESTS, cost=1
        )
        assert allowed is True
        
        # But not 2
        allowed, metadata = await quota_manager.check_quota(
            user_id, QuotaType.REQUESTS, cost=2
        )
        assert allowed is False
        assert metadata['remaining'] == 1
    
    @pytest.mark.asyncio
    async def test_consume_quota(self, quota_manager, mock_redis):
        """Test quota consumption tracking"""
        user_id = str(uuid4())
        api_id = str(uuid4())
        
        # Consume some quota
        success = await quota_manager.consume_quota(
            user_id, api_id,
            request_count=10,
            bandwidth_bytes=1024000,
            compute_ms=5000
        )
        
        assert success is True
        
        # Check Redis was updated
        current_month = datetime.utcnow().strftime("%Y-%m")
        quota_key = f"quota:{user_id}:{current_month}"
        
        # Verify increments were made
        assert quota_key in mock_redis.data
        assert mock_redis.data[quota_key]['requests'] == 10
        assert mock_redis.data[quota_key]['bandwidth'] == 1024000
        assert mock_redis.data[quota_key]['compute'] == 5000


if __name__ == "__main__":
    pytest.main([__file__, "-v"])