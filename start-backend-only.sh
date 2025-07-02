#!/bin/bash

echo "🚀 Starting API-Direct Backend Services"
echo "======================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Check Docker
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker Desktop.${NC}"
    exit 1
fi

echo -e "${BLUE}📦 Starting core services only...${NC}"

# Start only the working services
docker-compose up -d postgres redis

# Wait for services
echo -e "\n${YELLOW}⏳ Waiting for services to be ready...${NC}"
sleep 5

# Start FastAPI backend
echo -e "\n${BLUE}🐍 Starting FastAPI Backend...${NC}"
cd backend

# Create virtual environment if needed
if [ ! -d "venv" ]; then
    python3 -m venv venv
    source venv/bin/activate
    pip install -r requirements.txt
else
    source venv/bin/activate
fi

# Copy the mock auth version
cp api/main_with_mock.py api/main.py

# Start backend
source ../.env
export DATABASE_URL="postgresql://apidirect:localpassword@localhost:5432/apidirect"
export REDIS_URL="redis://localhost:6379"
export USE_MOCK_AUTH="true"

echo -e "\n${GREEN}Starting backend on http://localhost:8000${NC}"
uvicorn api.main:app --host 0.0.0.0 --port 8000 --reload &
BACKEND_PID=$!

cd ..

# Save PID
echo $BACKEND_PID > .backend.pid

echo -e "\n${GREEN}✨ Backend Services Running${NC}"
echo "============================"
echo ""
echo "🌐 Active Services:"
echo "  - Backend API: http://localhost:8000"
echo "  - API Docs: http://localhost:8000/docs"
echo "  - PostgreSQL: localhost:5432"
echo "  - Redis: localhost:6379"
echo ""
echo "🌍 Web Applications:"
echo "  - Console: https://console.apidirect.dev"
echo "  - Marketplace: https://marketplace.apidirect.dev"
echo ""
echo "📝 Login Credentials:"
echo "  - Email: demo@apidirect.dev"
echo "  - Password: secret"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop${NC}"

# Trap Ctrl+C
trap 'echo -e "\n${YELLOW}Stopping services...${NC}"; kill $BACKEND_PID 2>/dev/null; docker-compose down; rm -f .backend.pid; exit' INT

wait