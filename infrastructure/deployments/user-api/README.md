# API-Direct BYOA (Bring Your Own AWS) Deployment

This directory contains the complete Terraform infrastructure for deploying APIs to your own AWS account using the API-Direct BYOA model. This provides enterprise-grade infrastructure with full control and zero vendor lock-in.

## 🚀 Overview

The BYOA deployment model allows you to:
- **Deploy APIs to your own AWS account** with production-ready infrastructure
- **Maintain full control** of your data and infrastructure
- **Scale automatically** with enterprise-grade monitoring and alerting
- **Integrate seamlessly** with the API-Direct marketplace and billing platform
- **Comply with enterprise requirements** (SOC2, HIPAA, PCI-DSS, etc.)

## 🏗️ Architecture

This deployment creates a complete, production-ready infrastructure stack:

```
┌─────────────────────────────────────────────────────────────┐
│                        Your AWS Account                     │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │   Public Subnets │    │  Private Subnets │                │
│  │                 │    │                 │                │
│  │  ┌─────────────┐│    │ ┌─────────────┐ │                │
│  │  │     ALB     ││    │ │ ECS Fargate │ │                │
│  │  │             ││    │ │   Service   │ │                │
│  │  └─────────────┘│    │ └─────────────┘ │                │
│  │                 │    │                 │                │
│  │  ┌─────────────┐│    │ ┌─────────────┐ │                │
│  │  │ NAT Gateway ││    │ │     RDS     │ │                │
│  │  │             ││    │ │ PostgreSQL  │ │                │
│  │  └─────────────┘│    │ └─────────────┘ │                │
│  └─────────────────┘    └─────────────────┘                │
│                                                             │
│  ┌─────────────────────────────────────────────────────────┐│
│  │              Cross-Account IAM Role                     ││
│  │         (for API-Direct management)                     ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## 📦 Infrastructure Components

### 🌐 Networking Module
- **VPC** with public and private subnets across multiple AZs
- **Application Load Balancer** with SSL termination
- **NAT Gateways** for secure internet access from private subnets
- **Security Groups** with least-privilege access controls

### 🗄️ Database Module
- **RDS PostgreSQL** with encryption at rest and in transit
- **Automated backups** with configurable retention
- **Performance Insights** for monitoring and optimization
- **Read replicas** for high availability (optional)

### 🐳 API Fargate Module
- **ECS Fargate** cluster with auto-scaling
- **ECR repository** for container images
- **CloudWatch logging** with configurable retention
- **Health checks** for both containers and load balancer

### 🔐 IAM Cross-Account Module
- **Secure cross-account role** for API-Direct management
- **External ID rotation** for enhanced security
- **Least-privilege policies** for specific AWS services
- **Compliance controls** (MFA, IP restrictions, etc.)

## 🚀 Quick Start

### Prerequisites

1. **AWS Account** with appropriate permissions
2. **Terraform** >= 1.0 installed
3. **AWS CLI** configured with your credentials
4. **API-Direct Account** (for platform integration)

### Step 1: Clone and Configure

```bash
# Clone the repository
git clone https://github.com/your-org/api-direct-infrastructure.git
cd api-direct-infrastructure/infrastructure/deployments/user-api

# Copy the example configuration
cp terraform.tfvars.example terraform.tfvars
```

### Step 2: Customize Configuration

Edit `terraform.tfvars` with your specific settings:

```hcl
# Core Configuration
project_name = "my-api"
environment  = "prod"
aws_region   = "us-east-1"
owner_email  = "your-email@example.com"

# API-Direct Platform Configuration
api_direct_account_id = "123456789012"  # Provided by API-Direct

# Resource Configuration
db_instance_class = "db.t3.small"
cpu              = 512
memory           = 1024
desired_count    = 2

# Add your specific configuration...
```

### Step 3: Initialize and Deploy

```bash
# Initialize Terraform
terraform init

# Review the deployment plan
terraform plan

# Deploy the infrastructure
terraform apply
```

### Step 4: Configure Your Application

After deployment, you'll receive outputs including:
- **ECR Repository URL** - Push your container image here
- **Database Endpoint** - Configure your application connection
- **Load Balancer DNS** - Your API endpoint
- **Cross-Account Role ARN** - For API-Direct integration

## 📋 Configuration Reference

### Core Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `project_name` | Name of your project | - | ✅ |
| `environment` | Environment (dev/staging/prod) | `prod` | ✅ |
| `aws_region` | AWS region for deployment | `us-east-1` | ✅ |
| `owner_email` | Your email address | - | ✅ |
| `api_direct_account_id` | API-Direct AWS account ID | - | ✅ |

### Networking Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `vpc_cidr` | CIDR block for VPC | `10.0.0.0/16` |
| `az_count` | Number of Availability Zones | `2` |
| `ssl_certificate_arn` | SSL certificate ARN for HTTPS | `""` |

### Database Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `db_instance_class` | RDS instance class | `db.t3.micro` |
| `db_allocated_storage` | Initial storage (GB) | `20` |
| `db_backup_retention_period` | Backup retention (days) | `7` |

### Container Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `container_image` | Container image URL | `""` (uses ECR) |
| `container_port` | Port your app listens on | `8000` |
| `cpu` | CPU units (256/512/1024/2048/4096) | `256` |
| `memory` | Memory in MB | `512` |
| `desired_count` | Number of tasks | `2` |

### Auto Scaling Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `enable_auto_scaling` | Enable auto scaling | `true` |
| `min_capacity` | Minimum tasks | `1` |
| `max_capacity` | Maximum tasks | `10` |
| `cpu_target_value` | CPU target % | `70` |
| `memory_target_value` | Memory target % | `80` |

## 🔒 Security Features

### Cross-Account Access
- **External ID** for secure role assumption
- **MFA requirements** for enhanced security
- **IP restrictions** for access control
- **Session duration limits** for temporary access

### Data Protection
- **Encryption at rest** for database and logs
- **Encryption in transit** for all communications
- **Secrets Manager** for secure credential storage
- **VPC isolation** for network security

### Compliance Support
- **SOC2, HIPAA, PCI-DSS** compliance frameworks
- **Data residency** controls (US, EU, APAC)
- **Audit logging** with CloudTrail integration
- **Access controls** with least-privilege policies

## 📊 Monitoring and Alerting

### CloudWatch Integration
- **Custom dashboards** for API metrics
- **Automated alerts** for CPU, memory, and errors
- **Log aggregation** with configurable retention
- **Performance insights** for database monitoring

### Cost Management
- **Cost estimation** in Terraform outputs
- **Resource tagging** for cost allocation
- **Auto-scaling** to optimize costs
- **Storage optimization** with lifecycle policies

## 🔄 Deployment Workflows

### Development Environment
```hcl
project_name = "my-api-dev"
environment  = "dev"
db_instance_class = "db.t3.micro"
cpu = 256
memory = 512
desired_count = 1
enable_auto_scaling = false
db_deletion_protection = false
```

### Staging Environment
```hcl
project_name = "my-api-staging"
environment  = "staging"
db_instance_class = "db.t3.small"
cpu = 512
memory = 1024
desired_count = 2
db_backup_retention_period = 3
```

### Production Environment
```hcl
project_name = "my-api-prod"
environment  = "prod"
db_instance_class = "db.t3.medium"
cpu = 1024
memory = 2048
desired_count = 3
min_capacity = 2
max_capacity = 20
db_create_read_replica = true
db_backup_retention_period = 30
```

## 🚀 Container Deployment

### Building and Pushing Images

```bash
# Get ECR login token
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <ecr-repository-url>

# Build your container
docker build -t my-api .

# Tag for ECR
docker tag my-api:latest <ecr-repository-url>:latest

# Push to ECR
docker push <ecr-repository-url>:latest
```

### Updating Your Service

```bash
# Update the ECS service to use the new image
aws ecs update-service \
  --cluster <cluster-name> \
  --service <service-name> \
  --force-new-deployment
```

## 🔧 Troubleshooting

### Common Issues

**Deployment Fails with Permission Errors**
- Ensure your AWS credentials have sufficient permissions
- Check that the API-Direct account ID is correct
- Verify the external ID is properly configured

**Container Health Checks Failing**
- Ensure your application responds to health check endpoint
- Check container logs in CloudWatch
- Verify security group allows health check traffic

**Database Connection Issues**
- Check security group rules for database access
- Verify database credentials in Secrets Manager
- Ensure your application is in the correct VPC

### Getting Help

1. **Check CloudWatch Logs** for application errors
2. **Review Terraform State** for resource status
3. **Contact API-Direct Support** for platform issues
4. **Check AWS Console** for service health

## 📚 Additional Resources

- [Terraform AWS Provider Documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [ECS Fargate Best Practices](https://docs.aws.amazon.com/AmazonECS/latest/bestpracticesguide/)
- [RDS Security Best Practices](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_BestPractices.Security.html)
- [API-Direct Platform Documentation](https://docs.api-direct.com)

## 🤝 Support

For issues related to:
- **Infrastructure deployment**: Check this documentation and AWS console
- **API-Direct platform**: Contact API-Direct support
- **Application issues**: Check your application logs and configuration

## 📄 License

This infrastructure code is provided under the MIT License. See LICENSE file for details.
