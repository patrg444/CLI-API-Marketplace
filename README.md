# API-Direct Platform

A command-line-first platform for deploying, managing, and monetizing APIs with minimal DevOps overhead.

## 📖 Table of Contents

- [Overview](#-overview)
- [Feature Highlights](#-feature-highlights)
- [CLI Installation](#-cli-installation)
- [Quick Start Guide](#-quick-start-guide)
- [Documentation & Resources](#-documentation--resources)
- [Common Use Cases](#-common-use-cases)
- [Architecture](#%EF%B8%8F-architecture)
- [Technology Stack](#%EF%B8%8F-technology-stack)
- [Advanced Setup](#%EF%B8%8F-advanced-setup)
- [Security & Authentication](#-security--authentication)
- [Development Phases](#-development-phases)
- [Production Readiness](#-production-readiness-status)
- [Contributing](#-contributing)

## 🚀 Overview

API-Direct empowers developers to effortlessly create, deploy, manage, and monetize APIs directly from their command line. The platform abstracts away infrastructure complexity while fostering a vibrant marketplace of diverse and accessible API services.

## 🎯 Feature Highlights

- **Zero DevOps Setup**: Deploy APIs in minutes without infrastructure knowledge
- **Multi-Language Support**: Python, Node.js, Go, Ruby, Java, PHP templates
- **Built-in Marketplace**: Instantly monetize your APIs with subscriptions and usage-based pricing
- **Auto-scaling**: Handle traffic spikes without manual intervention
- **Real-time Monitoring**: Built-in metrics, logs, and performance tracking
- **Security First**: JWT authentication, API keys, rate limiting, and HTTPS by default
- **Import Existing APIs**: Bring your existing APIs without code changes

## ⚡ Quick Preview

See the CLI in action:

```bash
# Create a new API in seconds
$ apidirect init my-api --template python-fastapi
✅ Created new API project: my-api

# Deploy with one command
$ apidirect deploy
🚀 Deploying my-api...
✅ Deployed successfully!
🌐 API URL: https://my-api-abc123.api-direct.com

# Check status and logs
$ apidirect status
✅ my-api is running (2 replicas)
📊 24h: 1.2k requests, 99.9% uptime

# Publish to marketplace
$ apidirect publish --price 9.99
💰 Published to marketplace!
📈 Start earning from API subscriptions
```

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

## 📦 CLI Installation

### Quick Install (Recommended)

The fastest way to get started:

```bash
curl -fsSL https://raw.githubusercontent.com/patrg444/CLI-API-Marketplace/main/cli/install.sh | bash
```

### Platform-Specific Installation

#### macOS
```bash
# Homebrew (recommended)
brew install apidirect

# Manual download
curl -L https://github.com/patrg444/CLI-API-Marketplace/releases/latest/download/apidirect-darwin-amd64 -o apidirect
chmod +x apidirect && sudo mv apidirect /usr/local/bin/
```

#### Windows
```powershell
# Chocolatey
choco install apidirect

# PowerShell (manual)
Invoke-WebRequest -Uri "https://github.com/patrg444/CLI-API-Marketplace/releases/latest/download/apidirect-windows-amd64.exe" -OutFile "apidirect.exe"
```

#### Linux
```bash
# Ubuntu/Debian
curl -sSL https://apt.apidirect.io/key.gpg | sudo apt-key add -
echo "deb https://apt.apidirect.io stable main" | sudo tee /etc/apt/sources.list.d/apidirect.list
sudo apt update && sudo apt install apidirect

# Manual download
curl -L https://github.com/patrg444/CLI-API-Marketplace/releases/latest/download/apidirect-linux-amd64 -o apidirect
chmod +x apidirect && sudo mv apidirect /usr/local/bin/
```

#### Docker
```bash
docker run -it apidirect/cli:latest --help
```

### Verify Installation

```bash
apidirect --version
apidirect --help
```

## 🚀 Quick Start Guide

Get your API deployed in under 5 minutes! For a detailed walkthrough, see our [Quick Start Guide](cli/QUICK_START.md).

### 1. Create Your First API

```bash
# Initialize a new API project
apidirect init my-weather-api

# Choose your framework (Python/FastAPI, Node.js/Express, Go, etc.)
# Follow the interactive prompts
```

### 2. Deploy to Production

```bash
cd my-weather-api

# Authenticate with the platform
apidirect auth login

# Deploy your API
apidirect deploy

# Your API is now live! 🎉
```

### 3. Monetize (Optional)

```bash
# Set pricing and publish to marketplace
apidirect pricing set --plan basic --price 9.99 --calls 1000
apidirect publish
```

### Import Existing APIs

Already have an API? Import it without code changes:

```bash
apidirect import /path/to/your/existing/api
apidirect validate  # Verify configuration
apidirect deploy    # Deploy to production
```

## 📚 Documentation & Resources

### CLI Documentation
- **[Quick Start Guide](cli/QUICK_START.md)** - Get started in 5 minutes
- **[CLI Reference](docs/CLI_REFERENCE.md)** - Complete command documentation
- **[API Import Guide](docs/IMPORT_GUIDE.md)** - Bring existing APIs to the platform
- **[Deployment Guide](docs/DEPLOYMENT_GUIDE.md)** - Advanced deployment options

### Platform Documentation
- **[Architecture Overview](#%EF%B8%8F-architecture)** - Technical architecture details
- **[Security Guide](docs/SECURITY.md)** - Authentication and security features
- **[Marketplace Guide](docs/MARKETPLACE.md)** - Publishing and monetization
- **[API Reference](https://api.api-direct.io/docs)** - REST API documentation

### Developer Resources
- **[Contributing Guide](CONTRIBUTING.md)** - How to contribute to the project
- **[Examples Repository](examples/)** - Sample API implementations
- **[Testing Guide](testing/OVERALL_TEST_SUMMARY.md)** - Testing strategies and reports
- **[Launch Checklist](LAUNCH_CHECKLIST.md)** - Production deployment checklist

### Support & Community
- **[GitHub Issues](https://github.com/patrg444/CLI-API-Marketplace/issues)** - Bug reports and feature requests
- **[Discussions](https://github.com/patrg444/CLI-API-Marketplace/discussions)** - Community support
- **Email Support**: support@api-direct.io

## 💡 Common Use Cases

### For API Creators
- **Rapid Prototyping**: Get APIs online quickly for testing and validation
- **Production Deployment**: Scale from prototype to production seamlessly
- **Monetization**: Turn your APIs into revenue streams
- **Multi-environment Management**: Separate dev, staging, and production environments

### For API Consumers
- **API Discovery**: Find APIs in the marketplace by category and functionality
- **Subscription Management**: Easy subscription and billing management
- **Usage Monitoring**: Track API usage and costs in real-time
- **Integration Testing**: Test APIs before committing to subscriptions

### For Development Teams
- **Microservices Architecture**: Deploy and manage multiple interconnected APIs
- **CI/CD Integration**: Automate deployments with GitHub Actions and other tools
- **Team Collaboration**: Share APIs and environments across development teams
- **Cost Optimization**: Pay only for what you use with auto-scaling

### Prerequisites (for platform deployment)

- Docker 20.10+ and Docker Compose 2.0+
- Domain name (for production)
- Stripe account (for payments)
- 4GB+ RAM, 2+ CPU cores

### 🎯 One-Command Platform Launch

**Development:**
```bash
git clone <repository-url>
cd CLI-API-Marketplace
./scripts/deploy-development.sh
```

**Production:**
```bash
git clone <repository-url>
cd CLI-API-Marketplace
cp web/marketplace/.env.example .env.production
# Edit .env.production with your settings
./scripts/deploy-production.sh
```

### 📊 Access Your Platform
- **Marketplace**: http://localhost:3001
- **Creator Portal**: http://localhost:3000
- **Health Check**: http://localhost:3001/api/health

## 🛠️ Advanced Setup

### For Development Teams

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

#### For API Creators

1. **Import existing API** (NEW):
   ```bash
   apidirect import ./my-existing-api  # Auto-detect configuration
   apidirect validate                   # Verify before deploying
   ```

2. **Deploy to cloud**:
   ```bash
   apidirect deploy                     # Deploy using manifest
   apidirect deploy --hosted            # Use managed infrastructure
   ```

3. **Local development** (NEW):
   ```bash
   apidirect run                        # Run locally
   apidirect run --watch                # Auto-reload on changes
   ```

4. **Environment management** (NEW):
   ```bash
   apidirect env set DATABASE_URL=postgres://...
   apidirect env list --production
   apidirect env pull --staging
   ```

5. **Monitor & scale** (NEW):
   ```bash
   apidirect status                     # Deployment status
   apidirect logs --follow              # Stream logs
   apidirect scale --replicas 5         # Scale up
   ```

6. **Publish to marketplace**:
   ```bash
   apidirect publish                    # List on marketplace
   apidirect pricing set --plan-file pricing.json
   apidirect marketplace info           # View analytics
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

### Phase 2: Monetization & Consumer Experience ✅
- [x] Consumer subscriptions (data model)
- [x] API key management service
- [x] Billing service with Stripe integration
- [x] Metering service for usage tracking
- [x] Creator payouts service
- [x] Review and rating system
- [x] Complete marketplace frontend with Next.js
- [x] E2E testing coverage (90%+ pass rate)
- [x] Production deployment configuration

### Phase 3: Enhanced Developer Experience ✅
- [x] Import existing APIs without code changes
- [x] Auto-detection of frameworks and configuration
- [x] Interactive manifest generation
- [x] Local development with hot-reload
- [x] Advanced environment management
- [x] Log streaming with filtering and search
- [x] Real-time status monitoring
- [x] Dynamic scaling controls
- [x] Pre-deployment validation
- [x] Marketplace analytics dashboard
- [x] Revenue tracking and payouts
- [x] Consumer subscription management
- [x] Review and rating system
- [x] Advanced marketplace search
- [ ] Auto-documentation generation
- [ ] Version management

### Phase 4: Distribution & Deployment ✅
- [x] Multi-platform CLI builds (macOS, Windows, Linux)
- [x] Package manager support (Homebrew, Chocolatey, apt)
- [x] Docker images for containerized usage
- [x] GitHub Actions for automated releases
- [x] Install scripts for quick setup
- [x] Comprehensive test coverage

### Phase 5: Scaling & Advanced Features
- [x] Advanced search with Elasticsearch
- [x] Interactive API documentation (Swagger UI)
- [ ] Community features
- [ ] Performance optimization

## 🚀 Production Readiness Status

### ✅ Ready for Launch
- **Complete Marketplace Frontend**: Next.js with TypeScript, fully responsive
- **Authentication & Security**: Cognito JWT validation implemented
- **Core Services**: All microservices deployed and configured
- **Database**: PostgreSQL with complete schema migrations
- **API Gateway**: Rate limiting, authentication, and routing
- **Storage**: S3 integration for code packages
- **Search**: Elasticsearch for API discovery
- **Monitoring**: Health checks, Prometheus metrics, Grafana dashboards
- **Payment Processing**: Stripe integration with subscriptions and payouts
- **Deployment**: Docker containers with production optimization
- **Testing**: 90%+ E2E test coverage across all user flows

### 🎯 Launch Checklist
See [LAUNCH_CHECKLIST.md](LAUNCH_CHECKLIST.md) for complete pre-launch requirements including:
- Environment configuration
- Domain and SSL setup  
- Payment provider configuration
- Deployment verification

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

- **Repository**: https://github.com/patrg444/CLI-API-Marketplace
- **Documentation**: https://docs.api-direct.io
- **API Reference**: https://api.api-direct.io/docs
- **Marketplace**: https://marketplace.api-direct.io
- **Support**: support@api-direct.io

## 🙏 Acknowledgments

Built with love by the API-Direct team, leveraging best practices from modern cloud-native architectures.
