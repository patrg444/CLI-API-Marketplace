#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}CLI-API Marketplace Deployment Verifier${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to check if a service is running
check_service() {
    local service_name=$1
    local port=$2
    local endpoint=${3:-""}
    
    echo -n "Checking $service_name (port $port)... "
    
    # Check if container is running
    container_status=$(docker-compose ps | grep $service_name | grep Up)
    
    if [ -z "$container_status" ]; then
        echo -e "${RED}Container not running${NC}"
        return 1
    fi
    
    # Check if port is accessible
    if nc -z localhost $port 2>/dev/null; then
        echo -e "${GREEN}Running${NC}"
        
        # If endpoint provided, check health
        if [ ! -z "$endpoint" ]; then
            echo -n "  Health check... "
            response=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$port$endpoint" 2>/dev/null)
            if [ "$response" = "200" ]; then
                echo -e "${GREEN}Healthy${NC}"
            else
                echo -e "${YELLOW}Response: $response${NC}"
            fi
        fi
        return 0
    else
        echo -e "${YELLOW}Container up but port not accessible yet${NC}"
        return 1
    fi
}

# Function to check all services
check_all_services() {
    local all_up=0
    
    echo -e "${BLUE}Infrastructure Services:${NC}"
    echo "------------------------"
    
    check_service "postgres" 5432 || ((all_up++))
    check_service "redis" 6379 || ((all_up++))
    check_service "elasticsearch" 9200 "/_cluster/health" || ((all_up++))
    check_service "kibana" 5601 "/api/status" || ((all_up++))
    
    echo ""
    echo -e "${BLUE}Application Services:${NC}"
    echo "---------------------"
    
    check_service "gateway" 8082 "/health" || ((all_up++))
    check_service "apikey" 8081 "/health" || ((all_up++))
    check_service "marketplace" 8080 "/health" || ((all_up++))
    check_service "billing" 8083 "/health" || ((all_up++))
    check_service "metering" 8084 "/health" || ((all_up++))
    check_service "payout" 8085 "/health" || ((all_up++))
    check_service "storage" 8086 "/health" || ((all_up++))
    check_service "deployment" 8087 "/health" || ((all_up++))
    
    echo ""
    echo -e "${BLUE}Web Applications:${NC}"
    echo "-----------------"
    
    check_service "marketplace-ui" 3000 || ((all_up++))
    check_service "creator-portal" 3001 || ((all_up++))
    
    return $all_up
}

# Function to show container logs for debugging
show_problematic_containers() {
    echo ""
    echo -e "${YELLOW}Checking for containers with issues...${NC}"
    
    # Get containers that are not running properly
    problem_containers=$(docker-compose ps | grep -E "(Exit|Restarting)" | awk '{print $1}')
    
    if [ ! -z "$problem_containers" ]; then
        echo -e "${RED}Found problematic containers:${NC}"
        for container in $problem_containers; do
            echo ""
            echo -e "${YELLOW}Container: $container${NC}"
            echo "Last 10 log lines:"
            docker logs --tail 10 $container 2>&1 | sed 's/^/  /'
        done
    fi
}

# Function to wait for all services
wait_for_services() {
    local max_attempts=60  # 5 minutes (60 * 5 seconds)
    local attempt=0
    
    echo -e "${BLUE}Waiting for all services to be ready...${NC}"
    echo ""
    
    while [ $attempt -lt $max_attempts ]; do
        clear
        echo -e "${BLUE}========================================${NC}"
        echo -e "${BLUE}Deployment Status (Attempt $((attempt+1))/$max_attempts)${NC}"
        echo -e "${BLUE}========================================${NC}"
        echo ""
        
        check_all_services
        services_down=$?
        
        if [ $services_down -eq 0 ]; then
            echo ""
            echo -e "${GREEN}✓ All services are up and running!${NC}"
            echo ""
            echo -e "${BLUE}Access your services at:${NC}"
            echo "  - Marketplace UI: http://localhost:3000"
            echo "  - Creator Portal: http://localhost:3001"
            echo "  - API Gateway: http://localhost:8082"
            echo "  - Elasticsearch: http://localhost:9200"
            echo "  - Kibana: http://localhost:5601"
            return 0
        else
            echo ""
            echo -e "${YELLOW}$services_down services are not ready yet...${NC}"
            
            if [ $((attempt % 12)) -eq 0 ] && [ $attempt -gt 0 ]; then
                show_problematic_containers
            fi
            
            echo ""
            echo "Waiting 5 seconds before next check..."
            sleep 5
            ((attempt++))
        fi
    done
    
    echo ""
    echo -e "${RED}Timeout: Not all services came up within 5 minutes${NC}"
    show_problematic_containers
    return 1
}

# Main execution
case "${1:-}" in
    "wait")
        wait_for_services
        ;;
    "once")
        check_all_services
        services_down=$?
        if [ $services_down -gt 0 ]; then
            echo ""
            echo -e "${YELLOW}$services_down services are not ready${NC}"
            show_problematic_containers
            exit 1
        else
            echo ""
            echo -e "${GREEN}✓ All services are running!${NC}"
        fi
        ;;
    *)
        echo "Usage: $0 [wait|once]"
        echo ""
        echo "  wait  - Wait for all services to be ready (default)"
        echo "  once  - Check services once and exit"
        echo ""
        echo "Running default check..."
        echo ""
        check_all_services
        services_down=$?
        if [ $services_down -gt 0 ]; then
            echo ""
            echo -e "${YELLOW}$services_down services are not ready${NC}"
            echo ""
            echo "Run '$0 wait' to wait for all services to be ready"
        fi
        ;;
esac
