#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}====================================${NC}"
echo -e "${BLUE}Fixing import paths in services${NC}"
echo -e "${BLUE}====================================${NC}"
echo ""

# Function to fix imports in a file
fix_imports() {
    local file=$1
    echo -e "${YELLOW}Fixing imports in: $file${NC}"
    
    # Fix shared/auth imports
    sed -i.bak 's|"shared/auth"|"github.com/api-direct/services/shared/auth"|g' "$file"
    sed -i.bak 's|"shared/store"|"github.com/api-direct/services/shared/store"|g' "$file"
    sed -i.bak 's|"shared/models"|"github.com/api-direct/services/shared/models"|g' "$file"
    
    # Remove backup files
    rm -f "${file}.bak"
}

# Find all Go files that might have incorrect imports
echo -e "${BLUE}Searching for files with incorrect imports...${NC}"

# Find files with incorrect shared imports
files_to_fix=$(grep -r '"shared/' services/ --include="*.go" | cut -d: -f1 | sort | uniq || true)

if [ -z "$files_to_fix" ]; then
    echo -e "${GREEN}No files with incorrect imports found!${NC}"
else
    echo -e "${YELLOW}Found files with incorrect imports:${NC}"
    echo "$files_to_fix"
    echo ""
    
    for file in $files_to_fix; do
        fix_imports "$file"
    done
    
    echo ""
    echo -e "${GREEN}✓ Import paths fixed!${NC}"
fi

# Update go.sum files after fixing imports
echo ""
echo -e "${BLUE}Updating go.sum files...${NC}"

# Services to update
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

for service in "${SERVICES[@]}"; do
    if [ -d "services/$service" ]; then
        echo -e "${YELLOW}Updating go.sum for $service...${NC}"
        docker run --rm -v "$(pwd)/services/$service":/app -w /app golang:1.21-alpine sh -c "go mod tidy" > /dev/null 2>&1 || true
    fi
done

echo ""
echo -e "${GREEN}Import fixes complete!${NC}"
echo ""
echo "Next step: Run deployment"
echo "  docker-compose up -d --build"
