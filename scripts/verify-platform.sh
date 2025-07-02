#!/bin/bash

# API-Direct Platform Verification Script
# Ensures all components run correctly

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Kill any existing processes
cleanup() {
    print_info "Cleaning up existing processes..."
    pkill -f "python.*backend/api/main.py" || true
    pkill -f "python.*web/console/app.py" || true
    pkill -f "python.*http.server" || true
    docker-compose down 2>/dev/null || true
    sleep 2
}

# Start backend API
start_backend() {
    print_info "Starting FastAPI backend..."
    cd backend
    
    # Create virtual environment if it doesn't exist
    if [ ! -d "venv" ]; then
        print_info "Creating Python virtual environment..."
        python3 -m venv venv
    fi
    
    # Activate virtual environment
    source venv/bin/activate || . venv/bin/activate
    
    # Install requirements
    print_info "Installing backend requirements..."
    pip install -r requirements.txt > /dev/null 2>&1
    
    # Start backend in background
    PYTHONPATH=. uvicorn api.main:app --host 0.0.0.0 --port 8000 --reload > ../logs/backend.log 2>&1 &
    BACKEND_PID=$!
    
    cd ..
    sleep 5
    
    # Check if backend started
    if curl -s http://localhost:8000/health > /dev/null; then
        print_status "Backend API running on http://localhost:8000"
        return 0
    else
        print_error "Backend failed to start"
        return 1
    fi
}

# Start frontend
start_frontend() {
    print_info "Starting frontend server..."
    
    # Use Python's built-in server for simplicity
    cd web/console
    python3 -m http.server 8080 > ../../logs/frontend.log 2>&1 &
    FRONTEND_PID=$!
    cd ../..
    
    sleep 3
    
    # Check if frontend is accessible
    if curl -s http://localhost:8080 > /dev/null; then
        print_status "Frontend running on http://localhost:8080"
        return 0
    else
        print_error "Frontend failed to start"
        return 1
    fi
}

# Start landing page
start_landing() {
    print_info "Starting landing page server..."
    
    cd web/landing
    python3 -m http.server 3003 > ../../logs/landing.log 2>&1 &
    LANDING_PID=$!
    cd ../..
    
    sleep 2
    
    if curl -s http://localhost:3003 > /dev/null; then
        print_status "Landing page running on http://localhost:3003"
        return 0
    else
        print_error "Landing page failed to start"
        return 1
    fi
}

# Test API endpoints
test_api_endpoints() {
    print_info "Testing API endpoints..."
    
    # Test health endpoint
    if curl -s http://localhost:8000/health | grep -q "healthy"; then
        print_status "Health endpoint working"
    else
        print_error "Health endpoint not responding"
    fi
    
    # Test docs endpoint
    if curl -s http://localhost:8000/docs > /dev/null; then
        print_status "API documentation available at http://localhost:8000/docs"
    else
        print_error "API documentation not accessible"
    fi
}

# Test frontend pages
test_frontend_pages() {
    print_info "Testing frontend pages..."
    
    pages=("login.html" "register.html" "pages/dashboard.html" "pages/apis.html" "pages/marketplace.html")
    
    for page in "${pages[@]}"; do
        if curl -s http://localhost:8080/$page > /dev/null; then
            print_status "Page accessible: $page"
        else
            print_error "Page not accessible: $page"
        fi
    done
}

# Main execution
main() {
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║         API-Direct Platform Verification Script          ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    # Create logs directory
    mkdir -p logs
    
    # Cleanup any existing processes
    cleanup
    
    # Start services
    print_info "Starting API-Direct services..."
    echo ""
    
    if start_backend; then
        BACKEND_RUNNING=true
    else
        BACKEND_RUNNING=false
    fi
    
    if start_frontend; then
        FRONTEND_RUNNING=true
    else
        FRONTEND_RUNNING=false
    fi
    
    if start_landing; then
        LANDING_RUNNING=true
    else
        LANDING_RUNNING=false
    fi
    
    echo ""
    
    # Run tests if services are running
    if [ "$BACKEND_RUNNING" = true ]; then
        test_api_endpoints
    fi
    
    echo ""
    
    if [ "$FRONTEND_RUNNING" = true ]; then
        test_frontend_pages
    fi
    
    echo ""
    print_info "Service Status Summary:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    if [ "$LANDING_RUNNING" = true ]; then
        echo -e "🌐 Landing Page:    ${GREEN}http://localhost:3003${NC}"
    else
        echo -e "🌐 Landing Page:    ${RED}Not Running${NC}"
    fi
    
    if [ "$FRONTEND_RUNNING" = true ]; then
        echo -e "💻 Creator Portal:  ${GREEN}http://localhost:8080${NC}"
        echo -e "   ├─ Login:       ${GREEN}http://localhost:8080/login.html${NC}"
        echo -e "   ├─ Register:    ${GREEN}http://localhost:8080/register.html${NC}"
        echo -e "   └─ Dashboard:   ${GREEN}http://localhost:8080/pages/dashboard.html${NC}"
    else
        echo -e "💻 Creator Portal:  ${RED}Not Running${NC}"
    fi
    
    if [ "$BACKEND_RUNNING" = true ]; then
        echo -e "🔧 Backend API:     ${GREEN}http://localhost:8000${NC}"
        echo -e "   ├─ Health:      ${GREEN}http://localhost:8000/health${NC}"
        echo -e "   └─ API Docs:    ${GREEN}http://localhost:8000/docs${NC}"
    else
        echo -e "🔧 Backend API:     ${RED}Not Running${NC}"
    fi
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    if [ "$BACKEND_RUNNING" = true ] && [ "$FRONTEND_RUNNING" = true ] && [ "$LANDING_RUNNING" = true ]; then
        print_status "All services running successfully! 🎉"
        echo ""
        print_info "Next steps:"
        echo "  1. Visit http://localhost:3003 to see the landing page"
        echo "  2. Visit http://localhost:8080/register.html to create an account"
        echo "  3. Visit http://localhost:8000/docs to explore the API"
        echo ""
        print_warning "Press Ctrl+C to stop all services"
        
        # Keep script running
        trap "cleanup; exit" INT TERM
        while true; do
            sleep 1
        done
    else
        print_error "Some services failed to start. Check logs/ directory for details."
        cleanup
        exit 1
    fi
}

# Run main function
main