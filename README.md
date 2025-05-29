# API-Direct Platform

A command-line-first platform for deploying, managing, and monetizing APIs with minimal DevOps overhead.

## 🚀 Overview

API-Direct empowers developers to effortlessly create, deploy, manage, and monetize APIs directly from their command line. The platform abstracts away infrastructure complexity while fostering a vibrant marketplace of diverse and accessible API services.

## 🏗️ Architecture

The platform consists of several key components:

### Infrastructure (AWS-based)
- **Networking**: VPC with public/private subnets across multiple AZs
- **Authentication**: AWS Cognito for user management and OAuth2
- **Container Orchestration**: Amazon EKS with Knative for serverless functions
- **Database**: RDS PostgreSQL for metadata and user data
- **Storage**: S3 for code packages, ECR for container images
- **Load Balancing**: Application Load Balancer for routing

### Core Components
1. **CLI Tool** (`apidirect`)
   - Written in Go for cross-platform support
   - OAuth2 authentication flow
   - Project scaffolding and local testing
   - Deployment and management commands

2. **Backend Services**
   - Gateway Service: API routing and authentication
   - Storage Service: Code package management
   - Deployment Service: Container building and Knative deployment
   - Marketplace Service: API discovery and monetization

3. **Web Applications**
   - Marketplace Frontend: Browse and subscribe to APIs
   - Creator Portal: Dashboard for API creators

## 🛠️ Technology Stack

- **CLI**: Go with Cobra framework
- **Backend Services**: Go with Gin/Echo
- **Frontend**: React with TypeScript
- **Infrastructure**: Terraform for IaC
- **Container Runtime**: Docker with Knative
- **Cloud Provider**: AWS

## 📁 Project Structure

```
CLI-API-Marketplace/
├── cli/                    # Go CLI tool
│   ├── cmd/               # CLI commands
│   ├── pkg/               # Shared packages
│   │   ├── auth/         # Authentication logic
│   │   ├── config/       # Configuration management
│   │   └── scaffold/     # Project templates
│   └── main.go
├── services/              # Backend microservices
│   ├── gateway/          # API Gateway service
│   ├── storage/          # Code storage service
│   ├── deployment/       # Deployment service
│   └── marketplace/      # Marketplace service
├── web/                   # Frontend applications
│   ├── marketplace/      # Public marketplace
│   └── creator-portal/   # Creator dashboard
├── infrastructure/        # Infrastructure as Code
│   └── terraform/        # Terraform configurations
└── docs/                  # Documentation
```

## 🚀 Getting Started

### Prerequisites

- AWS Account with appropriate permissions
- Terraform >= 1.0
- Go >= 1.21
- Node.js >= 18 (for marketplace frontend)
- Docker

### Infrastructure Setup

1. **Configure AWS credentials**:
   ```bash
   aws configure
   ```

2. **Deploy infrastructure**:
   ```bash
   cd infrastructure/terraform
   terraform init
   terraform plan -out=tfplan
   terraform apply tfplan
   ```

3. **Note the outputs** (especially Cognito configuration)

### Building the CLI

1. **Build the CLI tool**:
   ```bash
   cd cli
   go build -o apidirect
   ```

2. **Install globally** (optional):
   ```bash
   go install
   ```

3. **Configure the CLI**:
   ```bash
   export APIDIRECT_REGION=us-east-1
   export APIDIRECT_COGNITO_POOL=<your-pool-id>
   export APIDIRECT_COGNITO_CLIENT=<your-client-id>
   export APIDIRECT_AUTH_DOMAIN=<your-auth-domain>
   ```

### Using the CLI

1. **Login**:
   ```bash
   apidirect login
   ```

2. **Create a new API**:
   ```bash
   apidirect init my-api --runtime python3.9
   cd my-api
   ```

3. **Deploy**:
   ```bash
   apidirect deploy
   ```

4. **View logs**:
   ```bash
   apidirect logs my-api
   ```

5. **Manage API in marketplace**:
   ```bash
   apidirect marketplace publish my-api
   apidirect pricing set my-api --plan free --limit 1000
   apidirect pricing set my-api --plan pro --price 29.99
   ```

## 🔐 Security & Authentication

**Status: ✅ Production Ready**

The platform implements enterprise-grade security:
- **JWT Authentication**: AWS Cognito integration with JWKS validation
- **Role-Based Access Control**: Creator, Consumer, and Admin roles
- **API Ownership Verification**: Enforced across all services
- **Secure Token Handling**: Token expiration and client ID validation

## 📊 Development Phases

### Phase 1: Core Platform & MVP ✅
- [x] Terraform infrastructure setup
- [x] Basic CLI with authentication
- [x] Project scaffolding (Python/Node.js)
- [x] JWT authentication with Cognito
- [x] Role-based access control
- [x] API ownership verification
- [x] Storage service with S3 integration
- [x] Deployment service framework
- [x] Basic marketplace website

### Phase 2: Monetization & Consumer Experience 🔄 (In Progress)
- [x] Consumer subscriptions (data model)
- [x] API key management service
- [x] Billing service with Stripe integration
- [x] Metering service for usage tracking
- [x] Creator payouts service
- [x] Review and rating system
- [ ] Payment processing webhooks (final integration)

### Phase 3: Enhanced Developer Experience
- [ ] Advanced CLI features
- [ ] Pre-deployment analysis
- [ ] Auto-documentation
- [ ] Version management
- [ ] Log streaming from Kubernetes
- [ ] Metrics collection

### Phase 4: Scaling & Advanced Features
- [x] Advanced search with Elasticsearch
- [x] Interactive API documentation (Swagger UI)
- [ ] Community features
- [ ] Performance analytics dashboard

## 🚀 Production Readiness Status

### ✅ Ready for Production
- **Authentication & Security**: Cognito JWT validation implemented
- **Core Services**: All microservices deployed and configured
- **Database**: PostgreSQL with complete schema migrations
- **API Gateway**: Rate limiting, authentication, and routing
- **Storage**: S3 integration for code packages
- **Search**: Elasticsearch for API discovery
- **Monitoring**: Basic health checks and logging

### ⚠️ Pre-Production Checklist
1. Set required environment variables:
   - `COGNITO_USER_POOL_ID`
   - `COGNITO_CLIENT_ID`
   - `AWS_REGION`
   - `STRIPE_SECRET_KEY`
   - `STRIPE_WEBHOOK_SECRET`
2. Run database migrations
3. Configure Elasticsearch indices
4. Set up SSL certificates
5. Configure backup strategies

### 📈 Testing Status
- **Unit Tests**: Core business logic covered
- **Integration Tests**: API endpoints tested
- **Security Tests**: Authentication flows verified
- **Load Tests**: Basic performance benchmarks completed

For detailed testing reports, see `/testing/OVERALL_TEST_SUMMARY.md`

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- Documentation: https://docs.api-direct.io
- API Reference: https://api.api-direct.io/docs
- Support: support@api-direct.io

## 🙏 Acknowledgments

Built with love by the API-Direct team, leveraging best practices from modern cloud-native architectures.
