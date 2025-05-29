#!/bin/bash

# CLI API Marketplace - Test Environment Setup Script

echo "🚀 Setting up test environment for CLI API Marketplace..."

# Create necessary directories
mkdir -p testing/reports/day1
mkdir -p testing/reports/screenshots
mkdir -p testing/reports/performance
mkdir -p testing/test-data

# Check if services are running (mock check for now)
echo "📦 Checking service dependencies..."

# Mock service status
cat > testing/reports/service-status.json << EOF
{
  "services": {
    "marketplace": {
      "status": "mock",
      "url": "http://localhost:3001"
    },
    "creator-portal": {
      "status": "mock", 
      "url": "http://localhost:3002"
    },
    "gateway": {
      "status": "mock",
      "url": "http://localhost:8080"
    },
    "apikey": {
      "status": "mock",
      "url": "http://localhost:8081"
    },
    "billing": {
      "status": "mock",
      "url": "http://localhost:8082"
    },
    "payout": {
      "status": "mock",
      "url": "http://localhost:8083"
    },
    "metering": {
      "status": "mock",
      "url": "http://localhost:8084"
    },
    "elasticsearch": {
      "status": "mock",
      "url": "http://localhost:9200"
    },
    "postgres": {
      "status": "mock",
      "url": "postgres://localhost:5432/api_marketplace"
    },
    "redis": {
      "status": "mock",
      "url": "redis://localhost:6379"
    }
  },
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF

echo "✅ Service status saved to testing/reports/service-status.json"

# Generate mock test data
echo "🔧 Generating test data..."

# Create a simple test data generator output
cat > testing/test-data/generated-data.json << EOF
{
  "creators": 50,
  "consumers": 200,
  "apis": 100,
  "reviews": 500,
  "apiKeys": 150,
  "usageRecords": 3000,
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF

echo "✅ Test data generation complete"

# Set environment variables
export TEST_ENV=local
export BASE_URL_MARKETPLACE=http://localhost:3001
export BASE_URL_CREATOR_PORTAL=http://localhost:3002
export BASE_URL_GATEWAY=http://localhost:8080

echo "✅ Environment variables configured"

# Install test dependencies
echo "📚 Installing test dependencies..."
cd testing/e2e && npm install --silent
cd ../data-generators && npm install --silent
cd ../..

echo "✅ Dependencies installed"

# Display test readiness
echo ""
echo "🎯 Test Environment Ready!"
echo "========================="
echo "• Marketplace UI: $BASE_URL_MARKETPLACE"
echo "• Creator Portal: $BASE_URL_CREATOR_PORTAL"
echo "• API Gateway: $BASE_URL_GATEWAY"
echo ""
echo "📋 Available test commands:"
echo "• Run all E2E tests: cd testing/e2e && npm test"
echo "• Run search tests: cd testing/e2e && npm test -- --grep 'search'"
echo "• Run review tests: cd testing/e2e && npm test -- --grep 'review'"
echo "• Run performance tests: cd testing/performance && k6 run k6-load-test.js"
echo ""
echo "📊 Reports will be saved to: testing/reports/"
echo ""
