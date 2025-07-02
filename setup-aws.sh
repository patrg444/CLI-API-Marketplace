#!/bin/bash
# AWS Setup Script for API Direct Marketplace

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}API Direct Marketplace - AWS Setup${NC}"
echo "===================================="

# Check prerequisites
echo -e "\n${YELLOW}Checking prerequisites...${NC}"

# Check AWS CLI
if ! command -v aws &> /dev/null; then
    echo -e "${RED}AWS CLI not found. Please install it first:${NC}"
    echo "https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"
    exit 1
fi

# Check Terraform
if ! command -v terraform &> /dev/null; then
    echo -e "${RED}Terraform not found. Please install it first:${NC}"
    echo "https://learn.hashicorp.com/tutorials/terraform/install-cli"
    exit 1
fi

# Check AWS credentials
if ! aws sts get-caller-identity &> /dev/null; then
    echo -e "${RED}AWS credentials not configured. Run: aws configure${NC}"
    exit 1
fi

# Get AWS account info
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
AWS_REGION=$(aws configure get region)

echo -e "${GREEN}✓ AWS CLI installed${NC}"
echo -e "${GREEN}✓ Terraform installed${NC}"
echo -e "${GREEN}✓ AWS credentials configured${NC}"
echo "  Account ID: $AWS_ACCOUNT_ID"
echo "  Region: $AWS_REGION"

# Prompt for configuration
echo -e "\n${YELLOW}Configuration${NC}"
echo "============="

# Environment
echo -n "Environment (dev/staging/production) [dev]: "
read ENVIRONMENT
ENVIRONMENT=${ENVIRONMENT:-dev}

# Domain
echo -n "Domain name (e.g., api-marketplace.com): "
read DOMAIN_NAME
if [ -z "$DOMAIN_NAME" ]; then
    echo -e "${RED}Domain name is required${NC}"
    exit 1
fi

# Email
echo -n "Email address for notifications: "
read EMAIL_FROM
if [ -z "$EMAIL_FROM" ]; then
    echo -e "${RED}Email address is required${NC}"
    exit 1
fi

# Generate passwords
echo -e "\n${YELLOW}Generating secure passwords...${NC}"
DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-32)
REDIS_TOKEN=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-32)

# Create terraform.tfvars
cd infrastructure/terraform/aws
cat > terraform.tfvars <<EOF
# Auto-generated configuration
project_name = "apidirect"
environment  = "$ENVIRONMENT"
aws_region   = "$AWS_REGION"

# Domain Configuration
domain_name = "$DOMAIN_NAME"
email_from_address = "$EMAIL_FROM"

# Database Configuration
db_password = "$DB_PASSWORD"

# Redis Configuration
redis_auth_token = "$REDIS_TOKEN"

# Instance Sizing
rds_instance_class = "$( [ "$ENVIRONMENT" = "production" ] && echo "db.t3.small" || echo "db.t3.micro" )"
redis_node_type    = "$( [ "$ENVIRONMENT" = "production" ] && echo "cache.t3.small" || echo "cache.t3.micro" )"
EOF

echo -e "${GREEN}✓ Configuration file created${NC}"

# Cost estimate
echo -e "\n${YELLOW}Estimated Monthly Costs:${NC}"
if [ "$ENVIRONMENT" = "production" ]; then
    echo "  RDS PostgreSQL (t3.small): ~\$30"
    echo "  ElastiCache Redis (t3.small): ~\$26"
    echo "  ECS Fargate: ~\$50-200"
    echo "  Load Balancer: ~\$25"
    echo "  Other services: ~\$20-50"
    echo -e "  ${GREEN}Total: ~\$150-350/month${NC}"
else
    echo "  RDS PostgreSQL (t3.micro): ~\$15"
    echo "  ElastiCache Redis (t3.micro): ~\$13"
    echo "  ECS Fargate: ~\$20-50"
    echo "  Load Balancer: ~\$25"
    echo "  Other services: ~\$10-20"
    echo -e "  ${GREEN}Total: ~\$80-120/month${NC}"
fi

# Confirm
echo -e "\n${YELLOW}Ready to create AWS infrastructure?${NC}"
echo "This will create:"
echo "  - VPC with public/private subnets"
echo "  - RDS PostgreSQL database"
echo "  - ElastiCache Redis"
echo "  - Cognito User Pool"
echo "  - S3 buckets"
echo "  - ECS cluster"
echo "  - Application Load Balancer"
echo ""
echo -n "Continue? (yes/no): "
read CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    echo "Setup cancelled"
    exit 0
fi

# Initialize Terraform
echo -e "\n${YELLOW}Initializing Terraform...${NC}"
terraform init

# Create infrastructure
echo -e "\n${YELLOW}Creating infrastructure...${NC}"
terraform plan -out=tfplan

echo -e "\n${YELLOW}Applying changes...${NC}"
terraform apply tfplan

# Get outputs
echo -e "\n${GREEN}Infrastructure created successfully!${NC}"
echo "====================================="

# Save outputs to env file
terraform output -json > outputs.json

# Create .env.production file
cat > ../../../.env.production <<EOF
# AWS Infrastructure Outputs
# Generated on $(date)

# Database
DATABASE_URL=postgresql://apidirect:$DB_PASSWORD@$(terraform output -raw rds_endpoint)/apidirect
REDIS_URL=redis://:$REDIS_TOKEN@$(terraform output -raw redis_endpoint)

# AWS Cognito
NEXT_PUBLIC_AWS_REGION=$AWS_REGION
NEXT_PUBLIC_AWS_USER_POOL_ID=$(terraform output -raw cognito_user_pool_id)
NEXT_PUBLIC_AWS_USER_POOL_WEB_CLIENT_ID=$(terraform output -raw cognito_client_id)

# S3
S3_BUCKET_NAME=$(terraform output -raw s3_assets_bucket)

# API Gateway
NEXT_PUBLIC_API_URL=$(terraform output -raw api_gateway_url)

# Domain
DOMAIN=$DOMAIN_NAME
NEXTAUTH_URL=https://$DOMAIN_NAME

# Email
EMAIL_FROM=$EMAIL_FROM
EOF

echo -e "${GREEN}✓ Environment file created${NC}"

# Next steps
echo -e "\n${YELLOW}Next Steps:${NC}"
echo "1. Update your DNS records:"
echo "   - A record: $(terraform output -raw alb_dns_name)"
echo ""
echo "2. Deploy your services:"
echo "   cd ../../../"
echo "   ./deploy-aws.sh"
echo ""
echo "3. Access your services:"
echo "   - Marketplace: https://$DOMAIN_NAME"
echo "   - API: $(terraform output -raw api_gateway_url)"
echo ""
echo -e "${GREEN}Setup complete!${NC}"

# Save credentials
echo -e "\n${YELLOW}Important: Save these credentials securely:${NC}"
echo "Database Password: $DB_PASSWORD"
echo "Redis Auth Token: $REDIS_TOKEN"