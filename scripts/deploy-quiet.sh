#!/bin/bash

# Quiet deployment script with progress indicators
# Usage: ./deploy-quiet.sh [options]
#   -v, --verbose    Show detailed output
#   -h, --help       Show this help message

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default to quiet mode
VERBOSE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo "  -v, --verbose    Show detailed output"
            echo "  -h, --help       Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h for help"
            exit 1
            ;;
    esac
done

# Function to show progress
show_progress() {
    local message=$1
    echo -e "${BLUE}► ${message}${NC}"
}

# Function to show success
show_success() {
    local message=$1
    echo -e "${GREEN}✓ ${message}${NC}"
}

# Function to show error
show_error() {
    local message=$1
    echo -e "${RED}✗ ${message}${NC}"
}

# Function to run command quietly or verbosely
run_command() {
    local message=$1
    shift
    
    show_progress "$message"
    
    if [ "$VERBOSE" = true ]; then
        if "$@"; then
            show_success "$message"
        else
            show_error "$message"
            return 1
        fi
    else
        if "$@" > /dev/null 2>&1; then
            show_success "$message"
        else
            show_error "$message"
            return 1
        fi
    fi
}

# Main deployment process
echo -e "${YELLOW}Starting quiet deployment...${NC}"

# Step 1: Generate go.sum files
if [ -x "scripts/generate-go-sums-improved.sh" ]; then
    run_command "Generating go.sum files" ./scripts/generate-go-sums-improved.sh
elif [ -x "scripts/docker-generate-go-sums.sh" ]; then
    run_command "Generating go.sum files" ./scripts/docker-generate-go-sums.sh
else
    show_error "go.sum generation script not found or not executable"
    exit 1
fi

# Step 2: Build and deploy with docker-compose
echo -e "\n${YELLOW}Building and deploying services...${NC}"

if [ "$VERBOSE" = true ]; then
    # Verbose mode - show Docker output
    COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 docker-compose up -d --build
else
    # Quiet mode - use BuildKit inline progress
    export COMPOSE_DOCKER_CLI_BUILD=1
    export DOCKER_BUILDKIT=1
    export BUILDKIT_PROGRESS=plain
    
    # Run docker-compose with minimal output
    show_progress "Building Docker images"
    if docker-compose build --quiet 2>&1 | grep -E "(ERROR|error|failed)" > /tmp/docker-errors.log; then
        if [ -s /tmp/docker-errors.log ]; then
            show_error "Build errors detected:"
            cat /tmp/docker-errors.log
            rm -f /tmp/docker-errors.log
            exit 1
        fi
    fi
    show_success "Docker images built"
    
    show_progress "Starting containers"
    if docker-compose up -d --quiet-pull 2>&1 | grep -E "(ERROR|error|failed)" > /tmp/docker-errors.log; then
        if [ -s /tmp/docker-errors.log ]; then
            show_error "Deployment errors detected:"
            cat /tmp/docker-errors.log
            rm -f /tmp/docker-errors.log
            exit 1
        fi
    fi
    show_success "Containers started"
    rm -f /tmp/docker-errors.log
fi

# Step 3: Check container status
echo -e "\n${YELLOW}Checking container status...${NC}"
sleep 3  # Give containers time to start

# Get container status
RUNNING_CONTAINERS=$(docker-compose ps --services --filter "status=running" 2>/dev/null | wc -l)
TOTAL_CONTAINERS=$(docker-compose ps --services 2>/dev/null | wc -l)

if [ "$RUNNING_CONTAINERS" -eq "$TOTAL_CONTAINERS" ]; then
    show_success "All $RUNNING_CONTAINERS containers are running"
else
    show_error "Only $RUNNING_CONTAINERS of $TOTAL_CONTAINERS containers are running"
    echo -e "\n${YELLOW}Container status:${NC}"
    docker-compose ps
fi

# Step 4: Show access URLs
echo -e "\n${GREEN}Deployment complete!${NC}"
echo -e "${YELLOW}Services are available at:${NC}"
echo "  • API Gateway: http://localhost:8082"
echo "  • Storage Service: http://localhost:8080"
echo "  • Deployment Service: http://localhost:8081"
echo "  • API Key Service: http://localhost:8083"
echo "  • Metering Service: http://localhost:8084"
echo "  • Billing Service: http://localhost:8085"
echo "  • Marketplace UI: http://localhost:3000"
echo "  • Creator Portal: http://localhost:3001"

# Show logs command
echo -e "\n${YELLOW}To view logs:${NC}"
echo "  docker-compose logs -f [service-name]"

# Show verbose tip if in quiet mode
if [ "$VERBOSE" = false ]; then
    echo -e "\n${BLUE}Tip: Use -v flag for verbose output${NC}"
fi
