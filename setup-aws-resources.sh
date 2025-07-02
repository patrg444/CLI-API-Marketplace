#!/bin/bash

echo "🚀 Setting up AWS Resources for API-Direct"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Load environment
source .env

echo "Account ID: 723595141930"
echo "Region: $AWS_REGION"
echo ""

# Generate unique suffixes
TIMESTAMP=$(date +%s)
RANDOM_SUFFIX=$(echo $RANDOM | md5sum | head -c 8)

echo "📦 Creating S3 Buckets..."
echo "========================="

# Create code storage bucket
BUCKET_NAME="apidirect-code-storage-${RANDOM_SUFFIX}"
echo "Creating bucket: $BUCKET_NAME"
aws s3 mb s3://$BUCKET_NAME --region $AWS_REGION

# Create artifacts bucket
ARTIFACTS_BUCKET="apidirect-artifacts-${RANDOM_SUFFIX}"
echo "Creating bucket: $ARTIFACTS_BUCKET"
aws s3 mb s3://$ARTIFACTS_BUCKET --region $AWS_REGION

echo ""
echo "🔐 Creating Cognito User Pool..."
echo "================================"

# Create Cognito User Pool
POOL_NAME="apidirect-users-${RANDOM_SUFFIX}"
echo "Creating user pool: $POOL_NAME"

USER_POOL_RESPONSE=$(aws cognito-idp create-user-pool \
  --pool-name "$POOL_NAME" \
  --auto-verified-attributes email \
  --username-attributes email \
  --mfa-configuration "OFF" \
  --email-configuration EmailSendingAccount=COGNITO_DEFAULT \
  --schema Name=email,Required=true,Mutable=false \
  --policies '{
    "PasswordPolicy": {
      "MinimumLength": 8,
      "RequireUppercase": true,
      "RequireLowercase": true,
      "RequireNumbers": true,
      "RequireSymbols": false
    }
  }' \
  --region $AWS_REGION)

USER_POOL_ID=$(echo $USER_POOL_RESPONSE | grep -o '"Id": "[^"]*' | grep -o '[^"]*$')
echo "Created User Pool: $USER_POOL_ID"

# Create App Client
echo "Creating app client..."
CLIENT_RESPONSE=$(aws cognito-idp create-user-pool-client \
  --user-pool-id $USER_POOL_ID \
  --client-name "apidirect-cli" \
  --explicit-auth-flows ALLOW_USER_PASSWORD_AUTH ALLOW_REFRESH_TOKEN_AUTH \
  --region $AWS_REGION)

CLIENT_ID=$(echo $CLIENT_RESPONSE | grep -o '"ClientId": "[^"]*' | grep -o '[^"]*$')
echo "Created Client: $CLIENT_ID"

echo ""
echo "✅ AWS Resources Created Successfully!"
echo "====================================="
echo ""
echo "Add these to your .env file:"
echo ""
echo "# S3 Buckets"
echo "CODE_STORAGE_BUCKET=$BUCKET_NAME"
echo "ARTIFACTS_BUCKET=$ARTIFACTS_BUCKET"
echo ""
echo "# Cognito"
echo "COGNITO_USER_POOL_ID=$USER_POOL_ID"
echo "COGNITO_CLIENT_ID=$CLIENT_ID"
echo "COGNITO_REGION=$AWS_REGION"
echo ""
echo "# For CLI"
echo "APIDIRECT_COGNITO_POOL=$USER_POOL_ID"
echo "APIDIRECT_COGNITO_CLIENT=$CLIENT_ID"
echo "APIDIRECT_REGION=$AWS_REGION"
echo ""
echo "# Switch to real auth when ready"
echo "USE_MOCK_AUTH=false"
echo ""
echo "Save these values! They're needed for the platform to work."