#!/bin/bash

# 🚀 Complete Production Deployment Script for API-Direct
# This script handles the entire production deployment process

set -e  # Exit on error

echo "🚀 API-Direct Production Deployment"
echo "==================================="

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to prompt for input with default
prompt_with_default() {
    local prompt="$1"
    local default="$2"
    local var_name="$3"
    
    read -p "$prompt [$default]: " value
    value="${value:-$default}"
    eval "$var_name='$value'"
}

# Check prerequisites
echo "📋 Checking prerequisites..."

if ! command_exists docker; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    exit 1
fi

if ! command_exists docker-compose; then
    echo -e "${RED}❌ Docker Compose is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✅ All prerequisites met${NC}"

# Step 1: Environment Configuration
echo ""
echo "📝 Step 1: Environment Configuration"
echo "-----------------------------------"

if [ ! -f .env.production ]; then
    echo "Creating .env.production from template..."
    cp .env.production.example .env.production
fi

# Generate secure secrets if not already done
if grep -q "CHANGE_THIS" .env.production; then
    echo "🔐 Generating secure secrets..."
    ./scripts/generate-secrets.sh > secrets.txt
    echo -e "${YELLOW}⚠️  Secrets generated in secrets.txt - update .env.production manually${NC}"
    echo "Press Enter when you've updated .env.production with the secrets..."
    read
fi

# Prompt for critical configuration
echo ""
echo "Please provide the following configuration:"

prompt_with_default "Your domain (e.g., apidirect.dev)" "apidirect.dev" DOMAIN
prompt_with_default "Server IP address" "" SERVER_IP
prompt_with_default "Stripe Secret Key" "sk_live_..." STRIPE_KEY
prompt_with_default "AWS Access Key ID" "" AWS_ACCESS_KEY
prompt_with_default "AWS Secret Access Key" "" AWS_SECRET_KEY

# Update .env.production with provided values
sed -i.bak "s/DOMAIN=.*/DOMAIN=$DOMAIN/" .env.production
sed -i.bak "s/STRIPE_SECRET_KEY=.*/STRIPE_SECRET_KEY=$STRIPE_KEY/" .env.production
sed -i.bak "s/AWS_ACCESS_KEY_ID=.*/AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY/" .env.production
sed -i.bak "s/AWS_SECRET_ACCESS_KEY=.*/AWS_SECRET_ACCESS_KEY=$AWS_SECRET_KEY/" .env.production

# Step 2: Server Setup (if remote deployment)
if [ -n "$SERVER_IP" ]; then
    echo ""
    echo "📡 Step 2: Remote Server Setup"
    echo "-----------------------------"
    
    echo "Setting up production server at $SERVER_IP..."
    
    # Create deployment script for remote server
    cat > remote-setup.sh << 'EOF'
#!/bin/bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
if ! command -v docker >/dev/null 2>&1; then
    curl -fsSL https://get.docker.com | sudo sh
    sudo usermod -aG docker $USER
fi

# Install Docker Compose
if ! command -v docker-compose >/dev/null 2>&1; then
    sudo apt install docker-compose -y
fi

# Install Certbot for SSL
sudo apt install certbot python3-certbot-nginx -y

# Create app directory
mkdir -p ~/api-direct
EOF

    # Copy to server and execute
    scp remote-setup.sh root@$SERVER_IP:/tmp/
    ssh root@$SERVER_IP "chmod +x /tmp/remote-setup.sh && /tmp/remote-setup.sh"
    
    # Copy application files
    echo "Copying application files..."
    rsync -avz --exclude='.git' --exclude='node_modules' --exclude='venv' \
        ./ root@$SERVER_IP:~/api-direct/
else
    echo ""
    echo "📡 Step 2: Local Deployment"
    echo "--------------------------"
fi

# Step 3: SSL Certificate Setup
echo ""
echo "🔒 Step 3: SSL Certificate Setup"
echo "--------------------------------"

if [ -n "$SERVER_IP" ]; then
    echo "Setting up SSL certificates on remote server..."
    ssh root@$SERVER_IP << EOF
cd ~/api-direct
sudo certbot certonly --standalone \
    -d $DOMAIN \
    -d www.$DOMAIN \
    -d console.$DOMAIN \
    -d marketplace.$DOMAIN \
    -d api.$DOMAIN \
    --email admin@$DOMAIN \
    --agree-tos \
    --no-eff-email
EOF
else
    echo "For local deployment, using self-signed certificates..."
    ./scripts/generate-local-certs.sh
fi

# Step 4: Build and Deploy
echo ""
echo "🏗️  Step 4: Building and Deploying Services"
echo "-------------------------------------------"

# Build all Docker images
echo "Building Docker images..."
docker-compose -f docker-compose.production.yml build

# Run database migrations
echo "Running database migrations..."
docker-compose -f docker-compose.production.yml run --rm backend \
    python -m alembic upgrade head

# Start all services
echo "Starting all services..."
docker-compose -f docker-compose.production.yml up -d

# Wait for services to be healthy
echo "Waiting for services to be healthy..."
sleep 30

# Step 5: Verify Deployment
echo ""
echo "✅ Step 5: Verifying Deployment"
echo "-------------------------------"

# Check service health
./scripts/verify-deployment.sh

# Step 6: Post-Deployment Setup
echo ""
echo "🔧 Step 6: Post-Deployment Configuration"
echo "---------------------------------------"

# Set up cron jobs
echo "Setting up automated backups..."
(crontab -l 2>/dev/null; echo "0 2 * * * cd $(pwd) && ./scripts/backup.sh") | crontab -

# Set up monitoring alerts
echo "Configuring monitoring..."
docker-compose -f docker-compose.production.yml exec grafana \
    grafana-cli admin reset-admin-password admin

echo ""
echo "🎉 Deployment Complete!"
echo "======================"
echo ""
echo "Access your services at:"
echo "  Main Site: https://$DOMAIN"
echo "  Console: https://console.$DOMAIN"
echo "  Marketplace: https://marketplace.$DOMAIN"
echo "  API: https://api.$DOMAIN"
echo "  Grafana: https://$DOMAIN:3000 (admin/admin)"
echo ""
echo "Next Steps:"
echo "1. Update DNS records to point to: $SERVER_IP"
echo "2. Test all endpoints"
echo "3. Configure Stripe webhooks"
echo "4. Set up email service"
echo "5. Enable backups"
echo ""
echo "📚 Documentation: https://docs.$DOMAIN"
echo "🐛 Issues: https://github.com/yourusername/CLI-API-Marketplace/issues"

# Create deployment summary
cat > deployment-summary.md << EOF
# API-Direct Deployment Summary

**Date**: $(date)
**Domain**: $DOMAIN
**Server**: ${SERVER_IP:-localhost}

## Services Deployed
- ✅ PostgreSQL Database
- ✅ Redis Cache
- ✅ Backend API
- ✅ Frontend Applications
- ✅ Monitoring Stack

## Configuration
- SSL: Let's Encrypt
- Backups: Daily at 2 AM
- Monitoring: Prometheus + Grafana

## Post-Deployment Checklist
- [ ] DNS records updated
- [ ] Stripe webhooks configured
- [ ] Email service tested
- [ ] Health checks passing
- [ ] Monitoring alerts configured
EOF

echo ""
echo "📄 Deployment summary saved to deployment-summary.md"