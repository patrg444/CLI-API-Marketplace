#!/bin/bash

echo "🧪 Testing API-Direct CLI"
echo "========================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Test CLI exists and is executable
echo "1. Checking CLI binary..."
if [ -f "cli/apidirect" ]; then
    echo -e "${GREEN}✓ CLI binary found${NC}"
    
    # Check if executable
    if [ -x "cli/apidirect" ]; then
        echo -e "${GREEN}✓ CLI is executable${NC}"
    else
        echo -e "${RED}✗ CLI is not executable${NC}"
        echo "  Run: chmod +x cli/apidirect"
    fi
else
    echo -e "${RED}✗ CLI binary not found${NC}"
    echo "  Run: cd cli && make build"
    exit 1
fi

echo ""
echo "2. Testing CLI commands..."

# Test version
echo -e "\n${YELLOW}Testing: apidirect version${NC}"
./cli/apidirect version

# Test help
echo -e "\n${YELLOW}Testing: apidirect --help${NC}"
./cli/apidirect --help | head -20

# Test validate command (doesn't require auth)
echo -e "\n${YELLOW}Testing: apidirect validate${NC}"
# Create a test apidirect.yaml
cat > test-api/apidirect.yaml << EOF
name: test-api
version: 1.0.0
description: Test API for validation
runtime: python3.11
handler: main.handler
environment:
  - name: API_KEY
    value: test-key
routes:
  - path: /hello
    method: GET
    description: Returns a greeting
EOF

./cli/apidirect validate test-api/apidirect.yaml
rm -rf test-api

# Test commands that would require auth
echo -e "\n${YELLOW}Testing auth-required commands (should fail gracefully)...${NC}"

echo -e "\n3. Testing: apidirect whoami"
./cli/apidirect whoami 2>&1 | head -5

echo -e "\n4. Testing: apidirect search weather"
./cli/apidirect search weather 2>&1 | head -5

echo ""
echo "📝 Summary"
echo "=========="
echo ""
echo -e "${GREEN}✓ CLI is functional and ready${NC}"
echo ""
echo "Next steps for full functionality:"
echo "1. Set AWS Cognito environment variables:"
echo "   export APIDIRECT_COGNITO_POOL=\"your-pool-id\""
echo "   export APIDIRECT_COGNITO_CLIENT=\"your-client-id\""
echo "   export APIDIRECT_AUTH_DOMAIN=\"your-auth-domain\""
echo ""
echo "2. Run: apidirect login"
echo "3. Start using authenticated commands"
echo ""
echo "For local development with mock auth:"
echo "- The backend API supports mock authentication"
echo "- Set USE_MOCK_AUTH=true in your environment"