"""
Webhook management system for API-Direct.
Handles webhook subscriptions, delivery, and retry logic.
"""

import asyncio
import hashlib
import hmac
import json
import os
import time
import uuid
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any, Set
from enum import Enum
import httpx
from pydantic import BaseModel, HttpUrl, Field
from sqlalchemy import select, update, delete
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.dialects.postgresql import UUID
import redis.asyncio as redis
from circuitbreaker import circuit
import logging

from .database import get_db, Webhook, WebhookDelivery, WebhookEvent

logger = logging.getLogger(__name__)

# Settings (could be moved to config.py)
class Settings:
    REDIS_URL = os.getenv("REDIS_URL", "redis://localhost:6379")
    WEBHOOK_WORKERS = int(os.getenv("WEBHOOK_WORKERS", "3"))

settings = Settings()


class WebhookEventType(str, Enum):
    """Webhook event types"""
    API_DEPLOYED = "api.deployed"
    API_UPDATED = "api.updated"
    API_DELETED = "api.deleted"
    API_STATUS_CHANGED = "api.status_changed"
    API_ERROR = "api.error"
    API_CALL_MADE = "api.call_made"
    API_LIMIT_REACHED = "api.limit_reached"
    DEPLOYMENT_STARTED = "deployment.started"
    DEPLOYMENT_COMPLETED = "deployment.completed"
    DEPLOYMENT_FAILED = "deployment.failed"
    SUBSCRIPTION_CREATED = "subscription.created"
    SUBSCRIPTION_UPDATED = "subscription.updated"
    SUBSCRIPTION_CANCELLED = "subscription.cancelled"
    PAYMENT_SUCCEEDED = "payment.succeeded"
    PAYMENT_FAILED = "payment.failed"


class WebhookStatus(str, Enum):
    """Webhook status"""
    ACTIVE = "active"
    PAUSED = "paused"
    FAILED = "failed"
    DISABLED = "disabled"


class DeliveryStatus(str, Enum):
    """Webhook delivery status"""
    PENDING = "pending"
    DELIVERED = "delivered"
    FAILED = "failed"
    RETRYING = "retrying"


class WebhookCreate(BaseModel):
    """Schema for creating a webhook"""
    url: HttpUrl
    events: List[WebhookEventType]
    description: Optional[str] = None
    headers: Optional[Dict[str, str]] = Field(default_factory=dict)
    retry_enabled: bool = True
    max_retries: int = Field(default=3, ge=0, le=10)
    timeout_seconds: int = Field(default=30, ge=5, le=120)


class WebhookUpdate(BaseModel):
    """Schema for updating a webhook"""
    url: Optional[HttpUrl] = None
    events: Optional[List[WebhookEventType]] = None
    description: Optional[str] = None
    headers: Optional[Dict[str, str]] = None
    status: Optional[WebhookStatus] = None
    retry_enabled: Optional[bool] = None
    max_retries: Optional[int] = Field(None, ge=0, le=10)
    timeout_seconds: Optional[int] = Field(None, ge=5, le=120)


class WebhookResponse(BaseModel):
    """Schema for webhook response"""
    id: str
    url: str
    events: List[str]
    status: str
    description: Optional[str]
    secret: str
    created_at: datetime
    updated_at: datetime
    last_triggered_at: Optional[datetime]
    failure_count: int
    success_count: int


class WebhookDeliveryResponse(BaseModel):
    """Schema for webhook delivery response"""
    id: str
    webhook_id: str
    event_type: str
    status: str
    attempt_count: int
    response_status: Optional[int]
    response_body: Optional[str]
    error_message: Optional[str]
    delivered_at: Optional[datetime]
    next_retry_at: Optional[datetime]
    created_at: datetime


class WebhookManager:
    """Manages webhook subscriptions and deliveries"""
    
    def __init__(self):
        self.redis_client: Optional[redis.Redis] = None
        self.http_client = httpx.AsyncClient(
            timeout=httpx.Timeout(30.0),
            limits=httpx.Limits(max_keepalive_connections=20, max_connections=100)
        )
        self._delivery_queue: asyncio.Queue = asyncio.Queue()
        self._workers: List[asyncio.Task] = []
        self._running = False
        
    async def initialize(self):
        """Initialize webhook manager"""
        self.redis_client = redis.from_url(
            settings.REDIS_URL,
            encoding="utf-8",
            decode_responses=True
        )
        
        # Start delivery workers
        self._running = True
        for i in range(settings.WEBHOOK_WORKERS):
            worker = asyncio.create_task(self._delivery_worker(i))
            self._workers.append(worker)
            
        logger.info(f"Webhook manager initialized with {settings.WEBHOOK_WORKERS} workers")
        
    async def shutdown(self):
        """Shutdown webhook manager"""
        self._running = False
        
        # Cancel all workers
        for worker in self._workers:
            worker.cancel()
            
        # Wait for workers to finish
        await asyncio.gather(*self._workers, return_exceptions=True)
        
        # Close connections
        await self.http_client.aclose()
        if self.redis_client:
            await self.redis_client.close()
            
        logger.info("Webhook manager shutdown complete")
        
    async def create_webhook(
        self,
        db: AsyncSession,
        user_id: str,
        webhook_data: WebhookCreate
    ) -> Webhook:
        """Create a new webhook subscription"""
        # Generate webhook secret
        secret = self._generate_secret()
        
        # Create webhook
        webhook = Webhook(
            id=str(uuid.uuid4()),
            user_id=user_id,
            url=str(webhook_data.url),
            events=webhook_data.events,
            description=webhook_data.description,
            secret=secret,
            headers=webhook_data.headers,
            retry_enabled=webhook_data.retry_enabled,
            max_retries=webhook_data.max_retries,
            timeout_seconds=webhook_data.timeout_seconds,
            status=WebhookStatus.ACTIVE,
            failure_count=0,
            success_count=0
        )
        
        db.add(webhook)
        await db.commit()
        await db.refresh(webhook)
        
        # Cache webhook in Redis
        await self._cache_webhook(webhook)
        
        logger.info(f"Created webhook {webhook.id} for user {user_id}")
        return webhook
        
    async def update_webhook(
        self,
        db: AsyncSession,
        webhook_id: str,
        user_id: str,
        update_data: WebhookUpdate
    ) -> Optional[Webhook]:
        """Update an existing webhook"""
        # Get webhook
        result = await db.execute(
            select(Webhook).where(
                Webhook.id == webhook_id,
                Webhook.user_id == user_id
            )
        )
        webhook = result.scalar_one_or_none()
        
        if not webhook:
            return None
            
        # Update fields
        for field, value in update_data.dict(exclude_unset=True).items():
            if value is not None:
                if field == "url":
                    value = str(value)
                setattr(webhook, field, value)
                
        webhook.updated_at = datetime.utcnow()
        
        await db.commit()
        await db.refresh(webhook)
        
        # Update cache
        await self._cache_webhook(webhook)
        
        logger.info(f"Updated webhook {webhook_id}")
        return webhook
        
    async def delete_webhook(
        self,
        db: AsyncSession,
        webhook_id: str,
        user_id: str
    ) -> bool:
        """Delete a webhook"""
        result = await db.execute(
            delete(Webhook).where(
                Webhook.id == webhook_id,
                Webhook.user_id == user_id
            )
        )
        
        if result.rowcount > 0:
            await db.commit()
            
            # Remove from cache
            await self._remove_webhook_from_cache(webhook_id)
            
            logger.info(f"Deleted webhook {webhook_id}")
            return True
            
        return False
        
    async def get_webhook(
        self,
        db: AsyncSession,
        webhook_id: str,
        user_id: str
    ) -> Optional[Webhook]:
        """Get a specific webhook"""
        result = await db.execute(
            select(Webhook).where(
                Webhook.id == webhook_id,
                Webhook.user_id == user_id
            )
        )
        return result.scalar_one_or_none()
        
    async def list_webhooks(
        self,
        db: AsyncSession,
        user_id: str,
        skip: int = 0,
        limit: int = 20
    ) -> List[Webhook]:
        """List user's webhooks"""
        result = await db.execute(
            select(Webhook)
            .where(Webhook.user_id == user_id)
            .order_by(Webhook.created_at.desc())
            .offset(skip)
            .limit(limit)
        )
        return result.scalars().all()
        
    async def get_webhook_deliveries(
        self,
        db: AsyncSession,
        webhook_id: str,
        user_id: str,
        skip: int = 0,
        limit: int = 50
    ) -> List[WebhookDelivery]:
        """Get webhook delivery history"""
        # Verify webhook ownership
        webhook = await self.get_webhook(db, webhook_id, user_id)
        if not webhook:
            return []
            
        result = await db.execute(
            select(WebhookDelivery)
            .where(WebhookDelivery.webhook_id == webhook_id)
            .order_by(WebhookDelivery.created_at.desc())
            .offset(skip)
            .limit(limit)
        )
        return result.scalars().all()
        
    async def trigger_event(
        self,
        db: AsyncSession,
        event_type: WebhookEventType,
        user_id: str,
        payload: Dict[str, Any],
        api_id: Optional[str] = None
    ):
        """Trigger a webhook event"""
        # Create event record
        event = WebhookEvent(
            id=str(uuid.uuid4()),
            event_type=event_type,
            user_id=user_id,
            api_id=api_id,
            payload=payload,
            created_at=datetime.utcnow()
        )
        
        db.add(event)
        await db.commit()
        
        # Get webhooks subscribed to this event
        webhooks = await self._get_webhooks_for_event(user_id, event_type)
        
        # Queue deliveries
        for webhook in webhooks:
            delivery = WebhookDelivery(
                id=str(uuid.uuid4()),
                webhook_id=webhook.id,
                event_id=event.id,
                event_type=event_type,
                payload=payload,
                status=DeliveryStatus.PENDING,
                attempt_count=0,
                created_at=datetime.utcnow()
            )
            
            db.add(delivery)
            await self._delivery_queue.put(delivery)
            
        await db.commit()
        
        logger.info(f"Triggered {event_type} event for user {user_id}, queued {len(webhooks)} deliveries")
        
    async def retry_delivery(
        self,
        db: AsyncSession,
        delivery_id: str,
        user_id: str
    ) -> bool:
        """Manually retry a failed delivery"""
        # Get delivery
        result = await db.execute(
            select(WebhookDelivery)
            .join(Webhook)
            .where(
                WebhookDelivery.id == delivery_id,
                Webhook.user_id == user_id
            )
        )
        delivery = result.scalar_one_or_none()
        
        if not delivery or delivery.status == DeliveryStatus.DELIVERED:
            return False
            
        # Reset delivery status
        delivery.status = DeliveryStatus.PENDING
        delivery.attempt_count = 0
        delivery.next_retry_at = None
        
        await db.commit()
        
        # Queue for delivery
        await self._delivery_queue.put(delivery)
        
        logger.info(f"Manually retrying delivery {delivery_id}")
        return True
        
    async def _delivery_worker(self, worker_id: int):
        """Worker to process webhook deliveries"""
        logger.info(f"Webhook delivery worker {worker_id} started")
        
        while self._running:
            try:
                # Get delivery from queue
                delivery = await asyncio.wait_for(
                    self._delivery_queue.get(),
                    timeout=1.0
                )
                
                # Process delivery
                await self._process_delivery(delivery)
                
            except asyncio.TimeoutError:
                continue
            except Exception as e:
                logger.error(f"Worker {worker_id} error: {e}")
                await asyncio.sleep(1)
                
        logger.info(f"Webhook delivery worker {worker_id} stopped")
        
    @circuit(failure_threshold=5, recovery_timeout=60)
    async def _process_delivery(self, delivery: WebhookDelivery):
        """Process a single webhook delivery"""
        async with get_db() as db:
            # Get webhook
            result = await db.execute(
                select(Webhook).where(Webhook.id == delivery.webhook_id)
            )
            webhook = result.scalar_one_or_none()
            
            if not webhook or webhook.status != WebhookStatus.ACTIVE:
                delivery.status = DeliveryStatus.FAILED
                delivery.error_message = "Webhook not found or inactive"
                await db.commit()
                return
                
            # Increment attempt count
            delivery.attempt_count += 1
            
            # Prepare payload
            payload = {
                "id": delivery.id,
                "event": delivery.event_type,
                "created": delivery.created_at.isoformat(),
                "data": delivery.payload
            }
            
            # Calculate signature
            signature = self._calculate_signature(webhook.secret, payload)
            
            # Prepare headers
            headers = {
                "Content-Type": "application/json",
                "X-Webhook-ID": webhook.id,
                "X-Webhook-Signature": signature,
                "X-Webhook-Event": delivery.event_type,
                "X-Webhook-Delivery": delivery.id
            }
            
            # Add custom headers
            if webhook.headers:
                headers.update(webhook.headers)
                
            try:
                # Send webhook
                response = await self.http_client.post(
                    webhook.url,
                    json=payload,
                    headers=headers,
                    timeout=webhook.timeout_seconds
                )
                
                # Update delivery status
                delivery.response_status = response.status_code
                delivery.response_body = response.text[:1000]  # Store first 1000 chars
                
                if 200 <= response.status_code < 300:
                    delivery.status = DeliveryStatus.DELIVERED
                    delivery.delivered_at = datetime.utcnow()
                    webhook.success_count += 1
                    webhook.last_triggered_at = datetime.utcnow()
                    logger.info(f"Successfully delivered webhook {delivery.id} to {webhook.url}")
                else:
                    raise Exception(f"HTTP {response.status_code}: {response.text[:200]}")
                    
            except Exception as e:
                logger.error(f"Failed to deliver webhook {delivery.id}: {e}")
                
                delivery.error_message = str(e)[:500]
                webhook.failure_count += 1
                
                # Check if we should retry
                if webhook.retry_enabled and delivery.attempt_count < webhook.max_retries:
                    delivery.status = DeliveryStatus.RETRYING
                    # Exponential backoff: 2^attempt * 60 seconds
                    retry_delay = (2 ** delivery.attempt_count) * 60
                    delivery.next_retry_at = datetime.utcnow() + timedelta(seconds=retry_delay)
                    
                    # Schedule retry
                    asyncio.create_task(self._schedule_retry(delivery, retry_delay))
                else:
                    delivery.status = DeliveryStatus.FAILED
                    
                    # Disable webhook if too many failures
                    if webhook.failure_count >= 10:
                        webhook.status = WebhookStatus.FAILED
                        logger.warning(f"Disabled webhook {webhook.id} due to excessive failures")
                        
            await db.commit()
            
    async def _schedule_retry(self, delivery: WebhookDelivery, delay: int):
        """Schedule a delivery retry"""
        await asyncio.sleep(delay)
        await self._delivery_queue.put(delivery)
        
    async def _get_webhooks_for_event(
        self,
        user_id: str,
        event_type: WebhookEventType
    ) -> List[Webhook]:
        """Get all active webhooks subscribed to an event"""
        # Try cache first
        cache_key = f"webhooks:{user_id}:{event_type}"
        cached = await self.redis_client.get(cache_key)
        
        if cached:
            webhook_ids = json.loads(cached)
            webhooks = []
            
            for webhook_id in webhook_ids:
                webhook_data = await self.redis_client.hgetall(f"webhook:{webhook_id}")
                if webhook_data:
                    webhooks.append(self._webhook_from_cache(webhook_data))
                    
            return webhooks
            
        # Query database
        async with get_db() as db:
            result = await db.execute(
                select(Webhook).where(
                    Webhook.user_id == user_id,
                    Webhook.status == WebhookStatus.ACTIVE,
                    Webhook.events.contains([event_type])
                )
            )
            webhooks = result.scalars().all()
            
            # Cache results
            webhook_ids = [w.id for w in webhooks]
            await self.redis_client.setex(
                cache_key,
                300,  # 5 minutes
                json.dumps(webhook_ids)
            )
            
            return webhooks
            
    async def _cache_webhook(self, webhook: Webhook):
        """Cache webhook in Redis"""
        cache_key = f"webhook:{webhook.id}"
        
        webhook_data = {
            "id": webhook.id,
            "user_id": webhook.user_id,
            "url": webhook.url,
            "events": json.dumps(webhook.events),
            "secret": webhook.secret,
            "status": webhook.status,
            "headers": json.dumps(webhook.headers or {}),
            "retry_enabled": str(webhook.retry_enabled),
            "max_retries": str(webhook.max_retries),
            "timeout_seconds": str(webhook.timeout_seconds)
        }
        
        await self.redis_client.hset(cache_key, mapping=webhook_data)
        await self.redis_client.expire(cache_key, 3600)  # 1 hour
        
        # Also update event subscriptions cache
        for event_type in webhook.events:
            cache_key = f"webhooks:{webhook.user_id}:{event_type}"
            await self.redis_client.delete(cache_key)  # Invalidate cache
            
    async def _remove_webhook_from_cache(self, webhook_id: str):
        """Remove webhook from cache"""
        await self.redis_client.delete(f"webhook:{webhook_id}")
        
    def _webhook_from_cache(self, cache_data: Dict[str, str]) -> Webhook:
        """Reconstruct webhook from cache data"""
        return Webhook(
            id=cache_data["id"],
            user_id=cache_data["user_id"],
            url=cache_data["url"],
            events=json.loads(cache_data["events"]),
            secret=cache_data["secret"],
            status=cache_data["status"],
            headers=json.loads(cache_data["headers"]),
            retry_enabled=cache_data["retry_enabled"] == "True",
            max_retries=int(cache_data["max_retries"]),
            timeout_seconds=int(cache_data["timeout_seconds"])
        )
        
    def _generate_secret(self) -> str:
        """Generate a webhook secret"""
        return f"whsec_{uuid.uuid4().hex}"
        
    def _calculate_signature(self, secret: str, payload: Dict[str, Any]) -> str:
        """Calculate webhook signature"""
        payload_bytes = json.dumps(payload, sort_keys=True).encode("utf-8")
        signature = hmac.new(
            secret.encode("utf-8"),
            payload_bytes,
            hashlib.sha256
        ).hexdigest()
        return f"sha256={signature}"


# Global webhook manager instance
webhook_manager = WebhookManager()