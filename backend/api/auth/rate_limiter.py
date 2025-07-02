"""
Rate limiting for authentication attempts
"""
import time
from typing import Dict, Optional
import redis.asyncio as redis
from fastapi import HTTPException, status
import logging

logger = logging.getLogger(__name__)

class AuthRateLimiter:
    """Rate limiter for authentication attempts"""
    
    def __init__(self, redis_client: redis.Redis):
        self.redis = redis_client
        self.max_attempts = 5  # Maximum attempts per window
        self.window_seconds = 300  # 5 minute window
        self.lockout_seconds = 900  # 15 minute lockout after max attempts
    
    async def check_rate_limit(self, identifier: str, attempt_type: str = "api_key") -> None:
        """
        Check if an identifier (IP, API key prefix, etc.) has exceeded rate limits
        
        Args:
            identifier: The identifier to check (e.g., IP address, API key prefix)
            attempt_type: Type of attempt (api_key, login, etc.)
            
        Raises:
            HTTPException: If rate limit exceeded
        """
        key = f"rate_limit:{attempt_type}:{identifier}"
        lockout_key = f"lockout:{attempt_type}:{identifier}"
        
        # Check if currently locked out
        if await self.redis.exists(lockout_key):
            ttl = await self.redis.ttl(lockout_key)
            raise HTTPException(
                status_code=status.HTTP_429_TOO_MANY_REQUESTS,
                detail=f"Too many failed attempts. Try again in {ttl} seconds.",
                headers={"Retry-After": str(ttl)}
            )
        
        # Get current attempt count
        attempts = await self.redis.get(key)
        current_attempts = int(attempts) if attempts else 0
        
        if current_attempts >= self.max_attempts:
            # Set lockout
            await self.redis.setex(lockout_key, self.lockout_seconds, "1")
            await self.redis.delete(key)  # Reset counter
            
            logger.warning(f"Rate limit exceeded for {attempt_type}:{identifier}")
            
            raise HTTPException(
                status_code=status.HTTP_429_TOO_MANY_REQUESTS,
                detail=f"Too many failed attempts. Account locked for {self.lockout_seconds // 60} minutes.",
                headers={"Retry-After": str(self.lockout_seconds)}
            )
    
    async def record_attempt(self, identifier: str, attempt_type: str = "api_key", success: bool = False) -> None:
        """
        Record an authentication attempt
        
        Args:
            identifier: The identifier (e.g., IP address, API key prefix)
            attempt_type: Type of attempt
            success: Whether the attempt was successful
        """
        if success:
            # Clear rate limit on successful attempt
            key = f"rate_limit:{attempt_type}:{identifier}"
            await self.redis.delete(key)
            return
        
        # Record failed attempt
        key = f"rate_limit:{attempt_type}:{identifier}"
        
        # Increment counter with expiry
        pipe = self.redis.pipeline()
        pipe.incr(key)
        pipe.expire(key, self.window_seconds)
        await pipe.execute()
    
    async def get_attempt_info(self, identifier: str, attempt_type: str = "api_key") -> Dict[str, any]:
        """
        Get current attempt information for an identifier
        
        Returns:
            Dict with attempts count and lockout status
        """
        key = f"rate_limit:{attempt_type}:{identifier}"
        lockout_key = f"lockout:{attempt_type}:{identifier}"
        
        attempts = await self.redis.get(key)
        current_attempts = int(attempts) if attempts else 0
        
        is_locked = await self.redis.exists(lockout_key)
        lockout_ttl = await self.redis.ttl(lockout_key) if is_locked else 0
        
        return {
            "attempts": current_attempts,
            "max_attempts": self.max_attempts,
            "is_locked": bool(is_locked),
            "lockout_ttl": lockout_ttl,
            "attempts_remaining": max(0, self.max_attempts - current_attempts)
        }


# Global rate limiter instance (initialized in main.py)
rate_limiter: Optional[AuthRateLimiter] = None

def get_rate_limiter() -> Optional[AuthRateLimiter]:
    """Get the global rate limiter instance"""
    return rate_limiter

def set_rate_limiter(limiter: AuthRateLimiter) -> None:
    """Set the global rate limiter instance"""
    global rate_limiter
    rate_limiter = limiter