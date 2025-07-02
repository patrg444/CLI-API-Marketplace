"""
Database models for API-Direct
"""
from pydantic import BaseModel, EmailStr
from typing import Optional, List, Dict, Any
from datetime import datetime

class User(BaseModel):
    id: str
    email: EmailStr
    name: str
    role: str = "creator"
    created_at: datetime
    api_count: int = 0
    total_requests: int = 0
    
    class Config:
        orm_mode = True

class APIKey(BaseModel):
    id: str
    name: str
    key: str
    user_id: str
    created_at: datetime
    last_used: Optional[datetime] = None
    permissions: List[str] = ["read", "write"]
    
    class Config:
        orm_mode = True

class API(BaseModel):
    id: str
    name: str
    description: str
    user_id: str
    endpoint: str
    status: str = "draft"
    created_at: datetime
    updated_at: datetime
    
    class Config:
        orm_mode = True