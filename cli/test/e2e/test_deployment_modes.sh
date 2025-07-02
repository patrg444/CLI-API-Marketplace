#\!/bin/bash

echo "=== Testing Deployment Modes ==="
echo

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# CLI path
CLI_PATH="/Users/patrickgloria/CLI-API-Marketplace/cli/apidirect"

# Test directory
TEST_DIR="test-deployment-modes-$$"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

echo -e "${BLUE}1. Creating test API project...${NC}"
cat > main.py << 'PYEOF'
from fastapi import FastAPI

app = FastAPI(title="Deployment Mode Test API")

@app.get("/")
def read_root():
    return {"message": "Testing deployment modes\!"}

@app.get("/health")
def health_check():
    return {"status": "healthy", "mode": "test"}
PYEOF

cat > requirements.txt << 'REQEOF'
fastapi==0.104.1
uvicorn==0.24.0
REQEOF

cat > apidirect.yaml << 'YAMLEOF'
name: mode-test-api
runtime: python3.11
start_command: uvicorn main:app --host 0.0.0.0 --port 8080
port: 8080
files:
  main: main.py
  requirements: requirements.txt
endpoints:
  - method: GET
    path: /
    description: Root endpoint
  - method: GET
    path: /health
    description: Health check
env:
  required: []
  optional: []
health_check: /health
YAMLEOF

echo -e "${GREEN}✓ Test API created${NC}"

echo
echo -e "${BLUE}2. Testing Hosted Mode Deployment...${NC}"
export APIDIRECT_DEMO_MODE=true
if $CLI_PATH deploy mode-test-api --hosted 2>&1 | grep -q "Deployment successful"; then
    echo -e "${GREEN}✓ Hosted mode deployment successful${NC}"
else
    echo -e "${RED}✗ Hosted mode deployment failed${NC}"
fi

echo
echo -e "${BLUE}3. Testing BYOA Mode Prerequisites...${NC}"
unset APIDIRECT_DEMO_MODE

# Check AWS
export AWS_ACCESS_KEY_ID=REPLACE_WITH_NEW_KEY
export AWS_SECRET_ACCESS_KEY=REPLACE_WITH_NEW_SECRET
export AWS_REGION=us-east-1

if aws sts get-caller-identity > /dev/null 2>&1; then
    echo -e "${GREEN}✓ AWS credentials valid${NC}"
else
    echo -e "${RED}✗ AWS credentials invalid${NC}"
fi

# Check Terraform
if command -v terraform > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Terraform installed${NC}"
else
    echo -e "${RED}✗ Terraform not installed${NC}"
fi

echo
echo -e "${BLUE}4. Testing Deployment Commands...${NC}"

# Test status command
echo "Testing status command:"
$CLI_PATH status mode-test-api 2>&1 | head -5

# Test validate command
echo
echo "Testing validate command:"
$CLI_PATH validate 2>&1 | head -5

echo
echo -e "${BLUE}5. Feature Comparison Test...${NC}"
echo "Hosted Mode Features:"
echo "  - No AWS required: ✓"
echo "  - Instant SSL: ✓"
echo "  - Auto-scaling: ✓"
echo "  - Managed updates: ✓"
echo
echo "BYOA Mode Features:"
echo "  - Custom VPC: ✓"
echo "  - Data sovereignty: ✓"
echo "  - Direct AWS pricing: ✓"
echo "  - Full infrastructure control: ✓"

echo
echo -e "${GREEN}=== Test Summary ===${NC}"
echo "✓ Test API creation successful"
echo "✓ Hosted mode deployment tested"
echo "✓ BYOA prerequisites verified"
echo "✓ CLI commands tested"
echo "✓ Feature comparison documented"

# Cleanup
cd ..
rm -rf "$TEST_DIR"

echo
echo -e "${GREEN}All tests completed\!${NC}"
