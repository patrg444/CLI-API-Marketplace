# Marketplace Service

The Marketplace Service provides advanced API discovery, search, and review functionality for the API-Direct platform.

## Features

### 1. Elasticsearch-Powered Search
- **Full-text search** across API names, descriptions, tags, and creator names
- **Fuzzy matching** for typo tolerance
- **Faceted filtering** by category, price range, tags, and rating
- **Search suggestions** with autocomplete
- **Relevance scoring** based on popularity, ratings, and boost factors
- **Synonym support** (e.g., "ml" matches "machine learning")

### 2. Review and Rating System
- **Authenticated reviews** with 1-5 star ratings
- **Verified purchase badges** for active subscribers
- **Helpful/not helpful voting** on reviews
- **Creator responses** to customer reviews
- **Review statistics** with rating distribution
- **Sort options**: most recent, most helpful, highest/lowest rating
- **Automatic rating aggregation** using materialized views

### 3. API Discovery
- **Browse APIs** by category with pagination
- **API details** including pricing plans and documentation
- **Real-time statistics** for subscriptions and ratings
- **Published API management** with indexing on publish/update

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Marketplace   │────▶│ Marketplace API  │────▶│  Elasticsearch  │
│    Frontend     │     │     Service      │     │     Cluster     │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                               │                          ▲
                               ▼                          │
                        ┌──────────────────┐             │
                        │    PostgreSQL    │─────────────┘
                        └──────────────────┘   (Periodic Sync)
```

## API Endpoints

### Public Endpoints (No Auth Required)

#### API Discovery
- `GET /api/v1/marketplace/apis` - List published APIs
- `GET /api/v1/marketplace/apis/:id` - Get API details
- `GET /api/v1/marketplace/apis/:id/documentation` - Get API documentation

#### Search
- `POST /api/v1/marketplace/search` - Advanced search with filters
- `GET /api/v1/marketplace/search/suggestions?q=...` - Search suggestions

#### Reviews (Read-only)
- `GET /api/v1/marketplace/apis/:id/reviews` - Get API reviews
- `GET /api/v1/marketplace/apis/:id/reviews/stats` - Get review statistics

### Authenticated Endpoints

#### Review Management
- `POST /api/v1/marketplace/apis/:id/reviews` - Submit a review
- `POST /api/v1/marketplace/reviews/:id/vote` - Vote on review helpfulness

#### Creator-Only Endpoints
- `POST /api/v1/marketplace/reviews/:id/response` - Respond to a review

#### Admin Endpoints
- `POST /api/v1/marketplace/admin/reindex` - Trigger full reindex
- `POST /api/v1/marketplace/admin/index/:id` - Index specific API

## Search Request Format

```json
{
  "query": "payment processing",
  "category": "Financial Services",
  "tags": ["stripe", "payments"],
  "price_range": "low",
  "min_rating": 4.0,
  "has_free_tier": true,
  "sort_by": "relevance",
  "page": 1,
  "limit": 20
}
```

## Review Submission Format

```json
{
  "rating": 5,
  "title": "Excellent API",
  "comment": "Great documentation and easy to integrate!"
}
```

## Environment Variables

- `PORT` - Service port (default: 8086)
- `DATABASE_URL` - PostgreSQL connection string
- `ELASTICSEARCH_URL` - Elasticsearch URL (default: http://localhost:9200)
- `COGNITO_USER_POOL_ID` - AWS Cognito user pool ID
- `COGNITO_REGION` - AWS region for Cognito

## Development

### Prerequisites
- Go 1.21+
- PostgreSQL with migrations applied
- Elasticsearch 8.x
- Docker (for containerized development)

### Running Locally

```bash
# Install dependencies
go mod download

# Run the service
go run main.go
```

### Running with Docker Compose

```bash
# Start all services including Elasticsearch
docker-compose up marketplace-api elasticsearch
```

### Testing

```bash
# Unit tests
go test ./...

# Integration tests (requires running services)
go test -tags=integration ./...
```

## Database Schema

The service uses the following key tables:
- `apis` - API metadata and statistics
- `api_pricing_plans` - Pricing information
- `api_reviews` - Customer reviews
- `review_votes` - Review helpfulness votes
- `api_rating_stats` - Materialized view for performance

## Elasticsearch Index

The `apis` index contains:
- Basic metadata (name, description, category, tags)
- Pricing information (has_free_tier, price_range)
- Statistics (ratings, reviews, subscriptions)
- Boost score for relevance tuning
- Completion suggester for autocomplete

## Monitoring

The service exposes:
- `/health` - Health check endpoint
- Structured logging with request IDs
- Elasticsearch query performance metrics (TODO)
- Review submission rate limiting (TODO)

## Future Enhancements

1. **Machine Learning Integration**
   - Personalized API recommendations
   - Review sentiment analysis
   - Search query understanding

2. **Advanced Analytics**
   - Popular search terms tracking
   - Conversion funnel analysis
   - Creator analytics dashboard

3. **Performance Optimizations**
   - Redis caching for popular searches
   - Read replicas for database queries
   - CDN integration for static content

4. **Enhanced Review System**
   - Review moderation workflow
   - Verified organization badges
   - API comparison features
