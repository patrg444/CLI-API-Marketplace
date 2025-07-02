"""
Rate Limiting and Quota Management for API-Direct

Provides flexible rate limiting with:
- Per-user, per-API, and global limits
- Multiple time windows (minute, hour, day)
- Quota management with billing integration
- Distributed rate limiting using Redis
"""

import asyncio
import json
from datetime import datetime, timedelta
from typing import Dict, Optional, Tuple, List, Any
from enum import Enum
import logging
from fastapi import HTTPException, Request
from fastapi.responses import JSONResponse
import redis.asyncio as aioredis
import asyncpg

logger = logging.getLogger(__name__)


class RateLimitWindow(Enum):
    SECOND = ("second", 1)
    MINUTE = ("minute", 60)
    HOUR = ("hour", 3600)
    DAY = ("day", 86400)
    MONTH = ("month", 2592000)  # 30 days
    
    def __init__(self, name: str, seconds: int):
        self.window_name = name
        self.seconds = seconds


class RateLimitTier(Enum):
    FREE = "free"
    BASIC = "basic"
    PRO = "pro"
    ENTERPRISE = "enterprise"
    CUSTOM = "custom"


class QuotaType(Enum):
    REQUESTS = "requests"
    BANDWIDTH = "bandwidth"
    COMPUTE_TIME = "compute_time"
    STORAGE = "storage"


# Default rate limits by tier
DEFAULT_RATE_LIMITS = {
    RateLimitTier.FREE: {
        RateLimitWindow.MINUTE: 10,
        RateLimitWindow.HOUR: 100,
        RateLimitWindow.DAY: 1000,
        RateLimitWindow.MONTH: 10000
    },
    RateLimitTier.BASIC: {
        RateLimitWindow.MINUTE: 60,
        RateLimitWindow.HOUR: 1000,
        RateLimitWindow.DAY: 10000,
        RateLimitWindow.MONTH: 100000
    },
    RateLimitTier.PRO: {
        RateLimitWindow.MINUTE: 300,
        RateLimitWindow.HOUR: 5000,
        RateLimitWindow.DAY: 50000,
        RateLimitWindow.MONTH: 1000000
    },
    RateLimitTier.ENTERPRISE: {
        RateLimitWindow.MINUTE: 1000,
        RateLimitWindow.HOUR: 20000,
        RateLimitWindow.DAY: 200000,
        RateLimitWindow.MONTH: 5000000
    }
}


class RateLimiter:
    """Redis-based distributed rate limiter using sliding window algorithm"""
    
    def __init__(self, redis_client: aioredis.Redis, db_pool: asyncpg.Pool):
        self.redis = redis_client
        self.db_pool = db_pool
        self._limits_cache = {}
        self._cache_ttl = 300  # 5 minutes
    
    async def check_rate_limit(
        self,
        key: str,
        window: RateLimitWindow,
        limit: int,
        cost: int = 1
    ) -> Tuple[bool, Dict[str, Any]]:
        """
        Check if request is within rate limit using sliding window
        
        Returns: (allowed, metadata)
        """
        now = datetime.utcnow()
        window_start = now - timedelta(seconds=window.seconds)
        window_key = f"rate_limit:{key}:{window.window_name}"
        
        # Use Redis sorted set for sliding window
        pipe = self.redis.pipeline()
        
        # Remove old entries
        pipe.zremrangebyscore(window_key, 0, window_start.timestamp())
        
        # Count current window
        pipe.zcard(window_key)
        
        # Add current request if allowed
        pipe.zadd(window_key, {str(now.timestamp()): now.timestamp()})
        
        # Set expiry
        pipe.expire(window_key, window.seconds + 60)
        
        results = await pipe.execute()
        current_count = results[1]
        
        # Check if limit exceeded
        if current_count + cost > limit:
            # Remove the added entry
            await self.redis.zrem(window_key, str(now.timestamp()))
            
            # Calculate reset time
            oldest_entry = await self.redis.zrange(window_key, 0, 0, withscores=True)
            if oldest_entry:
                reset_timestamp = oldest_entry[0][1] + window.seconds
                reset_time = datetime.fromtimestamp(reset_timestamp)
            else:
                reset_time = now + timedelta(seconds=window.seconds)
            
            return False, {
                "limit": limit,
                "remaining": max(0, limit - current_count),
                "reset": reset_time.isoformat(),
                "retry_after": int((reset_time - now).total_seconds()),
                "window": window.window_name
            }
        
        return True, {
            "limit": limit,
            "remaining": limit - current_count - cost,
            "reset": (now + timedelta(seconds=window.seconds)).isoformat(),
            "window": window.window_name
        }
    
    async def get_user_limits(self, user_id: str) -> Dict[RateLimitWindow, int]:
        """Get rate limits for a user based on their tier"""
        
        # Check cache
        cache_key = f"user_limits:{user_id}"
        if cache_key in self._limits_cache:
            cached, timestamp = self._limits_cache[cache_key]
            if (datetime.utcnow() - timestamp).seconds < self._cache_ttl:
                return cached
        
        async with self.db_pool.acquire() as conn:
            # Get user tier and custom limits
            user_data = await conn.fetchrow("""
                SELECT 
                    u.tier,
                    u.custom_rate_limits,
                    s.rate_limit_multiplier
                FROM users u
                LEFT JOIN subscriptions s ON u.id = s.user_id 
                    AND s.status = 'active'
                WHERE u.id = $1
            """, user_id)
            
            if not user_data:
                tier = RateLimitTier.FREE
                custom_limits = None
                multiplier = 1.0
            else:
                tier = RateLimitTier(user_data['tier'] or 'free')
                custom_limits = user_data['custom_rate_limits']
                multiplier = user_data['rate_limit_multiplier'] or 1.0
            
            # Start with default limits for tier
            limits = DEFAULT_RATE_LIMITS.get(tier, DEFAULT_RATE_LIMITS[RateLimitTier.FREE]).copy()
            
            # Apply multiplier
            for window in limits:
                limits[window] = int(limits[window] * multiplier)
            
            # Apply custom limits if any
            if custom_limits:
                for window_name, limit in custom_limits.items():
                    try:
                        window = RateLimitWindow[window_name.upper()]
                        limits[window] = limit
                    except KeyError:
                        pass
            
            # Cache the result
            self._limits_cache[cache_key] = (limits, datetime.utcnow())
            
            return limits
    
    async def get_api_limits(self, api_id: str) -> Dict[RateLimitWindow, int]:
        """Get rate limits configured for a specific API"""
        
        cache_key = f"api_limits:{api_id}"
        if cache_key in self._limits_cache:
            cached, timestamp = self._limits_cache[cache_key]
            if (datetime.utcnow() - timestamp).seconds < self._cache_ttl:
                return cached
        
        async with self.db_pool.acquire() as conn:
            api_data = await conn.fetchrow("""
                SELECT rate_limits, tier
                FROM apis
                WHERE id = $1 AND status = 'active'
            """, api_id)
            
            if not api_data:
                return {}
            
            if api_data['rate_limits']:
                limits = {}
                for window_name, limit in api_data['rate_limits'].items():
                    try:
                        window = RateLimitWindow[window_name.upper()]
                        limits[window] = limit
                    except KeyError:
                        pass
                
                self._limits_cache[cache_key] = (limits, datetime.utcnow())
                return limits
            
            # Use tier-based defaults
            tier = RateLimitTier(api_data['tier'] or 'basic')
            limits = DEFAULT_RATE_LIMITS.get(tier, DEFAULT_RATE_LIMITS[RateLimitTier.BASIC])
            
            self._limits_cache[cache_key] = (limits, datetime.utcnow())
            return limits
    
    async def check_user_rate_limit(
        self,
        user_id: str,
        cost: int = 1
    ) -> Tuple[bool, Dict[str, Any]]:
        """Check rate limits for a user across all windows"""
        
        limits = await self.get_user_limits(user_id)
        
        for window, limit in limits.items():
            allowed, metadata = await self.check_rate_limit(
                f"user:{user_id}",
                window,
                limit,
                cost
            )
            
            if not allowed:
                return False, metadata
        
        # All limits passed
        smallest_window = min(limits.keys(), key=lambda w: w.seconds)
        return True, {
            "limit": limits[smallest_window],
            "window": smallest_window.window_name
        }
    
    async def check_api_rate_limit(
        self,
        api_id: str,
        user_id: str,
        cost: int = 1
    ) -> Tuple[bool, Dict[str, Any]]:
        """Check rate limits for API usage by a user"""
        
        # Check user's global limits
        user_allowed, user_metadata = await self.check_user_rate_limit(user_id, cost)
        if not user_allowed:
            return False, user_metadata
        
        # Check API-specific limits
        api_limits = await self.get_api_limits(api_id)
        
        for window, limit in api_limits.items():
            allowed, metadata = await self.check_rate_limit(
                f"api:{api_id}:user:{user_id}",
                window,
                limit,
                cost
            )
            
            if not allowed:
                return False, metadata
        
        return True, {"status": "allowed"}
    
    async def check_global_api_limit(
        self,
        api_id: str,
        cost: int = 1
    ) -> Tuple[bool, Dict[str, Any]]:
        """Check global rate limits for an API (across all users)"""
        
        api_limits = await self.get_api_limits(api_id)
        
        # Check global API limits (10x individual limits)
        for window, limit in api_limits.items():
            global_limit = limit * 10
            allowed, metadata = await self.check_rate_limit(
                f"api:{api_id}:global",
                window,
                global_limit,
                cost
            )
            
            if not allowed:
                metadata["scope"] = "global"
                return False, metadata
        
        return True, {"status": "allowed"}
    
    def clear_cache(self, pattern: Optional[str] = None):
        """Clear rate limit cache"""
        if pattern:
            self._limits_cache = {
                k: v for k, v in self._limits_cache.items()
                if pattern not in k
            }
        else:
            self._limits_cache.clear()


class QuotaManager:
    """Manages API usage quotas and billing integration"""
    
    def __init__(self, db_pool: asyncpg.Pool, redis_client: aioredis.Redis):
        self.db_pool = db_pool
        self.redis = redis_client
    
    async def get_user_quota(self, user_id: str) -> Dict[str, Any]:
        """Get current quota usage and limits for a user"""
        
        async with self.db_pool.acquire() as conn:
            # Get user's subscription and quota
            quota_data = await conn.fetchrow("""
                SELECT 
                    u.tier,
                    s.quota_limits,
                    s.quota_period,
                    s.period_start,
                    s.period_end,
                    COALESCE(SUM(ac.request_count), 0) as used_requests,
                    COALESCE(SUM(ac.bandwidth_bytes), 0) as used_bandwidth,
                    COALESCE(SUM(ac.compute_ms), 0) as used_compute
                FROM users u
                LEFT JOIN subscriptions s ON u.id = s.user_id AND s.status = 'active'
                LEFT JOIN api_calls ac ON u.id = ac.user_id 
                    AND ac.created_at >= COALESCE(s.period_start, NOW() - INTERVAL '30 days')
                WHERE u.id = $1
                GROUP BY u.tier, s.quota_limits, s.quota_period, s.period_start, s.period_end
            """, user_id)
            
            if not quota_data:
                # Default free tier quotas
                return {
                    "tier": "free",
                    "quotas": {
                        QuotaType.REQUESTS.value: {"limit": 10000, "used": 0},
                        QuotaType.BANDWIDTH.value: {"limit": 1073741824, "used": 0},  # 1GB
                        QuotaType.COMPUTE_TIME.value: {"limit": 3600000, "used": 0}  # 1 hour
                    },
                    "period": "month",
                    "reset_date": (datetime.utcnow() + timedelta(days=30)).isoformat()
                }
            
            # Parse quota limits
            quota_limits = quota_data['quota_limits'] or self._get_default_quotas(quota_data['tier'])
            
            return {
                "tier": quota_data['tier'],
                "quotas": {
                    QuotaType.REQUESTS.value: {
                        "limit": quota_limits.get('requests', 10000),
                        "used": int(quota_data['used_requests'])
                    },
                    QuotaType.BANDWIDTH.value: {
                        "limit": quota_limits.get('bandwidth', 1073741824),
                        "used": int(quota_data['used_bandwidth'])
                    },
                    QuotaType.COMPUTE_TIME.value: {
                        "limit": quota_limits.get('compute_time', 3600000),
                        "used": int(quota_data['used_compute'])
                    }
                },
                "period": quota_data['quota_period'] or "month",
                "period_start": quota_data['period_start'].isoformat() if quota_data['period_start'] else None,
                "reset_date": quota_data['period_end'].isoformat() if quota_data['period_end'] else None
            }
    
    async def check_quota(
        self,
        user_id: str,
        quota_type: QuotaType,
        cost: int = 1
    ) -> Tuple[bool, Dict[str, Any]]:
        """Check if user has quota available"""
        
        quota_info = await self.get_user_quota(user_id)
        quota = quota_info['quotas'].get(quota_type.value, {})
        
        limit = quota.get('limit', 0)
        used = quota.get('used', 0)
        
        if used + cost > limit:
            return False, {
                "quota_type": quota_type.value,
                "limit": limit,
                "used": used,
                "remaining": max(0, limit - used),
                "reset_date": quota_info['reset_date']
            }
        
        return True, {
            "quota_type": quota_type.value,
            "remaining": limit - used - cost
        }
    
    async def consume_quota(
        self,
        user_id: str,
        api_id: str,
        request_count: int = 1,
        bandwidth_bytes: int = 0,
        compute_ms: int = 0
    ) -> bool:
        """Consume quota for API usage"""
        
        # This is typically done when recording API calls
        # The actual consumption is tracked in the api_calls table
        # This method can be used for real-time quota updates if needed
        
        # Update cached quota counters in Redis for fast checking
        pipe = self.redis.pipeline()
        
        current_month = datetime.utcnow().strftime("%Y-%m")
        
        if request_count > 0:
            pipe.hincrby(f"quota:{user_id}:{current_month}", "requests", request_count)
        if bandwidth_bytes > 0:
            pipe.hincrby(f"quota:{user_id}:{current_month}", "bandwidth", bandwidth_bytes)
        if compute_ms > 0:
            pipe.hincrby(f"quota:{user_id}:{current_month}", "compute", compute_ms)
        
        # Expire at end of month
        days_in_month = 31
        pipe.expire(f"quota:{user_id}:{current_month}", days_in_month * 86400)
        
        await pipe.execute()
        return True
    
    def _get_default_quotas(self, tier: str) -> Dict[str, int]:
        """Get default quotas by tier"""
        
        defaults = {
            "free": {
                "requests": 10000,
                "bandwidth": 1073741824,  # 1GB
                "compute_time": 3600000   # 1 hour
            },
            "basic": {
                "requests": 100000,
                "bandwidth": 10737418240,  # 10GB
                "compute_time": 36000000   # 10 hours
            },
            "pro": {
                "requests": 1000000,
                "bandwidth": 107374182400,  # 100GB
                "compute_time": 360000000   # 100 hours
            },
            "enterprise": {
                "requests": 10000000,
                "bandwidth": 1073741824000,  # 1TB
                "compute_time": 3600000000   # 1000 hours
            }
        }
        
        return defaults.get(tier, defaults["free"])


# Rate limiting middleware
async def rate_limit_middleware(request: Request, call_next):
    """FastAPI middleware for rate limiting"""
    
    # Skip rate limiting for certain paths
    skip_paths = ["/health", "/metrics", "/docs", "/openapi.json"]
    if request.url.path in skip_paths:
        return await call_next(request)
    
    # Get rate limiter from app state
    rate_limiter = request.app.state.rate_limiter
    if not rate_limiter:
        return await call_next(request)
    
    # Get user ID from request (from auth)
    user_id = getattr(request.state, "user_id", None)
    if not user_id:
        # Use IP-based rate limiting for unauthenticated requests
        user_id = f"ip:{request.client.host}"
    
    # Check rate limit
    allowed, metadata = await rate_limiter.check_user_rate_limit(user_id)
    
    if not allowed:
        return JSONResponse(
            status_code=429,
            content={
                "error": "Rate limit exceeded",
                "message": f"Too many requests. Please retry after {metadata['retry_after']} seconds.",
                "retry_after": metadata['retry_after'],
                "limit": metadata['limit'],
                "window": metadata['window'],
                "reset": metadata['reset']
            },
            headers={
                "X-RateLimit-Limit": str(metadata['limit']),
                "X-RateLimit-Remaining": str(metadata['remaining']),
                "X-RateLimit-Reset": metadata['reset'],
                "Retry-After": str(metadata['retry_after'])
            }
        )
    
    # Process request
    response = await call_next(request)
    
    # Add rate limit headers
    response.headers["X-RateLimit-Limit"] = str(metadata.get('limit', 0))
    response.headers["X-RateLimit-Remaining"] = str(metadata.get('remaining', 0))
    
    return response