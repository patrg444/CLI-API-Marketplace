"""
Mock authentication for local development
Provides a simple auth system without AWS Cognito dependency
"""
from datetime import datetime, timedelta
from typing import Optional, Dict, Any
import jwt
from fastapi import HTTPException, status
from passlib.context import CryptContext
import uuid

# Password hashing
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")

# Mock user storage (in production, this comes from Cognito/Database)
MOCK_USERS = {
    "demo@apidirect.dev": {
        "id": "user_demo123",
        "email": "demo@apidirect.dev",
        "password": "$2b$12$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPGga31lW",  # secret
        "name": "Demo User",
        "role": "creator",
        "created_at": "2024-01-01T00:00:00Z"
    }
}

class MockAuthService:
    def __init__(self, secret_key: str, algorithm: str = "HS256"):
        self.secret_key = secret_key
        self.algorithm = algorithm
        self.access_token_expire_minutes = 30
        self.refresh_token_expire_days = 7

    def verify_password(self, plain_password: str, hashed_password: str) -> bool:
        """Verify a password against its hash"""
        return pwd_context.verify(plain_password, hashed_password)

    def get_password_hash(self, password: str) -> str:
        """Hash a password"""
        return pwd_context.hash(password)

    def authenticate_user(self, email: str, password: str) -> Optional[Dict[str, Any]]:
        """Authenticate a user with email and password"""
        user = MOCK_USERS.get(email)
        if not user:
            return None
        if not self.verify_password(password, user["password"]):
            return None
        return user

    def create_access_token(self, data: dict, expires_delta: Optional[timedelta] = None):
        """Create a JWT access token"""
        to_encode = data.copy()
        if expires_delta:
            expire = datetime.utcnow() + expires_delta
        else:
            expire = datetime.utcnow() + timedelta(minutes=self.access_token_expire_minutes)
        
        to_encode.update({"exp": expire, "type": "access"})
        encoded_jwt = jwt.encode(to_encode, self.secret_key, algorithm=self.algorithm)
        return encoded_jwt

    def create_refresh_token(self, data: dict):
        """Create a JWT refresh token"""
        to_encode = data.copy()
        expire = datetime.utcnow() + timedelta(days=self.refresh_token_expire_days)
        to_encode.update({"exp": expire, "type": "refresh"})
        encoded_jwt = jwt.encode(to_encode, self.secret_key, algorithm=self.algorithm)
        return encoded_jwt

    def decode_token(self, token: str) -> Dict[str, Any]:
        """Decode and verify a JWT token"""
        try:
            payload = jwt.decode(token, self.secret_key, algorithms=[self.algorithm])
            return payload
        except jwt.ExpiredSignatureError:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Token has expired",
                headers={"WWW-Authenticate": "Bearer"},
            )
        except jwt.JWTError:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Could not validate credentials",
                headers={"WWW-Authenticate": "Bearer"},
            )

    def register_user(self, email: str, password: str, name: str) -> Dict[str, Any]:
        """Register a new user (mock implementation)"""
        if email in MOCK_USERS:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="User already exists"
            )
        
        user_id = f"user_{uuid.uuid4().hex[:8]}"
        user = {
            "id": user_id,
            "email": email,
            "password": self.get_password_hash(password),
            "name": name,
            "role": "creator",
            "created_at": datetime.utcnow().isoformat()
        }
        
        MOCK_USERS[email] = user
        
        # Return user without password
        return {k: v for k, v in user.items() if k != "password"}

    def get_current_user(self, token: str) -> Dict[str, Any]:
        """Get current user from token"""
        payload = self.decode_token(token)
        email = payload.get("sub")
        if email is None:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Could not validate credentials",
            )
        
        user = MOCK_USERS.get(email)
        if user is None:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="User not found",
            )
        
        # Return user without password
        return {k: v for k, v in user.items() if k != "password"}

# Mock API key management
MOCK_API_KEYS = {
    "demo_ak_12345": {
        "id": "key_123",
        "name": "Demo API Key",
        "user_id": "user_demo123",
        "created_at": "2024-01-01T00:00:00Z",
        "last_used": None,
        "permissions": ["read", "write"]
    }
}

def validate_api_key(api_key: str) -> Optional[Dict[str, Any]]:
    """Validate an API key"""
    return MOCK_API_KEYS.get(api_key)

# FastAPI dependency function
from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials

security = HTTPBearer()

# Initialize the auth service with a default secret
auth_service = MockAuthService(secret_key="local-development-secret-key")

async def get_current_user(credentials: HTTPAuthorizationCredentials = Depends(security)) -> Dict[str, Any]:
    """FastAPI dependency to get current user from JWT token"""
    token = credentials.credentials
    try:
        return auth_service.get_current_user(token)
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Could not validate credentials",
            headers={"WWW-Authenticate": "Bearer"},
        )