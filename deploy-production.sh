#!/bin/bash
# API Direct Marketplace - Production Deployment Script
# This script automates the deployment process for production

set -e  # Exit on error
set -u  # Exit on undefined variable

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$SCRIPT_DIR"
ENV_FILE="$PROJECT_ROOT/.env.production"
DOCKER_COMPOSE_FILE="$PROJECT_ROOT/docker-compose.production.yml"
BACKUP_DIR="$PROJECT_ROOT/backups/pre-deployment"
LOG_FILE="$PROJECT_ROOT/logs/deployment-$(date +%Y%m%d-%H%M%S).log"

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check if running as root
    if [[ $EUID -eq 0 ]]; then
       print_error "This script should not be run as root!"
       exit 1
    fi
    
    # Check required commands
    local required_commands=("docker" "docker-compose" "git" "curl" "openssl")
    for cmd in "${required_commands[@]}"; do
        if ! command -v "$cmd" &> /dev/null; then
            print_error "$cmd is not installed. Please install it first."
            exit 1
        fi
    done
    
    # Check Docker daemon
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running. Please start Docker first."
        exit 1
    fi
    
    # Check environment file
    if [[ ! -f "$ENV_FILE" ]]; then
        print_error "Environment file not found: $ENV_FILE"
        print_info "Creating from template..."
        if [[ -f "$PROJECT_ROOT/.env.production.example" ]]; then
            cp "$PROJECT_ROOT/.env.production.example" "$ENV_FILE"
            print_warning "Please edit $ENV_FILE with your production values before continuing."
            exit 1
        else
            print_error "Template file not found: $PROJECT_ROOT/.env.production.example"
            exit 1
        fi
    fi
    
    # Verify critical environment variables
    source "$ENV_FILE"
    local critical_vars=(
        "POSTGRES_PASSWORD"
        "JWT_SECRET"
        "STRIPE_SECRET_KEY"
        "DOMAIN"
    )
    
    for var in "${critical_vars[@]}"; do
        if [[ -z "${!var:-}" ]] || [[ "${!var}" == "CHANGE_ME"* ]]; then
            print_error "Critical environment variable not set: $var"
            print_warning "Please update $ENV_FILE with production values."
            exit 1
        fi
    done
    
    print_success "Prerequisites check passed"
}

# Function to create necessary directories
create_directories() {
    print_info "Creating necessary directories..."
    
    local dirs=(
        "$PROJECT_ROOT/logs"
        "$PROJECT_ROOT/logs/nginx"
        "$BACKUP_DIR"
        "$PROJECT_ROOT/nginx/ssl"
    )
    
    for dir in "${dirs[@]}"; do
        mkdir -p "$dir"
        print_info "Created directory: $dir"
    done
    
    print_success "Directories created"
}

# Function to backup current deployment
backup_current_deployment() {
    print_info "Backing up current deployment..."
    
    local backup_timestamp=$(date +%Y%m%d-%H%M%S)
    local backup_path="$BACKUP_DIR/backup-$backup_timestamp"
    
    mkdir -p "$backup_path"
    
    # Backup database if running
    if docker ps | grep -q apidirect-postgres; then
        print_info "Backing up PostgreSQL database..."
        docker exec apidirect-postgres pg_dump -U apidirect apidirect | gzip > "$backup_path/database.sql.gz"
        print_success "Database backed up to: $backup_path/database.sql.gz"
    fi
    
    # Backup environment file
    if [[ -f "$ENV_FILE" ]]; then
        cp "$ENV_FILE" "$backup_path/.env.production.backup"
        print_info "Environment file backed up"
    fi
    
    # Backup docker volumes list
    docker volume ls --format "{{.Name}}" > "$backup_path/docker-volumes.txt"
    
    print_success "Backup completed: $backup_path"
}

# Function to pull latest code
update_code() {
    print_info "Updating code from repository..."
    
    cd "$PROJECT_ROOT"
    
    # Check for uncommitted changes
    if [[ -n $(git status --porcelain) ]]; then
        print_warning "Uncommitted changes detected. Stashing..."
        git stash push -m "Deployment stash $(date +%Y%m%d-%H%M%S)"
    fi
    
    # Pull latest changes
    git pull origin main || {
        print_error "Failed to pull latest code"
        exit 1
    }
    
    print_success "Code updated successfully"
}

# Function to build Docker images
build_images() {
    print_info "Building Docker images..."
    
    cd "$PROJECT_ROOT"
    
    # Build all services
    docker-compose -f "$DOCKER_COMPOSE_FILE" build --no-cache || {
        print_error "Failed to build Docker images"
        exit 1
    }
    
    print_success "Docker images built successfully"
}

# Function to run database migrations
run_migrations() {
    print_info "Running database migrations..."
    
    # Start only database service
    docker-compose -f "$DOCKER_COMPOSE_FILE" up -d postgres
    
    # Wait for database to be ready
    print_info "Waiting for database to be ready..."
    sleep 10
    
    local max_attempts=30
    local attempt=0
    
    while ! docker exec apidirect-postgres pg_isready -U apidirect &> /dev/null; do
        attempt=$((attempt + 1))
        if [[ $attempt -gt $max_attempts ]]; then
            print_error "Database failed to start within timeout"
            exit 1
        fi
        print_info "Waiting for database... (attempt $attempt/$max_attempts)"
        sleep 2
    done
    
    # Run migrations
    docker exec apidirect-postgres psql -U apidirect -d apidirect -f /docker-entrypoint-initdb.d/schema.sql || {
        print_warning "Schema already exists or migration failed - continuing..."
    }
    
    print_success "Database migrations completed"
}

# Function to start services
start_services() {
    print_info "Starting all services..."
    
    cd "$PROJECT_ROOT"
    
    # Start all services
    docker-compose -f "$DOCKER_COMPOSE_FILE" up -d || {
        print_error "Failed to start services"
        exit 1
    }
    
    print_success "All services started"
}

# Function to setup SSL certificates
setup_ssl() {
    print_info "Setting up SSL certificates..."
    
    source "$ENV_FILE"
    
    if [[ -z "${DOMAIN:-}" ]]; then
        print_warning "DOMAIN not set in environment. Skipping SSL setup."
        return
    fi
    
    # Check if certificates already exist
    if [[ -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ]]; then
        print_info "SSL certificates already exist for $DOMAIN"
        return
    fi
    
    # Install certbot if not present
    if ! command -v certbot &> /dev/null; then
        print_info "Installing certbot..."
        sudo apt-get update
        sudo apt-get install -y certbot python3-certbot-nginx
    fi
    
    # Generate certificates
    print_info "Generating SSL certificates for $DOMAIN..."
    sudo certbot certonly --standalone \
        --non-interactive \
        --agree-tos \
        --email "${LETSENCRYPT_EMAIL:-admin@$DOMAIN}" \
        -d "$DOMAIN" \
        -d "www.$DOMAIN" \
        -d "api.$DOMAIN" \
        -d "console.$DOMAIN" || {
        print_warning "Failed to generate SSL certificates. You may need to do this manually."
    }
    
    print_success "SSL setup completed"
}

# Function to verify deployment
verify_deployment() {
    print_info "Verifying deployment..."
    
    # Wait for services to stabilize
    sleep 20
    
    # Check all containers are running
    local expected_services=(
        "apidirect-postgres"
        "apidirect-redis"
        "apidirect-influxdb"
        "apidirect-backend"
        "apidirect-marketplace"
        "apidirect-nginx"
        "apidirect-prometheus"
        "apidirect-grafana"
    )
    
    local all_healthy=true
    
    for service in "${expected_services[@]}"; do
        if docker ps | grep -q "$service"; then
            print_success "$service is running"
        else
            print_error "$service is not running"
            all_healthy=false
        fi
    done
    
    # Check health endpoints
    print_info "Checking health endpoints..."
    
    # Backend health check
    if curl -f -s http://localhost:8000/health > /dev/null; then
        print_success "Backend API is healthy"
    else
        print_error "Backend API health check failed"
        all_healthy=false
    fi
    
    # Marketplace health check
    if curl -f -s http://localhost:3001/api/health > /dev/null; then
        print_success "Marketplace is healthy"
    else
        print_error "Marketplace health check failed"
        all_healthy=false
    fi
    
    if $all_healthy; then
        print_success "Deployment verification passed!"
    else
        print_error "Deployment verification failed. Check logs for details."
        print_info "Logs: docker-compose -f $DOCKER_COMPOSE_FILE logs"
        exit 1
    fi
}

# Function to show post-deployment steps
show_post_deployment_steps() {
    print_info "Post-deployment steps:"
    echo
    echo "1. Configure your DNS to point to this server:"
    echo "   - A record for $DOMAIN -> $(curl -s ifconfig.me)"
    echo "   - A record for www.$DOMAIN -> $(curl -s ifconfig.me)"
    echo "   - A record for api.$DOMAIN -> $(curl -s ifconfig.me)"
    echo "   - A record for console.$DOMAIN -> $(curl -s ifconfig.me)"
    echo
    echo "2. Access your services:"
    echo "   - Marketplace: https://$DOMAIN"
    echo "   - API: https://api.$DOMAIN"
    echo "   - Console: https://console.$DOMAIN"
    echo "   - Grafana: http://$(hostname -I | awk '{print $1}'):3000"
    echo "   - Prometheus: http://$(hostname -I | awk '{print $1}'):9090"
    echo
    echo "3. Configure monitoring alerts in Grafana"
    echo "4. Set up automated backups (see backup-automation.sh)"
    echo "5. Configure log rotation"
    echo "6. Review security settings"
    echo
    print_success "Deployment completed successfully!"
}

# Function to rollback deployment
rollback_deployment() {
    print_error "Rolling back deployment..."
    
    # Stop all services
    docker-compose -f "$DOCKER_COMPOSE_FILE" down
    
    # Find latest backup
    local latest_backup=$(ls -t "$BACKUP_DIR" | head -1)
    
    if [[ -n "$latest_backup" ]]; then
        print_info "Restoring from backup: $latest_backup"
        
        # Restore database if backup exists
        if [[ -f "$BACKUP_DIR/$latest_backup/database.sql.gz" ]]; then
            docker-compose -f "$DOCKER_COMPOSE_FILE" up -d postgres
            sleep 10
            gunzip -c "$BACKUP_DIR/$latest_backup/database.sql.gz" | docker exec -i apidirect-postgres psql -U apidirect apidirect
            print_success "Database restored"
        fi
        
        # Restore environment file
        if [[ -f "$BACKUP_DIR/$latest_backup/.env.production.backup" ]]; then
            cp "$BACKUP_DIR/$latest_backup/.env.production.backup" "$ENV_FILE"
            print_success "Environment file restored"
        fi
    fi
    
    print_warning "Rollback completed. Please manually verify the system state."
}

# Main deployment flow
main() {
    print_info "Starting API Direct Marketplace deployment..."
    print_info "Deployment started at: $(date)"
    print_info "Logging to: $LOG_FILE"
    echo
    
    # Create log file
    mkdir -p "$(dirname "$LOG_FILE")"
    touch "$LOG_FILE"
    
    # Set trap for rollback on error
    trap 'rollback_deployment' ERR
    
    # Execute deployment steps
    check_prerequisites
    create_directories
    backup_current_deployment
    update_code
    build_images
    run_migrations
    start_services
    setup_ssl
    verify_deployment
    
    # Remove error trap after successful deployment
    trap - ERR
    
    show_post_deployment_steps
    
    print_info "Deployment completed at: $(date)"
}

# Handle command line arguments
case "${1:-}" in
    "rollback")
        rollback_deployment
        ;;
    "verify")
        verify_deployment
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo "Commands:"
        echo "  (none)    - Run full deployment"
        echo "  rollback  - Rollback to previous deployment"
        echo "  verify    - Verify current deployment"
        echo "  help      - Show this help message"
        ;;
    *)
        main
        ;;
esac