#!/bin/bash

# API-Direct Health Check Script
# Run this script to verify all services are running correctly

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_status() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

# Health check function
check_service() {
    local service_name=$1
    local url=$2
    local expected_status=${3:-200}
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null)
    
    if [ "$response" = "$expected_status" ]; then
        print_status "$service_name is healthy"
        return 0
    else
        print_error "$service_name is unhealthy (HTTP $response)"
        return 1
    fi
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running"
    exit 1
fi

print_info "🔍 Starting API-Direct Health Check..."
echo ""

# Check Docker containers
print_info "Checking Docker containers..."
healthy_containers=0
total_containers=0

containers=("apidirect-postgres" "apidirect-redis" "apidirect-influxdb" "apidirect-backend" "apidirect-nginx")

for container in "${containers[@]}"; do
    total_containers=$((total_containers + 1))
    if docker ps --format "table {{.Names}}" | grep -q "$container"; then
        status=$(docker inspect --format='{{.State.Health.Status}}' "$container" 2>/dev/null || echo "no-healthcheck")
        
        if [ "$status" = "healthy" ] || [ "$status" = "no-healthcheck" ]; then
            print_status "$container is running"
            healthy_containers=$((healthy_containers + 1))
        else
            print_error "$container is unhealthy (status: $status)"
        fi
    else
        print_error "$container is not running"
    fi
done

echo ""

# Check service endpoints
print_info "Checking service endpoints..."

# Backend API
check_service "Backend API" "http://localhost:8000/health"

# Main website
check_service "Main Website" "https://api-direct.io" 200

# Creator Portal
check_service "Creator Portal" "https://console.api-direct.io" 200

# Database connection
print_info "Checking database connection..."
if docker-compose -f docker-compose.production.yml exec -T postgres pg_isready -U apidirect > /dev/null 2>&1; then
    print_status "PostgreSQL is ready"
else
    print_error "PostgreSQL is not ready"
fi

# Redis connection
print_info "Checking Redis connection..."
if docker-compose -f docker-compose.production.yml exec -T redis redis-cli ping | grep -q "PONG"; then
    print_status "Redis is ready"
else
    print_error "Redis is not ready"
fi

# InfluxDB connection
print_info "Checking InfluxDB connection..."
if docker-compose -f docker-compose.production.yml exec -T influxdb influx ping > /dev/null 2>&1; then
    print_status "InfluxDB is ready"
else
    print_error "InfluxDB is not ready"
fi

echo ""

# SSL certificate check
print_info "Checking SSL certificates..."
domains=("api-direct.io" "console.api-direct.io" "api.api-direct.io")

for domain in "${domains[@]}"; do
    if openssl s_client -connect "$domain:443" -servername "$domain" < /dev/null 2>/dev/null | openssl x509 -noout -dates > /dev/null 2>&1; then
        expiry=$(openssl s_client -connect "$domain:443" -servername "$domain" < /dev/null 2>/dev/null | openssl x509 -noout -enddate 2>/dev/null | cut -d= -f2)
        print_status "SSL certificate for $domain is valid (expires: $expiry)"
    else
        print_error "SSL certificate for $domain is invalid or not found"
    fi
done

echo ""

# Disk space check
print_info "Checking disk space..."
disk_usage=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$disk_usage" -lt 80 ]; then
    print_status "Disk usage is healthy ($disk_usage%)"
elif [ "$disk_usage" -lt 90 ]; then
    print_warning "Disk usage is getting high ($disk_usage%)"
else
    print_error "Disk usage is critical ($disk_usage%)"
fi

# Memory check
print_info "Checking memory usage..."
memory_usage=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
if [ "$memory_usage" -lt 80 ]; then
    print_status "Memory usage is healthy ($memory_usage%)"
elif [ "$memory_usage" -lt 90 ]; then
    print_warning "Memory usage is getting high ($memory_usage%)"
else
    print_error "Memory usage is critical ($memory_usage%)"
fi

echo ""

# Summary
print_info "📊 Health Check Summary:"
echo "  Healthy containers: $healthy_containers/$total_containers"
echo "  Disk usage: $disk_usage%"
echo "  Memory usage: $memory_usage%"

if [ "$healthy_containers" -eq "$total_containers" ]; then
    print_status "🎉 All systems are healthy!"
    exit 0
else
    print_error "⚠️ Some systems need attention"
    exit 1
fi