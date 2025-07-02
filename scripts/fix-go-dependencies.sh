#!/bin/bash

# Fix Go Dependencies Script
# Run this script when Go is available to fix all Go service dependencies

set -e

echo "🔧 Fixing Go service dependencies..."
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    print_warning "Please install Go 1.21+ and run this script again"
    exit 1
fi

print_status "Go found ($(go version))"

# List of Go services
GO_SERVICES=(
    "services/billing"
    "services/deployment" 
    "services/gateway"
    "services/marketplace"
    "services/metering"
    "services/payout"
    "services/storage"
    "services/apikey"
    "services/shared"
)

echo ""
echo "Fixing dependencies for Go services..."

for service in "${GO_SERVICES[@]}"; do
    if [ -d "$service" ] && [ -f "$service/go.mod" ]; then
        echo ""
        echo "Processing $service..."
        cd "$service"
        
        # Clean and download dependencies
        go mod tidy
        go mod download
        
        # Verify the module
        go mod verify
        
        print_status "$service dependencies fixed"
        cd - > /dev/null
    else
        print_warning "$service not found or no go.mod file"
    fi
done

echo ""
echo "===================================="
print_status "All Go service dependencies fixed!"
echo ""
echo "You can now run the full test suite with:"
echo "  ./scripts/run-e2e-tests.sh"