#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}CLI-API Marketplace Final Deployment${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to update Dockerfile to copy shared code
update_dockerfile_with_shared() {
    local service=$1
    local dockerfile="services/$service/Dockerfile"
    
    # Check if we already updated this Dockerfile
    if grep -q "COPY ../shared" "$dockerfile" 2>/dev/null; then
        return
    fi
    
    echo -e "${YELLOW}Updating Dockerfile for $service to include shared code...${NC}"
    
    # Create a new Dockerfile that copies shared code
    cat > "$dockerfile" << 'EOF'
FROM golang:1.21-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git

# Copy shared code first
COPY shared ./shared

# Copy service-specific files
COPY go.mod go.sum* ./
RUN go mod download
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
EOF
}

# Function to create docker-compose override
create_docker_compose_override() {
    echo -e "${YELLOW}Creating docker-compose override for build contexts...${NC}"
    
    cat > docker-compose.override.yml << 'EOF'
services:
  storage:
    build:
      context: ./services
      dockerfile: storage/Dockerfile
  
  deployment:
    build:
      context: ./services
      dockerfile: deployment/Dockerfile
  
  apikey:
    build:
      context: ./services
      dockerfile: apikey/Dockerfile
  
  gateway:
    build:
      context: ./services
      dockerfile: gateway/Dockerfile
  
  metering:
    build:
      context: ./services
      dockerfile: metering/Dockerfile
  
  billing:
    build:
      context: ./services
      dockerfile: billing/Dockerfile
  
  marketplace:
    build:
      context: ./services
      dockerfile: marketplace/Dockerfile
  
  payout:
    build:
      context: ./services
      dockerfile: payout/Dockerfile
EOF
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

echo -e "${BLUE}Step 1: Updating Dockerfiles${NC}"
echo "============================="

for service in "${SERVICES[@]}"; do
    if [ -d "services/$service" ]; then
        update_dockerfile_with_shared "$service"
    fi
done

echo ""
echo -e "${BLUE}Step 2: Creating docker-compose override${NC}"
echo "========================================"

create_docker_compose_override

echo ""
echo -e "${BLUE}Step 3: Starting deployment${NC}"
echo "==========================="

# Clean up any existing containers
echo "Cleaning up existing containers..."
docker-compose down -v > /dev/null 2>&1 || true

# Build and start all services
echo "Building and starting services..."
echo "This may take 10-20 minutes on first run..."

if DOCKER_BUILDKIT=1 docker-compose up -d --build; then
    echo -e "${GREEN}✓ Services started successfully${NC}"
    
    echo ""
    echo -e "${BLUE}Step 4: Waiting for services to initialize${NC}"
    echo "=========================================="
    
    # Wait a bit for services to start
    echo "Waiting 30 seconds for services to initialize..."
    sleep 30
    
    # Run verification
    echo ""
    echo -e "${BLUE}Step 5: Verifying deployment${NC}"
    echo "============================"
    ./scripts/verify-deployment.sh once
    
    echo ""
    echo -e "${GREEN}Deployment process complete!${NC}"
    echo ""
    echo "To monitor services:"
    echo "  ./scripts/verify-deployment.sh wait"
    echo ""
    echo "To view logs:"
    echo "  docker-compose logs -f [service-name]"
    echo ""
    echo "Access your services:"
    echo "  - Marketplace UI: http://localhost:3000"
    echo "  - Creator Portal: http://localhost:3001"
    echo "  - API Gateway: http://localhost:8082"
else
    echo -e "${RED}✗ Deployment failed${NC}"
    echo ""
    echo "Check logs with:"
    echo "  docker-compose logs"
    exit 1
fi
