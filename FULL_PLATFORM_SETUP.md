# API-Direct Full Platform Setup Guide

## 🏗️ Complete Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         API-Direct Platform                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  Frontend Apps                    Backend Services                    │
│  ┌─────────────┐                 ┌─────────────────┐                │
│  │  Landing    │                 │  Gateway (8082) │◄──── API Calls │
│  │  (Vercel)   │                 └────────┬────────┘                │
│  └─────────────┘                          │                         │
│  ┌─────────────┐                 ┌────────▼────────┐                │
│  │  Console    │◄───────────────►│ FastAPI Backend │                │
│  │  (Vercel)   │                 │     (8000)      │                │
│  └─────────────┘                 └────────┬────────┘                │
│  ┌─────────────┐                          │                         │
│  │ Marketplace │                 ┌────────▼────────────────┐        │
│  │  (Vercel)   │                 │   Microservices         │        │
│  └─────────────┘                 ├─────────────────────────┤        │
│                                  │ • API Key Service (8083) │        │
│  ┌─────────────┐                 │ • Billing Service       │        │
│  │   CLI Tool  │◄───────────────►│ • Deployment (8081)     │        │
│  │ (apidirect) │                 │ • Marketplace           │        │
│  └─────────────┘                 │ • Metering              │        │
│                                  │ • Payout Service        │        │
│                                  │ • Storage (8087)        │        │
│                                  └─────────┬───────────────┘        │
│                                            │                         │
│  Infrastructure                   ┌────────▼────────┐                │
│  ┌─────────────┐                 │   PostgreSQL    │                │
│  │ Kubernetes  │                 │   Redis         │                │
│  │ Terraform   │                 │   Elasticsearch │                │
│  │ Monitoring  │                 └─────────────────┘                │
│  └─────────────┘                                                    │
└─────────────────────────────────────────────────────────────────────┘
```

## 📋 Complete Component List

### 1. **Microservices** (Go-based)
- **Gateway Service** - API routing, rate limiting, auth validation
- **API Key Service** - Manages API keys for consumers
- **Billing Service** - Stripe subscriptions and payments
- **Deployment Service** - Deploys APIs to Kubernetes/Knative
- **Marketplace Service** - API discovery and search
- **Metering Service** - Usage tracking for billing
- **Payout Service** - Creator earnings distribution
- **Storage Service** - S3 code package management

### 2. **Web Applications**
- **Landing Page** - Marketing site (deployed)
- **Console** - Creator dashboard (deployed)
- **Marketplace** - API discovery (deployed)
- **Docs Site** - Documentation (deployed)

### 3. **Backend API** (FastAPI)
- Main orchestration layer
- WebSocket support
- Dashboard endpoints
- Analytics APIs

### 4. **CLI Tool**
- Full-featured command-line interface
- 20+ commands for API management
- AWS Cognito authentication

### 5. **Infrastructure**
- **Terraform** - Complete AWS infrastructure
- **Kubernetes** - Service orchestration
- **Docker** - Containerization
- **Monitoring** - Prometheus + Grafana

### 6. **Testing**
- 830+ E2E tests (Playwright)
- Performance tests (K6)
- Unit tests for all services

## 🚀 Full Local Development Setup

### Step 1: Start All Services

```bash
# Start core infrastructure
docker-compose -f docker-compose.yml up -d

# This starts:
# - PostgreSQL (5432)
# - Redis (6379)
# - Elasticsearch (9200)
# - All microservices
# - FastAPI backend (8000)
# - Gateway (8082)
# - Monitoring stack
```

### Step 2: Initialize Database

```bash
# Run migrations
docker-compose exec postgres psql -U apidirect -d apidirect -f /docker-entrypoint-initdb.d/001_base_schema.sql
docker-compose exec postgres psql -U apidirect -d apidirect -f /docker-entrypoint-initdb.d/002_marketplace.sql
docker-compose exec postgres psql -U apidirect -d apidirect -f /docker-entrypoint-initdb.d/003_payouts.sql
docker-compose exec postgres psql -U apidirect -d apidirect -f /docker-entrypoint-initdb.d/004_api_reviews.sql
```

### Step 3: Configure Environment

```bash
# Copy environment template
cp .env.example .env

# For local development with mock auth
cp .env.local .env
```

### Step 4: Test Services

```bash
# Check all services are running
docker-compose ps

# Test gateway
curl http://localhost:8082/health

# Test backend API
curl http://localhost:8000/health

# Test storage service
curl http://localhost:8087/health
```

## 🌐 Service Endpoints

### Local Development
- **Backend API**: http://localhost:8000
- **API Gateway**: http://localhost:8082
- **Storage Service**: http://localhost:8087
- **API Key Service**: http://localhost:8083
- **Deployment Service**: http://localhost:8081
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379
- **Elasticsearch**: http://localhost:9200
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000

### Production URLs
- **Console**: https://console.apidirect.dev
- **Marketplace**: https://marketplace.apidirect.dev
- **Docs**: https://docs.apidirect.dev
- **API**: https://api.apidirect.dev (future)

## 🔧 Using the Complete Platform

### 1. Creator Flow (Publishing an API)

```bash
# Initialize a new API project
apidirect init my-weather-api

# Develop locally
cd my-weather-api
apidirect run

# Deploy to platform
apidirect deploy

# Publish to marketplace
apidirect publish --pricing free

# Monitor performance
apidirect analytics
apidirect logs --follow
```

### 2. Consumer Flow (Using an API)

```bash
# Search for APIs
apidirect search weather

# Subscribe to an API
apidirect subscribe weather-api

# Get API key
apidirect keys list

# Use the API
curl -H "X-API-Key: your-key" https://api.apidirect.dev/weather-api/forecast
```

### 3. Platform Admin Flow

```bash
# Monitor all services
docker-compose logs -f

# Check service health
./scripts/health-check.sh

# View Grafana dashboards
open http://localhost:3000

# Scale services
docker-compose scale gateway=3
```

## 📊 Monitoring & Observability

### Grafana Dashboards
1. **Marketplace Overview** - API metrics, popular APIs
2. **Service Health** - All microservice status
3. **Business Metrics** - Revenue, subscriptions, usage

### Prometheus Metrics
- Request rates and latencies
- Error rates by service
- Resource utilization
- Business KPIs

## 🧪 Testing the Full Platform

```bash
# Run unit tests
make test

# Run E2E tests
cd testing/e2e
npm test

# Run performance tests
cd testing/performance
k6 run load-test.js

# Generate test data
cd testing/data-generators
node generate-test-data.js
```

## 🚢 Production Deployment

### Using Terraform

```bash
cd infrastructure/terraform
terraform init
terraform plan
terraform apply
```

This provisions:
- AWS EKS cluster
- RDS PostgreSQL
- ElastiCache Redis
- S3 buckets
- ALB/NLB
- Cognito User Pool
- VPC and networking

### Using Kubernetes

```bash
# Deploy all services
kubectl apply -f infrastructure/k8s/

# Check deployments
kubectl get pods -n apidirect
kubectl get services -n apidirect
```

## 🔐 Security Configuration

### AWS Cognito
- User authentication
- API key validation
- Role-based access

### Stripe Integration
- Set `STRIPE_SECRET_KEY`
- Configure webhooks
- Test with Stripe CLI

## 📈 Next Steps

1. **Start Local Environment**
   ```bash
   docker-compose up -d
   ./start-local-backend.sh
   ```

2. **Configure AWS Resources** (when available)
   - Run Terraform
   - Set up Cognito
   - Configure domains

3. **Deploy to Production**
   - Push to GitHub (auto-deploys frontend)
   - Deploy backend to EKS
   - Configure SSL certificates

4. **Launch Platform**
   - Onboard creators
   - Enable monetization
   - Monitor metrics

The platform is **production-ready** with all components built and tested! 🎉