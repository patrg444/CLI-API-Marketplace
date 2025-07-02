#!/bin/bash

# Integration test runner for API-Direct marketplace
# Runs cross-service integration tests with proper setup and teardown

set -e

echo "🚀 Starting Integration Tests"
echo "============================"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test configuration
export TEST_ENV=integration
export LOG_LEVEL=INFO

# Check if services are running
check_service() {
    local service_name=$1
    local health_url=$2
    
    echo -n "Checking $service_name... "
    
    if curl -s -f "$health_url" > /dev/null; then
        echo -e "${GREEN}✓${NC}"
        return 0
    else
        echo -e "${RED}✗${NC}"
        return 1
    fi
}

# Start test environment
start_test_env() {
    echo -e "\n${YELLOW}Starting test environment...${NC}"
    
    # Start services with docker-compose
    docker-compose -f docker-compose.test.yml up -d
    
    # Wait for services to be ready
    echo "Waiting for services to be ready..."
    sleep 10
    
    # Check all services
    local all_ready=true
    
    check_service "API Gateway" "http://localhost:8000/health" || all_ready=false
    check_service "Auth Service" "http://localhost:8001/health" || all_ready=false
    check_service "Billing Service" "http://localhost:8002/health" || all_ready=false
    check_service "Metering Service" "http://localhost:8003/health" || all_ready=false
    check_service "PostgreSQL" "http://localhost:5432" || all_ready=false
    check_service "Redis" "http://localhost:6379" || all_ready=false
    
    if [ "$all_ready" = false ]; then
        echo -e "${RED}Some services are not ready. Check logs with: docker-compose -f docker-compose.test.yml logs${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}All services are ready!${NC}"
}

# Run database migrations
run_migrations() {
    echo -e "\n${YELLOW}Running database migrations...${NC}"
    
    # Run migrations for each service
    docker-compose -f docker-compose.test.yml exec -T backend alembic upgrade head
    docker-compose -f docker-compose.test.yml exec -T auth-service npm run migrate
    docker-compose -f docker-compose.test.yml exec -T billing-service npm run migrate
    
    echo -e "${GREEN}Migrations completed!${NC}"
}

# Seed test data
seed_test_data() {
    echo -e "\n${YELLOW}Seeding test data...${NC}"
    
    # Run seed scripts
    docker-compose -f docker-compose.test.yml exec -T backend python scripts/seed_test_data.py
    
    echo -e "${GREEN}Test data seeded!${NC}"
}

# Run integration tests
run_tests() {
    echo -e "\n${YELLOW}Running integration tests...${NC}"
    
    # Set test environment variables
    export TEST_API_URL=http://localhost:8000
    export TEST_DB_URL=postgresql://test:test@localhost:5432/testdb
    export TEST_REDIS_URL=redis://localhost:6379/1
    
    # Run pytest with coverage
    pytest tests/integration/ \
        -v \
        --asyncio-mode=auto \
        --cov=services \
        --cov-report=html \
        --cov-report=term-missing \
        --tb=short \
        -m "integration" \
        "$@"
    
    local test_result=$?
    
    if [ $test_result -eq 0 ]; then
        echo -e "\n${GREEN}✅ All integration tests passed!${NC}"
    else
        echo -e "\n${RED}❌ Some integration tests failed!${NC}"
    fi
    
    return $test_result
}

# Generate test report
generate_report() {
    echo -e "\n${YELLOW}Generating test report...${NC}"
    
    # Create reports directory
    mkdir -p reports/integration
    
    # Copy coverage report
    cp -r htmlcov reports/integration/
    
    # Generate summary report
    cat > reports/integration/summary.txt << EOF
Integration Test Summary
========================
Date: $(date)
Environment: $TEST_ENV

Test Results:
$(pytest tests/integration/ --tb=no -q 2>&1 | tail -n 10)

Coverage Report:
$(coverage report --include="services/*" 2>&1)

Service Health:
$(curl -s http://localhost:8000/health | jq '.')
EOF
    
    echo -e "${GREEN}Report generated at: reports/integration/${NC}"
}

# Cleanup test environment
cleanup() {
    echo -e "\n${YELLOW}Cleaning up test environment...${NC}"
    
    # Stop and remove containers
    docker-compose -f docker-compose.test.yml down -v
    
    # Clean up test data
    rm -rf .pytest_cache
    rm -rf __pycache__
    
    echo -e "${GREEN}Cleanup completed!${NC}"
}

# Main execution
main() {
    # Parse command line arguments
    local skip_setup=false
    local skip_cleanup=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-setup)
                skip_setup=true
                shift
                ;;
            --skip-cleanup)
                skip_cleanup=true
                shift
                ;;
            *)
                break
                ;;
        esac
    done
    
    # Trap to ensure cleanup runs
    if [ "$skip_cleanup" = false ]; then
        trap cleanup EXIT
    fi
    
    # Run test workflow
    if [ "$skip_setup" = false ]; then
        start_test_env
        run_migrations
        seed_test_data
    fi
    
    run_tests "$@"
    local test_result=$?
    
    generate_report
    
    exit $test_result
}

# Run main function
main "$@"