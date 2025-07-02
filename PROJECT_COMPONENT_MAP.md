# CLI-API-Marketplace Component Map

## Project Overview
API-Direct is a command-line-first platform for deploying, managing, and monetizing APIs with minimal DevOps overhead. The platform abstracts away infrastructure complexity while fostering a marketplace for API services.

## 1. Go Microservices (/services)

### Core Services (All written in Go):

#### 1.1 API Key Service (`/services/apikey`)
- **Purpose**: Manages API key generation, validation, and lifecycle
- **Port**: 8083
- **Dependencies**: PostgreSQL, Cognito
- **Key Features**:
  - API key generation and storage
  - Key validation middleware
  - CORS support
  - PostgreSQL store for key persistence

#### 1.2 Billing Service (`/services/billing`)
- **Purpose**: Handles subscriptions, invoicing, and payment processing
- **Port**: Not specified in main.go
- **Dependencies**: PostgreSQL, Stripe
- **Key Features**:
  - Subscription management
  - Invoice generation
  - Consumer billing
  - Stripe webhook integration
  - Background workers for payment processing

#### 1.3 Deployment Service (`/services/deployment`)
- **Purpose**: Manages API deployments to Kubernetes/Knative
- **Port**: 8081
- **Dependencies**: Kubernetes cluster, Cognito
- **Key Features**:
  - K8s client integration
  - Deployment lifecycle management
  - Authentication middleware
  - API deployment models

#### 1.4 Gateway Service (`/services/gateway`)
- **Purpose**: API Gateway for routing, rate limiting, and authentication
- **Port**: 8082 (assumed)
- **Dependencies**: Redis (for rate limiting)
- **Key Features**:
  - Request proxying
  - API key validation
  - Rate limiting with Redis
  - CORS handling
  - Request logging

#### 1.5 Marketplace Service (`/services/marketplace`)
- **Purpose**: API marketplace functionality including search and reviews
- **Dependencies**: PostgreSQL, Elasticsearch
- **Key Features**:
  - API listing and search
  - Review system
  - Elasticsearch integration for advanced search
  - API indexing
  - Review storage and retrieval

#### 1.6 Metering Service (`/services/metering`)
- **Purpose**: Tracks API usage and aggregates metrics
- **Dependencies**: PostgreSQL
- **Key Features**:
  - Usage tracking
  - Metric aggregation
  - CORS middleware
  - Nginx configuration included

#### 1.7 Payout Service (`/services/payout`)
- **Purpose**: Manages creator earnings and payouts
- **Dependencies**: PostgreSQL, Stripe
- **Key Features**:
  - Earnings tracking
  - Payout processing
  - Stripe Connect integration
  - Account management
  - Background workers for payout processing

#### 1.8 Storage Service (`/services/storage`)
- **Purpose**: Manages code package storage in S3
- **Port**: 8087 (mapped from 8080)
- **Dependencies**: AWS S3, Cognito
- **Key Features**:
  - S3 client integration
  - Code package upload/download
  - Authentication via Cognito
  - API metadata storage

#### 1.9 Shared Components (`/services/shared`)
- **Purpose**: Common code shared across services
- **Features**:
  - Cognito authentication helpers
  - Common store interfaces
  - Consumer data models

#### 1.10 Hosted Service (`/services/hosted`)
- **Purpose**: Appears to be for hosted API execution
- **Files**: Contains both Go and Python code (app.py)

## 2. Infrastructure Components

### 2.1 Terraform Configuration (`/infrastructure/terraform`)
- **Cloud Provider**: AWS
- **Key Resources**:
  - VPC with public/private subnets
  - Amazon EKS cluster
  - RDS PostgreSQL
  - S3 buckets (code storage)
  - ECR (container registry)
  - Application Load Balancer
  - AWS Cognito (authentication)

### 2.2 Kubernetes Manifests (`/infrastructure/k8s`)
- Deployment configurations for all services
- Ingress configuration
- Namespace setup
- Redis service
- Storage configurations

### 2.3 Database (`/infrastructure/database`)
- **Migrations**:
  - 001: Base schema (users, APIs)
  - 002: Marketplace schema
  - 003: Payout schema
  - 004: Review system updates
- **Schema includes**:
  - Users table
  - APIs table
  - Reviews
  - Subscriptions
  - Payouts
  - API keys

### 2.4 Infrastructure Modules (`/infrastructure/modules`)
- Reusable Terraform modules:
  - API Fargate module
  - Database module
  - IAM cross-account module
  - Networking module

## 3. Monitoring Setup (`/monitoring`)

### 3.1 Prometheus Configuration
- Scrape configs for all services
- Alert rules
- Service discovery
- Metrics collection from:
  - Backend services
  - PostgreSQL
  - Redis
  - Nginx
  - Node exporters

### 3.2 Grafana Dashboards
- API marketplace overview dashboard
- Service metrics visualization

### 3.3 Alertmanager
- Alert routing and notifications

## 4. Testing Framework (`/testing`)

### 4.1 E2E Tests (`/testing/e2e`)
- Playwright-based tests
- 830 tests across multiple browsers
- Test fixtures and results
- Button functionality tests
- Scroll verification tests

### 4.2 Test Data Generators (`/testing/data-generators`)
- JavaScript-based data generation
- Creates test data for:
  - APIs
  - Users (creators/consumers)
  - Reviews
  - Subscriptions
  - Usage data
  - API keys

### 4.3 Performance Tests (`/testing/performance`)
- K6 load testing scripts

### 4.4 Test Reports
- Daily test execution reports
- Bug fix verification
- Service status tracking

## 5. CLI Tool (`/cli`)

### 5.1 Commands (`/cli/cmd`)
Built with Cobra framework, includes:
- **auth**: Login/logout functionality
- **init**: Project initialization
- **deploy**: API deployment (v1 and v2)
- **import**: Import existing APIs
- **run**: Local development server
- **status**: Deployment status
- **logs**: Log streaming (v1 and v2)
- **scale**: Scaling controls
- **publish**: Marketplace publishing
- **pricing**: Pricing management
- **marketplace**: Marketplace operations
- **search**: API search
- **subscribe**: Subscription management
- **earnings**: View earnings
- **analytics**: Usage analytics
- **validate**: Configuration validation
- **env**: Environment variable management
- **docs**: Documentation generation
- **completion**: Shell completion
- **version**: Version information
- **self-update**: CLI updates

### 5.2 Packages (`/cli/pkg`)
- **auth**: Authentication logic (Cognito integration)
- **config**: Configuration management
- **detector**: Framework auto-detection
- **errors**: Error handling
- **manifest**: API manifest handling
- **scaffold**: Project scaffolding and ML templates
- **wizard**: Interactive setup wizard

### 5.3 Templates (`/cli/templates`)
- ML-specific templates
- Template configuration

## 6. Scripts and Automation (`/scripts`)

### 6.1 Deployment Scripts
- **deploy-development.sh**: Development environment setup
- **deploy-production.sh**: Production deployment
- **deploy-simple.sh**: Simplified deployment (handles go.sum)
- **deploy-quiet.sh**: Less verbose deployment
- **deploy-infrastructure.sh**: Infrastructure-only deployment
- **deploy-services.sh**: Service deployment

### 6.2 Utility Scripts
- **backup.sh**: Database backup
- **backup-automation.sh**: Automated backup setup
- **health-check.sh**: Service health verification
- **verify-deployment.sh**: Deployment verification
- **configure-cli-env.sh**: CLI environment setup

### 6.3 Fix Scripts
- **fix-dockerfiles.sh**: Dockerfile corrections
- **fix-go-dependencies.sh**: Go dependency fixes
- **generate-go-sums.sh**: Generate missing go.sum files

### 6.4 Testing Scripts
- **run-e2e-tests.sh**: E2E test execution
- **test-platform.py**: Platform testing (Python)
- **verify-platform.sh**: Platform verification

## 7. Web Applications (`/web`)

### 7.1 Marketplace Frontend (`/web/marketplace`)
- **Framework**: Next.js 14 with TypeScript
- **Styling**: Tailwind CSS
- **Features**:
  - API browsing and search
  - Subscription management
  - API playground
  - Review system
  - Responsive design

### 7.2 Creator Portal (`/web/creator-portal`)
- **Framework**: React
- **Purpose**: Dashboard for API creators
- **Features**:
  - API management
  - Analytics viewing
  - Earnings tracking

### 7.3 Landing Page (`/web/landing`)
- Static landing pages
- Documentation site
- Templates showcase

### 7.4 Console (`/web/console`)
- Legacy console application
- Includes both creator and consumer portals

## 8. Docker Configurations

### 8.1 Main Docker Compose Files
- **docker-compose.yml**: Primary configuration
- **docker-compose.dev.yml**: Development overrides
- **docker-compose.local.yml**: Local development
- **docker-compose.production.yml**: Production configuration
- **docker-compose.test.yml**: Testing configuration

### 8.2 Service Configuration
- All services have individual Dockerfiles
- Multi-stage builds for optimization
- Health checks configured
- Volume mounts for development

## 9. Database Schema

### Core Tables:
- **users**: API creators with Cognito integration
- **apis**: API metadata and configuration
- **api_keys**: Generated API keys
- **subscriptions**: Consumer subscriptions
- **usage_records**: API usage tracking
- **invoices**: Billing records
- **payouts**: Creator payouts
- **reviews**: API reviews and ratings
- **earnings**: Creator earnings tracking

## 10. Current Status

### ✅ What's Built:
1. **Complete microservice architecture** (9 Go services)
2. **CLI tool** with comprehensive commands
3. **Web applications** (marketplace, creator portal)
4. **Database schema** with migrations
5. **Terraform infrastructure** as code
6. **Docker configurations** for all services
7. **Monitoring setup** (Prometheus/Grafana)
8. **Testing framework** with E2E tests
9. **Authentication** via AWS Cognito
10. **Payment integration** with Stripe

### ⚠️ What Needs Connection/Deployment:
1. **Docker services need to be running** locally
2. **AWS infrastructure** needs to be provisioned via Terraform
3. **Environment variables** need to be configured
4. **Stripe account** needs to be connected
5. **Domain names** need to be configured
6. **SSL certificates** need to be set up
7. **Kubernetes cluster** needs to be configured

### 📊 Architecture Summary:
- **Architecture Pattern**: Microservices
- **Communication**: REST APIs with authentication
- **Data Storage**: PostgreSQL + Redis + S3 + Elasticsearch
- **Container Orchestration**: Kubernetes/Docker Compose
- **Cloud Provider**: AWS
- **Authentication**: AWS Cognito with JWT
- **Payment Processing**: Stripe
- **Monitoring**: Prometheus + Grafana
- **Testing**: Comprehensive E2E with Playwright

This is a production-ready platform that requires deployment and configuration to become operational. All core components are built and tested.