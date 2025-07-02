#!/bin/bash

echo "🔍 Testing AWS Connection"
echo "========================"
echo ""

# Load environment variables
source .env

# Export AWS credentials
export AWS_ACCESS_KEY_ID
export AWS_SECRET_ACCESS_KEY
export AWS_REGION

# Test AWS connection
echo "Testing AWS credentials..."
aws sts get-caller-identity 2>&1 | head -10

echo ""
echo "Checking S3 access..."
aws s3 ls 2>&1 | head -10

echo ""
echo "Checking Cognito access..."
aws cognito-idp list-user-pools --max-results 10 2>&1 | head -10

echo ""
echo "If you see errors above, the credentials may be invalid or lack permissions."
echo "For now, we can continue with mock authentication for local development."