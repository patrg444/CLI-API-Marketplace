#!/bin/bash

# Generate go.sum files for all Go services

echo "Generating go.sum files for all services..."

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
    cd "services/$service" || continue
    
    # Initialize go module if needed and download dependencies
    if [ -f "go.mod" ]; then
        echo "  Downloading dependencies for $service..."
        go mod download
        go mod tidy
        echo "  ✓ Generated go.sum for $service"
    else
        echo "  ⚠️  No go.mod found for $service"
    fi
    
    cd ../..
done

echo "Done! All go.sum files generated."
