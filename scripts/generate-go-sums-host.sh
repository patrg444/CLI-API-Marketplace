#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}====================================${NC}"
echo -e "${BLUE}Generating go.sum files on host${NC}"
echo -e "${BLUE}====================================${NC}"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}Go is not installed. Attempting to generate go.sum files using Docker...${NC}"
    USE_DOCKER=true
else
    echo -e "${GREEN}Go is installed. Using local Go to generate go.sum files.${NC}"
    USE_DOCKER=false
fi

# Function to generate go.sum for a service
generate_go_sum() {
    local service=$1
    local service_path="services/$service"
    
    echo -e "${YELLOW}Processing $service...${NC}"
    
    if [ ! -f "$service_path/go.mod" ]; then
        echo -e "${RED}✗ go.mod not found for $service${NC}"
        return 1
    fi
    
    if [ "$USE_DOCKER" = true ]; then
        # Use Docker to generate go.sum
        docker run --rm -v "$(pwd)/$service_path":/app -w /app golang:1.21-alpine sh -c "go mod download && go mod tidy" > /dev/null 2>&1
    else
        # Use local Go
        (cd "$service_path" && go mod download && go mod tidy) > /dev/null 2>&1
    fi
    
    if [ -s "$service_path/go.sum" ]; then
        echo -e "${GREEN}✓ Generated go.sum for $service${NC}"
    else
        echo -e "${YELLOW}⚠ Empty or missing go.sum for $service${NC}"
    fi
}

# Services to process
SERVICES=(
    "storage"
    "deployment"
    "apikey"
    "gateway"
    "metering"
    "billing"
    "marketplace"
    "payout"
)

echo -e "${BLUE}Generating go.sum files...${NC}"
echo "=========================="

for service in "${SERVICES[@]}"; do
    if [ -d "services/$service" ]; then
        generate_go_sum "$service"
    else
        echo -e "${RED}✗ Service directory not found: $service${NC}"
    fi
done

echo ""
echo -e "${GREEN}go.sum generation complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Run the deployment: docker-compose up -d --build"
echo "2. Or use the deployment script: ./scripts/deploy-simple.sh"
