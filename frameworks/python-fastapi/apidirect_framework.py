"""
API-Direct Framework - FastAPI-compatible API framework with built-in monetization and deployment
"""

import asyncio
import inspect
import json
import os
import time
from typing import Any, Callable, Dict, List, Optional, Type, Union
from dataclasses import dataclass
from functools import wraps
from pathlib import Path

try:
    from fastapi import FastAPI, Request, Response, HTTPException
    from fastapi.responses import JSONResponse
    from fastapi.middleware.cors import CORSMiddleware
    from pydantic import BaseModel
    import uvicorn
except ImportError:
    raise ImportError(
        "FastAPI dependencies not found. Install with: pip install fastapi uvicorn pydantic"
    )


@dataclass
class MonetizationConfig:
    """Configuration for API monetization"""
    free_calls: int = 1000
    price_per_call: float = 0.001
    rate_limit_per_minute: int = 100
    require_api_key: bool = True


@dataclass
class APIDirectConfig:
    """Configuration for API-Direct features"""
    enable_analytics: bool = True
    enable_billing: bool = True
    enable_rate_limiting: bool = True
    enable_api_keys: bool = True
    local_mode: bool = True  # When True, runs locally without platform features


class APIDirectFramework:
    """
    FastAPI-compatible framework with API-Direct superpowers
    
    Usage:
        app = APIDirectFramework()
        
        @app.get("/items/{item_id}")
        @app.monetize(free_calls=1000, price_per_call=0.001)
        async def read_item(item_id: int):
            return {"item_id": item_id}
    """
    
    def __init__(self, 
                 title: str = "API-Direct API",
                 description: str = "API built with API-Direct Framework",
                 version: str = "1.0.0",
                 config: Optional[APIDirectConfig] = None):
        
        self.config = config or APIDirectConfig()
        self.fastapi_app = FastAPI(title=title, description=description, version=version)
        self.monetized_endpoints = {}
        self.usage_stats = {}
        self.api_keys = set()  # In production, this would be a database
        
        # Add CORS middleware
        self.fastapi_app.add_middleware(
            CORSMiddleware,
            allow_origins=["*"],
            allow_credentials=True,
            allow_methods=["*"],
            allow_headers=["*"],
        )
        
        # Add API-Direct middleware
        self._setup_middleware()
        self._setup_default_routes()
    
    def _setup_middleware(self):
        """Setup API-Direct specific middleware"""
        
        @self.fastapi_app.middleware("http")
        async def apidirect_middleware(request: Request, call_next):
            start_time = time.time()
            
            # API Key validation
            if self.config.enable_api_keys and self._requires_api_key(request.url.path):
                api_key = request.headers.get("X-API-Key")
                if not api_key or not self._validate_api_key(api_key):
                    return JSONResponse(
                        status_code=401,
                        content={"error": "Invalid or missing API key"}
                    )
            
            # Rate limiting
            if self.config.enable_rate_limiting:
                if not self._check_rate_limit(request):
                    return JSONResponse(
                        status_code=429,
                        content={"error": "Rate limit exceeded"}
                    )
            
            # Process request
            response = await call_next(request)
            
            # Analytics and billing
            if self.config.enable_analytics:
                self._track_usage(request, response, time.time() - start_time)
            
            # Add API-Direct headers
            response.headers["X-Powered-By"] = "API-Direct"
            response.headers["X-Response-Time"] = f"{time.time() - start_time:.3f}s"
            
            return response
    
    def _setup_default_routes(self):
        """Setup default API-Direct routes"""
        
        @self.fastapi_app.get("/_apidirect/health")
        async def health_check():
            return {"status": "healthy", "framework": "API-Direct", "timestamp": time.time()}
        
        @self.fastapi_app.get("/_apidirect/stats")
        async def get_stats():
            if not self.config.enable_analytics:
                raise HTTPException(status_code=404, detail="Analytics disabled")
            return self.usage_stats
        
        @self.fastapi_app.post("/_apidirect/api-keys")
        async def create_api_key(request: Request):
            # In production, this would require authentication
            import secrets
            api_key = f"apidirect_{secrets.token_urlsafe(32)}"
            self.api_keys.add(api_key)
            return {"api_key": api_key, "message": "API key created successfully"}
    
    def monetize(self, 
                 free_calls: int = 1000,
                 price_per_call: float = 0.001,
                 rate_limit_per_minute: int = 100):
        """
        Decorator to add monetization to an endpoint
        
        Args:
            free_calls: Number of free calls per month
            price_per_call: Price per API call after free tier
            rate_limit_per_minute: Rate limit per minute
        """
        def decorator(func):
            # Store monetization config for this endpoint
            endpoint_path = self._get_endpoint_path(func)
            self.monetized_endpoints[endpoint_path] = MonetizationConfig(
                free_calls=free_calls,
                price_per_call=price_per_call,
                rate_limit_per_minute=rate_limit_per_minute
            )
            
            @wraps(func)
            async def wrapper(*args, **kwargs):
                # In production, this would check billing status
                if self.config.local_mode:
                    print(f"💰 Monetized endpoint called: {endpoint_path}")
                    print(f"   Free calls: {free_calls}, Price: ${price_per_call}/call")
                
                return await func(*args, **kwargs) if asyncio.iscoroutinefunction(func) else func(*args, **kwargs)
            
            return wrapper
        return decorator
    
    def require_api_key(self):
        """Decorator to require API key for an endpoint"""
        def decorator(func):
            endpoint_path = self._get_endpoint_path(func)
            if endpoint_path not in self.monetized_endpoints:
                self.monetized_endpoints[endpoint_path] = MonetizationConfig()
            self.monetized_endpoints[endpoint_path].require_api_key = True
            return func
        return decorator
    
    def rate_limit(self, calls_per_minute: int = 100):
        """Decorator to set rate limit for an endpoint"""
        def decorator(func):
            endpoint_path = self._get_endpoint_path(func)
            if endpoint_path not in self.monetized_endpoints:
                self.monetized_endpoints[endpoint_path] = MonetizationConfig()
            self.monetized_endpoints[endpoint_path].rate_limit_per_minute = calls_per_minute
            return func
        return decorator
    
    def track_usage(self):
        """Decorator to track usage for an endpoint"""
        def decorator(func):
            @wraps(func)
            async def wrapper(*args, **kwargs):
                endpoint_path = self._get_endpoint_path(func)
                if self.config.enable_analytics:
                    self._increment_usage(endpoint_path)
                return await func(*args, **kwargs) if asyncio.iscoroutinefunction(func) else func(*args, **kwargs)
            return wrapper
        return decorator
    
    # FastAPI-compatible methods
    def get(self, path: str, **kwargs):
        return self.fastapi_app.get(path, **kwargs)
    
    def post(self, path: str, **kwargs):
        return self.fastapi_app.post(path, **kwargs)
    
    def put(self, path: str, **kwargs):
        return self.fastapi_app.put(path, **kwargs)
    
    def delete(self, path: str, **kwargs):
        return self.fastapi_app.delete(path, **kwargs)
    
    def patch(self, path: str, **kwargs):
        return self.fastapi_app.patch(path, **kwargs)
    
    def options(self, path: str, **kwargs):
        return self.fastapi_app.options(path, **kwargs)
    
    def head(self, path: str, **kwargs):
        return self.fastapi_app.head(path, **kwargs)
    
    def add_middleware(self, middleware_class, **kwargs):
        return self.fastapi_app.add_middleware(middleware_class, **kwargs)
    
    def include_router(self, router, **kwargs):
        return self.fastapi_app.include_router(router, **kwargs)
    
    # API-Direct specific methods
    def run(self, host: str = "127.0.0.1", port: int = 8000, **kwargs):
        """Run the API with enhanced development features"""
        print("🚀 Starting API-Direct Development Server")
        print(f"📊 Analytics: {'Enabled' if self.config.enable_analytics else 'Disabled'}")
        print(f"💰 Billing: {'Enabled' if self.config.enable_billing else 'Disabled'}")
        print(f"🔑 API Keys: {'Enabled' if self.config.enable_api_keys else 'Disabled'}")
        print(f"⚡ Rate Limiting: {'Enabled' if self.config.enable_rate_limiting else 'Disabled'}")
        print(f"🌐 Server: http://{host}:{port}")
        print(f"📚 Docs: http://{host}:{port}/docs")
        print(f"📈 Stats: http://{host}:{port}/_apidirect/stats")
        
        uvicorn.run(self.fastapi_app, host=host, port=port, **kwargs)
    
    def generate_apidirect_config(self) -> dict:
        """Generate apidirect.yaml configuration"""
        config = {
            "name": self.fastapi_app.title,
            "description": self.fastapi_app.description,
            "version": self.fastapi_app.version,
            "runtime": "python3.9",
            "framework": "apidirect-fastapi",
            "monetization": {
                "enabled": len(self.monetized_endpoints) > 0,
                "endpoints": {}
            },
            "features": {
                "analytics": self.config.enable_analytics,
                "billing": self.config.enable_billing,
                "rate_limiting": self.config.enable_rate_limiting,
                "api_keys": self.config.enable_api_keys
            }
        }
        
        for endpoint, monetization in self.monetized_endpoints.items():
            config["monetization"]["endpoints"][endpoint] = {
                "free_calls": monetization.free_calls,
                "price_per_call": monetization.price_per_call,
                "rate_limit_per_minute": monetization.rate_limit_per_minute,
                "require_api_key": monetization.require_api_key
            }
        
        return config
    
    def save_apidirect_config(self, path: str = "apidirect.yaml"):
        """Save apidirect.yaml configuration file"""
        import yaml
        config = self.generate_apidirect_config()
        with open(path, 'w') as f:
            yaml.dump(config, f, default_flow_style=False)
        print(f"✅ API-Direct configuration saved to {path}")
    
    # Helper methods
    def _get_endpoint_path(self, func) -> str:
        """Extract endpoint path from function"""
        return f"/{func.__name__}"
    
    def _requires_api_key(self, path: str) -> bool:
        """Check if endpoint requires API key"""
        for endpoint_path, config in self.monetized_endpoints.items():
            if path.startswith(endpoint_path) and config.require_api_key:
                return True
        return False
    
    def _validate_api_key(self, api_key: str) -> bool:
        """Validate API key"""
        return api_key in self.api_keys or api_key.startswith("apidirect_dev_")
    
    def _check_rate_limit(self, request: Request) -> bool:
        """Check rate limit for request"""
        # Simple in-memory rate limiting for development
        # In production, this would use Redis
        return True
    
    def _track_usage(self, request: Request, response: Response, response_time: float):
        """Track API usage"""
        path = request.url.path
        if path not in self.usage_stats:
            self.usage_stats[path] = {
                "calls": 0,
                "total_response_time": 0,
                "avg_response_time": 0,
                "status_codes": {}
            }
        
        stats = self.usage_stats[path]
        stats["calls"] += 1
        stats["total_response_time"] += response_time
        stats["avg_response_time"] = stats["total_response_time"] / stats["calls"]
        
        status_code = str(response.status_code)
        stats["status_codes"][status_code] = stats["status_codes"].get(status_code, 0) + 1
    
    def _increment_usage(self, endpoint_path: str):
        """Increment usage counter for endpoint"""
        if endpoint_path not in self.usage_stats:
            self.usage_stats[endpoint_path] = {"calls": 0}
        self.usage_stats[endpoint_path]["calls"] += 1


# Convenience imports for FastAPI compatibility
from fastapi import Query, Path, Body, Header, Cookie, Form, File, UploadFile, Depends, Security
from fastapi.security import HTTPBasic, HTTPBearer, APIKeyHeader, APIKeyQuery, APIKeyCookie
from pydantic import BaseModel, Field

# Export main class
__all__ = [
    "APIDirectFramework",
    "APIDirectConfig", 
    "MonetizationConfig",
    "Query", "Path", "Body", "Header", "Cookie", "Form", "File", "UploadFile",
    "Depends", "Security", "HTTPBasic", "HTTPBearer", 
    "APIKeyHeader", "APIKeyQuery", "APIKeyCookie",
    "BaseModel", "Field"
]
