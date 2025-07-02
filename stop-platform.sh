#!/bin/bash

echo "🛑 Stopping API-Direct Platform"
echo "==============================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Stop backend if running
if [ -f .backend.pid ]; then
    BACKEND_PID=$(cat .backend.pid)
    if kill -0 $BACKEND_PID 2>/dev/null; then
        echo -e "${YELLOW}Stopping FastAPI backend...${NC}"
        kill $BACKEND_PID
        rm -f .backend.pid
    fi
fi

# Stop Docker services
echo -e "${YELLOW}Stopping Docker services...${NC}"
docker-compose down

# Optional: Remove volumes (uncomment if needed)
# echo -e "${YELLOW}Removing data volumes...${NC}"
# docker-compose down -v

echo -e "\n${GREEN}✅ All services stopped${NC}"