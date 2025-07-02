#!/bin/bash

# Run CLI tests using Docker
echo "🐳 Running CLI tests in Docker container..."

# Create a temporary directory for test results
mkdir -p test-results

# Build and run tests
docker run --rm \
  -v "$(pwd)":/app \
  -w /app \
  golang:1.21-alpine \
  sh -c '
    # Install dependencies
    apk add --no-cache git gcc musl-dev
    
    # Install testify
    go get github.com/stretchr/testify/assert
    go get github.com/stretchr/testify/require
    
    # Run tests
    echo "🧪 Running marketplace command tests..."
    
    # First, let'\''s compile to check for syntax errors
    echo "📦 Checking compilation..."
    go build ./cmd/... || exit 1
    
    echo "✅ Compilation successful!"
    
    # Now run the actual tests
    echo "🏃 Running tests..."
    go test ./cmd -v
  '