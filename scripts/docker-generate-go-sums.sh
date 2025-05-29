#!/bin/bash

# Generate go.sum files using Docker
# This script runs go mod download inside Docker containers to generate go.sum files

echo "Generating go.sum files using Docker..."

# List of services with Go modules
services=(
    "storage"
    "deployment"
    "apikey"
    "gateway"
    "metering"
    "billing"
    "marketplace"
    "payout"
)

# Generate go.sum for each service
for service in "${services[@]}"; do
    echo "Processing service: $service"
    
    if [ -d "services/$service" ] && [ -f "services/$service/go.mod" ]; then
        # Create a temporary Dockerfile for generating go.sum
        cat > "services/$service/Dockerfile.gosum" <<EOF
FROM golang:1.21-alpine
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod ./
RUN go mod download && go mod tidy
EOF
        
        # Build and run container to generate go.sum
        docker build -f "services/$service/Dockerfile.gosum" -t "gosum-$service" "services/$service" > /dev/null 2>&1
        
        # Extract go.sum from container
        docker run --rm "gosum-$service" cat go.sum > "services/$service/go.sum" 2>/dev/null
        
        # Clean up
        docker rmi "gosum-$service" > /dev/null 2>&1
        rm -f "services/$service/Dockerfile.gosum"
        
        if [ -s "services/$service/go.sum" ]; then
            echo "  ✓ Generated go.sum for $service"
        else
            echo "  ⚠️  Failed to generate go.sum for $service"
        fi
    else
        echo "  ⚠️  No go.mod found for $service"
    fi
done

echo "Done! All go.sum files generated."
