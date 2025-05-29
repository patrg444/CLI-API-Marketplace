#!/bin/bash

# Deploy with automatic go.sum generation
# This script handles the go.sum issue and deploys with reduced verbosity

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${YELLOW}Starting deployment with go.sum generation...${NC}"

# Step 1: Create go.sum files by modifying Dockerfiles temporarily
echo -e "\n${BLUE}► Generating go.sum files...${NC}"

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

for service in "${services[@]}"; do
    if [ -d "services/$service" ]; then
        echo -e "  Processing $service..."
        
        # Backup original Dockerfile
        cp "services/$service/Dockerfile" "services/$service/Dockerfile.backup"
        
        # Modify Dockerfile to handle missing go.sum
        sed -i.bak 's/COPY go.mod go.sum \.\//COPY go.mod* go.sum* .\//g' "services/$service/Dockerfile"
        
        # Add go mod download after COPY
        sed -i.bak '/COPY go.mod\*/a\
RUN go mod download && go mod tidy' "services/$service/Dockerfile"
        
        # Create empty go.sum if it doesn't exist
        touch "services/$service/go.sum"
    fi
done

echo -e "${GREEN}✓ Dockerfiles prepared for go.sum generation${NC}"

# Step 2: Build with docker-compose
echo -e "\n${BLUE}► Building services...${NC}"

# Set environment variables for cleaner output
export COMPOSE_DOCKER_CLI_BUILD=1
export DOCKER_BUILDKIT=1
export BUILDKIT_PROGRESS=plain

# Build quietly but capture errors
if DOCKER_BUILDKIT=1 docker-compose build --progress=plain 2>&1 | tee /tmp/build.log | grep -E "^(#[0-9]+ (ERROR|DONE)|error:|failed)" ; then
    # Check if there were actual errors
    if grep -q "ERROR\|error:\|failed" /tmp/build.log; then
        echo -e "${RED}✗ Build failed. Check /tmp/build.log for details${NC}"
        
        # Restore original Dockerfiles
        for service in "${services[@]}"; do
            if [ -f "services/$service/Dockerfile.backup" ]; then
                mv "services/$service/Dockerfile.backup" "services/$service/Dockerfile"
            fi
        done
        exit 1
    fi
fi

echo -e "${GREEN}✓ Services built successfully${NC}"

# Step 3: Start services
echo -e "\n${BLUE}► Starting services...${NC}"
docker-compose up -d

# Step 4: Restore original Dockerfiles
echo -e "\n${BLUE}► Cleaning up...${NC}"
for service in "${services[@]}"; do
    if [ -f "services/$service/Dockerfile.backup" ]; then
        mv "services/$service/Dockerfile.backup" "services/$service/Dockerfile"
    fi
    # Remove .bak files
    rm -f "services/$service/Dockerfile.bak"
done

# Step 5: Check status
echo -e "\n${BLUE}► Checking service status...${NC}"
sleep 5

# Show running containers
echo -e "\n${YELLOW}Running containers:${NC}"
docker-compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"

# Count running services
RUNNING=$(docker-compose ps -q | wc -l)
echo -e "\n${GREEN}✓ $RUNNING containers are running${NC}"

# Show access URLs
echo -e "\n${GREEN}Deployment complete!${NC}"
echo -e "${YELLOW}Services are available at:${NC}"
echo "  • API Gateway: http://localhost:8082"
echo "  • Storage Service: http://localhost:8080"
echo "  • Deployment Service: http://localhost:8081"
echo "  • API Key Service: http://localhost:8083"
echo "  • Metering Service: http://localhost:8084"
echo "  • Billing Service: http://localhost:8085"
echo "  • Marketplace UI: http://localhost:3000"
echo "  • Creator Portal: http://localhost:3001"
echo "  • PostgreSQL: localhost:5432"
echo "  • Redis: localhost:6379"
echo "  • Elasticsearch: http://localhost:9200"
echo "  • Kibana: http://localhost:5601"

echo -e "\n${YELLOW}To view logs:${NC}"
echo "  docker-compose logs -f [service-name]"

echo -e "\n${YELLOW}To stop all services:${NC}"
echo "  docker-compose down"

# Clean up temp file
rm -f /tmp/build.log
