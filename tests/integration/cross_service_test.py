"""
Cross-service integration tests for API-Direct marketplace
Tests service-to-service communication, data flow, and end-to-end scenarios
"""

import pytest
import asyncio
import aiohttp
import websockets
import json
import time
from datetime import datetime, timedelta
from unittest.mock import patch, AsyncMock
import redis.asyncio as redis
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker

from tests.fixtures import (
    test_user, test_api, test_subscription,
    auth_headers, api_client, db_session
)


class TestUserRegistrationFlow:
    """Test complete user registration and onboarding flow"""
    
    @pytest.mark.asyncio
    async def test_complete_user_registration(self, api_client):
        """Test user registration across all services"""
        # 1. Register user via Auth Service
        registration_data = {
            "email": "newuser@example.com",
            "password": "SecurePass123!",
            "username": "newuser",
            "company": "Test Corp"
        }
        
        async with api_client.post("/auth/register", json=registration_data) as resp:
            assert resp.status == 201
            user_data = await resp.json()
            user_id = user_data["user_id"]
            
        # 2. Verify user created in Database
        async with api_client.get(f"/users/{user_id}") as resp:
            assert resp.status == 200
            user = await resp.json()
            assert user["email"] == registration_data["email"]
            
        # 3. Check welcome email sent via Notification Service
        async with api_client.get(f"/notifications/users/{user_id}/recent") as resp:
            assert resp.status == 200
            notifications = await resp.json()
            assert any(n["type"] == "welcome_email" for n in notifications)
            
        # 4. Verify free tier subscription created via Billing Service
        async with api_client.get(f"/billing/users/{user_id}/subscription") as resp:
            assert resp.status == 200
            subscription = await resp.json()
            assert subscription["plan"] == "free"
            assert subscription["status"] == "active"
            
        # 5. Check initial API quota set via Metering Service
        async with api_client.get(f"/metering/users/{user_id}/quota") as resp:
            assert resp.status == 200
            quota = await resp.json()
            assert quota["daily_limit"] == 1000
            assert quota["used_today"] == 0
    
    @pytest.mark.asyncio
    async def test_user_verification_flow(self, api_client):
        """Test email verification across services"""
        # Register user
        registration_data = {
            "email": "verify@example.com",
            "password": "SecurePass123!",
            "username": "verifyuser"
        }
        
        async with api_client.post("/auth/register", json=registration_data) as resp:
            user_data = await resp.json()
            user_id = user_data["user_id"]
        
        # Get verification token from notification service
        async with api_client.get(f"/notifications/users/{user_id}/emails") as resp:
            emails = await resp.json()
            verification_email = next(e for e in emails if e["type"] == "verification")
            token = verification_email["data"]["token"]
        
        # Verify email
        async with api_client.post(f"/auth/verify-email", json={"token": token}) as resp:
            assert resp.status == 200
        
        # Check user status updated
        async with api_client.get(f"/users/{user_id}") as resp:
            user = await resp.json()
            assert user["email_verified"] is True
        
        # Verify additional features unlocked
        async with api_client.get(f"/metering/users/{user_id}/quota") as resp:
            quota = await resp.json()
            assert quota["daily_limit"] == 5000  # Increased for verified users


class TestAPIPublishingFlow:
    """Test API publishing and marketplace listing flow"""
    
    @pytest.mark.asyncio
    async def test_publish_new_api(self, api_client, auth_headers):
        """Test publishing a new API across all services"""
        # 1. Create API via API Service
        api_data = {
            "name": "Weather API v2",
            "description": "Advanced weather data API",
            "base_url": "https://api.weather.example.com",
            "documentation_url": "https://docs.weather.example.com",
            "categories": ["weather", "data"],
            "pricing_model": "freemium"
        }
        
        async with api_client.post("/apis", json=api_data, headers=auth_headers) as resp:
            assert resp.status == 201
            api_info = await resp.json()
            api_id = api_info["id"]
        
        # 2. Upload OpenAPI spec
        openapi_spec = {
            "openapi": "3.0.0",
            "info": {"title": "Weather API", "version": "2.0"},
            "paths": {
                "/current": {"get": {"summary": "Get current weather"}},
                "/forecast": {"get": {"summary": "Get weather forecast"}}
            }
        }
        
        async with api_client.put(
            f"/apis/{api_id}/openapi",
            json=openapi_spec,
            headers=auth_headers
        ) as resp:
            assert resp.status == 200
        
        # 3. Configure pricing tiers via Billing Service
        pricing_config = {
            "tiers": [
                {"name": "free", "monthly_calls": 1000, "price": 0},
                {"name": "starter", "monthly_calls": 10000, "price": 29},
                {"name": "pro", "monthly_calls": 100000, "price": 99}
            ]
        }
        
        async with api_client.post(
            f"/billing/apis/{api_id}/pricing",
            json=pricing_config,
            headers=auth_headers
        ) as resp:
            assert resp.status == 200
        
        # 4. Submit for marketplace review
        async with api_client.post(
            f"/apis/{api_id}/submit-review",
            headers=auth_headers
        ) as resp:
            assert resp.status == 200
            review_data = await resp.json()
            assert review_data["status"] == "pending_review"
        
        # 5. Simulate approval process
        admin_headers = {"Authorization": "Bearer admin-token"}
        async with api_client.post(
            f"/admin/apis/{api_id}/approve",
            headers=admin_headers
        ) as resp:
            assert resp.status == 200
        
        # 6. Verify API is live in marketplace
        async with api_client.get(f"/marketplace/apis/{api_id}") as resp:
            assert resp.status == 200
            marketplace_listing = await resp.json()
            assert marketplace_listing["status"] == "active"
            assert marketplace_listing["visibility"] == "public"
        
        # 7. Check metrics tracking initialized
        async with api_client.get(
            f"/analytics/apis/{api_id}/metrics",
            headers=auth_headers
        ) as resp:
            assert resp.status == 200
            metrics = await resp.json()
            assert metrics["total_subscribers"] == 0
            assert metrics["total_calls"] == 0


class TestSubscriptionFlow:
    """Test API subscription and usage flow"""
    
    @pytest.mark.asyncio
    async def test_subscribe_and_use_api(self, api_client, auth_headers):
        """Test subscribing to an API and using it"""
        # 1. Browse marketplace and find API
        async with api_client.get("/marketplace/apis?category=weather") as resp:
            apis = await resp.json()
            selected_api = apis["results"][0]
            api_id = selected_api["id"]
        
        # 2. Subscribe to API
        subscription_data = {
            "api_id": api_id,
            "plan": "starter",
            "payment_method_id": "pm_test_visa"
        }
        
        async with api_client.post(
            "/subscriptions",
            json=subscription_data,
            headers=auth_headers
        ) as resp:
            assert resp.status == 201
            subscription = await resp.json()
            subscription_id = subscription["id"]
            api_key = subscription["api_key"]
        
        # 3. Verify subscription active in billing
        async with api_client.get(
            f"/billing/subscriptions/{subscription_id}",
            headers=auth_headers
        ) as resp:
            billing_info = await resp.json()
            assert billing_info["status"] == "active"
            assert billing_info["current_period_end"] > time.time()
        
        # 4. Make API calls through gateway
        gateway_headers = {"X-API-Key": api_key}
        async with api_client.get(
            f"/gateway/{api_id}/current?city=London",
            headers=gateway_headers
        ) as resp:
            assert resp.status == 200
            weather_data = await resp.json()
            assert "temperature" in weather_data
        
        # 5. Verify usage tracked
        await asyncio.sleep(1)  # Wait for async processing
        
        async with api_client.get(
            f"/metering/subscriptions/{subscription_id}/usage",
            headers=auth_headers
        ) as resp:
            usage = await resp.json()
            assert usage["total_calls"] == 1
            assert usage["remaining_quota"] == 9999
        
        # 6. Check analytics updated
        async with api_client.get(
            f"/analytics/subscriptions/{subscription_id}/stats",
            headers=auth_headers
        ) as resp:
            stats = await resp.json()
            assert stats["calls_today"] == 1
            assert stats["average_latency"] > 0


class TestRateLimitingAcrossServices:
    """Test rate limiting coordination across services"""
    
    @pytest.mark.asyncio
    async def test_rate_limit_enforcement(self, api_client, auth_headers):
        """Test rate limits are enforced across all services"""
        # Get free tier API key with low rate limit
        api_key = "test_free_tier_key"
        
        # Make requests up to rate limit
        successful_calls = 0
        rate_limited = False
        
        for i in range(15):  # Free tier allows 10 calls/minute
            async with api_client.get(
                "/gateway/weather-api/current",
                headers={"X-API-Key": api_key}
            ) as resp:
                if resp.status == 200:
                    successful_calls += 1
                elif resp.status == 429:
                    rate_limited = True
                    break
        
        assert successful_calls == 10
        assert rate_limited is True
        
        # Verify rate limit info in response headers
        async with api_client.get(
            "/gateway/weather-api/current",
            headers={"X-API-Key": api_key}
        ) as resp:
            assert resp.status == 429
            assert "X-RateLimit-Limit" in resp.headers
            assert "X-RateLimit-Remaining" in resp.headers
            assert resp.headers["X-RateLimit-Remaining"] == "0"
            assert "X-RateLimit-Reset" in resp.headers
    
    @pytest.mark.asyncio
    async def test_rate_limit_across_different_apis(self, api_client):
        """Test rate limits are tracked per API"""
        api_key = "test_multi_api_key"
        
        # Make calls to first API
        for i in range(5):
            async with api_client.get(
                "/gateway/weather-api/current",
                headers={"X-API-Key": api_key}
            ) as resp:
                assert resp.status == 200
        
        # Make calls to second API (should not be rate limited)
        for i in range(5):
            async with api_client.get(
                "/gateway/geocoding-api/search",
                headers={"X-API-Key": api_key}
            ) as resp:
                assert resp.status == 200
        
        # Verify separate quotas via metering service
        async with api_client.get(
            f"/metering/keys/{api_key}/usage-by-api"
        ) as resp:
            usage = await resp.json()
            assert usage["weather-api"]["calls"] == 5
            assert usage["geocoding-api"]["calls"] == 5


class TestWebSocketIntegration:
    """Test WebSocket connections across services"""
    
    @pytest.mark.asyncio
    async def test_realtime_notifications(self, api_client, auth_headers):
        """Test real-time notifications via WebSocket"""
        # Extract token from headers
        token = auth_headers["Authorization"].split(" ")[1]
        
        # Connect to WebSocket
        async with websockets.connect(
            f"ws://localhost:8000/ws?token={token}"
        ) as websocket:
            # Subscribe to API updates
            await websocket.send(json.dumps({
                "type": "subscribe",
                "api_id": "weather-api"
            }))
            
            # Trigger an event that should notify via WebSocket
            async with api_client.post(
                "/apis/weather-api/incidents",
                json={"type": "maintenance", "message": "Scheduled maintenance"},
                headers={"Authorization": "Bearer api-owner-token"}
            ) as resp:
                assert resp.status == 201
            
            # Receive notification via WebSocket
            message = await asyncio.wait_for(websocket.recv(), timeout=5.0)
            notification = json.loads(message)
            
            assert notification["type"] == "api_update"
            assert notification["api_id"] == "weather-api"
            assert notification["incident"]["type"] == "maintenance"
    
    @pytest.mark.asyncio
    async def test_collaborative_features(self, api_client):
        """Test team collaboration via WebSocket"""
        # Create team workspace
        team_data = {
            "name": "API Dev Team",
            "members": ["user1@example.com", "user2@example.com"]
        }
        
        async with api_client.post("/teams", json=team_data) as resp:
            team = await resp.json()
            team_id = team["id"]
        
        # Connect both users to WebSocket
        ws1 = await websockets.connect(f"ws://localhost:8000/ws?token=user1_token")
        ws2 = await websockets.connect(f"ws://localhost:8000/ws?token=user2_token")
        
        try:
            # User 1 makes changes
            await ws1.send(json.dumps({
                "type": "api_update",
                "team_id": team_id,
                "changes": {"description": "Updated description"}
            }))
            
            # User 2 should receive the update
            message = await asyncio.wait_for(ws2.recv(), timeout=5.0)
            update = json.loads(message)
            
            assert update["type"] == "team_update"
            assert update["team_id"] == team_id
            assert update["changes"]["description"] == "Updated description"
            
        finally:
            await ws1.close()
            await ws2.close()


class TestFailoverAndResilience:
    """Test service resilience and failover scenarios"""
    
    @pytest.mark.asyncio
    async def test_service_degradation(self, api_client, auth_headers):
        """Test graceful degradation when services fail"""
        # Simulate database service being slow
        with patch('services.database.query', new_callable=AsyncMock) as mock_query:
            mock_query.side_effect = asyncio.TimeoutError()
            
            # API calls should still work with cached data
            async with api_client.get(
                "/apis/popular",
                headers=auth_headers
            ) as resp:
                assert resp.status == 200
                data = await resp.json()
                assert data["source"] == "cache"
                assert len(data["results"]) > 0
    
    @pytest.mark.asyncio
    async def test_circuit_breaker(self, api_client):
        """Test circuit breaker pattern across services"""
        # Make failing requests to trigger circuit breaker
        for i in range(10):
            async with api_client.get("/unstable-endpoint") as resp:
                pass  # Ignore errors
        
        # Circuit should be open now
        async with api_client.get("/unstable-endpoint") as resp:
            assert resp.status == 503
            error = await resp.json()
            assert error["error"] == "Circuit breaker is open"
            assert "retry_after" in error
    
    @pytest.mark.asyncio
    async def test_message_queue_resilience(self, api_client, auth_headers):
        """Test message queue handles service failures"""
        # Send analytics event
        event_data = {
            "api_id": "weather-api",
            "event_type": "api_call",
            "timestamp": datetime.utcnow().isoformat(),
            "metadata": {"endpoint": "/current", "status": 200}
        }
        
        # Simulate analytics service being down
        with patch('services.analytics.process_event', side_effect=Exception("Service down")):
            async with api_client.post(
                "/events/analytics",
                json=event_data,
                headers=auth_headers
            ) as resp:
                assert resp.status == 202  # Accepted for processing
        
        # Event should be queued
        redis_client = redis.from_url("redis://localhost:6379")
        queue_length = await redis_client.llen("analytics_events_queue")
        assert queue_length > 0
        
        # Verify event is retried when service recovers
        await asyncio.sleep(5)  # Wait for retry
        queue_length = await redis_client.llen("analytics_events_dlq")
        assert queue_length == 0  # Should not be in dead letter queue yet


class TestDataConsistency:
    """Test data consistency across services"""
    
    @pytest.mark.asyncio
    async def test_distributed_transaction(self, api_client, auth_headers):
        """Test distributed transaction across multiple services"""
        # Create a complex operation that touches multiple services
        operation_data = {
            "user_id": "user-123",
            "api_id": "new-api",
            "subscription_plan": "pro",
            "payment_method": "pm_test_visa"
        }
        
        # Start distributed transaction
        async with api_client.post(
            "/transactions/start",
            json={"type": "api_subscription"},
            headers=auth_headers
        ) as resp:
            transaction = await resp.json()
            tx_id = transaction["id"]
        
        try:
            # Step 1: Create API
            async with api_client.post(
                f"/apis?tx_id={tx_id}",
                json={"name": "Transaction Test API"},
                headers=auth_headers
            ) as resp:
                assert resp.status == 201
            
            # Step 2: Create subscription
            async with api_client.post(
                f"/subscriptions?tx_id={tx_id}",
                json=operation_data,
                headers=auth_headers
            ) as resp:
                assert resp.status == 201
            
            # Step 3: Process payment
            async with api_client.post(
                f"/billing/charge?tx_id={tx_id}",
                json={"amount": 9900, "currency": "usd"},
                headers=auth_headers
            ) as resp:
                assert resp.status == 200
            
            # Commit transaction
            async with api_client.post(
                f"/transactions/{tx_id}/commit",
                headers=auth_headers
            ) as resp:
                assert resp.status == 200
            
        except Exception as e:
            # Rollback on failure
            async with api_client.post(
                f"/transactions/{tx_id}/rollback",
                headers=auth_headers
            ) as resp:
                assert resp.status == 200
            raise e
    
    @pytest.mark.asyncio
    async def test_eventual_consistency(self, api_client, auth_headers):
        """Test eventual consistency between services"""
        # Update user profile
        profile_update = {
            "company": "New Company Inc",
            "timezone": "America/New_York"
        }
        
        async with api_client.patch(
            "/users/profile",
            json=profile_update,
            headers=auth_headers
        ) as resp:
            assert resp.status == 200
        
        # Check propagation to other services
        max_wait = 10  # seconds
        start_time = time.time()
        
        while time.time() - start_time < max_wait:
            # Check if billing service has updated info
            async with api_client.get(
                "/billing/customer/info",
                headers=auth_headers
            ) as resp:
                billing_info = await resp.json()
                if billing_info.get("company") == "New Company Inc":
                    break
            
            await asyncio.sleep(0.5)
        
        # Verify all services eventually consistent
        services_to_check = [
            "/billing/customer/info",
            "/notifications/preferences",
            "/analytics/user/settings"
        ]
        
        for endpoint in services_to_check:
            async with api_client.get(endpoint, headers=auth_headers) as resp:
                data = await resp.json()
                assert data.get("company") == "New Company Inc"


class TestEndToEndScenarios:
    """Test complete end-to-end user scenarios"""
    
    @pytest.mark.asyncio
    async def test_api_provider_journey(self, api_client):
        """Test complete journey of an API provider"""
        # 1. Register as API provider
        provider_data = {
            "email": "provider@example.com",
            "password": "SecurePass123!",
            "username": "apiprovider",
            "account_type": "provider"
        }
        
        async with api_client.post("/auth/register", json=provider_data) as resp:
            provider = await resp.json()
            provider_token = provider["token"]
        
        provider_headers = {"Authorization": f"Bearer {provider_token}"}
        
        # 2. Complete provider verification
        async with api_client.post(
            "/providers/verify",
            json={"business_name": "API Solutions Inc", "tax_id": "12-3456789"},
            headers=provider_headers
        ) as resp:
            assert resp.status == 200
        
        # 3. Create and publish multiple APIs
        api_ids = []
        for i in range(3):
            api_data = {
                "name": f"Service API v{i+1}",
                "description": f"Service {i+1} for testing",
                "base_url": f"https://api{i+1}.example.com"
            }
            
            async with api_client.post(
                "/apis",
                json=api_data,
                headers=provider_headers
            ) as resp:
                api = await resp.json()
                api_ids.append(api["id"])
        
        # 4. Monitor API performance
        await asyncio.sleep(2)  # Simulate some usage
        
        async with api_client.get(
            "/analytics/provider/dashboard",
            headers=provider_headers
        ) as resp:
            dashboard = await resp.json()
            assert dashboard["total_apis"] == 3
            assert dashboard["total_revenue"] >= 0
            assert "top_apis" in dashboard
        
        # 5. Handle support ticket
        async with api_client.post(
            "/support/tickets",
            json={
                "subject": "API Integration Issue",
                "description": "Customer having trouble with authentication",
                "api_id": api_ids[0]
            },
            headers=provider_headers
        ) as resp:
            ticket = await resp.json()
            ticket_id = ticket["id"]
        
        # 6. Withdraw earnings
        async with api_client.post(
            "/billing/withdrawals",
            json={"amount": 50000, "currency": "usd"},
            headers=provider_headers
        ) as resp:
            withdrawal = await resp.json()
            assert withdrawal["status"] == "pending"
    
    @pytest.mark.asyncio
    async def test_api_consumer_journey(self, api_client):
        """Test complete journey of an API consumer"""
        # 1. Register and explore marketplace
        consumer_data = {
            "email": "developer@startup.com",
            "password": "SecurePass123!",
            "username": "developer"
        }
        
        async with api_client.post("/auth/register", json=consumer_data) as resp:
            consumer = await resp.json()
            consumer_token = consumer["token"]
        
        consumer_headers = {"Authorization": f"Bearer {consumer_token}"}
        
        # 2. Search and compare APIs
        async with api_client.get(
            "/marketplace/search?q=weather&sort=popularity"
        ) as resp:
            search_results = await resp.json()
            selected_apis = search_results["results"][:3]
        
        # 3. Test APIs with sandbox
        for api in selected_apis:
            async with api_client.post(
                f"/sandbox/test/{api['id']}",
                json={"endpoint": "/current", "params": {"city": "London"}},
                headers=consumer_headers
            ) as resp:
                test_result = await resp.json()
                assert "response" in test_result
                assert "latency" in test_result
        
        # 4. Subscribe to preferred API
        chosen_api = selected_apis[0]
        async with api_client.post(
            "/subscriptions",
            json={"api_id": chosen_api["id"], "plan": "starter"},
            headers=consumer_headers
        ) as resp:
            subscription = await resp.json()
            api_key = subscription["api_key"]
        
        # 5. Integrate and use API
        app_data = {
            "name": "Weather Dashboard",
            "description": "Real-time weather monitoring",
            "api_keys": [api_key]
        }
        
        async with api_client.post(
            "/applications",
            json=app_data,
            headers=consumer_headers
        ) as resp:
            app = await resp.json()
            app_id = app["id"]
        
        # 6. Monitor usage and costs
        await asyncio.sleep(1)  # Simulate usage
        
        async with api_client.get(
            f"/applications/{app_id}/metrics",
            headers=consumer_headers
        ) as resp:
            metrics = await resp.json()
            assert "total_requests" in metrics
            assert "total_cost" in metrics
            assert "error_rate" in metrics


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--asyncio-mode=auto"])