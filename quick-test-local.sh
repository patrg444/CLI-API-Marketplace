#!/bin/bash

echo "🚀 API-Direct Local Quick Test"
echo "=============================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check Docker
echo "1. Checking Docker..."
if command -v docker &> /dev/null; then
    echo -e "${GREEN}✓ Docker installed${NC}"
    if docker ps &> /dev/null; then
        echo -e "${GREEN}✓ Docker daemon running${NC}"
    else
        echo -e "${RED}✗ Docker daemon not running${NC}"
        echo -e "${YELLOW}  Starting Docker...${NC}"
        open -a Docker
        echo "  Waiting 20 seconds for Docker to start..."
        sleep 20
    fi
else
    echo -e "${RED}✗ Docker not installed${NC}"
    exit 1
fi

echo ""
echo "2. Checking environment files..."
if [ -f ".env" ]; then
    echo -e "${GREEN}✓ .env file exists${NC}"
else
    echo -e "${YELLOW}! .env file missing, creating from example...${NC}"
    cp .env.example .env
fi

if [ -f ".env.local" ]; then
    echo -e "${GREEN}✓ .env.local file exists${NC}"
else
    echo -e "${YELLOW}! .env.local file missing, creating from example...${NC}"
    cp .env.example .env.local
fi

echo ""
echo "3. Starting local services..."
echo -e "${YELLOW}This may take a few minutes on first run...${NC}"

# Start services
docker-compose -f docker-compose.local.yml up -d

echo ""
echo "4. Waiting for services to be ready..."
sleep 10

echo ""
echo "5. Checking service status..."
docker-compose -f docker-compose.local.yml ps

echo ""
echo "6. Running health checks..."
echo ""

# Function to check URL
check_url() {
    local url=$1
    local name=$2
    if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "200\|301\|302"; then
        echo -e "${GREEN}✓ $name is accessible${NC}"
    else
        echo -e "${RED}✗ $name is not accessible${NC}"
    fi
}

# Check local services
check_url "http://localhost:3000" "Frontend (Marketplace)"
check_url "http://localhost:3001" "Console"
check_url "http://localhost:8000/health" "Backend API"
check_url "http://localhost:8001/health" "Gateway"

echo ""
echo "7. Service URLs:"
echo "   - Marketplace: http://localhost:3000"
echo "   - Console: http://localhost:3001"
echo "   - API Docs: http://localhost:8000/docs"
echo "   - Gateway: http://localhost:8001"
echo ""

echo "8. Useful commands:"
echo "   - View logs: docker-compose -f docker-compose.local.yml logs -f"
echo "   - Stop services: docker-compose -f docker-compose.local.yml down"
echo "   - Restart services: docker-compose -f docker-compose.local.yml restart"
echo ""

echo "✅ Local setup complete!"
echo ""
echo "Next steps:"
echo "1. Open http://localhost:3001 to access the console"
echo "2. Create a test account"
echo "3. Deploy a test API"
echo "4. Test the monetization features"
echo ""