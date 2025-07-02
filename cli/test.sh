#!/bin/bash

# CLI Command Tests Runner
# This script runs all tests for the new marketplace commands

set -e

echo "🧪 Running CLI Command Tests..."
echo "================================"

# Navigate to CLI directory
cd "$(dirname "$0")"

# Ensure dependencies are installed
echo "📦 Checking dependencies..."
go mod download

# Run tests with coverage
echo ""
echo "🏃 Running unit tests..."
go test ./cmd -v -coverprofile=coverage.out -covermode=atomic

# Show coverage summary
echo ""
echo "📊 Coverage Summary:"
go tool cover -func=coverage.out | grep total || true

# Generate HTML coverage report
echo ""
echo "📄 Generating coverage report..."
go tool cover -html=coverage.out -o coverage.html
echo "Coverage report saved to: coverage.html"

# Run specific test suites
echo ""
echo "🎯 Running marketplace command tests..."
echo ""

echo "Testing Analytics Commands..."
go test ./cmd -run TestAnalytics -v

echo ""
echo "Testing Earnings Commands..."
go test ./cmd -run TestEarnings -v

echo ""
echo "Testing Subscriptions Commands..."
go test ./cmd -run TestSubscriptions -v

echo ""
echo "Testing Review Commands..."
go test ./cmd -run TestReview -v

echo ""
echo "Testing Search Commands..."
go test ./cmd -run TestSearch -v

# Run benchmarks if any
echo ""
echo "⚡ Running benchmarks..."
go test ./cmd -bench=. -benchmem -run=^$ || true

# Check for race conditions
echo ""
echo "🔍 Checking for race conditions..."
go test ./cmd -race -short

echo ""
echo "✅ All tests completed!"
echo ""

# Show test statistics
echo "📈 Test Statistics:"
go test ./cmd -json | go run github.com/jstemmer/go-junit-report -set-exit-code > test-report.xml 2>/dev/null || true

# Count tests
TOTAL_TESTS=$(go test ./cmd -list . | grep -E "^Test" | wc -l | tr -d ' ')
echo "Total test cases: $TOTAL_TESTS"

# Exit with appropriate code
if [ -f coverage.out ]; then
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    echo "Total coverage: ${COVERAGE}%"
    
    # Fail if coverage is below threshold
    THRESHOLD=70
    if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
        echo "❌ Coverage ${COVERAGE}% is below threshold of ${THRESHOLD}%"
        exit 1
    fi
fi

echo ""
echo "🎉 All tests passed!"