#!/usr/bin/env python3
"""
API-Direct Backend API with Mock Authentication
FastAPI application serving the Creator Portal and CLI
"""

from fastapi import FastAPI, HTTPException, Depends, Security, WebSocket
from fastapi.middleware.cors import CORSMiddleware
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from .auth.mock_auth import MockAuthService, validate_api_key
from .routes.api_creation import router as api_creation_router
import os
from datetime import datetime, timedelta
from typing import List, Optional, Dict, Any
import json
import logging
from pydantic import BaseModel, EmailStr

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize FastAPI app
app = FastAPI(
    title="API-Direct Backend",
    description="Backend API for API-Direct platform",
    version="1.0.0"
)

# CORS configuration
origins = os.getenv("CORS_ORIGINS", "http://localhost:3000,http://localhost:8080").split(",")
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Security
security = HTTPBearer()

# Initialize mock auth service
USE_MOCK_AUTH = os.getenv("USE_MOCK_AUTH", "false").lower() == "true"
JWT_SECRET = os.getenv("JWT_SECRET", "local-development-secret")
auth_service = MockAuthService(JWT_SECRET) if USE_MOCK_AUTH else None

# Include routers
app.include_router(api_creation_router, prefix="/api")

# Import and include deployment router
from .routes.deployments import router as deployment_router
app.include_router(deployment_router, prefix="/api")

# Models
class UserRegister(BaseModel):
    name: str
    email: EmailStr
    password: str
    company: Optional[str] = None

class UserLogin(BaseModel):
    email: EmailStr
    password: str

class UserResponse(BaseModel):
    id: str
    name: str
    email: str
    company: Optional[str] = None
    profile_image: Optional[str] = None
    created_at: str

class TokenResponse(BaseModel):
    access_token: str
    token_type: str
    expires_in: int
    user: UserResponse

class APIKeyCreate(BaseModel):
    name: str
    permissions: List[str] = ["read", "write"]

class APIKeyResponse(BaseModel):
    id: str
    key: str
    name: str
    created_at: str
    last_used: Optional[str] = None
    permissions: List[str]

# Dependency to get current user
async def get_current_user(credentials: HTTPAuthorizationCredentials = Security(security)) -> Dict[str, Any]:
    """Get current user from JWT token"""
    if not USE_MOCK_AUTH:
        raise HTTPException(status_code=503, detail="Authentication service not available")
    
    token = credentials.credentials
    try:
        user = auth_service.get_current_user(token)
        return user
    except Exception as e:
        raise HTTPException(status_code=401, detail=str(e))

# Health check
@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "timestamp": datetime.utcnow().isoformat(),
        "version": "1.0.0",
        "mock_auth": USE_MOCK_AUTH
    }

# Authentication endpoints
@app.post("/auth/register", response_model=TokenResponse)
async def register(user_data: UserRegister):
    """Register a new user"""
    if not USE_MOCK_AUTH:
        raise HTTPException(status_code=503, detail="Registration not available")
    
    try:
        # Register user
        user = auth_service.register_user(
            email=user_data.email,
            password=user_data.password,
            name=user_data.name
        )
        
        # Create tokens
        access_token = auth_service.create_access_token(data={"sub": user["email"]})
        
        return {
            "access_token": access_token,
            "token_type": "bearer",
            "expires_in": 1800,
            "user": UserResponse(**user)
        }
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Registration error: {e}")
        raise HTTPException(status_code=400, detail="Registration failed")

@app.post("/auth/login", response_model=TokenResponse)
async def login(user_data: UserLogin):
    """Login user"""
    if not USE_MOCK_AUTH:
        raise HTTPException(status_code=503, detail="Login not available")
    
    user = auth_service.authenticate_user(user_data.email, user_data.password)
    if not user:
        raise HTTPException(status_code=401, detail="Invalid credentials")
    
    access_token = auth_service.create_access_token(data={"sub": user["email"]})
    
    return {
        "access_token": access_token,
        "token_type": "bearer",
        "expires_in": 1800,
        "user": UserResponse(**{k: v for k, v in user.items() if k != "password"})
    }

@app.get("/auth/me", response_model=UserResponse)
async def get_me(current_user: Dict = Depends(get_current_user)):
    """Get current user info"""
    return UserResponse(**current_user)

# API Key Management
@app.post("/api-keys", response_model=APIKeyResponse)
async def create_api_key(
    key_data: APIKeyCreate,
    current_user: Dict = Depends(get_current_user)
):
    """Create a new API key"""
    # Generate a demo API key
    api_key = f"ak_{current_user['id']}_{datetime.utcnow().strftime('%Y%m%d%H%M%S')}"
    
    return {
        "id": f"key_{datetime.utcnow().timestamp()}",
        "key": api_key,
        "name": key_data.name,
        "created_at": datetime.utcnow().isoformat(),
        "last_used": None,
        "permissions": key_data.permissions
    }

@app.get("/api-keys", response_model=List[APIKeyResponse])
async def list_api_keys(current_user: Dict = Depends(get_current_user)):
    """List user's API keys"""
    # Return mock data for demo
    return [
        {
            "id": "key_1",
            "key": "ak_demo_12345****",
            "name": "Production Key",
            "created_at": "2024-01-01T00:00:00Z",
            "last_used": "2024-01-15T12:00:00Z",
            "permissions": ["read", "write"]
        }
    ]

# Dashboard Stats
@app.get("/dashboard/stats")
async def get_dashboard_stats(current_user: Dict = Depends(get_current_user)):
    """Get dashboard statistics"""
    return {
        "total_apis": 3,
        "active_subscriptions": 127,
        "monthly_revenue": 2847.50,
        "api_calls_today": 15420,
        "growth_percentage": 23.5,
        "popular_apis": [
            {"name": "Weather API", "calls": 5420},
            {"name": "Translation API", "calls": 3890},
            {"name": "Image Recognition API", "calls": 2150}
        ]
    }

# Analytics
@app.get("/analytics/overview")
async def get_analytics_overview(
    current_user: Dict = Depends(get_current_user),
    period: str = "7d"
):
    """Get analytics overview"""
    # Generate mock data based on period
    data_points = 7 if period == "7d" else 30
    
    return {
        "period": period,
        "total_calls": 108420,
        "unique_users": 89,
        "average_latency": 145,
        "error_rate": 0.02,
        "daily_stats": [
            {
                "date": (datetime.utcnow() - timedelta(days=i)).strftime("%Y-%m-%d"),
                "calls": 15000 + (i * 1000),
                "errors": 30 + (i * 5),
                "latency": 140 + (i * 2)
            }
            for i in range(data_points)
        ]
    }

# My APIs
@app.get("/apis")
async def list_my_apis(current_user: Dict = Depends(get_current_user)):
    """List user's APIs"""
    return {
        "apis": [
            {
                "id": "api_1",
                "name": "Weather Forecast API",
                "description": "Real-time weather data and forecasts",
                "status": "active",
                "version": "1.2.0",
                "created_at": "2024-01-01T00:00:00Z",
                "monthly_calls": 45320,
                "subscribers": 42
            },
            {
                "id": "api_2",
                "name": "Translation API",
                "description": "Multi-language translation service",
                "status": "active",
                "version": "2.0.1",
                "created_at": "2024-01-15T00:00:00Z",
                "monthly_calls": 28900,
                "subscribers": 31
            }
        ]
    }

# Billing
@app.get("/billing/summary")
async def get_billing_summary(current_user: Dict = Depends(get_current_user)):
    """Get billing summary"""
    return {
        "current_balance": 2847.50,
        "pending_payout": 450.00,
        "next_payout_date": "2024-02-01",
        "lifetime_earnings": 15420.75,
        "recent_transactions": [
            {
                "id": "txn_1",
                "date": "2024-01-20T00:00:00Z",
                "description": "API calls - Weather Forecast API",
                "amount": 127.50,
                "type": "earning"
            },
            {
                "id": "txn_2",
                "date": "2024-01-19T00:00:00Z",
                "description": "API calls - Translation API",
                "amount": 89.25,
                "type": "earning"
            }
        ]
    }

# WebSocket endpoint for real-time updates
@app.websocket("/ws")
async def websocket_endpoint_handler(websocket: WebSocket):
    """WebSocket endpoint for real-time updates"""
    await websocket.accept()
    
    try:
        while True:
            # Send mock real-time data
            await websocket.send_json({
                "type": "stats_update",
                "data": {
                    "api_calls": 15420 + int(datetime.utcnow().timestamp() % 100),
                    "active_users": 89 + int(datetime.utcnow().timestamp() % 10),
                    "timestamp": datetime.utcnow().isoformat()
                }
            })
            
            # Wait for 5 seconds before next update
            await asyncio.sleep(5)
            
    except Exception as e:
        logger.error(f"WebSocket error: {e}")
    finally:
        await websocket.close()

# Add asyncio import for websocket
import asyncio

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)