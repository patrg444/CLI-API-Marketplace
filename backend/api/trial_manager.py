"""
API Trial and Sandbox Management
Handles free trials, test mode, and mock responses for APIs
"""

import asyncpg
import logging
from datetime import datetime, timedelta
from typing import Dict, Any, Optional, List
from fastapi import HTTPException
import re
import json
import httpx
from uuid import UUID

logger = logging.getLogger(__name__)


class TrialManager:
    """Manages API trials and sandbox testing"""
    
    def __init__(self, db_pool: asyncpg.Pool):
        self.db_pool = db_pool
    
    async def start_trial(self, user_id: str, api_id: str) -> Dict[str, Any]:
        """
        Start a free trial for a user on an API
        
        Returns:
            Trial details including limits and expiration
        """
        async with self.db_pool.acquire() as conn:
            # Check if API has trial enabled
            api = await conn.fetchrow("""
                SELECT trial_enabled, trial_requests, trial_duration_days, 
                       trial_rate_limit, name
                FROM apis 
                WHERE id = $1
            """, api_id)
            
            if not api or not api['trial_enabled']:
                raise HTTPException(
                    status_code=400,
                    detail="This API does not offer a free trial"
                )
            
            # Check if user already has/had a trial for this API
            existing_trial = await conn.fetchrow("""
                SELECT id, status FROM api_trials
                WHERE user_id = $1 AND api_id = $2
            """, user_id, api_id)
            
            if existing_trial:
                if existing_trial['status'] == 'active':
                    raise HTTPException(
                        status_code=400,
                        detail="You already have an active trial for this API"
                    )
                else:
                    raise HTTPException(
                        status_code=400,
                        detail="You have already used your free trial for this API"
                    )
            
            # Calculate expiration
            expires_at = None
            if api['trial_duration_days']:
                expires_at = datetime.utcnow() + timedelta(days=api['trial_duration_days'])
            
            # Create trial
            trial_id = await conn.fetchval("""
                INSERT INTO api_trials (
                    user_id, api_id, requests_limit, expires_at
                ) VALUES ($1, $2, $3, $4)
                RETURNING id
            """, user_id, api_id, api['trial_requests'], expires_at)
            
            logger.info(f"Started trial {trial_id} for user {user_id} on API {api_id}")
            
            return {
                "trial_id": str(trial_id),
                "api_id": api_id,
                "api_name": api['name'],
                "status": "active",
                "requests_limit": api['trial_requests'],
                "requests_used": 0,
                "rate_limit": api['trial_rate_limit'],
                "expires_at": expires_at.isoformat() if expires_at else None,
                "started_at": datetime.utcnow().isoformat()
            }
    
    async def check_trial_access(self, user_id: str, api_id: str) -> Optional[Dict[str, Any]]:
        """
        Check if user has valid trial access to an API
        
        Returns:
            Trial details if valid, None otherwise
        """
        async with self.db_pool.acquire() as conn:
            trial = await conn.fetchrow("""
                SELECT t.*, a.trial_rate_limit
                FROM api_trials t
                JOIN apis a ON t.api_id = a.id
                WHERE t.user_id = $1 AND t.api_id = $2 AND t.status = 'active'
            """, user_id, api_id)
            
            if not trial:
                return None
            
            # Check if expired
            if trial['expires_at'] and trial['expires_at'] < datetime.utcnow():
                await conn.execute("""
                    UPDATE api_trials 
                    SET status = 'expired'
                    WHERE id = $1
                """, trial['id'])
                return None
            
            # Check if requests exhausted
            if trial['requests_limit'] and trial['requests_used'] >= trial['requests_limit']:
                await conn.execute("""
                    UPDATE api_trials 
                    SET status = 'expired'
                    WHERE id = $1
                """, trial['id'])
                return None
            
            return {
                "trial_id": str(trial['id']),
                "requests_remaining": (trial['requests_limit'] - trial['requests_used']) if trial['requests_limit'] else None,
                "rate_limit": trial['trial_rate_limit'],
                "expires_at": trial['expires_at'].isoformat() if trial['expires_at'] else None
            }
    
    async def record_trial_usage(self, user_id: str, api_id: str) -> bool:
        """
        Record usage of trial request
        
        Returns:
            True if recorded, False if trial invalid
        """
        async with self.db_pool.acquire() as conn:
            result = await conn.execute("""
                UPDATE api_trials 
                SET requests_used = requests_used + 1
                WHERE user_id = $1 AND api_id = $2 AND status = 'active'
                AND (expires_at IS NULL OR expires_at > NOW())
                AND (requests_limit IS NULL OR requests_used < requests_limit)
            """, user_id, api_id)
            
            return result.split()[-1] == '1'
    
    async def convert_trial(self, user_id: str, api_id: str) -> Dict[str, Any]:
        """
        Convert trial to paid subscription
        """
        async with self.db_pool.acquire() as conn:
            result = await conn.execute("""
                UPDATE api_trials 
                SET status = 'converted',
                    converted_at = NOW()
                WHERE user_id = $1 AND api_id = $2 AND status = 'active'
            """, user_id, api_id)
            
            if result.split()[-1] != '1':
                raise HTTPException(
                    status_code=400,
                    detail="No active trial found to convert"
                )
            
            return {"message": "Trial converted successfully"}
    
    async def create_sandbox_request(
        self, 
        api_id: str, 
        user_id: str,
        endpoint: str,
        method: str,
        request_data: Dict[str, Any],
        use_mock: bool = True
    ) -> Dict[str, Any]:
        """
        Handle sandbox/test mode request
        """
        async with self.db_pool.acquire() as conn:
            # Check if API has sandbox enabled
            api = await conn.fetchrow("""
                SELECT sandbox_enabled, sandbox_base_url, base_url
                FROM apis WHERE id = $1
            """, api_id)
            
            if not api or not api['sandbox_enabled']:
                raise HTTPException(
                    status_code=400,
                    detail="Sandbox mode not available for this API"
                )
            
            # Try to find mock response first
            if use_mock:
                mock_response = await self._find_mock_response(
                    conn, api_id, endpoint, method
                )
                
                if mock_response:
                    # Record sandbox request
                    await conn.execute("""
                        INSERT INTO sandbox_requests (
                            api_id, user_id, endpoint, method,
                            request_headers, request_body,
                            response_status, response_headers, response_body,
                            response_time_ms, is_mocked
                        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
                    """, api_id, user_id, endpoint, method,
                        json.dumps(request_data.get('headers', {})),
                        json.dumps(request_data.get('body', {})),
                        mock_response['response_status'],
                        mock_response['response_headers'],
                        mock_response['response_body'],
                        10,  # Mock response time
                        True
                    )
                    
                    return {
                        "status": mock_response['response_status'],
                        "headers": mock_response['response_headers'] or {},
                        "data": mock_response['response_body'],
                        "is_sandbox": True,
                        "is_mocked": True
                    }
            
            # Make real request to sandbox URL if available
            if api['sandbox_base_url']:
                response = await self._make_sandbox_request(
                    api['sandbox_base_url'],
                    endpoint,
                    method,
                    request_data
                )
                
                # Record sandbox request
                await conn.execute("""
                    INSERT INTO sandbox_requests (
                        api_id, user_id, endpoint, method,
                        request_headers, request_body,
                        response_status, response_headers, response_body,
                        response_time_ms, is_mocked
                    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
                """, api_id, user_id, endpoint, method,
                    json.dumps(request_data.get('headers', {})),
                    json.dumps(request_data.get('body', {})),
                    response['status'],
                    json.dumps(response['headers']),
                    json.dumps(response['data']),
                    response['response_time_ms'],
                    False
                )
                
                return response
            
            raise HTTPException(
                status_code=503,
                detail="No sandbox endpoint or mock data available"
            )
    
    async def _find_mock_response(
        self, 
        conn: asyncpg.Connection, 
        api_id: str, 
        endpoint: str, 
        method: str
    ) -> Optional[Dict[str, Any]]:
        """Find matching mock response for endpoint"""
        mocks = await conn.fetch("""
            SELECT * FROM api_mock_responses
            WHERE api_id = $1 AND method = $2 AND is_active = true
            ORDER BY created_at DESC
        """, api_id, method.upper())
        
        for mock in mocks:
            if re.match(mock['endpoint_pattern'], endpoint):
                return mock
        
        return None
    
    async def _make_sandbox_request(
        self,
        base_url: str,
        endpoint: str,
        method: str,
        request_data: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Make actual HTTP request to sandbox endpoint"""
        import time
        start_time = time.time()
        
        async with httpx.AsyncClient() as client:
            try:
                response = await client.request(
                    method=method,
                    url=f"{base_url}{endpoint}",
                    headers=request_data.get('headers', {}),
                    json=request_data.get('body') if method in ['POST', 'PUT', 'PATCH'] else None,
                    params=request_data.get('params') if method == 'GET' else None,
                    timeout=30.0
                )
                
                response_time_ms = int((time.time() - start_time) * 1000)
                
                return {
                    "status": response.status_code,
                    "headers": dict(response.headers),
                    "data": response.json() if response.headers.get('content-type', '').startswith('application/json') else response.text,
                    "is_sandbox": True,
                    "is_mocked": False,
                    "response_time_ms": response_time_ms
                }
            except Exception as e:
                logger.error(f"Sandbox request failed: {e}")
                raise HTTPException(
                    status_code=503,
                    detail=f"Sandbox request failed: {str(e)}"
                )
    
    async def create_mock_response(
        self,
        api_id: str,
        endpoint_pattern: str,
        method: str,
        response_data: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Create a mock response for API sandbox"""
        async with self.db_pool.acquire() as conn:
            mock_id = await conn.fetchval("""
                INSERT INTO api_mock_responses (
                    api_id, endpoint_pattern, method,
                    response_status, response_headers, response_body,
                    description
                ) VALUES ($1, $2, $3, $4, $5, $6, $7)
                RETURNING id
            """, api_id, endpoint_pattern, method.upper(),
                response_data.get('status', 200),
                response_data.get('headers', {}),
                response_data.get('body', {}),
                response_data.get('description')
            )
            
            return {
                "mock_id": str(mock_id),
                "message": "Mock response created successfully"
            }
    
    async def get_trial_analytics(self, api_id: str) -> Dict[str, Any]:
        """Get analytics for API trials"""
        async with self.db_pool.acquire() as conn:
            # Trial statistics
            stats = await conn.fetchrow("""
                SELECT 
                    COUNT(*) as total_trials,
                    COUNT(*) FILTER (WHERE status = 'active') as active_trials,
                    COUNT(*) FILTER (WHERE status = 'converted') as converted_trials,
                    COUNT(*) FILTER (WHERE status = 'expired') as expired_trials,
                    AVG(requests_used)::float as avg_requests_used,
                    (COUNT(*) FILTER (WHERE status = 'converted')::float / 
                     NULLIF(COUNT(*), 0) * 100)::float as conversion_rate
                FROM api_trials
                WHERE api_id = $1
            """, api_id)
            
            # Sandbox usage
            sandbox_stats = await conn.fetchrow("""
                SELECT 
                    COUNT(*) as total_sandbox_requests,
                    COUNT(*) FILTER (WHERE is_mocked = true) as mocked_requests,
                    COUNT(*) FILTER (WHERE is_mocked = false) as real_requests,
                    AVG(response_time_ms)::float as avg_response_time
                FROM sandbox_requests
                WHERE api_id = $1 AND created_at > NOW() - INTERVAL '30 days'
            """, api_id)
            
            return {
                "trial_stats": dict(stats) if stats else {},
                "sandbox_stats": dict(sandbox_stats) if sandbox_stats else {}
            }