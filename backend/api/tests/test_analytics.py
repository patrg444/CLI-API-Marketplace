"""
Test analytics functionality
"""

import pytest
from unittest.mock import Mock, AsyncMock, patch, MagicMock
from fastapi.testclient import TestClient
import sys
import os
import uuid
from decimal import Decimal
from datetime import datetime, timedelta

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

# Mock stripe before importing
mock_stripe = MagicMock()
sys.modules['stripe'] = mock_stripe

# Mock docker before importing
mock_docker = Mock()
sys.modules['docker'] = mock_docker

# Mock the database before importing main
with patch('asyncpg.create_pool', new_callable=AsyncMock) as mock_pool:
    with patch('redis.from_url') as mock_redis:
        mock_pool.return_value = AsyncMock()
        mock_redis.return_value = Mock()
        
        import main
        # Set the global db_pool variable
        main.db_pool = AsyncMock()
        main.payment_manager = Mock()
        main.deployment_manager = Mock()
        main.analytics_manager = Mock()
        
        from main import app, create_access_token
        from analytics import AnalyticsManager

client = TestClient(app)


class TestAnalyticsManager:
    """Test AnalyticsManager functionality"""
    
    @pytest.fixture
    def analytics_manager(self):
        mock_pool = AsyncMock()
        return AnalyticsManager(mock_pool)
    
    @pytest.fixture
    def auth_headers(self):
        """Create authentication headers"""
        user_id = str(uuid.uuid4())
        email = "test@example.com"
        access_token = create_access_token(user_id, email)
        return {"Authorization": f"Bearer {access_token}"}
    
    @pytest.mark.asyncio
    async def test_get_usage_by_consumer(self, analytics_manager):
        """Test getting usage analytics by consumer"""
        # Mock database response
        mock_conn = AsyncMock()
        mock_consumers = [
            {
                "consumer_id": uuid.uuid4(),
                "consumer_name": "Test Company",
                "consumer_company": "Test Corp",
                "total_calls": 1000,
                "active_days": 20,
                "avg_response_time": Decimal("150.5"),
                "error_count": 10,
                "revenue_generated": Decimal("100.00"),
                "last_call_at": datetime.utcnow()
            }
        ]
        mock_conn.fetch.return_value = mock_consumers
        mock_conn.fetchval.return_value = 5  # Total consumers
        
        # Mock acquire context manager
        mock_acquire = MagicMock()
        mock_acquire.__aenter__ = AsyncMock(return_value=mock_conn)
        mock_acquire.__aexit__ = AsyncMock(return_value=None)
        analytics_manager.db_pool.acquire = Mock(return_value=mock_acquire)
        
        # Test
        result = await analytics_manager.get_usage_by_consumer(
            user_id="test-user-id",
            period="30d"
        )
        
        assert result["period"] == "30d"
        assert result["total_consumers"] == 5
        assert len(result["consumers"]) == 1
        assert result["consumers"][0]["consumer_name"] == "Test Company"
        assert result["consumers"][0]["total_calls"] == 1000
        assert result["consumers"][0]["error_rate"] == 1.0  # 10/1000 * 100
    
    @pytest.mark.asyncio
    async def test_get_geographic_analytics(self, analytics_manager):
        """Test getting geographic analytics"""
        # Mock database response
        mock_conn = AsyncMock()
        mock_countries = [
            {
                "country": "US",
                "call_count": 5000,
                "unique_consumers": 50,
                "avg_response_time": Decimal("120.0"),
                "revenue": Decimal("500.00")
            },
            {
                "country": "UK",
                "call_count": 2000,
                "unique_consumers": 20,
                "avg_response_time": Decimal("150.0"),
                "revenue": Decimal("200.00")
            }
        ]
        mock_cities = [
            {"city": "New York", "country": "US", "call_count": 1500},
            {"city": "London", "country": "UK", "call_count": 1000}
        ]
        
        mock_conn.fetch.side_effect = [mock_countries, mock_cities]
        
        # Mock acquire context manager
        mock_acquire = MagicMock()
        mock_acquire.__aenter__ = AsyncMock(return_value=mock_conn)
        mock_acquire.__aexit__ = AsyncMock(return_value=None)
        analytics_manager.db_pool.acquire = Mock(return_value=mock_acquire)
        
        # Test
        result = await analytics_manager.get_geographic_analytics(
            user_id="test-user-id",
            period="30d"
        )
        
        assert result["period"] == "30d"
        assert len(result["countries"]) == 2
        assert result["countries"][0]["country"] == "US"
        assert result["countries"][0]["call_count"] == 5000
        assert len(result["top_cities"]) == 2
        assert result["top_cities"][0]["city"] == "New York"
    
    @pytest.mark.asyncio
    async def test_get_error_analytics(self, analytics_manager):
        """Test getting error analytics"""
        # Mock database response
        mock_conn = AsyncMock()
        mock_errors = [
            {
                "status_code": 404,
                "count": 100,
                "path": "/api/v1/users",
                "method": "GET"
            },
            {
                "status_code": 500,
                "count": 50,
                "path": "/api/v1/process",
                "method": "POST"
            }
        ]
        mock_trends = [
            {
                "time_bucket": datetime.utcnow() - timedelta(hours=2),
                "client_errors": 20,
                "server_errors": 5,
                "total_calls": 1000
            }
        ]
        mock_messages = [
            {"error_message": "User not found", "count": 80},
            {"error_message": "Internal server error", "count": 50}
        ]
        
        mock_conn.fetch.side_effect = [mock_errors, mock_trends, mock_messages]
        
        # Mock acquire context manager
        mock_acquire = MagicMock()
        mock_acquire.__aenter__ = AsyncMock(return_value=mock_conn)
        mock_acquire.__aexit__ = AsyncMock(return_value=None)
        analytics_manager.db_pool.acquire = Mock(return_value=mock_acquire)
        
        # Test
        result = await analytics_manager.get_error_analytics(
            user_id="test-user-id",
            period="7d"
        )
        
        assert result["period"] == "7d"
        assert len(result["error_details"]) == 2
        assert result["error_details"][0]["status_code"] == 404
        assert result["error_details"][0]["error_type"] == "Not Found"
        assert len(result["error_trends"]) == 1
        assert result["error_trends"][0]["error_rate"] == 2.5  # (20+5)/1000 * 100
        assert len(result["common_errors"]) == 2


class TestAnalyticsAPI:
    """Test analytics API endpoints"""
    
    def setup_method(self):
        """Setup test user"""
        self.test_user_id = str(uuid.uuid4())
        self.test_email = "test@example.com"
        self.access_token = create_access_token(self.test_user_id, self.test_email)
        self.auth_headers = {"Authorization": f"Bearer {self.access_token}"}
        
        # Mock analytics manager
        main.analytics_manager = Mock()
        
        # Mock database for authentication
        mock_conn = AsyncMock()
        mock_conn.fetchrow.return_value = {
            "id": self.test_user_id,
            "email": self.test_email,
            "name": "Test User",
            "company": "Test Corp",
            "profile_image": None,
            "created_at": datetime.utcnow()
        }
        
        # Mock acquire context manager
        mock_acquire = MagicMock()
        mock_acquire.__aenter__ = AsyncMock(return_value=mock_conn)
        mock_acquire.__aexit__ = AsyncMock(return_value=None)
        main.db_pool.acquire = Mock(return_value=mock_acquire)
    
    def test_get_consumer_analytics(self):
        """Test consumer analytics endpoint"""
        # Mock analytics manager response
        main.analytics_manager.get_usage_by_consumer = AsyncMock(return_value={
            "period": "30d",
            "total_consumers": 10,
            "consumers": [
                {
                    "consumer_id": "consumer-123",
                    "consumer_name": "Test Company",
                    "total_calls": 1000,
                    "error_rate": 1.5,
                    "revenue_generated": 100.0
                }
            ]
        })
        
        response = client.get(
            "/api/analytics/consumers?period=30d",
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["total_consumers"] == 10
        assert len(data["consumers"]) == 1
        assert data["consumers"][0]["consumer_name"] == "Test Company"
    
    def test_get_geographic_analytics(self):
        """Test geographic analytics endpoint"""
        # Mock analytics manager response
        main.analytics_manager.get_geographic_analytics = AsyncMock(return_value={
            "period": "30d",
            "countries": [
                {
                    "country": "US",
                    "call_count": 5000,
                    "unique_consumers": 50,
                    "revenue": 500.0
                }
            ],
            "top_cities": [
                {
                    "city": "New York",
                    "country": "US",
                    "call_count": 1500
                }
            ]
        })
        
        response = client.get(
            "/api/analytics/geographic",
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        assert len(data["countries"]) == 1
        assert data["countries"][0]["country"] == "US"
        assert len(data["top_cities"]) == 1
    
    def test_get_error_analytics(self):
        """Test error analytics endpoint"""
        # Mock analytics manager response
        main.analytics_manager.get_error_analytics = AsyncMock(return_value={
            "period": "7d",
            "error_details": [
                {
                    "status_code": 404,
                    "count": 100,
                    "endpoint": "GET /api/v1/users",
                    "error_type": "Not Found"
                }
            ],
            "error_trends": [
                {
                    "timestamp": datetime.utcnow().isoformat(),
                    "client_errors": 20,
                    "server_errors": 5,
                    "error_rate": 2.5
                }
            ],
            "common_errors": []
        })
        
        response = client.get(
            "/api/analytics/errors?period=7d",
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        assert len(data["error_details"]) == 1
        assert data["error_details"][0]["status_code"] == 404
    
    def test_get_revenue_analytics(self):
        """Test revenue analytics endpoint"""
        # Mock analytics manager response
        main.analytics_manager.get_revenue_by_api = AsyncMock(return_value={
            "period": "30d",
            "apis": [
                {
                    "api_id": "api-123",
                    "api_name": "Test API",
                    "revenue": 1000.0,
                    "total_calls": 10000,
                    "unique_consumers": 50
                }
            ],
            "revenue_trends": [
                {
                    "date": datetime.utcnow().isoformat(),
                    "total_revenue": 100.0,
                    "apis": {
                        "Test API": {
                            "revenue": 100.0,
                            "calls": 1000
                        }
                    }
                }
            ]
        })
        
        response = client.get(
            "/api/analytics/revenue?period=30d",
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        assert len(data["apis"]) == 1
        assert data["apis"][0]["api_name"] == "Test API"
        assert data["apis"][0]["revenue"] == 1000.0
    
    def test_get_endpoint_analytics(self):
        """Test endpoint analytics endpoint"""
        api_id = str(uuid.uuid4())
        
        # Mock analytics manager response
        main.analytics_manager.get_endpoint_analytics = AsyncMock(return_value={
            "period": "7d",
            "api_id": api_id,
            "endpoints": [
                {
                    "method": "GET",
                    "path": "/api/v1/users",
                    "call_count": 5000,
                    "performance": {
                        "avg_response_time": 120.0,
                        "p95_response_time": 250.0
                    },
                    "success_rate": 98.5,
                    "status_codes": {
                        "200": 4925,
                        "404": 50,
                        "500": 25
                    }
                }
            ]
        })
        
        response = client.get(
            f"/api/analytics/endpoints/{api_id}?period=7d",
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["api_id"] == api_id
        assert len(data["endpoints"]) == 1
        assert data["endpoints"][0]["method"] == "GET"
        assert data["endpoints"][0]["success_rate"] == 98.5
    
    def test_get_usage_patterns(self):
        """Test usage patterns endpoint"""
        # Mock analytics manager response
        main.analytics_manager.get_usage_patterns = AsyncMock(return_value={
            "hourly_pattern": [
                {"hour": 9, "call_count": 1000, "avg_response_time": 120.0},
                {"hour": 10, "call_count": 1500, "avg_response_time": 130.0}
            ],
            "daily_pattern": [
                {"day_of_week": 1, "day_name": "Monday", "call_count": 5000, "avg_response_time": 125.0}
            ],
            "peak_hours": [
                {"timestamp": datetime.utcnow().isoformat(), "call_count": 2000}
            ]
        })
        
        response = client.get(
            "/api/analytics/patterns",
            headers=self.auth_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        assert len(data["hourly_pattern"]) == 2
        assert data["hourly_pattern"][0]["hour"] == 9
        assert len(data["daily_pattern"]) == 1
        assert data["daily_pattern"][0]["day_name"] == "Monday"
    
    def test_analytics_without_auth(self):
        """Test analytics endpoints require authentication"""
        response = client.get("/api/analytics/consumers")
        assert response.status_code == 403
        
        response = client.get("/api/analytics/geographic")
        assert response.status_code == 403
        
        response = client.get("/api/analytics/errors")
        assert response.status_code == 403
    
    def test_analytics_service_unavailable(self):
        """Test when analytics service is unavailable"""
        main.analytics_manager = None
        
        response = client.get(
            "/api/analytics/consumers",
            headers=self.auth_headers
        )
        
        assert response.status_code == 503
        assert response.json()["detail"] == "Analytics service unavailable"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])