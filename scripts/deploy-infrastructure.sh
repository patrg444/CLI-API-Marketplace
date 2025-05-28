#!/bin/bash

# API-Direct Infrastructure Deployment Script
# This script handles the complete infrastructure deployment process

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== API-Direct Infrastructure Deployment ===${NC}"

# Check prerequisites
echo -e "\n${YELLOW}Checking prerequisites...${NC}"

# Check if AWS CLI is installed
if ! command -v aws &> /dev/null; then
    echo -e "${RED}AWS CLI is not installed. Please install it first.${NC}"
    exit 1
fi

# Check if Terraform is installed
if ! command -v terraform &> /dev/null; then
    echo -e "${RED}Terraform is not installed. Please install it first.${NC}"
    exit 1
fi

# Check AWS credentials
if ! aws sts get-caller-identity &> /dev/null; then
    echo -e "${RED}AWS credentials are not configured. Please run 'aws configure'.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ All prerequisites met${NC}"

# Navigate to terraform directory
cd "$(dirname "$0")/../infrastructure/terraform"

# Initialize Terraform
echo -e "\n${YELLOW}Initializing Terraform...${NC}"
terraform init

# Validate configuration
echo -e "\n${YELLOW}Validating Terraform configuration...${NC}"
terraform validate

# Check if terraform.tfvars exists
if [ ! -f "terraform.tfvars" ]; then
    echo -e "\n${YELLOW}terraform.tfvars not found. Creating from example...${NC}"
    if [ -f "terraform.tfvars.example" ]; then
        cp terraform.tfvars.example terraform.tfvars
        echo -e "${YELLOW}Please edit terraform.tfvars with your specific values before continuing.${NC}"
        exit 1
    else
        echo -e "${RED}terraform.tfvars.example not found. Please create terraform.tfvars manually.${NC}"
        exit 1
    fi
fi

# Plan the deployment
echo -e "\n${YELLOW}Planning infrastructure deployment...${NC}"
terraform plan -out=tfplan

# Ask for confirmation
echo -e "\n${YELLOW}Do you want to apply this plan? (yes/no)${NC}"
read -r response
if [[ ! "$response" =~ ^[Yy][Ee][Ss]$ ]]; then
    echo -e "${RED}Deployment cancelled.${NC}"
    exit 1
fi

# Apply the plan
echo -e "\n${YELLOW}Deploying infrastructure...${NC}"
terraform apply tfplan

# Save outputs
echo -e "\n${YELLOW}Saving Terraform outputs...${NC}"
terraform output -json > ../../outputs.json

echo -e "\n${GREEN}=== Infrastructure deployment complete! ===${NC}"
echo -e "${GREEN}Outputs saved to outputs.json${NC}"
echo -e "\n${YELLOW}Next steps:${NC}"
echo -e "1. Run ./scripts/configure-cli-env.sh to set up CLI environment"
echo -e "2. Build the CLI: cd cli && go build -o apidirect"
echo -e "3. Test authentication: ./apidirect login"
