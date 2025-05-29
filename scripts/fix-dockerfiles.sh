#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}====================================${NC}"
echo -e "${BLUE}Fixing Dockerfiles${NC}"
echo -e "${BLUE}====================================${NC}"
echo ""

# Function to create proper Dockerfile without shared copy
fix_dockerfile() {
    local service=$1
    local dockerfile="services/$service/Dockerfile"
    
    echo -e "${YELLOW}Fixing Dockerfile for $service...${NC}"
    
    # Different Dockerfile based on whether it's a Go service or Node service
    if [ "$service" == "marketplace" ] || [ "$service" == "creator-portal" ]; then
        # Node.js services
        cat > "$dockerfile" << 'EOF'
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
CMD ["npm", "start"]
EOF
    else
        # Go services - determine the main binary name
        local binary_name="main"
        if [ "$service" == "storage" ]; then
            binary_name="storage-service"
        elif [ "$service" == "deployment" ]; then
            binary_name="deployment-service"
        elif [ "$service" == "metering" ]; then
            binary_name="metering"
        fi
        
        cat > "$dockerfile" << EOF
FROM golang:1.21-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $binary_name .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/$binary_name .
CMD ["./$binary_name"]
EOF
    fi
}

# All services
SERVICES=(
    "storage"
    "deployment"
    "apikey"
    "gateway"
    "metering"
    "billing"
    "marketplace"
    "payout"
    "marketplace-api"
    "creator-portal"
)

for service in "${SERVICES[@]}"; do
    if [ -d "services/$service" ]; then
        fix_dockerfile "$service"
    elif [ "$service" == "marketplace-api" ] && [ -d "web/marketplace" ]; then
        # Handle marketplace-api which is actually web/marketplace
        echo -e "${YELLOW}Fixing Dockerfile for marketplace-api (web/marketplace)...${NC}"
        cat > "web/marketplace/Dockerfile" << 'EOF'
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

FROM node:18-alpine
WORKDIR /app
COPY --from=builder /app/node_modules ./node_modules
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["npm", "start"]
EOF
    elif [ "$service" == "creator-portal" ] && [ -d "web/creator-portal" ]; then
        # Handle creator-portal
        echo -e "${YELLOW}Fixing Dockerfile for creator-portal...${NC}"
        cat > "web/creator-portal/Dockerfile" << 'EOF'
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE 3001
CMD ["npm", "start"]
EOF
    fi
done

echo ""
echo -e "${GREEN}Dockerfiles fixed!${NC}"
echo ""
echo "Next step: Run deployment"
echo "  docker-compose up -d --build"
