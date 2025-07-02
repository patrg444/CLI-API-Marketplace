"""
Input validation models using Pydantic v2
Provides comprehensive validation for all API endpoints
"""

from pydantic import BaseModel, EmailStr, Field, validator, ConfigDict
from typing import Optional, List, Dict, Any, Literal
from datetime import datetime
import re

# API name validation pattern
API_NAME_PATTERN = re.compile(r'^[a-zA-Z][a-zA-Z0-9-_]{2,63}$')
# Secure password pattern (min 8 chars, 1 upper, 1 lower, 1 digit, 1 special)
PASSWORD_PATTERN = re.compile(r'^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$')


class UserRegistrationRequest(BaseModel):
    """User registration validation"""
    email: EmailStr
    password: str = Field(..., min_length=8, max_length=128)
    name: str = Field(..., min_length=2, max_length=100)
    company: Optional[str] = Field(None, max_length=100)
    
    @validator('password')
    def validate_password(cls, v):
        if not PASSWORD_PATTERN.match(v):
            raise ValueError(
                'Password must be at least 8 characters with 1 uppercase, '
                '1 lowercase, 1 digit, and 1 special character'
            )
        return v
    
    @validator('name')
    def validate_name(cls, v):
        if not v.strip():
            raise ValueError('Name cannot be empty')
        return v.strip()


class UserLoginRequest(BaseModel):
    """User login validation"""
    email: EmailStr
    password: str = Field(..., min_length=1)


class APICreateRequest(BaseModel):
    """API creation validation"""
    name: str = Field(..., min_length=3, max_length=64)
    description: str = Field(..., min_length=10, max_length=500)
    category: Literal['ai', 'data', 'financial', 'communication', 'utility', 'other']
    tags: List[str] = Field(default=[], max_items=10)
    visibility: Literal['public', 'private'] = 'private'
    
    @validator('name')
    def validate_api_name(cls, v):
        if not API_NAME_PATTERN.match(v):
            raise ValueError(
                'API name must start with a letter and contain only '
                'letters, numbers, hyphens, and underscores (3-64 chars)'
            )
        return v
    
    @validator('tags')
    def validate_tags(cls, v):
        # Ensure unique tags
        unique_tags = list(set(tag.lower().strip() for tag in v if tag.strip()))
        # Validate each tag
        for tag in unique_tags:
            if len(tag) < 2 or len(tag) > 30:
                raise ValueError(f'Tag "{tag}" must be 2-30 characters')
            if not re.match(r'^[a-z0-9-]+$', tag):
                raise ValueError(f'Tag "{tag}" can only contain lowercase letters, numbers, and hyphens')
        return unique_tags


class APIUpdateRequest(BaseModel):
    """API update validation"""
    name: Optional[str] = Field(None, min_length=3, max_length=64)
    description: Optional[str] = Field(None, min_length=10, max_length=500)
    category: Optional[Literal['ai', 'data', 'financial', 'communication', 'utility', 'other']] = None
    tags: Optional[List[str]] = Field(None, max_items=10)
    visibility: Optional[Literal['public', 'private']] = None
    
    @validator('name')
    def validate_api_name(cls, v):
        if v and not API_NAME_PATTERN.match(v):
            raise ValueError(
                'API name must start with a letter and contain only '
                'letters, numbers, hyphens, and underscores (3-64 chars)'
            )
        return v


class PricingConfigRequest(BaseModel):
    """API pricing configuration validation"""
    model: Literal['free', 'pay_per_use', 'subscription', 'freemium']
    
    # Pay-per-use pricing
    price_per_request: Optional[float] = Field(None, ge=0.001, le=100.0)
    free_requests: Optional[int] = Field(None, ge=0, le=10000)
    
    # Subscription pricing
    monthly_price: Optional[float] = Field(None, ge=0.99, le=9999.99)
    annual_price: Optional[float] = Field(None, ge=9.99, le=99999.99)
    trial_days: Optional[int] = Field(None, ge=0, le=90)
    
    # Rate limits
    requests_per_minute: Optional[int] = Field(None, ge=1, le=10000)
    requests_per_day: Optional[int] = Field(None, ge=1, le=1000000)
    
    @validator('price_per_request')
    def validate_pay_per_use(cls, v, values):
        if values.get('model') == 'pay_per_use' and v is None:
            raise ValueError('price_per_request is required for pay-per-use model')
        return v
    
    @validator('monthly_price')
    def validate_subscription(cls, v, values):
        if values.get('model') == 'subscription' and v is None:
            raise ValueError('monthly_price is required for subscription model')
        return v


class APIKeyCreateRequest(BaseModel):
    """API key creation validation"""
    name: str = Field(..., min_length=3, max_length=100)
    scopes: List[Literal['read', 'write', 'deploy', 'admin']] = ['read', 'write']
    expires_in_days: Optional[int] = Field(None, ge=1, le=365)
    
    @validator('scopes')
    def validate_scopes(cls, v):
        return list(set(v))  # Remove duplicates


class WebhookConfigRequest(BaseModel):
    """Webhook configuration validation"""
    url: str = Field(..., regex=r'^https://.*')
    events: List[str] = Field(..., min_items=1, max_items=20)
    secret: Optional[str] = Field(None, min_length=16, max_length=64)
    active: bool = True
    
    @validator('url')
    def validate_webhook_url(cls, v):
        # Additional security checks
        forbidden_hosts = ['localhost', '127.0.0.1', '0.0.0.0', '::1']
        from urllib.parse import urlparse
        parsed = urlparse(v)
        if parsed.hostname in forbidden_hosts:
            raise ValueError('Webhook URL cannot point to localhost')
        return v


class DeploymentRequest(BaseModel):
    """API deployment validation"""
    api_id: str
    environment: Literal['development', 'staging', 'production']
    deployment_type: Literal['hosted', 'byoa']
    
    # BYOA specific
    aws_account_id: Optional[str] = Field(None, regex=r'^\d{12}$')
    aws_region: Optional[str] = Field(None, regex=r'^[a-z]{2}-[a-z]+-\d{1}$')
    
    # Resource limits
    min_instances: int = Field(1, ge=1, le=10)
    max_instances: int = Field(2, ge=1, le=100)
    memory_mb: int = Field(512, ge=128, le=8192)
    
    @validator('max_instances')
    def validate_instance_range(cls, v, values):
        if v < values.get('min_instances', 1):
            raise ValueError('max_instances must be >= min_instances')
        return v
    
    @validator('aws_account_id')
    def validate_byoa_config(cls, v, values):
        if values.get('deployment_type') == 'byoa' and not v:
            raise ValueError('aws_account_id is required for BYOA deployments')
        return v


class UsageQueryRequest(BaseModel):
    """Usage metrics query validation"""
    api_id: Optional[str] = None
    start_date: datetime
    end_date: datetime
    granularity: Literal['minute', 'hour', 'day', 'month'] = 'hour'
    metrics: List[Literal['requests', 'errors', 'latency', 'bandwidth']] = ['requests']
    
    @validator('end_date')
    def validate_date_range(cls, v, values):
        start = values.get('start_date')
        if start and v <= start:
            raise ValueError('end_date must be after start_date')
        
        # Max 90 days for minute granularity
        if values.get('granularity') == 'minute':
            max_days = 1
        elif values.get('granularity') == 'hour':
            max_days = 30
        else:
            max_days = 365
            
        if start and (v - start).days > max_days:
            raise ValueError(f'Date range too large for {values.get("granularity")} granularity (max {max_days} days)')
        
        return v


class PaymentMethodRequest(BaseModel):
    """Payment method validation"""
    stripe_payment_method_id: str = Field(..., regex=r'^pm_[a-zA-Z0-9]+$')
    set_as_default: bool = True


class SubscriptionRequest(BaseModel):
    """Subscription request validation"""
    api_id: str
    plan_id: str
    payment_method_id: Optional[str] = Field(None, regex=r'^pm_[a-zA-Z0-9]+$')
    
    
class SearchRequest(BaseModel):
    """Marketplace search validation"""
    query: Optional[str] = Field(None, max_length=200)
    category: Optional[str] = None
    tags: Optional[List[str]] = Field(None, max_items=10)
    price_range: Optional[Dict[str, float]] = None
    sort_by: Literal['relevance', 'popularity', 'price', 'newest'] = 'relevance'
    page: int = Field(1, ge=1)
    per_page: int = Field(20, ge=1, le=100)
    
    @validator('query')
    def sanitize_query(cls, v):
        if v:
            # Remove potential SQL injection characters
            v = re.sub(r'[^\w\s-]', '', v)
        return v
    
    @validator('price_range')
    def validate_price_range(cls, v):
        if v:
            min_price = v.get('min', 0)
            max_price = v.get('max', 999999)
            if min_price < 0 or max_price < 0:
                raise ValueError('Price cannot be negative')
            if min_price > max_price:
                raise ValueError('min price cannot be greater than max price')
        return v


# Response models with proper configuration
class APIResponse(BaseModel):
    """Standard API response"""
    model_config = ConfigDict(from_attributes=True)
    
    id: str
    name: str
    description: str
    category: str
    tags: List[str]
    creator_id: str
    creator_name: Optional[str]
    status: str
    visibility: str
    created_at: datetime
    updated_at: datetime
    
    # Stats
    total_requests: int = 0
    monthly_requests: int = 0
    average_latency: Optional[float] = None
    uptime_percentage: float = 99.9


class ErrorResponse(BaseModel):
    """Standard error response"""
    error: str
    detail: Optional[str] = None
    request_id: Optional[str] = None
    timestamp: datetime = Field(default_factory=datetime.utcnow)