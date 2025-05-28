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

## Next Steps

1. **Deploy Kubernetes components**:
   ```bash
   kubectl apply -f infrastructure/k8s/
   ```

2. **Build backend services**:
   ```bash
   cd services/storage
   docker build -t storage-service .
   ```

3. **Set up CI/CD** for automated deployments

## Support

For issues or questions:
- GitHub Issues: [CLI-API-Marketplace/issues](https://github.com/patrg444/CLI-API-Marketplace/issues)
- Documentation: [docs/](./docs/)
