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

## 📊 Development Phases

### Phase 1: Core Platform & MVP ✅
- [x] Terraform infrastructure setup
- [x] Basic CLI with authentication
- [x] Project scaffolding (Python/Node.js)
- [ ] Deployment pipeline
- [ ] Managed execution environment
- [ ] Basic marketplace website

### Phase 2: Monetization & Consumer Experience
- [ ] Consumer subscriptions
- [ ] API key management
- [ ] Billing and metering
- [ ] Creator payouts

### Phase 3: Enhanced Developer Experience
- [ ] Advanced CLI features
- [ ] Pre-deployment analysis
- [ ] Auto-documentation
- [ ] Version management

### Phase 4: Scaling & Advanced Features
- [ ] Advanced search/discovery
- [ ] Interactive API documentation
- [ ] Community features
- [ ] Performance analytics

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
