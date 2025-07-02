#!/bin/bash

# Test build script for API Direct CLI
# This script tests the build process for all platforms

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🧪 Testing API Direct CLI Build Process${NC}"
echo "========================================"

# Change to CLI directory
cd "$(dirname "$0")/.."

# Run tests first
echo -e "\n${YELLOW}Running unit tests...${NC}"
go test -v -cover ./...

# Test build for current platform
echo -e "\n${YELLOW}Testing build for current platform...${NC}"
go build -o test-build/apidirect main.go

if [ -f "test-build/apidirect" ]; then
    echo -e "${GREEN}✅ Build successful${NC}"
    
    # Test the binary
    echo -e "\n${YELLOW}Testing binary...${NC}"
    ./test-build/apidirect --version
    ./test-build/apidirect --help
else
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi

# Test cross-compilation
echo -e "\n${YELLOW}Testing cross-compilation...${NC}"

PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    OUTPUT="test-build/apidirect-${GOOS}-${GOARCH}"
    
    if [ "$GOOS" = "windows" ]; then
        OUTPUT="${OUTPUT}.exe"
    fi
    
    echo -e "Building for ${GOOS}/${GOARCH}..."
    
    if GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT" main.go; then
        echo -e "  ${GREEN}✓ Success${NC}"
    else
        echo -e "  ${RED}✗ Failed${NC}"
    fi
done

# Test Docker build
if command -v docker &> /dev/null; then
    echo -e "\n${YELLOW}Testing Docker build...${NC}"
    
    # Build Docker image
    docker build -f ../docker/Dockerfile -t apidirect-test:latest ../
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Docker build successful${NC}"
        
        # Test Docker image
        echo -e "\n${YELLOW}Testing Docker image...${NC}"
        docker run --rm apidirect-test:latest --version
    else
        echo -e "${RED}❌ Docker build failed${NC}"
    fi
else
    echo -e "\n${YELLOW}⚠️  Docker not found, skipping Docker tests${NC}"
fi

# Test Makefile targets
if command -v make &> /dev/null; then
    echo -e "\n${YELLOW}Testing Makefile targets...${NC}"
    
    # Test various make targets
    make_targets=("fmt" "vet" "test")
    
    for target in "${make_targets[@]}"; do
        echo -e "Testing 'make $target'..."
        if make $target; then
            echo -e "  ${GREEN}✓ Success${NC}"
        else
            echo -e "  ${RED}✗ Failed${NC}"
        fi
    done
else
    echo -e "\n${YELLOW}⚠️  Make not found, skipping Makefile tests${NC}"
fi

# Test shell completions
echo -e "\n${YELLOW}Testing shell completions...${NC}"
./test-build/apidirect completion bash > /dev/null
if [ $? -eq 0 ]; then
    echo -e "  ${GREEN}✓ Bash completion works${NC}"
else
    echo -e "  ${RED}✗ Bash completion failed${NC}"
fi

./test-build/apidirect completion zsh > /dev/null
if [ $? -eq 0 ]; then
    echo -e "  ${GREEN}✓ Zsh completion works${NC}"
else
    echo -e "  ${RED}✗ Zsh completion failed${NC}"
fi

# Clean up
echo -e "\n${YELLOW}Cleaning up...${NC}"
rm -rf test-build

echo -e "\n${GREEN}✅ All build tests completed!${NC}"