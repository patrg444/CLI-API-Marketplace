# Webhook Management System

## Overview

The webhook management system allows API-Direct users to subscribe to real-time events and receive HTTP callbacks when specific events occur. This enables seamless integration with external systems and automation workflows.

## Features

### 🔔 Event Subscriptions
- Subscribe to multiple event types per webhook
- Granular event filtering
- Support for API-specific and account-wide events

### 🚀 Reliable Delivery
- Automatic retry with exponential backoff
- Circuit breaker pattern for failing endpoints
- Delivery status tracking
- Manual retry capability

### 🔒 Security
- HMAC-SHA256 signatures for request verification
- Unique webhook secrets
- Custom headers support
- Configurable timeouts

### 📊 Monitoring
- Delivery history and logs
- Success/failure metrics
- Real-time status updates
- Failed webhook auto-disable

## Supported Events

```python
class WebhookEventType(str, Enum):
    # API Lifecycle
    API_DEPLOYED = "api.deployed"
    API_UPDATED = "api.updated"
    API_DELETED = "api.deleted"
    API_STATUS_CHANGED = "api.status_changed"
    API_ERROR = "api.error"
    
    # API Usage
    API_CALL_MADE = "api.call_made"
    API_LIMIT_REACHED = "api.limit_reached"
    
    # Deployment
    DEPLOYMENT_STARTED = "deployment.started"
    DEPLOYMENT_COMPLETED = "deployment.completed"
    DEPLOYMENT_FAILED = "deployment.failed"
    
    # Subscription & Billing
    SUBSCRIPTION_CREATED = "subscription.created"
    SUBSCRIPTION_UPDATED = "subscription.updated"
    SUBSCRIPTION_CANCELLED = "subscription.cancelled"
    PAYMENT_SUCCEEDED = "payment.succeeded"
    PAYMENT_FAILED = "payment.failed"
```

## API Endpoints

### Create Webhook
```http
POST /api/webhooks
Authorization: Bearer <token>

{
  "url": "https://example.com/webhook",
  "events": ["api.deployed", "api.error"],
  "description": "Production deployment notifications",
  "headers": {
    "X-Custom-Header": "value"
  },
  "retry_enabled": true,
  "max_retries": 3,
  "timeout_seconds": 30
}
```

### List Webhooks
```http
GET /api/webhooks?skip=0&limit=20
Authorization: Bearer <token>
```

### Get Webhook
```http
GET /api/webhooks/{webhook_id}
Authorization: Bearer <token>
```

### Update Webhook
```http
PATCH /api/webhooks/{webhook_id}
Authorization: Bearer <token>

{
  "events": ["api.deployed", "api.updated", "api.error"],
  "status": "paused"
}
```

### Delete Webhook
```http
DELETE /api/webhooks/{webhook_id}
Authorization: Bearer <token>
```

### Get Delivery History
```http
GET /api/webhooks/{webhook_id}/deliveries?skip=0&limit=50
Authorization: Bearer <token>
```

### Retry Failed Delivery
```http
POST /api/webhooks/{webhook_id}/deliveries/{delivery_id}/retry
Authorization: Bearer <token>
```

### Test Webhook
```http
POST /api/webhooks/test/{webhook_id}
Authorization: Bearer <token>
```

## Webhook Payload Format

All webhooks receive a standardized payload:

```json
{
  "id": "delivery-uuid",
  "event": "api.deployed",
  "created": "2024-01-10T12:00:00Z",
  "data": {
    // Event-specific data
  }
}
```

### Example: API Deployed Event
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "event": "api.deployed",
  "created": "2024-01-10T12:00:00Z",
  "data": {
    "api_id": "api-123",
    "api_name": "sentiment-analysis",
    "version": "1.0.0",
    "endpoint_url": "https://api.apidirect.dev/sentiment-123",
    "deployment_type": "hosted",
    "user_id": "user-456"
  }
}
```

## Request Verification

Verify webhook authenticity using the signature header:

```python
import hmac
import hashlib
import json

def verify_webhook_signature(payload: bytes, signature: str, secret: str) -> bool:
    expected = hmac.new(
        secret.encode('utf-8'),
        payload,
        hashlib.sha256
    ).hexdigest()
    
    # Remove 'sha256=' prefix from signature
    actual = signature.replace('sha256=', '')
    
    return hmac.compare_digest(expected, actual)

# Usage
@app.post("/webhook")
async def handle_webhook(request: Request):
    payload = await request.body()
    signature = request.headers.get("X-Webhook-Signature")
    
    if not verify_webhook_signature(payload, signature, webhook_secret):
        raise HTTPException(status_code=401, detail="Invalid signature")
    
    # Process webhook...
```

## Headers

Every webhook request includes:

| Header | Description |
|--------|-------------|
| `Content-Type` | Always `application/json` |
| `X-Webhook-ID` | Unique webhook subscription ID |
| `X-Webhook-Signature` | HMAC-SHA256 signature |
| `X-Webhook-Event` | Event type |
| `X-Webhook-Delivery` | Unique delivery ID |

## Retry Logic

Failed webhooks are retried with exponential backoff:

- **Attempt 1**: Immediate
- **Attempt 2**: After 2 minutes
- **Attempt 3**: After 4 minutes
- **Attempt 4**: After 8 minutes (if max_retries >= 4)

### Failure Conditions
- HTTP status >= 400
- Connection timeout
- SSL errors
- DNS resolution failures

### Auto-disable
Webhooks are automatically disabled after 10 consecutive failures to prevent resource waste.

## Best Practices

### 1. Respond Quickly
- Return 2xx status within timeout period
- Process webhooks asynchronously
- Use queues for heavy processing

### 2. Idempotency
- Use delivery ID for deduplication
- Handle potential duplicate deliveries
- Store processed delivery IDs

### 3. Security
- Verify signatures on every request
- Use HTTPS endpoints only
- Rotate webhook secrets periodically
- Whitelist API-Direct IPs if possible

### 4. Error Handling
```python
@app.post("/webhook")
async def handle_webhook(request: Request):
    try:
        # Verify and process
        return {"status": "accepted"}, 200
    except Exception as e:
        # Log error but return success to prevent retries
        logger.error(f"Webhook processing error: {e}")
        return {"status": "accepted"}, 200
```

## Implementation Details

### Architecture
- **Async Workers**: Multiple workers process deliveries concurrently
- **Redis Queue**: Reliable delivery queue with persistence
- **Circuit Breaker**: Prevents cascading failures
- **Database**: PostgreSQL for webhook metadata and history

### Performance
- Handles 1000+ webhooks/second
- 30-second default timeout
- Automatic connection pooling
- Batch processing for high volume

### Monitoring
- Prometheus metrics for delivery rates
- OpenTelemetry tracing
- Real-time WebSocket updates
- Health check endpoints

## CLI Integration

```bash
# Create webhook
apidirect webhook create \
  --url https://example.com/webhook \
  --events api.deployed,api.error \
  --description "Production notifications"

# List webhooks
apidirect webhook list

# View delivery history
apidirect webhook deliveries <webhook-id>

# Test webhook
apidirect webhook test <webhook-id>

# Delete webhook
apidirect webhook delete <webhook-id>
```

## Troubleshooting

### Common Issues

1. **Webhook not receiving events**
   - Verify webhook is active
   - Check event subscriptions
   - Confirm endpoint is accessible
   - Review delivery history for errors

2. **Signature verification failing**
   - Ensure using raw request body
   - Check secret hasn't changed
   - Verify signature algorithm (SHA256)

3. **Timeouts**
   - Respond within timeout period
   - Process asynchronously
   - Return 200 immediately

4. **Too many retries**
   - Fix endpoint errors
   - Disable retry if not needed
   - Increase timeout if appropriate

## Future Enhancements

- [ ] Webhook templates
- [ ] Event filtering rules
- [ ] Batch delivery mode
- [ ] GraphQL subscriptions
- [ ] Webhook transformations
- [ ] Dead letter queue
- [ ] Analytics dashboard