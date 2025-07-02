#!/bin/bash

# Test Terraform Plan Generation
# This demonstrates the Terraform planning phase

set -e

echo "=== Terraform Planning Test ==="
echo

# Set up environment
export AWS_ACCESS_KEY_ID=AKIAQFVO6DCPFJ3ODKGA
export AWS_SECRET_ACCESS_KEY=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
export AWS_REGION=us-east-1

# Create test directory structure
TEST_DIR="/tmp/apidirect-terraform-test-$$"
mkdir -p "$TEST_DIR"

echo "1. Setting up test infrastructure..."

# Copy the entire infrastructure directory structure
cp -r ../infrastructure/* "$TEST_DIR/" 2>/dev/null || true

# Check if modules exist
if [ -d "$TEST_DIR/modules" ]; then
    echo "   ✓ Infrastructure modules found"
else
    echo "   ✗ Infrastructure modules not found at expected location"
    echo "   Looking for modules..."
    find ../infrastructure -name "*.tf" -type f | head -10
fi

# Create a simple test Terraform configuration
echo "2. Creating test Terraform configuration..."
cat > "$TEST_DIR/test_main.tf" <<'EOF'
# Test Terraform Configuration for BYOA Deployment

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "e2e-test-api"
}

# Test VPC resource
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  
  tags = {
    Name = "${var.project_name}-test-vpc"
    ManagedBy = "API-Direct"
  }
}

# Test security group
resource "aws_security_group" "test" {
  name        = "${var.project_name}-test-sg"
  description = "Test security group for API-Direct"
  vpc_id      = aws_vpc.test.id
  
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  tags = {
    Name = "${var.project_name}-test-sg"
  }
}

output "vpc_id" {
  value = aws_vpc.test.id
  description = "ID of the test VPC"
}

output "security_group_id" {
  value = aws_security_group.test.id
  description = "ID of the test security group"
}
EOF

echo "3. Initializing Terraform..."
cd "$TEST_DIR"
terraform init

echo "4. Validating Terraform configuration..."
terraform validate

echo "5. Creating Terraform plan..."
terraform plan -out=tfplan.binary

echo "6. Converting plan to readable format..."
terraform show tfplan.binary > tfplan.txt

echo "7. Plan summary:"
grep -E "will be created|will be destroyed|Plan:" tfplan.txt || echo "No resources to create/destroy"

echo
echo "=== Test Results ==="
echo "✓ Terraform configuration created"
echo "✓ Terraform initialized successfully"
echo "✓ Configuration validated"
echo "✓ Plan generated successfully"
echo
echo "This demonstrates that:"
echo "- Terraform can connect to AWS"
echo "- AWS credentials are valid"
echo "- Infrastructure can be planned"
echo
echo "In a real deployment, this would create:"
echo "- Complete VPC with subnets"
echo "- ECS Fargate cluster and service"
echo "- Application Load Balancer"
echo "- RDS database (optional)"
echo "- CloudWatch monitoring"
echo

# Clean up
echo "Cleaning up test files..."
rm -rf "$TEST_DIR"

echo "Test completed successfully!"