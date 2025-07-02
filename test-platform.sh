#!/bin/bash

echo "🧪 Testing API-Direct Platform Components"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "1. Testing AWS Connection..."
source .env
aws sts get-caller-identity 2>&1 | head -3
echo ""

echo "2. Testing S3 Buckets..."
aws s3 ls s3://$CODE_STORAGE_BUCKET 2>&1 | head -3
echo ""

echo "3. Testing Cognito..."
aws cognito-idp describe-user-pool --user-pool-id $COGNITO_USER_POOL_ID 2>&1 | head -10
echo ""

echo "4. Testing CLI..."
./cli/apidirect version
echo ""

echo "5. Web Applications Status:"
echo "  - Landing: https://apidirect.dev"
echo "  - Console: https://console.apidirect.dev"
echo "  - Marketplace: https://marketplace.apidirect.dev"
echo "  - Docs: https://docs.apidirect.dev"
echo ""

echo -e "${GREEN}✅ Platform Configuration Summary${NC}"
echo "=================================="
echo "AWS Account: 723595141930"
echo "Region: $AWS_REGION"
echo "Cognito Pool: $COGNITO_USER_POOL_ID"
echo "S3 Bucket: $CODE_STORAGE_BUCKET"
echo ""

echo "To start using the platform:"
echo "1. Visit https://console.apidirect.dev"
echo "2. Login with demo@apidirect.dev / secret (mock auth)"
echo "3. Or use CLI: ./cli/apidirect login (requires Cognito setup)"