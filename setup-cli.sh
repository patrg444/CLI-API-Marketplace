#!/bin/bash

echo "🚀 API-Direct CLI Setup"
echo "======================="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if CLI is built
if [ ! -f "cli/apidirect" ]; then
    echo -e "${YELLOW}CLI not built. Building now...${NC}"
    cd cli
    make build
    cd ..
fi

# Create config directory
CONFIG_DIR="$HOME/.apidirect"
mkdir -p "$CONFIG_DIR"

# Copy default config if it doesn't exist
if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
    cp cli/config/apidirect.yaml "$CONFIG_DIR/config.yaml"
    echo -e "${GREEN}✓ Created config file at $CONFIG_DIR/config.yaml${NC}"
fi

# Create CLI symlink for easy access
echo ""
echo "Would you like to install the CLI globally? (y/n)"
read -r response
if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    sudo cp cli/apidirect /usr/local/bin/
    echo -e "${GREEN}✓ CLI installed to /usr/local/bin/apidirect${NC}"
else
    echo -e "${YELLOW}You can run the CLI using: ./cli/apidirect${NC}"
fi

echo ""
echo "🔐 AWS Cognito Configuration"
echo "============================"
echo ""
echo "To use the CLI with AWS Cognito authentication, you'll need to set these environment variables:"
echo ""
echo "export APIDIRECT_COGNITO_POOL=\"your-user-pool-id\""
echo "export APIDIRECT_COGNITO_CLIENT=\"your-client-id\""
echo "export APIDIRECT_AUTH_DOMAIN=\"https://your-domain.auth.us-east-1.amazoncognito.com\""
echo "export APIDIRECT_REGION=\"us-east-1\""
echo ""
echo "Add these to your ~/.bashrc or ~/.zshrc file to make them permanent."
echo ""

# Check if environment variables are already set
if [ -n "$APIDIRECT_COGNITO_POOL" ]; then
    echo -e "${GREEN}✓ APIDIRECT_COGNITO_POOL is set${NC}"
else
    echo -e "${RED}✗ APIDIRECT_COGNITO_POOL is not set${NC}"
fi

if [ -n "$APIDIRECT_COGNITO_CLIENT" ]; then
    echo -e "${GREEN}✓ APIDIRECT_COGNITO_CLIENT is set${NC}"
else
    echo -e "${RED}✗ APIDIRECT_COGNITO_CLIENT is not set${NC}"
fi

if [ -n "$APIDIRECT_AUTH_DOMAIN" ]; then
    echo -e "${GREEN}✓ APIDIRECT_AUTH_DOMAIN is set${NC}"
else
    echo -e "${RED}✗ APIDIRECT_AUTH_DOMAIN is not set${NC}"
fi

echo ""
echo "📚 CLI Commands"
echo "==============="
echo ""
echo "Once configured with AWS Cognito:"
echo "  apidirect login          - Authenticate with API-Direct"
echo "  apidirect init           - Create a new API project"
echo "  apidirect deploy         - Deploy your API"
echo "  apidirect logs           - View API logs"
echo "  apidirect marketplace    - Browse the API marketplace"
echo ""
echo "For local development (before AWS setup):"
echo "  apidirect validate       - Validate your apidirect.yaml"
echo "  apidirect run           - Run your API locally"
echo "  apidirect docs          - Generate API documentation"
echo ""

# Test CLI
echo "🧪 Testing CLI..."
if command -v apidirect &> /dev/null; then
    echo -e "${GREEN}✓ CLI is accessible${NC}"
    apidirect version
else
    echo -e "${YELLOW}Run ./cli/apidirect to use the CLI${NC}"
    ./cli/apidirect version
fi

echo ""
echo "✅ CLI setup complete!"
echo ""
echo "Next steps:"
echo "1. Set up AWS Cognito environment variables"
echo "2. Run 'apidirect login' to authenticate"
echo "3. Create your first API with 'apidirect init'"