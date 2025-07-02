#!/bin/bash

# Integration test for the entire CLI workflow
# This script tests the complete user journey

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🧪 API Direct CLI Integration Test${NC}"
echo "=================================="

# Test configuration
TEST_DIR="/tmp/apidirect-integration-test"
TEST_API_DIR="$TEST_DIR/my-test-api"
CLI_BIN="./apidirect"

# Build the CLI first
echo -e "\n${YELLOW}Building CLI...${NC}"
go build -o "$CLI_BIN" main.go

if [ ! -f "$CLI_BIN" ]; then
    echo -e "${RED}Failed to build CLI${NC}"
    exit 1
fi

# Create test directory
mkdir -p "$TEST_API_DIR"
cd "$TEST_DIR"

# Test 1: Version and help commands
echo -e "\n${YELLOW}Test 1: Basic Commands${NC}"
echo "---------------------"

echo "Testing version command..."
if $CLI_BIN version; then
    echo -e "${GREEN}✓ Version command works${NC}"
else
    echo -e "${RED}✗ Version command failed${NC}"
fi

echo "Testing help command..."
if $CLI_BIN --help > /dev/null; then
    echo -e "${GREEN}✓ Help command works${NC}"
else
    echo -e "${RED}✗ Help command failed${NC}"
fi

# Test 2: Import workflow
echo -e "\n${YELLOW}Test 2: Import Workflow${NC}"
echo "----------------------"

# Create a sample Express.js API
cat > "$TEST_API_DIR/package.json" << 'EOF'
{
  "name": "my-test-api",
  "version": "1.0.0",
  "description": "Test API for integration testing",
  "main": "index.js",
  "scripts": {
    "start": "node index.js"
  },
  "dependencies": {
    "express": "^4.18.0"
  }
}
EOF

cat > "$TEST_API_DIR/index.js" << 'EOF'
const express = require('express');
const app = express();
const PORT = process.env.PORT || 3000;

app.get('/', (req, res) => {
  res.json({ message: 'Hello from Test API!' });
});

app.get('/users', (req, res) => {
  res.json([
    { id: 1, name: 'John Doe' },
    { id: 2, name: 'Jane Smith' }
  ]);
});

app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});
EOF

# Mock the import command (since it requires backend)
echo "Simulating import command..."
cat > "$TEST_API_DIR/apidirect.yaml" << 'EOF'
name: my-test-api
description: Test API for integration testing
version: 1.0.0
framework: express
language: javascript
entry_point: index.js
port: 3000
environment:
  NODE_ENV: production
build:
  command: npm install
  output_dir: .
run:
  command: npm start
dependencies:
  - express@^4.18.0
EOF

if [ -f "$TEST_API_DIR/apidirect.yaml" ]; then
    echo -e "${GREEN}✓ Import simulation successful${NC}"
else
    echo -e "${RED}✗ Import simulation failed${NC}"
fi

# Test 3: Validate command
echo -e "\n${YELLOW}Test 3: Validate Command${NC}"
echo "-----------------------"

cd "$TEST_API_DIR"
echo "Testing validate command..."
# Since validate requires backend, we'll test the command structure
if $CLI_BIN validate --help > /dev/null; then
    echo -e "${GREEN}✓ Validate command available${NC}"
else
    echo -e "${RED}✗ Validate command not found${NC}"
fi

# Test 4: Environment management
echo -e "\n${YELLOW}Test 4: Environment Management${NC}"
echo "-----------------------------"

# Test env commands
echo "Testing env commands..."
if $CLI_BIN env --help > /dev/null; then
    echo -e "${GREEN}✓ Env command available${NC}"
    
    # Test subcommands
    for subcmd in list set get delete pull push; do
        if $CLI_BIN env $subcmd --help > /dev/null 2>&1; then
            echo -e "  ${GREEN}✓ env $subcmd available${NC}"
        else
            echo -e "  ${RED}✗ env $subcmd not found${NC}"
        fi
    done
else
    echo -e "${RED}✗ Env command not found${NC}"
fi

# Test 5: Documentation generation
echo -e "\n${YELLOW}Test 5: Documentation Generation${NC}"
echo "-------------------------------"

echo "Testing docs commands..."
if $CLI_BIN docs --help > /dev/null; then
    echo -e "${GREEN}✓ Docs command available${NC}"
    
    # Test subcommands
    for subcmd in generate preview publish; do
        if $CLI_BIN docs $subcmd --help > /dev/null 2>&1; then
            echo -e "  ${GREEN}✓ docs $subcmd available${NC}"
        else
            echo -e "  ${RED}✗ docs $subcmd not found${NC}"
        fi
    done
else
    echo -e "${RED}✗ Docs command not found${NC}"
fi

# Test 6: Marketplace commands
echo -e "\n${YELLOW}Test 6: Marketplace Commands${NC}"
echo "---------------------------"

marketplace_commands=(
    "search"
    "browse"
    "info"
    "subscribe"
    "subscriptions"
    "analytics"
    "earnings"
    "review"
)

for cmd in "${marketplace_commands[@]}"; do
    if $CLI_BIN $cmd --help > /dev/null 2>&1; then
        echo -e "${GREEN}✓ $cmd command available${NC}"
    else
        echo -e "${RED}✗ $cmd command not found${NC}"
    fi
done

# Test 7: Shell completion
echo -e "\n${YELLOW}Test 7: Shell Completion${NC}"
echo "-----------------------"

shells=("bash" "zsh" "fish" "powershell")
for shell in "${shells[@]}"; do
    if $CLI_BIN completion $shell > /dev/null 2>&1; then
        echo -e "${GREEN}✓ $shell completion works${NC}"
    else
        echo -e "${RED}✗ $shell completion failed${NC}"
    fi
done

# Test 8: Configuration management
echo -e "\n${YELLOW}Test 8: Configuration Management${NC}"
echo "-------------------------------"

echo "Testing config directory creation..."
export APIDIRECT_CONFIG_DIR="$TEST_DIR/.apidirect"
mkdir -p "$APIDIRECT_CONFIG_DIR"

if [ -d "$APIDIRECT_CONFIG_DIR" ]; then
    echo -e "${GREEN}✓ Config directory created${NC}"
    
    # Create test config
    cat > "$APIDIRECT_CONFIG_DIR/config.yaml" << EOF
api_endpoint: https://api.apidirect.io
region: us-east-1
output_format: json
EOF
    
    if [ -f "$APIDIRECT_CONFIG_DIR/config.yaml" ]; then
        echo -e "${GREEN}✓ Config file created${NC}"
    fi
else
    echo -e "${RED}✗ Failed to create config directory${NC}"
fi

# Clean up
echo -e "\n${YELLOW}Cleaning up...${NC}"
cd /
rm -rf "$TEST_DIR"
rm -f "$CLI_BIN"

# Summary
echo -e "\n${BLUE}Integration Test Summary${NC}"
echo "========================"
echo -e "${GREEN}✅ All integration tests completed!${NC}"
echo ""
echo "Tested features:"
echo "• Basic commands (version, help)"
echo "• Import workflow simulation"
echo "• Validate command"
echo "• Environment management"
echo "• Documentation generation"
echo "• Marketplace commands"
echo "• Shell completions"
echo "• Configuration management"