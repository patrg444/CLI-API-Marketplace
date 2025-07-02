"""
Tests for webhook management system
"""

import pytest
import asyncio
import json
from datetime import datetime
from unittest.mock import Mock, AsyncMock, patch
import httpx
from sqlalchemy.ext.asyncio import AsyncSession

from ..webhooks import (
    WebhookManager,
    WebhookCreate,
    WebhookUpdate,
    WebhookEventType,
    WebhookStatus,
    DeliveryStatus
)
from ..database import Webhook, WebhookDelivery, WebhookEvent


@pytest.fixture
async def webhook_manager():
    """Create a webhook manager instance for testing"""
    manager = WebhookManager()
    # Mock Redis client
    manager.redis_client = AsyncMock()
    manager.redis_client.setex = AsyncMock()
    manager.redis_client.get = AsyncMock(return_value=None)
    manager.redis_client.hset = AsyncMock()
    manager.redis_client.hgetall = AsyncMock(return_value={})
    manager.redis_client.delete = AsyncMock()
    manager.redis_client.expire = AsyncMock()
    manager.redis_client.zcard = AsyncMock(return_value=0)
    
    # Mock HTTP client
    manager.http_client = AsyncMock()
    
    # Initialize with test settings
    manager._running = True
    manager._delivery_queue = asyncio.Queue()
    
    yield manager
    
    # Cleanup
    manager._running = False
    await manager.http_client.aclose()


@pytest.fixture
def sample_webhook_data():
    """Sample webhook creation data"""
    return WebhookCreate(
        url="https://example.com/webhook",
        events=[WebhookEventType.API_DEPLOYED, WebhookEventType.API_ERROR],
        description="Test webhook",
        headers={"X-Custom-Header": "test-value"},
        retry_enabled=True,
        max_retries=3,
        timeout_seconds=30
    )


class TestWebhookManager:
    """Test webhook manager functionality"""
    
    async def test_create_webhook(self, webhook_manager, mock_db, sample_webhook_data):
        """Test creating a new webhook"""
        user_id = "test-user-123"
        
        webhook = await webhook_manager.create_webhook(
            mock_db,
            user_id,
            sample_webhook_data
        )
        
        assert webhook.user_id == user_id
        assert webhook.url == str(sample_webhook_data.url)
        assert webhook.events == sample_webhook_data.events
        assert webhook.description == sample_webhook_data.description
        assert webhook.headers == sample_webhook_data.headers
        assert webhook.status == WebhookStatus.ACTIVE
        assert webhook.secret.startswith("whsec_")
        assert webhook.failure_count == 0
        assert webhook.success_count == 0
        
        # Verify webhook was added to database
        mock_db.add.assert_called_once()
        mock_db.commit.assert_called()
        
        # Verify webhook was cached
        webhook_manager.redis_client.hset.assert_called()
        
    async def test_update_webhook(self, webhook_manager, mock_db):
        """Test updating an existing webhook"""
        webhook_id = "webhook-123"
        user_id = "test-user-123"
        
        # Create mock webhook
        mock_webhook = Mock(spec=Webhook)
        mock_webhook.id = webhook_id
        mock_webhook.user_id = user_id
        mock_webhook.url = "https://old.example.com/webhook"
        mock_webhook.events = [WebhookEventType.API_DEPLOYED]
        mock_webhook.status = WebhookStatus.ACTIVE
        
        # Mock database query
        mock_result = Mock()
        mock_result.scalar_one_or_none.return_value = mock_webhook
        mock_db.execute.return_value = mock_result
        
        # Update data
        update_data = WebhookUpdate(
            url="https://new.example.com/webhook",
            events=[WebhookEventType.API_DEPLOYED, WebhookEventType.API_ERROR],
            status=WebhookStatus.PAUSED
        )
        
        updated_webhook = await webhook_manager.update_webhook(
            mock_db,
            webhook_id,
            user_id,
            update_data
        )
        
        assert updated_webhook is not None
        assert mock_webhook.url == str(update_data.url)
        assert mock_webhook.events == update_data.events
        assert mock_webhook.status == update_data.status
        
        mock_db.commit.assert_called()
        
    async def test_delete_webhook(self, webhook_manager, mock_db):
        """Test deleting a webhook"""
        webhook_id = "webhook-123"
        user_id = "test-user-123"
        
        # Mock successful deletion
        mock_result = Mock()
        mock_result.rowcount = 1
        mock_db.execute.return_value = mock_result
        
        deleted = await webhook_manager.delete_webhook(
            mock_db,
            webhook_id,
            user_id
        )
        
        assert deleted is True
        mock_db.commit.assert_called()
        webhook_manager.redis_client.delete.assert_called_with(f"webhook:{webhook_id}")
        
    async def test_trigger_event(self, webhook_manager, mock_db):
        """Test triggering a webhook event"""
        user_id = "test-user-123"
        event_type = WebhookEventType.API_DEPLOYED
        api_id = "api-456"
        payload = {
            "api_id": api_id,
            "status": "deployed",
            "timestamp": datetime.utcnow().isoformat()
        }
        
        # Mock webhooks subscribed to this event
        mock_webhooks = [
            Mock(
                id="webhook-1",
                user_id=user_id,
                url="https://example.com/webhook1",
                events=[event_type],
                status=WebhookStatus.ACTIVE,
                secret="whsec_test1"
            ),
            Mock(
                id="webhook-2",
                user_id=user_id,
                url="https://example.com/webhook2",
                events=[event_type, WebhookEventType.API_ERROR],
                status=WebhookStatus.ACTIVE,
                secret="whsec_test2"
            )
        ]
        
        # Mock _get_webhooks_for_event
        webhook_manager._get_webhooks_for_event = AsyncMock(return_value=mock_webhooks)
        
        await webhook_manager.trigger_event(
            mock_db,
            event_type,
            user_id,
            payload,
            api_id
        )
        
        # Verify event was created
        assert mock_db.add.call_count >= 1  # Event + deliveries
        mock_db.commit.assert_called()
        
        # Verify deliveries were queued
        assert webhook_manager._delivery_queue.qsize() == 2
        
    async def test_webhook_delivery_success(self, webhook_manager, mock_db):
        """Test successful webhook delivery"""
        # Create mock webhook
        webhook = Mock(
            id="webhook-123",
            url="https://example.com/webhook",
            secret="whsec_test",
            status=WebhookStatus.ACTIVE,
            headers={"X-Custom": "test"},
            timeout_seconds=30,
            retry_enabled=True,
            max_retries=3,
            success_count=0,
            failure_count=0
        )
        
        # Create mock delivery
        delivery = Mock(
            id="delivery-123",
            webhook_id=webhook.id,
            event_type=WebhookEventType.API_DEPLOYED,
            payload={"test": "data"},
            attempt_count=0,
            status=DeliveryStatus.PENDING
        )
        
        # Mock database query
        mock_result = Mock()
        mock_result.scalar_one_or_none.return_value = webhook
        mock_db.execute.return_value = mock_result
        
        # Mock successful HTTP response
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.text = "OK"
        webhook_manager.http_client.post.return_value = mock_response
        
        # Process delivery
        await webhook_manager._process_delivery(delivery)
        
        # Verify webhook was called
        webhook_manager.http_client.post.assert_called_once()
        call_args = webhook_manager.http_client.post.call_args
        
        # Check URL
        assert call_args[0][0] == webhook.url
        
        # Check headers
        headers = call_args[1]["headers"]
        assert headers["Content-Type"] == "application/json"
        assert headers["X-Webhook-ID"] == webhook.id
        assert headers["X-Webhook-Event"] == delivery.event_type
        assert headers["X-Custom"] == "test"
        assert "X-Webhook-Signature" in headers
        
        # Check payload
        payload = call_args[1]["json"]
        assert payload["id"] == delivery.id
        assert payload["event"] == delivery.event_type
        assert payload["data"] == delivery.payload
        
        # Verify delivery was marked as delivered
        assert delivery.status == DeliveryStatus.DELIVERED
        assert delivery.response_status == 200
        assert delivery.delivered_at is not None
        assert webhook.success_count == 1
        
    async def test_webhook_delivery_failure_with_retry(self, webhook_manager, mock_db):
        """Test webhook delivery failure with retry"""
        # Create mock webhook
        webhook = Mock(
            id="webhook-123",
            url="https://example.com/webhook",
            secret="whsec_test",
            status=WebhookStatus.ACTIVE,
            headers={},
            timeout_seconds=30,
            retry_enabled=True,
            max_retries=3,
            success_count=0,
            failure_count=0
        )
        
        # Create mock delivery
        delivery = Mock(
            id="delivery-123",
            webhook_id=webhook.id,
            event_type=WebhookEventType.API_ERROR,
            payload={"error": "test"},
            attempt_count=0,
            status=DeliveryStatus.PENDING
        )
        
        # Mock database query
        mock_result = Mock()
        mock_result.scalar_one_or_none.return_value = webhook
        mock_db.execute.return_value = mock_result
        
        # Mock failed HTTP response
        webhook_manager.http_client.post.side_effect = httpx.TimeoutException("Timeout")
        
        # Process delivery
        await webhook_manager._process_delivery(delivery)
        
        # Verify delivery was marked for retry
        assert delivery.status == DeliveryStatus.RETRYING
        assert delivery.attempt_count == 1
        assert delivery.next_retry_at is not None
        assert webhook.failure_count == 1
        
    async def test_webhook_delivery_max_retries_exceeded(self, webhook_manager, mock_db):
        """Test webhook delivery when max retries exceeded"""
        # Create mock webhook
        webhook = Mock(
            id="webhook-123",
            url="https://example.com/webhook",
            secret="whsec_test",
            status=WebhookStatus.ACTIVE,
            headers={},
            timeout_seconds=30,
            retry_enabled=True,
            max_retries=3,
            success_count=0,
            failure_count=5
        )
        
        # Create mock delivery at max attempts
        delivery = Mock(
            id="delivery-123",
            webhook_id=webhook.id,
            event_type=WebhookEventType.API_ERROR,
            payload={"error": "test"},
            attempt_count=2,  # Will be 3 after increment
            status=DeliveryStatus.RETRYING
        )
        
        # Mock database query
        mock_result = Mock()
        mock_result.scalar_one_or_none.return_value = webhook
        mock_db.execute.return_value = mock_result
        
        # Mock failed HTTP response
        webhook_manager.http_client.post.side_effect = Exception("Connection error")
        
        # Process delivery
        await webhook_manager._process_delivery(delivery)
        
        # Verify delivery was marked as failed
        assert delivery.status == DeliveryStatus.FAILED
        assert delivery.attempt_count == 3
        assert webhook.failure_count == 6
        
    async def test_webhook_signature_calculation(self, webhook_manager):
        """Test webhook signature calculation"""
        secret = "whsec_test123"
        payload = {
            "id": "delivery-123",
            "event": "api.deployed",
            "data": {"test": "value"}
        }
        
        signature = webhook_manager._calculate_signature(secret, payload)
        
        assert signature.startswith("sha256=")
        assert len(signature) == 71  # sha256= + 64 hex chars
        
    async def test_get_webhooks_for_event_cached(self, webhook_manager, mock_db):
        """Test getting webhooks for event with cache hit"""
        user_id = "test-user-123"
        event_type = WebhookEventType.API_DEPLOYED
        
        # Mock cached webhook IDs
        cached_ids = json.dumps(["webhook-1", "webhook-2"])
        webhook_manager.redis_client.get.return_value = cached_ids
        
        # Mock cached webhook data
        webhook_data_1 = {
            "id": "webhook-1",
            "user_id": user_id,
            "url": "https://example.com/webhook1",
            "events": json.dumps([event_type]),
            "secret": "whsec_1",
            "status": WebhookStatus.ACTIVE,
            "headers": "{}",
            "retry_enabled": "True",
            "max_retries": "3",
            "timeout_seconds": "30"
        }
        webhook_data_2 = {
            "id": "webhook-2",
            "user_id": user_id,
            "url": "https://example.com/webhook2",
            "events": json.dumps([event_type]),
            "secret": "whsec_2",
            "status": WebhookStatus.ACTIVE,
            "headers": "{}",
            "retry_enabled": "True",
            "max_retries": "3",
            "timeout_seconds": "30"
        }
        
        webhook_manager.redis_client.hgetall.side_effect = [
            webhook_data_1,
            webhook_data_2
        ]
        
        webhooks = await webhook_manager._get_webhooks_for_event(user_id, event_type)
        
        assert len(webhooks) == 2
        assert webhooks[0].id == "webhook-1"
        assert webhooks[1].id == "webhook-2"
        
        # Verify cache was used
        webhook_manager.redis_client.get.assert_called_once()
        assert webhook_manager.redis_client.hgetall.call_count == 2
        
    async def test_retry_delivery(self, webhook_manager, mock_db):
        """Test manual retry of failed delivery"""
        delivery_id = "delivery-123"
        user_id = "test-user-123"
        
        # Create mock delivery
        mock_delivery = Mock(
            id=delivery_id,
            webhook_id="webhook-123",
            status=DeliveryStatus.FAILED,
            attempt_count=2
        )
        
        # Mock database query
        mock_result = Mock()
        mock_result.scalar_one_or_none.return_value = mock_delivery
        mock_db.execute.return_value = mock_result
        
        retried = await webhook_manager.retry_delivery(
            mock_db,
            delivery_id,
            user_id
        )
        
        assert retried is True
        assert mock_delivery.status == DeliveryStatus.PENDING
        assert mock_delivery.attempt_count == 0
        assert mock_delivery.next_retry_at is None
        
        mock_db.commit.assert_called()
        assert webhook_manager._delivery_queue.qsize() == 1