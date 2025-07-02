#!/bin/bash

# Comprehensive Button Testing Suite Runner
# This script runs all button-related tests with proper configuration

set -e

echo "🔘 Starting Comprehensive Button Testing Suite..."

# Set test environment variables
export NODE_ENV=test
export PWTEST_DEBUG=0
export PLAYWRIGHT_BROWSERS_PATH=0

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test categories
TESTS=(
    "ui-components/button-interactions.spec.ts"
    "ui-components/form-button-validation.spec.ts" 
    "ui-components/accessibility-button-tests.spec.ts"
    "ui-components/button-performance.spec.ts"
    "ui-components/button-edge-cases.spec.ts"
)

# Test descriptions
declare -A TEST_DESCRIPTIONS=(
    ["button-interactions.spec.ts"]="Basic button interactions and functionality"
    ["form-button-validation.spec.ts"]="Form button states and validation"
    ["accessibility-button-tests.spec.ts"]="Button accessibility and screen reader support"
    ["button-performance.spec.ts"]="Button rendering and interaction performance"
    ["button-edge-cases.spec.ts"]="Edge cases and error handling"
)

# Results tracking
PASSED_TESTS=()
FAILED_TESTS=()
TOTAL_TESTS=${#TESTS[@]}

echo -e "${BLUE}📋 Running ${TOTAL_TESTS} button test suites...${NC}"
echo ""

# Function to run individual test
run_test() {
    local test_file=$1
    local test_name=$(basename "$test_file" .spec.ts)
    local description=${TEST_DESCRIPTIONS[$test_name.spec.ts]}
    
    echo -e "${YELLOW}🧪 Running: ${test_name}${NC}"
    echo -e "   📝 ${description}"
    echo ""
    
    # Run the test with detailed output
    if npx playwright test "$test_file" --reporter=list --timeout=60000; then
        echo -e "${GREEN}✅ PASSED: ${test_name}${NC}"
        PASSED_TESTS+=("$test_name")
    else
        echo -e "${RED}❌ FAILED: ${test_name}${NC}"
        FAILED_TESTS+=("$test_name")
    fi
    
    echo ""
    echo "----------------------------------------"
    echo ""
}

# Check if Playwright is installed
if ! command -v npx playwright &> /dev/null; then
    echo -e "${RED}❌ Playwright not found. Installing...${NC}"
    npm install --save-dev playwright @playwright/test
    npx playwright install
fi

# Ensure test environment is ready
echo -e "${BLUE}🔧 Setting up test environment...${NC}"

# Create test results directory
mkdir -p test-results/button-tests

# Check if development server is running
if ! curl -s http://localhost:3000 > /dev/null 2>&1 && ! curl -s http://localhost:3001 > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Development server not detected. Please start the server:${NC}"
    echo "   npm run dev"
    echo ""
    echo "Attempting to start server..."
    
    # Try to start server in background
    npm run dev &
    SERVER_PID=$!
    
    # Wait for server to start
    echo "Waiting for server to start..."
    for i in {1..30}; do
        if curl -s http://localhost:3000 > /dev/null 2>&1 || curl -s http://localhost:3001 > /dev/null 2>&1; then
            echo -e "${GREEN}✅ Server is running${NC}"
            break
        fi
        sleep 2
        echo -n "."
    done
    echo ""
fi

# Verify server is accessible
if curl -s http://localhost:3000 > /dev/null 2>&1; then
    BASE_URL="http://localhost:3000"
elif curl -s http://localhost:3001 > /dev/null 2>&1; then
    BASE_URL="http://localhost:3001"
else
    echo -e "${RED}❌ Cannot access development server. Please start it manually.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Using server at: ${BASE_URL}${NC}"
echo ""

# Set base URL for tests
export PLAYWRIGHT_BASE_URL=$BASE_URL

# Run each test suite
for test_file in "${TESTS[@]}"; do
    run_test "$test_file"
done

# Generate summary report
echo "========================================"
echo -e "${BLUE}📊 BUTTON TESTING SUMMARY${NC}"
echo "========================================"
echo ""

echo -e "Total Test Suites: ${TOTAL_TESTS}"
echo -e "${GREEN}Passed: ${#PASSED_TESTS[@]}${NC}"
echo -e "${RED}Failed: ${#FAILED_TESTS[@]}${NC}"
echo ""

if [ ${#PASSED_TESTS[@]} -gt 0 ]; then
    echo -e "${GREEN}✅ Passed Tests:${NC}"
    for test in "${PASSED_TESTS[@]}"; do
        echo -e "   • $test"
    done
    echo ""
fi

if [ ${#FAILED_TESTS[@]} -gt 0 ]; then
    echo -e "${RED}❌ Failed Tests:${NC}"
    for test in "${FAILED_TESTS[@]}"; do
        echo -e "   • $test"
    done
    echo ""
    echo -e "${RED}Please review the test output above for detailed error information.${NC}"
fi

# Calculate success rate
SUCCESS_RATE=$((${#PASSED_TESTS[@]} * 100 / TOTAL_TESTS))
echo -e "Success Rate: ${SUCCESS_RATE}%"

# Generate detailed HTML report
echo ""
echo -e "${BLUE}📋 Generating detailed HTML report...${NC}"
npx playwright show-report

# Clean up
if [ ! -z "$SERVER_PID" ]; then
    echo ""
    echo -e "${YELLOW}🛑 Stopping test server...${NC}"
    kill $SERVER_PID 2>/dev/null || true
fi

echo ""
if [ ${#FAILED_TESTS[@]} -eq 0 ]; then
    echo -e "${GREEN}🎉 All button tests passed! Your buttons are working perfectly.${NC}"
    exit 0
else
    echo -e "${RED}⚠️  Some button tests failed. Please review and fix the issues.${NC}"
    exit 1
fi