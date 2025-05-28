# Payout Service

The Payout Service manages creator earnings, Stripe Connect integration, and monthly payouts for the API Direct Marketplace.

## Overview

This service handles:
- Stripe Connect account onboarding for creators
- Real-time earnings tracking
- Monthly payout processing with 20% platform commission
- Platform revenue analytics
- Automated payout scheduling

## Architecture

### Components

1. **Stripe Connect Integration**
   - Creator account onboarding
   - Payout processing via transfers
   - Dashboard access for creators

2. **Earnings Tracking**
   - Real-time calculation from billing data
   - Per-API earnings breakdown
   - Monthly and lifetime totals

3. **Payout Processing**
   - Monthly automated payouts (1st of each month)
   - Minimum payout threshold: $25
   - Platform fee: 20%
   - Detailed line item tracking

4. **Background Workers**
   - **Earnings Worker**: Calculates earnings hourly from billing data
   - **Payout Worker**: Processes monthly payouts
   - **Report Worker**: Generates monthly platform reports

## API Endpoints

### Stripe Connect Account Management
- `POST /api/v1/accounts/onboard` - Start Stripe Connect onboarding
- `GET /api/v1/accounts/onboard/callback` - Handle onboarding return
- `GET /api/v1/accounts/status` - Get account status
- `GET /api/v1/accounts/dashboard` - Get Stripe dashboard link

### Earnings & Payouts
- `GET /api/v1/earnings` - Get creator earnings summary
- `GET /api/v1/earnings/{apiId}` - Get earnings for specific API
- `GET /api/v1/payouts` - List creator payouts
- `GET /api/v1/payouts/{payoutId}` - Get payout details
- `GET /api/v1/payouts/upcoming` - Get upcoming payout info

### Platform Analytics (Admin Only)
- `GET /api/v1/platform/revenue` - Platform revenue metrics
- `GET /api/v1/platform/analytics` - Comprehensive analytics

### Webhooks
- `POST /webhooks/stripe` - Handle Stripe webhooks

## Database Schema

### Tables
- `creator_payment_accounts` - Stripe Connect account details
- `payouts` - Payout records with status tracking
- `payout_line_items` - Detailed payout breakdown by API
- `creator_earnings` - Real-time earnings tracking
- `platform_revenue` - Platform-wide revenue metrics

## Configuration

### Environment Variables
```env
DATABASE_URL=postgresql://user:pass@host:5432/api_direct
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
```

### Payout Schedule
- **Frequency**: Monthly (1st of each month)
- **Minimum Amount**: $25.00
- **Platform Fee**: 20%
- **Currency**: USD

## Stripe Connect Flow

1. **Onboarding**
   ```
   Creator → Start Onboarding → Stripe Connect → Return to Platform
   ```

2. **Payout Processing**
   ```
   Calculate Earnings → Create Payout Record → Stripe Transfer → Update Status
   ```

## Security Considerations

- All endpoints require creator authentication
- Platform analytics restricted to admin users
- Stripe webhook signature verification
- Sensitive data (Stripe keys) stored in K8s secrets

## Development

### Running Locally
```bash
# Set environment variables
export DATABASE_URL=postgresql://localhost:5432/api_direct
export STRIPE_SECRET_KEY=sk_test_...
export STRIPE_WEBHOOK_SECRET=whsec_...

# Run the service
go run main.go
```

### Testing Webhooks
Use Stripe CLI for local webhook testing:
```bash
stripe listen --forward-to localhost:8086/webhooks/stripe
```

### Database Migrations
Apply the payout schema migration:
```bash
psql $DATABASE_URL < infrastructure/database/migrations/003_payout_schema.sql
```

## Monitoring

### Key Metrics
- Active Stripe Connect accounts
- Monthly payout volume
- Platform revenue (20% commission)
- Failed payout rate
- Earnings calculation lag

### Health Checks
- `GET /health` - Service health status

## Error Handling

### Common Issues
1. **Onboarding Incomplete**: Creator hasn't finished Stripe Connect setup
2. **Below Threshold**: Earnings below $25 minimum
3. **Account Restricted**: Stripe account issues
4. **Calculation Errors**: Billing data inconsistencies

### Retry Logic
- Failed payouts are retried with exponential backoff
- Webhook failures trigger automatic retries from Stripe

## Integration Points

### Dependencies
- **Billing Service**: Source of revenue data
- **Storage Service**: API ownership verification
- **PostgreSQL**: Data persistence
- **Stripe API**: Payment processing

### Consumers
- **Creator Portal**: Payout dashboard and onboarding
- **Admin Dashboard**: Platform analytics

## Future Enhancements

1. **Multi-currency Support**: Handle international creators
2. **Custom Payout Schedules**: Weekly/bi-weekly options
3. **Tax Documentation**: 1099 generation for US creators
4. **Payout Notifications**: Email/SMS alerts
5. **Advanced Analytics**: Creator performance metrics
