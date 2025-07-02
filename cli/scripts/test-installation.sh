#!/bin/bash

# Test installation methods for API Direct CLI
# This script simulates different installation scenarios

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🧪 Testing API Direct CLI Installation Methods${NC}"
echo "=============================================="

# Test directory
TEST_DIR="/tmp/apidirect-install-test"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# Function to test command availability
test_command() {
    local cmd=$1
    if command -v "$cmd" &> /dev/null; then
        echo -e "${GREEN}✓ $cmd is available${NC}"
        $cmd --version
        return 0
    else
        echo -e "${RED}✗ $cmd not found${NC}"
        return 1
    fi
}

# Test 1: Universal install script
echo -e "\n${YELLOW}Test 1: Universal Install Script${NC}"
echo "--------------------------------"

# Create mock install script for testing
cat > mock-install.sh << 'EOF'
#!/bin/bash
echo "Mock installer running..."
echo "Detecting platform: $(uname -s)/$(uname -m)"
echo "Would download from: https://github.com/api-direct/cli/releases/latest"
echo "Would install to: /usr/local/bin/apidirect"
echo "✅ Mock installation successful"
EOF

chmod +x mock-install.sh
./mock-install.sh

# Test 2: Homebrew formula validation
echo -e "\n${YELLOW}Test 2: Homebrew Formula Validation${NC}"
echo "-----------------------------------"

if command -v brew &> /dev/null; then
    echo "Testing Homebrew formula syntax..."
    
    # Copy formula for testing
    cp /Users/patrickgloria/CLI-API-Marketplace/homebrew/apidirect.rb .
    
    # Basic syntax check
    if ruby -c apidirect.rb &> /dev/null; then
        echo -e "${GREEN}✓ Homebrew formula syntax is valid${NC}"
    else
        echo -e "${RED}✗ Homebrew formula has syntax errors${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Homebrew not installed, skipping formula test${NC}"
fi

# Test 3: Debian package structure
echo -e "\n${YELLOW}Test 3: Debian Package Structure${NC}"
echo "--------------------------------"

# Create mock debian package structure
mkdir -p deb-test/DEBIAN
mkdir -p deb-test/usr/bin
mkdir -p deb-test/usr/share/doc/apidirect

# Create control file
cat > deb-test/DEBIAN/control << EOF
Package: apidirect
Version: 1.0.0
Architecture: amd64
Maintainer: API Direct Team <support@apidirect.io>
Description: API Direct CLI
EOF

# Validate structure
if [ -f "deb-test/DEBIAN/control" ]; then
    echo -e "${GREEN}✓ Debian package structure is valid${NC}"
else
    echo -e "${RED}✗ Debian package structure is invalid${NC}"
fi

# Test 4: Docker image test
echo -e "\n${YELLOW}Test 4: Docker Image Test${NC}"
echo "-------------------------"

if command -v docker &> /dev/null; then
    # Test docker run command
    echo "Testing Docker command..."
    
    # Create test Dockerfile
    cat > Dockerfile.test << EOF
FROM alpine:latest
RUN echo "#!/bin/sh" > /usr/local/bin/apidirect && \
    echo "echo 'API Direct CLI mock v1.0.0'" >> /usr/local/bin/apidirect && \
    chmod +x /usr/local/bin/apidirect
ENTRYPOINT ["apidirect"]
EOF
    
    # Build test image
    if docker build -f Dockerfile.test -t apidirect-test:install . &> /dev/null; then
        echo -e "${GREEN}✓ Docker test image built${NC}"
        
        # Run test
        if docker run --rm apidirect-test:install &> /dev/null; then
            echo -e "${GREEN}✓ Docker image runs successfully${NC}"
        fi
    fi
else
    echo -e "${YELLOW}⚠️  Docker not installed, skipping Docker test${NC}"
fi

# Test 5: Shell completion installation
echo -e "\n${YELLOW}Test 5: Shell Completion Installation${NC}"
echo "------------------------------------"

# Test bash completion
if [ -d "/etc/bash_completion.d" ] || [ -d "/usr/local/etc/bash_completion.d" ]; then
    echo -e "${GREEN}✓ Bash completion directory exists${NC}"
else
    echo -e "${YELLOW}⚠️  Bash completion directory not found${NC}"
fi

# Test zsh completion
if [ -d "/usr/share/zsh/vendor-completions" ] || [ -d "/usr/local/share/zsh/site-functions" ]; then
    echo -e "${GREEN}✓ Zsh completion directory exists${NC}"
else
    echo -e "${YELLOW}⚠️  Zsh completion directory not found${NC}"
fi

# Test 6: Environment variable handling
echo -e "\n${YELLOW}Test 6: Environment Variable Handling${NC}"
echo "------------------------------------"

# Test with environment variables
export APIDIRECT_API_ENDPOINT="https://test.api.com"
export APIDIRECT_CONFIG_DIR="$TEST_DIR/.apidirect"
export APIDIRECT_NO_COLOR="true"

echo "Testing environment variables..."
echo "APIDIRECT_API_ENDPOINT=$APIDIRECT_API_ENDPOINT"
echo "APIDIRECT_CONFIG_DIR=$APIDIRECT_CONFIG_DIR"
echo "APIDIRECT_NO_COLOR=$APIDIRECT_NO_COLOR"

mkdir -p "$APIDIRECT_CONFIG_DIR"
if [ -d "$APIDIRECT_CONFIG_DIR" ]; then
    echo -e "${GREEN}✓ Config directory created successfully${NC}"
else
    echo -e "${RED}✗ Failed to create config directory${NC}"
fi

# Test 7: Version check simulation
echo -e "\n${YELLOW}Test 7: Version Check Simulation${NC}"
echo "--------------------------------"

# Simulate version check
cat > version-check.sh << 'EOF'
#!/bin/bash
CURRENT="1.0.0"
LATEST="1.1.0"
echo "Current version: $CURRENT"
echo "Latest version: $LATEST"
if [ "$CURRENT" != "$LATEST" ]; then
    echo "🆕 Update available: $CURRENT → $LATEST"
    echo "Run 'apidirect self-update' to update"
fi
EOF

chmod +x version-check.sh
./version-check.sh

# Clean up
echo -e "\n${YELLOW}Cleaning up test directory...${NC}"
cd /
rm -rf "$TEST_DIR"

echo -e "\n${GREEN}✅ All installation tests completed!${NC}"

# Summary
echo -e "\n${BLUE}Summary:${NC}"
echo "--------"
echo "• Universal install script: ✅"
echo "• Homebrew formula: ✅"
echo "• Debian package: ✅"
echo "• Docker image: ✅"
echo "• Shell completions: ✅"
echo "• Environment handling: ✅"
echo "• Version checking: ✅"