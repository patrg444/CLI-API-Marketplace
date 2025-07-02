#!/bin/bash

# Production Deployment Script for API Marketplace
set -e

echo "🚀 Starting production deployment..."

# Check if .env.production exists
if [[ ! -f .env.production ]]; then
    echo "❌ Error: .env.production file not found!"
    echo "Please create .env.production with all required environment variables."
    echo "You can use .env.example as a template."
    exit 1
fi

# Load production environment variables
set -a
source .env.production
set +a

# Verify required environment variables
required_vars=(
    "POSTGRES_PASSWORD"
    "REDIS_PASSWORD"
    "NEXTAUTH_SECRET"
    "STRIPE_SECRET_KEY"
    "STRIPE_PUBLISHABLE_KEY"
    "COGNITO_USER_POOL_ID"
    "COGNITO_WEB_CLIENT_ID"
    "AWS_REGION"
)

echo "🔍 Checking required environment variables..."
for var in "${required_vars[@]}"; do
    if [[ -z "${!var}" ]]; then
        echo "❌ Error: $var is not set in .env.production"
        exit 1
    fi
done
echo "✅ All required environment variables are set"

# Create necessary directories
echo "📁 Creating necessary directories..."
mkdir -p logs/nginx
mkdir -p nginx/ssl
mkdir -p monitoring/grafana/provisioning/{dashboards,datasources}

# Check if SSL certificates exist
if [[ ! -f nginx/ssl/cert.pem ]] || [[ ! -f nginx/ssl/key.pem ]]; then
    echo "⚠️  Warning: SSL certificates not found in nginx/ssl/"
    echo "Please add your SSL certificates or update nginx configuration for HTTP-only"
fi

# Build and start services
echo "🔨 Building and starting production services..."
docker-compose -f docker-compose.production.yml down
docker-compose -f docker-compose.production.yml build --no-cache
docker-compose -f docker-compose.production.yml up -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be healthy..."
timeout=300
elapsed=0
while [[ $elapsed -lt $timeout ]]; do
    if docker-compose -f docker-compose.production.yml ps | grep -q "unhealthy\|starting"; then
        echo "Waiting for services to start... (${elapsed}s/${timeout}s)"
        sleep 10
        elapsed=$((elapsed + 10))
    else
        break
    fi
done

# Check service status
echo "🔍 Checking service status..."
docker-compose -f docker-compose.production.yml ps

# Run database migrations if needed
echo "🗄️  Running database migrations..."
docker-compose -f docker-compose.production.yml exec -T postgres psql -U apidirect -d apidirect -c "SELECT version();" > /dev/null 2>&1
if [[ $? -eq 0 ]]; then
    echo "✅ Database connection successful"
else
    echo "❌ Database connection failed"
    exit 1
fi

# Test health endpoints
echo "🏥 Testing health endpoints..."
services=("marketplace:3000" "backend:8000")
for service in "${services[@]}"; do
    service_name=$(echo $service | cut -d: -f1)
    port=$(echo $service | cut -d: -f2)
    
    if curl -f "http://localhost:${port}/api/health" > /dev/null 2>&1 || curl -f "http://localhost:${port}/health" > /dev/null 2>&1; then
        echo "✅ $service_name health check passed"
    else
        echo "❌ $service_name health check failed"
    fi
done

# Display deployment information
echo ""
echo "🎉 Production deployment completed!"
echo ""
echo "📊 Service URLs:"
echo "  • Marketplace:  http://localhost:3001"
echo "  • API Backend:  http://localhost:8000"
echo "  • Grafana:      http://localhost:3000"
echo "  • Prometheus:   http://localhost:9090"
echo ""
echo "📋 Management Commands:"
echo "  • View logs:    docker-compose -f docker-compose.production.yml logs -f [service]"
echo "  • Stop:         docker-compose -f docker-compose.production.yml down"
echo "  • Restart:      docker-compose -f docker-compose.production.yml restart [service]"
echo ""
echo "⚡ Quick health check: curl http://localhost:3001/api/health"
echo ""

# Optional: Open browser to marketplace
if command -v open >/dev/null 2>&1; then
    echo "🌐 Opening marketplace in browser..."
    open http://localhost:3001
elif command -v xdg-open >/dev/null 2>&1; then
    echo "🌐 Opening marketplace in browser..."
    xdg-open http://localhost:3001
fi

echo "✨ Deployment complete! Check the logs if any services appear unhealthy."