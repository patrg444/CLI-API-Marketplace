#!/bin/bash

# Real Terraform Deployment Test
# This tests the actual Terraform modules

set -e

echo "=== Terraform Deployment Test ==="
echo

# Set up environment
export AWS_ACCESS_KEY_ID=AKIAQFVO6DCPFJ3ODKGA
export AWS_SECRET_ACCESS_KEY=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
export AWS_REGION=us-east-1

# Create temporary deployment directory
DEPLOY_DIR="/tmp/apidirect-test-$$"
mkdir -p "$DEPLOY_DIR"

echo "1. Copying Terraform modules..."
cp -r ../infrastructure/deployments/user-api/* "$DEPLOY_DIR/"

echo "2. Creating terraform.tfvars..."
cat > "$DEPLOY_DIR/terraform.tfvars" <<EOF
project_name = "e2e-test-api"
environment = "test"
aws_region = "us-east-1"
owner_email = "test@example.com"
api_direct_account_id = "012178036894"

# Container configuration
container_image = "e2e-test-api:latest"
container_port = 8000
health_check_path = "/health"

# Scaling
min_capacity = 1
max_capacity = 3

# Resources
cpu = 256
memory = 512

# Disable database for test
enable_database = false
EOF

echo "3. Initializing Terraform..."
cd "$DEPLOY_DIR"
terraform init

echo "4. Creating deployment plan..."
terraform plan -out=tfplan

echo "5. Showing planned resources..."
terraform show tfplan | grep -E "will be created|Plan:" | head -20

echo
echo "=== Terraform Test Summary ==="
echo "✓ Terraform modules copied successfully"
echo "✓ Configuration variables set"
echo "✓ Terraform initialized"
echo "✓ Deployment plan created"
echo
echo "The plan shows all resources that would be created."
echo "To apply this plan, you would run: terraform apply tfplan"
echo
echo "Cleaning up test files..."
rm -rf "$DEPLOY_DIR"

echo "Test completed successfully!"