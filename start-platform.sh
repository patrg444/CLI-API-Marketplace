#!/bin/bash

echo "🚀 Starting API-Direct Full Platform"
echo "===================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Check Docker
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker Desktop.${NC}"
    exit 1
fi

echo -e "${BLUE}📦 Starting all services with Docker Compose...${NC}"
docker-compose up -d

# Wait for services to be healthy
echo -e "\n${YELLOW}⏳ Waiting for services to be ready...${NC}"
sleep 10

# Check service health
echo -e "\n${BLUE}🔍 Checking service status...${NC}"

# Function to check service
check_service() {
    local name=$1
    local url=$2
    if curl -s -f "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ $name is running${NC}"
        return 0
    else
        echo -e "${RED}✗ $name is not responding${NC}"
        return 1
    fi
}

# Check each service
check_service "PostgreSQL" "http://localhost:5432" || true
check_service "Redis" "http://localhost:6379" || true
check_service "API Gateway" "http://localhost:8082/health"
check_service "Storage Service" "http://localhost:8087/health"
check_service "API Key Service" "http://localhost:8083/health"
check_service "Deployment Service" "http://localhost:8081/health"

# Start backend API separately if needed
if [ -f "backend/api/main.py" ]; then
    echo -e "\n${BLUE}🐍 Starting FastAPI Backend...${NC}"
    cd backend
    if [ ! -d "venv" ]; then
        python3 -m venv venv
        source venv/bin/activate
        pip install -r requirements.txt
    else
        source venv/bin/activate
    fi
    
    # Start in background
    USE_MOCK_AUTH=true uvicorn api.main:app --host 0.0.0.0 --port 8000 --reload &
    BACKEND_PID=$!
    cd ..
    
    sleep 5
    check_service "Backend API" "http://localhost:8000/health"
fi

echo -e "\n${GREEN}✨ Platform Status Summary${NC}"
echo "================================"
echo ""
echo "🌐 Service URLs:"
echo "  - Backend API: http://localhost:8000"
echo "  - API Gateway: http://localhost:8082"
echo "  - API Docs: http://localhost:8000/docs"
echo ""
echo "🔧 Microservices:"
echo "  - Storage Service: http://localhost:8087"
echo "  - API Key Service: http://localhost:8083"
echo "  - Deployment Service: http://localhost:8081"
echo ""
echo "📊 Infrastructure:"
echo "  - PostgreSQL: localhost:5432"
echo "  - Redis: localhost:6379"
echo "  - Grafana: http://localhost:3000 (if configured)"
echo ""
echo "🌍 Web Applications:"
echo "  - Landing: https://apidirect.dev"
echo "  - Console: https://console.apidirect.dev"
echo "  - Marketplace: https://marketplace.apidirect.dev"
echo "  - Docs: https://docs.apidirect.dev"
echo ""
echo "📝 Default Credentials:"
echo "  - Email: demo@apidirect.dev"
echo "  - Password: secret"
echo ""
echo -e "${YELLOW}📚 Next Steps:${NC}"
echo "1. Visit https://console.apidirect.dev to access the dashboard"
echo "2. Use './cli/apidirect --help' to see CLI commands"
echo "3. Check logs with 'docker-compose logs -f'"
echo "4. Stop all services with './stop-platform.sh'"
echo ""
echo -e "${GREEN}✅ Platform is ready!${NC}"

# Save PID for stop script
echo $BACKEND_PID > .backend.pid

# Trap to handle Ctrl+C
trap 'echo -e "\n${YELLOW}Stopping services...${NC}"; docker-compose down; kill $BACKEND_PID 2>/dev/null; rm -f .backend.pid; exit' INT

# Keep script running
echo -e "\n${YELLOW}Press Ctrl+C to stop all services${NC}"
wait