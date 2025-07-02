"""
Caching layer for API documentation and frequently accessed data
Uses Redis for distributed caching with TTL support
"""

import json
import hashlib
from typing import Optional, Any, Callable, Union
from datetime import timedelta
import redis.asyncio as redis
from functools import wraps
import pickle
import logging

logger = logging.getLogger(__name__)


class CacheManager:
    """Manages caching operations with Redis"""
    
    def __init__(self, redis_client: redis.Redis, default_ttl: int = 300):
        """
        Initialize cache manager
        
        Args:
            redis_client: Redis async client
            default_ttl: Default TTL in seconds (5 minutes)
        """
        self.redis = redis_client
        self.default_ttl = default_ttl
        self.prefix = "apidirect:cache:"
    
    def _make_key(self, namespace: str, key: str) -> str:
        """Create a namespaced cache key"""
        return f"{self.prefix}{namespace}:{key}"
    
    async def get(self, namespace: str, key: str) -> Optional[Any]:
        """
        Get value from cache
        
        Args:
            namespace: Cache namespace (e.g., 'api_docs', 'user_data')
            key: Cache key
            
        Returns:
            Cached value or None if not found/expired
        """
        full_key = self._make_key(namespace, key)
        try:
            value = await self.redis.get(full_key)
            if value:
                return pickle.loads(value)
        except Exception as e:
            logger.error(f"Cache get error for {full_key}: {e}")
        return None
    
    async def set(
        self, 
        namespace: str, 
        key: str, 
        value: Any, 
        ttl: Optional[int] = None
    ) -> bool:
        """
        Set value in cache
        
        Args:
            namespace: Cache namespace
            key: Cache key
            value: Value to cache
            ttl: Time to live in seconds (uses default if None)
            
        Returns:
            True if successful
        """
        full_key = self._make_key(namespace, key)
        ttl = ttl or self.default_ttl
        
        try:
            serialized = pickle.dumps(value)
            await self.redis.setex(full_key, ttl, serialized)
            return True
        except Exception as e:
            logger.error(f"Cache set error for {full_key}: {e}")
            return False
    
    async def delete(self, namespace: str, key: str) -> bool:
        """Delete a cache entry"""
        full_key = self._make_key(namespace, key)
        try:
            result = await self.redis.delete(full_key)
            return result > 0
        except Exception as e:
            logger.error(f"Cache delete error for {full_key}: {e}")
            return False
    
    async def clear_namespace(self, namespace: str) -> int:
        """Clear all entries in a namespace"""
        pattern = self._make_key(namespace, "*")
        deleted = 0
        try:
            async for key in self.redis.scan_iter(match=pattern):
                if await self.redis.delete(key):
                    deleted += 1
        except Exception as e:
            logger.error(f"Cache clear error for namespace {namespace}: {e}")
        return deleted
    
    async def get_or_set(
        self,
        namespace: str,
        key: str,
        factory: Callable,
        ttl: Optional[int] = None
    ) -> Any:
        """
        Get from cache or compute and cache if missing
        
        Args:
            namespace: Cache namespace
            key: Cache key
            factory: Async function to compute value if not cached
            ttl: Time to live in seconds
            
        Returns:
            Cached or computed value
        """
        # Try to get from cache first
        value = await self.get(namespace, key)
        if value is not None:
            return value
        
        # Compute value
        value = await factory()
        
        # Cache it
        await self.set(namespace, key, value, ttl)
        
        return value


def cache_key(*args, **kwargs) -> str:
    """Generate a cache key from function arguments"""
    # Create a string representation of args and kwargs
    key_parts = [str(arg) for arg in args]
    key_parts.extend(f"{k}={v}" for k, v in sorted(kwargs.items()))
    key_string = ":".join(key_parts)
    
    # Hash it to ensure consistent length
    return hashlib.md5(key_string.encode()).hexdigest()


def cached(
    namespace: str,
    ttl: Union[int, timedelta] = 300,
    key_func: Optional[Callable] = None
):
    """
    Decorator for caching async function results
    
    Args:
        namespace: Cache namespace
        ttl: Time to live (seconds or timedelta)
        key_func: Optional function to generate cache key from args
        
    Usage:
        @cached("api_docs", ttl=600)
        async def get_api_documentation(api_id: str):
            # Expensive operation
            return await fetch_from_database(api_id)
    """
    if isinstance(ttl, timedelta):
        ttl = int(ttl.total_seconds())
    
    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            # Get cache manager from somewhere (e.g., global or dependency injection)
            from .main import cache_manager
            if not cache_manager:
                # No cache available, just call function
                return await func(*args, **kwargs)
            
            # Generate cache key
            if key_func:
                key = key_func(*args, **kwargs)
            else:
                # Skip 'self' argument for methods
                cache_args = args[1:] if args and hasattr(args[0], '__dict__') else args
                key = cache_key(*cache_args, **kwargs)
            
            # Try to get from cache
            result = await cache_manager.get(namespace, key)
            if result is not None:
                return result
            
            # Compute and cache
            result = await func(*args, **kwargs)
            await cache_manager.set(namespace, key, result, ttl)
            
            return result
        
        return wrapper
    return decorator


class DocumentationCache:
    """Specialized cache for API documentation"""
    
    def __init__(self, cache_manager: CacheManager):
        self.cache = cache_manager
        self.namespace = "api_docs"
        self.ttl = 3600  # 1 hour for documentation
    
    async def get_documentation(self, api_id: str, version: str = "latest") -> Optional[dict]:
        """Get cached API documentation"""
        key = f"{api_id}:{version}"
        return await self.cache.get(self.namespace, key)
    
    async def set_documentation(
        self, 
        api_id: str, 
        documentation: dict, 
        version: str = "latest"
    ) -> bool:
        """Cache API documentation"""
        key = f"{api_id}:{version}"
        return await self.cache.set(self.namespace, key, documentation, self.ttl)
    
    async def invalidate_documentation(self, api_id: str) -> int:
        """Invalidate all cached versions of API documentation"""
        pattern = f"{api_id}:*"
        deleted = 0
        # Note: This is a simplified version. In production, use SCAN
        # to avoid blocking Redis
        try:
            keys = await self.cache.redis.keys(
                self.cache._make_key(self.namespace, pattern)
            )
            if keys:
                deleted = await self.cache.redis.delete(*keys)
        except Exception as e:
            logger.error(f"Error invalidating docs for {api_id}: {e}")
        return deleted


class ResponseCache:
    """Cache for API endpoint responses"""
    
    def __init__(self, cache_manager: CacheManager):
        self.cache = cache_manager
        self.namespace = "responses"
    
    def cache_endpoint(self, ttl: int = 60):
        """
        Decorator to cache endpoint responses
        
        Usage:
            @router.get("/api/stats")
            @response_cache.cache_endpoint(ttl=300)
            async def get_stats():
                return expensive_calculation()
        """
        def decorator(func):
            @wraps(func)
            async def wrapper(*args, **kwargs):
                # Generate cache key from endpoint and params
                import inspect
                sig = inspect.signature(func)
                bound = sig.bind(*args, **kwargs)
                bound.apply_defaults()
                
                # Create key from function name and arguments
                key_parts = [func.__name__]
                for param_name, param_value in bound.arguments.items():
                    if param_name not in ['self', 'request']:
                        key_parts.append(f"{param_name}={param_value}")
                
                key = ":".join(key_parts)
                
                # Try cache
                result = await self.cache.get(self.namespace, key)
                if result is not None:
                    return result
                
                # Compute and cache
                result = await func(*args, **kwargs)
                await self.cache.set(self.namespace, key, result, ttl)
                
                return result
            
            return wrapper
        return decorator


# Global cache instances (initialized in main.py)
cache_manager: Optional[CacheManager] = None
documentation_cache: Optional[DocumentationCache] = None
response_cache: Optional[ResponseCache] = None


def init_cache(redis_client: redis.Redis) -> None:
    """Initialize global cache instances"""
    global cache_manager, documentation_cache, response_cache
    
    cache_manager = CacheManager(redis_client)
    documentation_cache = DocumentationCache(cache_manager)
    response_cache = ResponseCache(cache_manager)
    
    logger.info("Cache layer initialized")