# API-Direct Platform Deployment Guide

This guide walks you through deploying the API-Direct platform infrastructure and setting up your development environment.

## Prerequisites

Before deploying, ensure you have the following installed:

- **AWS CLI** (v2.x): [Installation Guide](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
- **Terraform** (v1.0+): [Installation Guide](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli)
- **Go** (v1.21+): [Installation Guide](https://golang.org/doc/install)
- **Docker**: [Installation Guide](https://docs.docker.com/get-docker/)
- **kubectl**: [Installation Guide](https://kubernetes.io/docs/tasks/tools/)
- **jq**: JSON processor for parsing Terraform outputs
  - macOS: `brew install jq`
  - Ubuntu: `sudo apt-get install jq`

## AWS Account Setup

1. **Create an AWS Account** if you don't have one
2. **Configure AWS CLI** with your credentials:
   ```bash
   aws configure
   ```
3. **Verify AWS access**:
   ```bash
   aws sts get-caller-identity
   ```

## Infrastructure Deployment

### 1. Clone the Repository

```bash
git clone https://github.com/patrg444/CLI-API-Marketplace.git
cd CLI-API-Marketplace
```

### 2. Configure Terraform Variables

```bash
cd infrastructure/terraform
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars` with your preferred settings:
- Set a strong database password
- Adjust instance types and node counts for your needs
- Configure region if not using us-east-1

### 3. Deploy Infrastructure

Use the automated deployment script:

```bash
cd ../..  # Return to project root
./scripts/deploy-infrastructure.sh
```

This script will:
- Initialize Terraform
- Validate the configuration
- Show you the deployment plan
- Apply the infrastructure (after confirmation)
- Save outputs to `outputs.json`

**⏱️ Expected Duration**: 15-20 minutes for full deployment

### 4. Configure CLI Environment

After infrastructure deployment:

```bash
./scripts/configure-cli-env.sh
```

This will:
- Extract Terraform outputs
- Create environment variable files
- Configure kubectl for EKS access

### 5. Load Environment Variables

```bash
source cli-env.sh
```

Add to your shell profile for persistence:
```bash
echo "source $(pwd)/cli-env.sh" >> ~/.bashrc  # or ~/.zshrc
```

### 6. Configure Security Environment Variables

Create a `.env` file for service configuration:

```bash
# Cognito Configuration (from Terraform outputs)
export COGNITO_USER_POOL_ID="your-user-pool-id"
export COGNITO_CLIENT_ID="your-client-id"
export AWS_REGION="us-east-1"

# Database Configuration
export DATABASE_URL="postgresql://user:password@rds-endpoint:5432/apiplatform"

# Stripe Configuration (for billing)
export STRIPE_SECRET_KEY="sk_test_..."
export STRIPE_WEBHOOK_SECRET="whsec_..."
export STRIPE_PUBLISHABLE_KEY="pk_test_..."

# Elasticsearch Configuration
export ELASTICSEARCH_URL="http://elasticsearch-service:9200"

# Redis Configuration
export REDIS_URL="redis://redis-service:6379"
```

## Building and Testing the CLI

### 1. Build the CLI Tool

```bash
cd cli
go mod download
go build -o apidirect
```

### 2. Test Authentication

```bash
./apidirect login
```

This will:
- Open your browser for authentication
- Create a Cognito user account
- Save authentication tokens locally

### 3. Create a Test API

```bash
./apidirect init test-api --runtime python3.9
cd test-api
cat README.md  # Review the generated project
```

## Terraform Outputs Reference

After deployment, key outputs are available:

| Output | Description | Usage |
|--------|-------------|-------|
| `cognito_user_pool_id` | Cognito User Pool ID | CLI authentication |
| `cognito_cli_client_id` | CLI OAuth client ID | CLI authentication |
| `cognito_auth_domain` | Cognito auth domain | OAuth2 flow |
| `api_gateway_endpoint` | ALB endpoint URL | API requests |
| `eks_cluster_id` | EKS cluster name | Kubernetes access |
| `code_storage_bucket_name` | S3 bucket for code | Code uploads |
| `ecr_registry_url` | ECR registry URL | Container images |

View all outputs:
```bash
cd infrastructure/terraform
terraform output
```

## Cost Optimization Tips

### Development Environment

To minimize costs during development:

1. **Reduce EKS nodes**:
   ```hcl
   min_node_count = 1
   max_node_count = 3
   ```

2. **Use smaller instances**:
   ```hcl
   node_instance_types = ["t3.small"]
   ```

3. **Destroy when not in use**:
   ```bash
   cd infrastructure/terraform
   terraform destroy
   ```

### Production Considerations

For production deployment:

1. Enable deletion protection on RDS
2. Use larger instance types
3. Configure backup retention
4. Set up monitoring and alerts
5. Use a custom domain with SSL certificate

## Troubleshooting

### Common Issues

1. **Terraform state lock**: If deployment is interrupted:
   ```bash
   terraform force-unlock <LOCK_ID>
   ```

2. **EKS authentication issues**:
   ```bash
   aws eks update-kubeconfig --region us-east-1 --name api-direct-dev-eks
   ```

3. **Cognito domain conflicts**: The domain must be globally unique. Modify in `cognito.tf` if needed.

### Cleanup

To completely remove all infrastructure:

```bash
cd infrastructure/terraform
terraform destroy -auto-approve
```

⚠️ **Warning**: This will delete all resources including databases and stored data.

## Service Deployment

### 1. Deploy Database Migrations

```bash
# Connect to RDS and run migrations
cd infrastructure/database/migrations
psql -h <rds-endpoint> -U postgres -d apiplatform < 001_base_schema.sql
psql -h <rds-endpoint> -U postgres -d apiplatform < 002_marketplace_schema.sql
psql -h <rds-endpoint> -U postgres -d apiplatform < 003_payout_schema.sql
psql -h <rds-endpoint> -U postgres -d apiplatform < 004_review_system_updates.sql
```

### 2. Deploy Kubernetes Services

```bash
# Create namespace and deploy all services
kubectl apply -f infrastructure/k8s/namespace.yaml
kubectl apply -f infrastructure/k8s/
```

### 3. Build and Push Service Images

Use the automated deployment script:

```bash
./scripts/deploy-services.sh
```

Or manually build each service:

```bash
# Get ECR login
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR_REGISTRY_URL

# Build and push each service
for service in gateway storage deployment marketplace apikey billing metering payout; do
  cd services/$service
  docker build -t $ECR_REGISTRY_URL/$service:latest .
  docker push $ECR_REGISTRY_URL/$service:latest
  cd ../..
done
```

### 4. Configure Service Secrets

```bash
# Create Kubernetes secrets for sensitive data
kubectl create secret generic cognito-config \
  --from-literal=user-pool-id=$COGNITO_USER_POOL_ID \
  --from-literal=client-id=$COGNITO_CLIENT_ID \
  -n api-platform

kubectl create secret generic stripe-config \
  --from-literal=secret-key=$STRIPE_SECRET_KEY \
  --from-literal=webhook-secret=$STRIPE_WEBHOOK_SECRET \
  -n api-platform

kubectl create secret generic db-config \
  --from-literal=url=$DATABASE_URL \
  -n api-platform
```

## Security Configuration

### JWT Authentication Setup

All services now implement JWT authentication with AWS Cognito:

1. **Shared Authentication Package**: `services/shared/auth/cognito.go`
   - JWKS validation against Cognito
   - Token expiration checks
   - Role-based access control

2. **Service Middleware**: Each service has updated auth middleware
   - Validates JWT tokens on protected endpoints
   - Enforces role requirements (Creator, Consumer, Admin)
   - API ownership verification

### Required Cognito Custom Attributes

Ensure your Cognito User Pool has these custom attributes:
- `custom:user_type` - Values: "creator", "consumer", "admin"
- `custom:stripe_customer_id` - For billing integration
- `custom:stripe_account_id` - For creator payouts

### API Ownership Verification

The platform now enforces API ownership across all services:
- Storage Service: Verifies ownership before delete/update operations
- Deployment Service: Checks ownership for deployment actions
- Marketplace Service: Ensures only owners can modify API settings

## Production Checklist

### Security Requirements ✅
- [ ] All environment variables set
- [ ] Cognito user pool configured with custom attributes
- [ ] SSL certificates installed on ALB
- [ ] Database connection uses SSL
- [ ] Kubernetes secrets created
- [ ] Network policies configured

### Service Dependencies ✅
- [ ] PostgreSQL database running and migrations applied
- [ ] Redis deployed for caching and rate limiting
- [ ] Elasticsearch cluster ready for search
- [ ] All microservices deployed and healthy
- [ ] Ingress configured with proper routing

### Monitoring Setup
- [ ] CloudWatch logs configured
- [ ] Service health checks enabled
- [ ] Alerts configured for critical failures
- [ ] Backup strategy implemented

## Next Steps

1. **Verify Service Health**:
   ```bash
   kubectl get pods -n api-platform
   kubectl get services -n api-platform
   ```

2. **Test Authentication Flow**:
   ```bash
   ./apidirect login
   ./apidirect init test-api --runtime python3.9
   ./apidirect deploy
   ```

3. **Set up CI/CD** for automated deployments

4. **Configure Monitoring** with CloudWatch and Prometheus

## Support

For issues or questions:
- GitHub Issues: [CLI-API-Marketplace/issues](https://github.com/patrg444/CLI-API-Marketplace/issues)
- Documentation: [docs/](./docs/)
