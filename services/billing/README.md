# Billing Service

The Billing Service handles all payment processing, subscription management, and billing operations for the API marketplace platform.

## Features

### Core Functionality
- **Stripe Integration**: Full integration with Stripe for payment processing
- **Subscription Management**: Create, update, cancel, and manage API subscriptions
- **Payment Methods**: Add, remove, and manage payment methods
- **Invoice Management**: Track and retrieve billing history
- **Webhook Processing**: Handle Stripe webhook events for real-time updates
- **Usage-Based Billing**: Support for metered/pay-per-use pricing models
- **Background Workers**: Automated usage aggregation and billing tasks

### API Endpoints

#### Consumer Management
- `POST /api/v1/consumers/register` - Register or retrieve consumer account
- `GET /api/v1/consumers/{consumerId}` - Get consumer details

#### Subscription Management
- `POST /api/v1/subscriptions` - Create new subscription
- `GET /api/v1/subscriptions` - List user's subscriptions
- `GET /api/v1/subscriptions/{subscriptionId}` - Get subscription details
- `PUT /api/v1/subscriptions/{subscriptionId}/cancel` - Cancel subscription
- `PUT /api/v1/subscriptions/{subscriptionId}/upgrade` - Upgrade/downgrade subscription
- `GET /api/v1/subscriptions/{subscriptionId}/usage` - Get subscription usage

#### Payment Methods
- `POST /api/v1/payment-methods` - Add payment method
- `GET /api/v1/payment-methods` - List payment methods
- `DELETE /api/v1/payment-methods/{paymentMethodId}` - Remove payment method
- `PUT /api/v1/payment-methods/{paymentMethodId}/default` - Set default payment method

#### Invoices
- `GET /api/v1/invoices` - List invoices
- `GET /api/v1/invoices/{invoiceId}` - Get invoice details
- `GET /api/v1/invoices/{invoiceId}/download` - Download invoice PDF

#### Creator Analytics
- `GET /api/v1/apis/{apiId}/usage` - Get API usage summary
- `GET /api/v1/apis/{apiId}/earnings` - Get API earnings

#### Webhooks
- `POST /webhooks/stripe` - Stripe webhook endpoint (no auth required)

## Architecture

### Components

1. **Stripe Client** (`stripe/client.go`)
   - Wrapper around Stripe Go SDK
   - Handles all Stripe API operations
   - Manages customers, subscriptions, payment methods, and invoices

2. **Store Layer** (`store/`)
   - `consumer.go` - Consumer account management
   - `subscription.go` - Subscription data operations
   - `invoice.go` - Invoice tracking
   - `billing.go` - Aggregated billing operations and pricing plans

3. **Webhook Handler** (`webhooks/stripe.go`)
   - Processes Stripe webhook events
   - Updates local database state
   - Handles subscription lifecycle events

4. **Background Workers** (`workers/workers.go`)
   - Usage aggregation worker
   - Invoice generation worker
   - Subscription sync worker

5. **HTTP Handlers** (`handlers/handlers.go`)
   - REST API endpoint implementations
   - Request validation and response formatting

## Database Schema

The service uses the following main tables:
- `consumers` - Consumer accounts with Stripe customer IDs
- `subscriptions` - Active subscriptions linking consumers to APIs
- `invoices` - Billing history and invoice records
- `api_pricing_plans` - Pricing plan configurations

## Configuration

### Environment Variables
```bash
# Server Configuration
PORT=8080

# Database
DATABASE_URL=postgresql://user:pass@host:5432/dbname

# Redis
REDIS_URL=redis://localhost:6379

# Stripe
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...

# JWT Authentication
JWT_SECRET=your-secret-key

# Service URLs
API_KEY_SERVICE_URL=http://apikey-service:8080
METERING_SERVICE_URL=http://metering-service:8080
```

## Integration Points

### API Key Service
- Generates API keys upon successful subscription
- Deactivates keys when subscriptions are cancelled

### Metering Service
- Fetches usage data for usage-based billing
- Aggregates API call counts for billing periods

### Marketplace Frontend
- Provides subscription creation endpoints
- Returns Stripe checkout sessions or direct subscription data
- Supplies billing history and subscription management

## Stripe Webhook Events Handled

- `customer.created`
- `customer.subscription.created`
- `customer.subscription.updated`
- `customer.subscription.deleted`
- `customer.subscription.trial_will_end`
- `invoice.created`
- `invoice.finalized`
- `invoice.paid`
- `invoice.payment_failed`
- `checkout.session.completed`
- `payment_intent.succeeded`
- `payment_intent.payment_failed`

## Development

### Running Locally
```bash
# Install dependencies
go mod download

# Run with environment variables
go run main.go
```

### Docker Build
```bash
docker build -t billing-service .
```

### Testing
```bash
go test ./...
```

## Deployment

The service is deployed as a Kubernetes deployment with:
- 2 replicas for high availability
- Health checks on `/health` endpoint
- Network policies for secure communication
- Secrets for sensitive configuration

See `infrastructure/k8s/billing-service.yaml` for full deployment configuration.

## Security Considerations

- JWT authentication for all API endpoints except webhooks
- Stripe webhook signature verification
- Network policies restrict communication
- Secrets stored in Kubernetes secrets
- PCI compliance handled by Stripe

## Future Enhancements

1. **Stripe Connect Integration** - For creator payouts
2. **Advanced Usage Analytics** - More detailed usage breakdowns
3. **Custom Invoice Generation** - For enterprise customers
4. **Multi-Currency Support** - Handle different currencies
5. **Tax Calculation** - Integration with tax services
6. **Dunning Management** - Automated payment retry logic
