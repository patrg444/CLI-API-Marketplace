#!/bin/bash

# API-Direct End-to-End Test Runner
# This script orchestrates the entire test environment setup and execution

set -e  # Exit on any error

echo "🚀 Starting API-Direct E2E Test Environment..."
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check prerequisites
echo "Checking prerequisites..."

# Check Docker
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed or not in PATH"
    exit 1
fi
print_status "Docker found"

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is not installed or not in PATH"
    exit 1
fi
print_status "Docker Compose found"

# Check Node.js
if ! command -v node &> /dev/null; then
    print_error "Node.js is not installed or not in PATH"
    exit 1
fi
print_status "Node.js found ($(node --version))"

# Check npm
if ! command -v npm &> /dev/null; then
    print_error "npm is not installed or not in PATH"
    exit 1
fi
print_status "npm found ($(npm --version))"

echo ""
echo "Starting services..."

# Step 1: Start backend services with test configuration
echo "1. Starting backend services (PostgreSQL, Redis, Elasticsearch)..."
docker-compose -f docker-compose.test.yml up -d

# Wait for services to be healthy
echo "   Waiting for services to be ready..."
sleep 10

# Check service health
if docker-compose -f docker-compose.test.yml ps | grep -q "unhealthy"; then
    print_warning "Some services may not be fully ready"
fi
print_status "Backend services started"

# Step 2: Install and start marketplace application
echo ""
echo "2. Setting up marketplace application..."

# Check if .env.local exists
if [ ! -f "web/marketplace/.env.local" ]; then
    print_warning ".env.local not found, creating from template..."
    if [ -f "web/marketplace/.env.template" ]; then
        cp web/marketplace/.env.template web/marketplace/.env.local
        print_status "Created .env.local from template"
    else
        print_error ".env.template not found"
        exit 1
    fi
fi

# Install marketplace dependencies if needed
if [ ! -d "web/marketplace/node_modules" ]; then
    echo "   Installing marketplace dependencies..."
    cd web/marketplace
    npm install
    cd ../..
    print_status "Marketplace dependencies installed"
fi

# Start marketplace in background
echo "   Starting marketplace on port 3001..."
cd web/marketplace
nohup npm run dev -- --port 3001 > ../../logs/marketplace.log 2>&1 &
MARKETPLACE_PID=$!
cd ../..

# Wait for marketplace to start
echo "   Waiting for marketplace to be ready..."
sleep 5  # Give it some initial time to start

for i in {1..60}; do
    if curl -s http://localhost:3001 > /dev/null 2>&1; then
        break
    fi
    if [ $i -eq 30 ]; then
        echo "   Still waiting... (this may take a bit for first startup)"
    fi
    sleep 2
done

if ! curl -s http://localhost:3001 > /dev/null 2>&1; then
    print_error "Marketplace failed to start on port 3001"
    echo "   Check logs/marketplace.log for details"
    if [ -f "logs/marketplace.log" ]; then
        echo "   Last few lines of marketplace log:"
        tail -n 10 logs/marketplace.log
    fi
    exit 1
fi
print_status "Marketplace started and ready"

# Step 3: Install test dependencies
echo ""
echo "3. Setting up test environment..."

if [ ! -d "testing/e2e/node_modules" ]; then
    echo "   Installing test dependencies..."
    cd testing/e2e
    npm install
    cd ../..
    print_status "Test dependencies installed"
fi

# Install Playwright browsers if needed
echo "   Installing Playwright browsers..."
cd testing/e2e
npx playwright install > /dev/null 2>&1 || true
cd ../..
print_status "Playwright browsers ready"

# Step 4: Run tests
echo ""
echo "4. Running end-to-end tests..."
echo "=================================================="

cd testing/e2e

# Run tests with specific configuration
MARKETPLACE_URL=http://localhost:3001 npm test

TEST_EXIT_CODE=$?

cd ../..

# Step 5: Cleanup
echo ""
echo "5. Cleaning up..."

# Stop marketplace
if kill -0 $MARKETPLACE_PID 2>/dev/null; then
    kill $MARKETPLACE_PID
    print_status "Marketplace stopped"
fi

# Stop docker services
docker-compose -f docker-compose.test.yml down > /dev/null 2>&1
print_status "Backend services stopped"

echo ""
echo "=================================================="
if [ $TEST_EXIT_CODE -eq 0 ]; then
    print_status "E2E tests completed successfully!"
    echo ""
    echo "📊 Test reports available at:"
    echo "   HTML Report: testing/e2e/playwright-report/index.html"
    echo "   JUnit XML:   testing/e2e/test-results/junit.xml"
    echo "   JSON:        testing/e2e/test-results/results.json"
else
    print_error "E2E tests failed with exit code $TEST_EXIT_CODE"
    echo ""
    echo "🔍 Check the following for debugging:"
    echo "   Marketplace logs: logs/marketplace.log"
    echo "   Test screenshots: testing/e2e/test-results/"
    echo "   Test videos:      testing/e2e/test-results/"
fi

exit $TEST_EXIT_CODE