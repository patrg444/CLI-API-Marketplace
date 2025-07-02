"""
Hybrid authentication supporting both AWS Cognito and mock auth
"""
import os
from typing import Optional, Dict, Any
from fastapi import HTTPException, Security, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials

from .cognito import get_cognito_auth
from .mock_auth import MockAuthService

# Security scheme
security = HTTPBearer()

# Initialize mock auth if enabled
USE_MOCK_AUTH = os.getenv("USE_MOCK_AUTH", "false").lower() == "true"
JWT_SECRET = os.getenv("JWT_SECRET", "local-development-secret")
mock_auth = MockAuthService(JWT_SECRET) if USE_MOCK_AUTH else None

async def get_current_user(
    credentials: HTTPAuthorizationCredentials = Security(security)
) -> Dict[str, Any]:
    """
    Get current user from token.
    Supports both Cognito tokens and mock auth tokens.
    """
    token = credentials.credentials
    
    # Try Cognito first if available
    cognito = get_cognito_auth()
    if cognito and not USE_MOCK_AUTH:
        try:
            return await cognito.get_user_from_token(token)
        except HTTPException:
            raise
        except Exception as e:
            # Log error and fall through to mock auth if enabled
            print(f"Cognito auth failed: {e}")
            if not USE_MOCK_AUTH:
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Authentication failed"
                )
    
    # Try mock auth if enabled
    if USE_MOCK_AUTH and mock_auth:
        try:
            return mock_auth.get_current_user(token)
        except HTTPException:
            raise
        except Exception as e:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail=f"Authentication failed: {str(e)}"
            )
    
    # No auth method available
    raise HTTPException(
        status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
        detail="No authentication service available"
    )

async def get_optional_user(
    credentials: Optional[HTTPAuthorizationCredentials] = Security(security)
) -> Optional[Dict[str, Any]]:
    """Get current user if authenticated, otherwise return None"""
    if not credentials:
        return None
    
    try:
        return await get_current_user(credentials)
    except:
        return None

def create_mock_token(user_data: Dict[str, Any]) -> str:
    """Create a mock token for testing"""
    if not mock_auth:
        raise ValueError("Mock auth not enabled")
    
    return mock_auth.create_access_token(data={"sub": user_data.get("email", "test@example.com")})

async def validate_api_key(api_key: str) -> Optional[Dict[str, Any]]:
    """
    Validate an API key against the database.
    Falls back to mock validation in development mode.
    """
    # Use mock validation in development
    if USE_MOCK_AUTH:
        from .mock_auth import validate_api_key as mock_validate
        return mock_validate(api_key)
    
    # Use real API key validation in production
    from ..database import get_db_pool
    from .api_keys import APIKeyManager
    
    db_pool = await get_db_pool()
    if not db_pool:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Database connection not available"
        )
    
    api_key_manager = APIKeyManager(db_pool)
    return await api_key_manager.validate_api_key(api_key)