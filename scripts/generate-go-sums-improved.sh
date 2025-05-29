#!/bin/bash

# Improved script to generate go.sum files using Docker
# This version actually runs go mod download inside containers

echo "Generating go.sum files using Docker (improved version)..."

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

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Generate go.sum for each service
for service in "${services[@]}"; do
    echo -e "${YELLOW}Processing service: $service${NC}"
    
    if [ -d "services/$service" ] && [ -f "services/$service/go.mod" ]; then
        cd "services/$service" || continue
        
        # Remove existing empty go.sum
        rm -f go.sum
        
        # Create a temporary Dockerfile that generates go.sum
        cat > Dockerfile.temp <<'EOF'
FROM golang:1.21-alpine
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod ./
RUN go mod download && go mod tidy
CMD ["cat", "go.sum"]
EOF
        
        # Build the temporary image
        echo "  Building temporary Docker image..."
        if docker build -f Dockerfile.temp -t "temp-gosum-$service" . > /dev/null 2>&1; then
            # Run container and capture go.sum
            echo "  Generating go.sum..."
            if docker run --rm "temp-gosum-$service" > go.sum 2>/dev/null; then
                # Check if go.sum has content
                if [ -s go.sum ]; then
                    echo -e "  ${GREEN}✓ Generated go.sum for $service ($(wc -l < go.sum) lines)${NC}"
                else
                    echo -e "  ${RED}✗ Generated empty go.sum for $service${NC}"
                    # Create a minimal go.sum to prevent build errors
                    echo "// Auto-generated minimal go.sum" > go.sum
                fi
            else
                echo -e "  ${RED}✗ Failed to extract go.sum for $service${NC}"
            fi
            
            # Clean up
            docker rmi "temp-gosum-$service" > /dev/null 2>&1
        else
            echo -e "  ${RED}✗ Failed to build Docker image for $service${NC}"
        fi
        
        # Clean up temporary Dockerfile
        rm -f Dockerfile.temp
        
        cd ../..
    else
        echo -e "  ${YELLOW}⚠️  No go.mod found for $service${NC}"
    fi
done

echo -e "${GREEN}Done! Check above for any errors.${NC}"

# Verify all go.sum files exist and are not empty
echo -e "\n${YELLOW}Verification:${NC}"
for service in "${services[@]}"; do
    if [ -f "services/$service/go.sum" ]; then
        size=$(wc -c < "services/$service/go.sum" 2>/dev/null || echo "0")
        if [ "$size" -gt 0 ]; then
            echo -e "  ${GREEN}✓ $service: go.sum exists (${size} bytes)${NC}"
        else
            echo -e "  ${RED}✗ $service: go.sum is empty${NC}"
        fi
    else
        echo -e "  ${RED}✗ $service: go.sum missing${NC}"
    fi
done
