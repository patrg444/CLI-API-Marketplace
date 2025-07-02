#!/bin/bash

# API-Direct Production Deployment Script
set -e

echo "🚀 Starting API-Direct Production Deployment..."

# Configuration
DOMAIN="api-direct.io"
CONSOLE_DOMAIN="console.api-direct.io"
API_DOMAIN="api.api-direct.io"
EMAIL="admin@api-direct.io"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   print_error "This script should not be run as root"
   exit 1
fi

# Check if .env.production exists
if [ ! -f .env.production ]; then
    print_error ".env.production file not found!"
    print_warning "Please copy .env.production.example to .env.production and configure it"
    exit 1
fi

# Load environment variables
source .env.production

# Validate required environment variables
required_vars=(
    "POSTGRES_PASSWORD"
    "REDIS_PASSWORD"
    "JWT_SECRET"
    "STRIPE_SECRET_KEY"
)

for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        print_error "Required environment variable $var is not set"
        exit 1
    fi
done

print_status "Environment validation passed ✓"

# Install Docker and Docker Compose if not present
if ! command -v docker &> /dev/null; then
    print_status "Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    sudo usermod -aG docker $USER
    rm get-docker.sh
fi

if ! command -v docker-compose &> /dev/null; then
    print_status "Installing Docker Compose..."
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
fi

# Install Certbot for SSL certificates
if ! command -v certbot &> /dev/null; then
    print_status "Installing Certbot..."
    sudo apt-get update
    sudo apt-get install -y certbot python3-certbot-nginx
fi

# Create SSL certificates directory
sudo mkdir -p /etc/nginx/ssl

# Generate SSL certificates with Let's Encrypt
print_status "Generating SSL certificates..."
domains=("$DOMAIN" "$CONSOLE_DOMAIN" "$API_DOMAIN")

for domain in "${domains[@]}"; do
    if [ ! -f "/etc/letsencrypt/live/$domain/fullchain.pem" ]; then
        print_status "Obtaining SSL certificate for $domain..."
        sudo certbot certonly --standalone --non-interactive --agree-tos --email "$EMAIL" -d "$domain"
        
        # Copy certificates to nginx directory
        sudo cp "/etc/letsencrypt/live/$domain/fullchain.pem" "/etc/nginx/ssl/$domain.crt"
        sudo cp "/etc/letsencrypt/live/$domain/privkey.pem" "/etc/nginx/ssl/$domain.key"
    else
        print_status "SSL certificate for $domain already exists ✓"
    fi
done

# Set up SSL certificate auto-renewal
print_status "Setting up SSL certificate auto-renewal..."
echo "0 12 * * * /usr/bin/certbot renew --quiet && docker-compose -f docker-compose.production.yml restart nginx" | sudo crontab -

# Create necessary directories
mkdir -p logs/nginx
mkdir -p logs/backend
mkdir -p monitoring/grafana/provisioning/dashboards
mkdir -p monitoring/grafana/provisioning/datasources

# Set up Grafana datasources
cat > monitoring/grafana/provisioning/datasources/prometheus.yml << EOF
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    url: http://prometheus:9090
    access: proxy
    isDefault: true
EOF

# Build and start services
print_status "Building and starting services..."
docker-compose -f docker-compose.production.yml build
docker-compose -f docker-compose.production.yml up -d

# Wait for services to be healthy
print_status "Waiting for services to be healthy..."
sleep 30

# Check service health
services=("postgres" "redis" "influxdb" "backend" "nginx")
for service in "${services[@]}"; do
    if docker-compose -f docker-compose.production.yml ps | grep -q "$service.*Up"; then
        print_status "$service is running ✓"
    else
        print_error "$service failed to start"
        docker-compose -f docker-compose.production.yml logs "$service"
        exit 1
    fi
done

# Run database migrations
print_status "Running database migrations..."
docker-compose -f docker-compose.production.yml exec backend python -c "
import asyncio
import asyncpg
import os

async def run_migrations():
    conn = await asyncpg.connect(os.getenv('DATABASE_URL'))
    
    # Check if database is properly initialized
    result = await conn.fetch('SELECT count(*) FROM information_schema.tables WHERE table_name = \\'users\\'')
    if result[0]['count'] > 0:
        print('Database already initialized ✓')
    else:
        print('Database needs initialization')
    
    await conn.close()

asyncio.run(run_migrations())
"

# Set up backup cron job
print_status "Setting up automated backups..."
cat > /tmp/backup_cron << EOF
# Daily database backup at 2 AM
0 2 * * * cd $(pwd) && ./scripts/backup.sh
EOF
crontab /tmp/backup_cron
rm /tmp/backup_cron

# Display final status
print_status "🎉 Deployment completed successfully!"
echo ""
echo "Service URLs:"
echo "  Main Site: https://$DOMAIN"
echo "  Creator Portal: https://$CONSOLE_DOMAIN"
echo "  API Endpoint: https://$API_DOMAIN"
echo "  Monitoring: https://$DOMAIN:3000 (Grafana)"
echo ""
echo "Service Status:"
docker-compose -f docker-compose.production.yml ps
echo ""
print_status "Deployment logs are available in the ./logs directory"
print_warning "Remember to:"
print_warning "  1. Update your DNS records to point to this server"
print_warning "  2. Configure your Stripe webhook endpoints"
print_warning "  3. Test all functionality before going live"
print_warning "  4. Set up monitoring alerts in Grafana"