#!/bin/bash

# Development Deployment Script for API Marketplace
set -e

echo "🚀 Starting development deployment..."

# Check if .env.local exists, create from example if not
if [[ ! -f .env.local ]]; then
    if [[ -f .env.example ]]; then
        echo "📝 Creating .env.local from .env.example..."
        cp .env.example .env.local
        echo "⚠️  Please update .env.local with your development configuration"
    else
        echo "❌ Error: Neither .env.local nor .env.example found!"
        exit 1
    fi
fi

# Load development environment variables
set -a
source .env.local
set +a

# Set development defaults
export NODE_ENV=${NODE_ENV:-development}
export MOUNT_SOURCE=${MOUNT_SOURCE:-rw}

# Create necessary directories
echo "📁 Creating necessary directories..."
mkdir -p logs
mkdir -p data/postgres
mkdir -p data/redis
mkdir -p data/elasticsearch

# Stop any existing containers
echo "🛑 Stopping existing containers..."
docker-compose down 2>/dev/null || true

# Build and start development services
echo "🔨 Building and starting development services..."
docker-compose up --build -d

# Wait for core services to be ready
echo "⏳ Waiting for core services to be ready..."
sleep 10

# Check database connection
echo "🗄️  Checking database connection..."
max_attempts=30
attempt=1
while [[ $attempt -le $max_attempts ]]; do
    if docker-compose exec -T postgres pg_isready -U apidirect > /dev/null 2>&1; then
        echo "✅ Database is ready"
        break
    else
        echo "Waiting for database... (attempt $attempt/$max_attempts)"
        sleep 2
        attempt=$((attempt + 1))
    fi
done

if [[ $attempt -gt $max_attempts ]]; then
    echo "❌ Database failed to start within expected time"
    docker-compose logs postgres
    exit 1
fi

# Check Redis connection
echo "🔧 Checking Redis connection..."
if docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; then
    echo "✅ Redis is ready"
else
    echo "❌ Redis connection failed"
    docker-compose logs redis
fi

# Display service status
echo "🔍 Service status:"
docker-compose ps

# Display useful information
echo ""
echo "🎉 Development environment is ready!"
echo ""
echo "📊 Service URLs:"
echo "  • Marketplace:       http://localhost:3001"
echo "  • Creator Portal:    http://localhost:3000"
echo "  • API Gateway:       http://localhost:8082"
echo "  • API Key Service:   http://localhost:8083"
echo "  • Metering Service:  http://localhost:8084"
echo "  • Billing Service:   http://localhost:8085"
echo "  • Marketplace API:   http://localhost:8086"
echo "  • Storage Service:   http://localhost:8087"
echo "  • Elasticsearch:     http://localhost:9200"
echo "  • Kibana:           http://localhost:5601"
echo ""
echo "🗄️  Database Info:"
echo "  • Host: localhost:5432"
echo "  • Database: apidirect"
echo "  • User: apidirect"
echo "  • Password: localpassword"
echo ""
echo "📋 Useful Commands:"
echo "  • View logs:         docker-compose logs -f [service]"
echo "  • Restart service:   docker-compose restart [service]"
echo "  • Stop all:          docker-compose down"
echo "  • Rebuild:           docker-compose up --build -d [service]"
echo "  • Database shell:    docker-compose exec postgres psql -U apidirect -d apidirect"
echo "  • Redis CLI:         docker-compose exec redis redis-cli"
echo ""
echo "🧪 Testing:"
echo "  • Run E2E tests:     cd testing/e2e && npm test"
echo "  • API health:        curl http://localhost:8082/health"
echo ""

# Check if we can access the marketplace
echo "🏥 Quick health check..."
sleep 5
if curl -f "http://localhost:3001" > /dev/null 2>&1; then
    echo "✅ Marketplace is accessible"
else
    echo "⚠️  Marketplace might still be starting up"
fi

# Optional: Open browser to marketplace
if command -v open >/dev/null 2>&1; then
    echo "🌐 Opening marketplace in browser..."
    open http://localhost:3001
elif command -v xdg-open >/dev/null 2>&1; then
    echo "🌐 Opening marketplace in browser..."
    xdg-open http://localhost:3001
fi

echo "✨ Development environment is ready for coding!"