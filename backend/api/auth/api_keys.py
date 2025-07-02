"""
API Key management for CLI authentication
Provides creation, validation, and management of API keys
"""

import secrets
import hashlib
import asyncpg
from datetime import datetime, timedelta
from typing import Optional, List, Dict, Any
import bcrypt
from fastapi import HTTPException, Security, Depends
from fastapi.security import APIKeyHeader
import logging

logger = logging.getLogger(__name__)

# API Key header
api_key_header = APIKeyHeader(name="X-API-Key", auto_error=False)


class APIKeyManager:
    """Manages API keys for CLI authentication"""
    
    def __init__(self, db_pool: asyncpg.Pool):
        self.db_pool = db_pool
    
    async def create_api_key(
        self, 
        user_id: str, 
        name: str, 
        scopes: List[str] = None,
        expires_in_days: Optional[int] = None
    ) -> Dict[str, Any]:
        """
        Create a new API key for a user
        
        Args:
            user_id: The user's ID
            name: A descriptive name for the key
            scopes: List of permitted operations (default: all)
            expires_in_days: Number of days until expiration (default: no expiration)
            
        Returns:
            Dict containing the key details and the actual key (shown only once)
        """
        # Generate a secure random key
        raw_key = secrets.token_urlsafe(32)
        key_prefix = raw_key[:8]  # First 8 chars for easy identification
        
        # Hash the key for storage
        key_hash = bcrypt.hashpw(raw_key.encode('utf-8'), bcrypt.gensalt()).decode('utf-8')
        
        # Calculate expiration
        expires_at = None
        if expires_in_days:
            expires_at = datetime.utcnow() + timedelta(days=expires_in_days)
        
        # Default scopes
        if scopes is None:
            scopes = ["read", "write", "deploy"]
        
        async with self.db_pool.acquire() as conn:
            # Check existing keys count
            key_count = await conn.fetchval(
                "SELECT COUNT(*) FROM user_api_keys WHERE user_id = $1",
                user_id
            )
            
            if key_count >= 10:
                raise HTTPException(
                    status_code=400,
                    detail="Maximum number of API keys (10) reached"
                )
            
            # Insert the new key
            key_id = await conn.fetchval("""
                INSERT INTO user_api_keys 
                (user_id, key_hash, name, scopes, expires_at)
                VALUES ($1, $2, $3, $4, $5)
                RETURNING id
            """, user_id, key_hash, name, scopes, expires_at)
            
            logger.info(f"Created API key '{name}' for user {user_id}")
            
            return {
                "id": str(key_id),
                "name": name,
                "key": raw_key,  # Only returned once!
                "key_prefix": key_prefix,
                "scopes": scopes,
                "expires_at": expires_at.isoformat() if expires_at else None,
                "created_at": datetime.utcnow().isoformat()
            }
    
    async def validate_api_key(self, api_key: str, client_ip: Optional[str] = None) -> Optional[Dict[str, Any]]:
        """
        Validate an API key and return user info with rate limiting
        
        Args:
            api_key: The API key to validate
            client_ip: Client IP address for rate limiting
            
        Returns:
            User info dict if valid, None otherwise
        """
        if not api_key:
            return None
        
        # Get rate limiter if available
        from .rate_limiter import get_rate_limiter
        rate_limiter = get_rate_limiter()
        
        # Use API key prefix for rate limiting (first 8 chars)
        identifier = client_ip or api_key[:8] if len(api_key) >= 8 else api_key
        
        # Check rate limit before attempting validation
        if rate_limiter:
            await rate_limiter.check_rate_limit(identifier, "api_key")
        
        try:
            async with self.db_pool.acquire() as conn:
                # First, try to find by key prefix for efficiency (if we implement prefix indexing)
                # For now, get all active keys (we need to check each hash)
                keys = await conn.fetch("""
                    SELECT 
                        k.id, k.user_id, k.key_hash, k.scopes, k.expires_at,
                        u.email, u.name, u.is_active
                    FROM user_api_keys k
                    JOIN users u ON k.user_id = u.id
                    WHERE u.is_active = TRUE
                    AND (k.expires_at IS NULL OR k.expires_at > NOW())
                """)
                
                # Check each key
                for key_record in keys:
                    if bcrypt.checkpw(api_key.encode('utf-8'), key_record['key_hash'].encode('utf-8')):
                        # Update last used timestamp
                        await conn.execute(
                            "UPDATE user_api_keys SET last_used_at = NOW() WHERE id = $1",
                            key_record['id']
                        )
                        
                        # Record successful attempt
                        if rate_limiter:
                            await rate_limiter.record_attempt(identifier, "api_key", success=True)
                        
                        return {
                            "user_id": str(key_record['user_id']),
                            "email": key_record['email'],
                            "name": key_record['name'],
                            "scopes": key_record['scopes'],
                            "key_id": str(key_record['id'])
                        }
                
                # No valid key found - record failed attempt
                if rate_limiter:
                    await rate_limiter.record_attempt(identifier, "api_key", success=False)
                
                return None
                
        except Exception as e:
            logger.error(f"Error validating API key: {e}")
            # Record failed attempt on error
            if rate_limiter:
                await rate_limiter.record_attempt(identifier, "api_key", success=False)
            raise
    
    async def list_api_keys(self, user_id: str) -> List[Dict[str, Any]]:
        """
        List all API keys for a user
        
        Args:
            user_id: The user's ID
            
        Returns:
            List of API key details (without the actual keys)
        """
        async with self.db_pool.acquire() as conn:
            keys = await conn.fetch("""
                SELECT 
                    id, name, scopes, last_used_at, expires_at, created_at,
                    SUBSTRING(key_hash, 1, 8) as key_prefix
                FROM user_api_keys
                WHERE user_id = $1
                ORDER BY created_at DESC
            """, user_id)
            
            return [
                {
                    "id": str(row['id']),
                    "name": row['name'],
                    "key_prefix": row['key_prefix'],
                    "scopes": row['scopes'],
                    "last_used_at": row['last_used_at'].isoformat() if row['last_used_at'] else None,
                    "expires_at": row['expires_at'].isoformat() if row['expires_at'] else None,
                    "created_at": row['created_at'].isoformat(),
                    "is_expired": row['expires_at'] and row['expires_at'] < datetime.utcnow()
                }
                for row in keys
            ]
    
    async def revoke_api_key(self, user_id: str, key_id: str) -> bool:
        """
        Revoke an API key
        
        Args:
            user_id: The user's ID
            key_id: The key ID to revoke
            
        Returns:
            True if revoked, False if not found
        """
        async with self.db_pool.acquire() as conn:
            result = await conn.execute("""
                DELETE FROM user_api_keys
                WHERE id = $1 AND user_id = $2
            """, key_id, user_id)
            
            revoked = result.split()[-1] == '1'
            if revoked:
                logger.info(f"Revoked API key {key_id} for user {user_id}")
            
            return revoked
    
    def check_scope(self, user_info: Dict[str, Any], required_scope: str) -> bool:
        """
        Check if a user has the required scope
        
        Args:
            user_info: User info from validate_api_key
            required_scope: The scope to check
            
        Returns:
            True if user has the scope
        """
        if not user_info or 'scopes' not in user_info:
            return False
        
        scopes = user_info.get('scopes', [])
        return required_scope in scopes or 'admin' in scopes


# Dependency for FastAPI routes
async def get_current_api_user(
    api_key: str = Security(api_key_header)
) -> Dict[str, Any]:
    """
    FastAPI dependency to get current user from API key
    
    Usage:
        @app.get("/api/protected")
        async def protected_route(user: dict = Depends(get_current_api_user)):
            return {"user": user}
    """
    if not api_key:
        raise HTTPException(
            status_code=401,
            detail="API key required",
            headers={"WWW-Authenticate": "ApiKey"},
        )
    
    # Import here to avoid circular imports
    from main import db_pool
    
    if not db_pool:
        raise HTTPException(
            status_code=500,
            detail="Database not initialized"
        )
    
    manager = APIKeyManager(db_pool)
    user_info = await manager.validate_api_key(api_key)
    
    if not user_info:
        raise HTTPException(
            status_code=401,
            detail="Invalid API key",
            headers={"WWW-Authenticate": "ApiKey"},
        )
    
    return user_info


def require_scope(scope: str):
    """
    Dependency to require a specific scope
    
    Usage:
        @app.post("/api/deploy", dependencies=[Depends(require_scope("deploy"))])
        async def deploy_api(user: dict = Depends(get_current_api_user)):
            return {"deployed": True}
    """
    async def scope_checker(user: dict = Depends(get_current_api_user)):
        if scope not in user.get('scopes', []):
            raise HTTPException(
                status_code=403,
                detail=f"Insufficient permissions. Required scope: {scope}"
            )
        return user
    
    return scope_checker