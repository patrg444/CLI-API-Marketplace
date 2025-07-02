#!/bin/bash
# Pre-flight check script for AWS deployment

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}API Direct Marketplace - Pre-flight Check${NC}"
echo "=========================================="

ERRORS=0
WARNINGS=0

# Function to check command exists
check_command() {
    if command -v $1 &> /dev/null; then
        echo -e "${GREEN}✓${NC} $2 found: $(command -v $1)"
        return 0
    else
        echo -e "${RED}✗${NC} $2 not found. Please install $1"
        ERRORS=$((ERRORS + 1))
        return 1
    fi
}

# Function to check version
check_version() {
    local cmd=$1
    local name=$2
    local min_version=$3
    
    if command -v $cmd &> /dev/null; then
        version=$($cmd --version | grep -oE '[0-9]+\.[0-9]+' | head -1)
        echo -e "${GREEN}✓${NC} $name version: $version"
    else
        echo -e "${RED}✗${NC} $name not found"
        ERRORS=$((ERRORS + 1))
    fi
}

# Function to run test
run_test() {
    if $1 &> /dev/null; then
        echo -e "${GREEN}✓${NC} $2"
        return 0
    else
        echo -e "${RED}✗${NC} $2"
        ERRORS=$((ERRORS + 1))
        return 1
    fi
}

# Function to run warning test
run_warning() {
    if $1 &> /dev/null; then
        echo -e "${GREEN}✓${NC} $2"
        return 0
    else
        echo -e "${YELLOW}⚠${NC} $2"
        WARNINGS=$((WARNINGS + 1))
        return 1
    fi
}

echo -e "\n${YELLOW}1. Checking Prerequisites${NC}"
echo "========================"

# Check required commands
check_command "aws" "AWS CLI"
check_command "terraform" "Terraform"
check_command "docker" "Docker"
check_command "node" "Node.js"
check_command "npm" "NPM"
check_command "git" "Git"

# Check versions
echo -e "\n${YELLOW}2. Checking Versions${NC}"
echo "==================="
check_version "aws" "AWS CLI" "2.0"
check_version "terraform" "Terraform" "1.0"
check_version "docker" "Docker" "20.0"
check_version "node" "Node.js" "16.0"

# Check AWS credentials
echo -e "\n${YELLOW}3. Checking AWS Configuration${NC}"
echo "============================"

if aws sts get-caller-identity &> /dev/null; then
    ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
    REGION=$(aws configure get region)
    USER=$(aws sts get-caller-identity --query Arn --output text | cut -d'/' -f2)
    
    echo -e "${GREEN}✓${NC} AWS credentials configured"
    echo "  Account ID: $ACCOUNT_ID"
    echo "  Region: $REGION"
    echo "  User: $USER"
    
    # Check if this is the correct account
    if [[ "$ACCOUNT_ID" == "0121-7803-6894" ]]; then
        echo -e "${GREEN}✓${NC} Correct AWS account"
    else
        echo -e "${YELLOW}⚠${NC} Different account than expected (0121-7803-6894)"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    echo -e "${RED}✗${NC} AWS credentials not configured"
    echo "  Run: aws configure"
    ERRORS=$((ERRORS + 1))
fi

# Check IAM permissions
echo -e "\n${YELLOW}4. Checking AWS Permissions${NC}"
echo "=========================="

run_test "aws iam get-user" "IAM read access"
run_test "aws ec2 describe-regions" "EC2 access"
run_test "aws s3 ls" "S3 access"
run_warning "aws rds describe-db-instances" "RDS access"
run_warning "aws ecs list-clusters" "ECS access"

# Check Docker
echo -e "\n${YELLOW}5. Checking Docker${NC}"
echo "================="

if docker info &> /dev/null; then
    echo -e "${GREEN}✓${NC} Docker daemon running"
    
    # Check Docker Compose
    if docker compose version &> /dev/null; then
        echo -e "${GREEN}✓${NC} Docker Compose v2 available"
    elif docker-compose --version &> /dev/null; then
        echo -e "${YELLOW}⚠${NC} Docker Compose v1 (consider upgrading)"
        WARNINGS=$((WARNINGS + 1))
    else
        echo -e "${RED}✗${NC} Docker Compose not found"
        ERRORS=$((ERRORS + 1))
    fi
else
    echo -e "${RED}✗${NC} Docker daemon not running"
    echo "  Start Docker Desktop or run: sudo systemctl start docker"
    ERRORS=$((ERRORS + 1))
fi

# Check project structure
echo -e "\n${YELLOW}6. Checking Project Structure${NC}"
echo "============================"

run_test "test -f setup-aws.sh" "AWS setup script exists"
run_test "test -d infrastructure/terraform/aws" "Terraform configuration exists"
run_test "test -d services" "Services directory exists"
run_test "test -d web/marketplace" "Web marketplace exists"
run_test "test -f docker-compose.production.yml" "Production Docker compose file exists"

# Check Node.js project
echo -e "\n${YELLOW}7. Checking Node.js Project${NC}"
echo "=========================="

if [[ -f "web/marketplace/package.json" ]]; then
    echo -e "${GREEN}✓${NC} package.json found"
    
    cd web/marketplace
    if [[ -d "node_modules" ]]; then
        echo -e "${GREEN}✓${NC} Node modules installed"
    else
        echo -e "${YELLOW}⚠${NC} Node modules not installed"
        echo "  Run: cd web/marketplace && npm install"
        WARNINGS=$((WARNINGS + 1))
    fi
    cd ../..
else
    echo -e "${RED}✗${NC} package.json not found"
    ERRORS=$((ERRORS + 1))
fi

# Check environment files
echo -e "\n${YELLOW}8. Checking Environment Files${NC}"
echo "============================"

if [[ -f ".env.production.example" ]]; then
    echo -e "${GREEN}✓${NC} Production environment template exists"
else
    echo -e "${YELLOW}⚠${NC} Production environment template missing"
    WARNINGS=$((WARNINGS + 1))
fi

# Check AWS service quotas
echo -e "\n${YELLOW}9. Checking AWS Service Quotas${NC}"
echo "============================="

# Check VPC quota
VPC_COUNT=$(aws ec2 describe-vpcs --query 'length(Vpcs)' --output text 2>/dev/null || echo "0")
echo "  VPCs used: $VPC_COUNT/5 (default limit)"

# Estimate costs
echo -e "\n${YELLOW}10. Estimated AWS Costs${NC}"
echo "======================"
echo "Development environment:"
echo "  - RDS (db.t3.micro): ~$15/month"
echo "  - ElastiCache (cache.t3.micro): ~$13/month"
echo "  - ECS Fargate: ~$20-50/month"
echo "  - Load Balancer: ~$25/month"
echo "  - Other services: ~$10-20/month"
echo -e "  ${BLUE}Total: ~$80-120/month${NC}"

# Summary
echo -e "\n${YELLOW}Summary${NC}"
echo "======="

if [[ $ERRORS -eq 0 ]]; then
    if [[ $WARNINGS -eq 0 ]]; then
        echo -e "${GREEN}✅ All checks passed! Ready to deploy.${NC}"
        echo -e "\nNext step: ${BLUE}./setup-aws.sh${NC}"
    else
        echo -e "${YELLOW}⚠ Passed with $WARNINGS warnings${NC}"
        echo -e "\nYou can proceed with: ${BLUE}./setup-aws.sh${NC}"
        echo "But review the warnings above first."
    fi
    exit 0
else
    echo -e "${RED}❌ Failed with $ERRORS errors and $WARNINGS warnings${NC}"
    echo -e "\nPlease fix the errors above before proceeding."
    exit 1
fi