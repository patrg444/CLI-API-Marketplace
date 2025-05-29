#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}====================================${NC}"
echo -e "${BLUE}CLI-API Marketplace Fixed Deployment${NC}"
echo -e "${BLUE}====================================${NC}"
echo ""

# Function to generate go.sum for a service
generate_go_sum() {
    local service=$1
    local service_path="services/$service"
    
    echo -e "${YELLOW}Generating go.sum for $service...${NC}"
    
    # Create a temporary Dockerfile for generating go.sum
    cat > "$service_path/Dockerfile.temp" << 'EOF'
FROM golang:1.21-alpine
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod ./
RUN go mod download
RUN go mod tidy
CMD ["cat", "go.sum"]
EOF
    
    # Build and run to get go.sum
    docker build -f "$service_path/Dockerfile.temp" -t "gosum-$service" "$service_path" > /dev/null 2>&1
    docker run --rm "gosum-$service" > "$service_path/go.sum" 2>/dev/null || echo "" > "$service_path/go.sum"
    
    # Clean up
    rm -f "$service_path/Dockerfile.temp"
    docker rmi "gosum-$service" > /dev/null 2>&1 || true
    
    if [ -s "$service_path/go.sum" ]; then
        echo -e "${GREEN}✓ Generated go.sum for $service${NC}"
    else
        echo -e "${YELLOW}⚠ Empty go.sum for $service (might be okay if no dependencies)${NC}"
    fi
}

# Function to update Dockerfile to handle go.sum
update_dockerfile() {
    local service=$1
    local dockerfile="services/$service/Dockerfile"
    
    # Check if Dockerfile already handles missing go.sum
    if ! grep -q "go mod download" "$dockerfile"; then
        # Backup original
        cp "$dockerfile" "$dockerfile.original" 2>/dev/null || true
        
        # Create updated Dockerfile
        cat > "$dockerfile" << 'EOF'
FROM golang:1.21-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod ./
COPY go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
EOF
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

echo -e "${BLUE}Step 1: Generating go.sum files${NC}"
echo "================================="

for service in "${SERVICES[@]}"; do
    if [ -d "services/$service" ]; then
        generate_go_sum "$service"
    fi
done

echo ""
echo -e "${BLUE}Step 2: Updating Dockerfiles${NC}"
echo "============================="

for service in "${SERVICES[@]}"; do
    if [ -d "services/$service" ]; then
        update_dockerfile "$service"
        echo -e "${GREEN}✓ Updated Dockerfile for $service${NC}"
    fi
done

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
    echo -e "${BLUE}Step 4: Verifying deployment${NC}"
    echo "============================"
    
    # Wait a bit for services to start
    echo "Waiting for services to initialize..."
    sleep 10
    
    # Run verification
    ./scripts/verify-deployment.sh once
    
    echo ""
    echo -e "${GREEN}Deployment complete!${NC}"
    echo ""
    echo "To monitor services:"
    echo "  ./scripts/verify-deployment.sh wait"
    echo ""
    echo "To view logs:"
    echo "  docker-compose logs -f [service-name]"
else
    echo -e "${RED}✗ Deployment failed${NC}"
    echo ""
    echo "Check logs with:"
    echo "  docker-compose logs"
    exit 1
fi
