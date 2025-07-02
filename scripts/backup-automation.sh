#!/bin/bash
# API Direct Marketplace - Automated Backup Script
# Run this script via cron for regular backups

set -e

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKUP_ROOT="/var/backups/apidirect"
S3_BUCKET="${BACKUP_S3_BUCKET:-apidirect-backups}"
RETENTION_DAYS=30
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BACKUP_DIR="$BACKUP_ROOT/$TIMESTAMP"
LOG_FILE="$PROJECT_ROOT/logs/backup-$TIMESTAMP.log"
ENV_FILE="$PROJECT_ROOT/.env.production"

# Load environment variables
if [[ -f "$ENV_FILE" ]]; then
    source "$ENV_FILE"
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

# Send notification (configure your preferred method)
send_notification() {
    local status=$1
    local message=$2
    
    # Email notification
    if command -v mail &> /dev/null && [[ -n "${ADMIN_EMAIL:-}" ]]; then
        echo "$message" | mail -s "API Direct Backup $status" "$ADMIN_EMAIL"
    fi
    
    # Slack webhook (if configured)
    if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"text\":\"Backup $status: $message\"}" \
            "$SLACK_WEBHOOK_URL" 2>/dev/null || true
    fi
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Create directories
    mkdir -p "$BACKUP_ROOT" "$BACKUP_DIR" "$(dirname "$LOG_FILE")"
    
    # Check required commands
    local required_commands=("docker" "tar" "gzip" "aws")
    for cmd in "${required_commands[@]}"; do
        if ! command -v "$cmd" &> /dev/null; then
            log_error "$cmd is not installed"
            if [[ "$cmd" == "aws" ]]; then
                log_warning "AWS CLI not found. S3 upload will be skipped."
            else
                exit 1
            fi
        fi
    done
    
    # Check AWS credentials if AWS CLI is available
    if command -v aws &> /dev/null; then
        if ! aws s3 ls &> /dev/null; then
            log_warning "AWS credentials not configured. S3 upload will be skipped."
        fi
    fi
    
    log_success "Prerequisites check passed"
}

# Backup PostgreSQL database
backup_postgres() {
    log_info "Backing up PostgreSQL database..."
    
    local db_backup_file="$BACKUP_DIR/postgres-$TIMESTAMP.sql.gz"
    
    if docker ps | grep -q apidirect-postgres; then
        # Dump all databases
        docker exec apidirect-postgres pg_dumpall -U apidirect | gzip > "$db_backup_file"
        
        # Also create individual database dumps
        docker exec apidirect-postgres psql -U apidirect -t -c "SELECT datname FROM pg_database WHERE datistemplate = false;" | while read db; do
            if [[ -n "$db" && "$db" != "postgres" ]]; then
                log_info "Backing up database: $db"
                docker exec apidirect-postgres pg_dump -U apidirect "$db" | gzip > "$BACKUP_DIR/postgres-$db-$TIMESTAMP.sql.gz"
            fi
        done
        
        log_success "PostgreSQL backup completed: $db_backup_file"
    else
        log_warning "PostgreSQL container not running. Skipping database backup."
    fi
}

# Backup Redis data
backup_redis() {
    log_info "Backing up Redis data..."
    
    if docker ps | grep -q apidirect-redis; then
        # Force Redis to save
        docker exec apidirect-redis redis-cli BGSAVE
        sleep 5
        
        # Copy Redis dump file
        docker cp apidirect-redis:/data/dump.rdb "$BACKUP_DIR/redis-$TIMESTAMP.rdb"
        
        log_success "Redis backup completed"
    else
        log_warning "Redis container not running. Skipping Redis backup."
    fi
}

# Backup InfluxDB data
backup_influxdb() {
    log_info "Backing up InfluxDB data..."
    
    if docker ps | grep -q apidirect-influxdb; then
        # Create InfluxDB backup
        docker exec apidirect-influxdb influx backup /tmp/influx-backup
        docker cp apidirect-influxdb:/tmp/influx-backup "$BACKUP_DIR/influxdb-$TIMESTAMP"
        docker exec apidirect-influxdb rm -rf /tmp/influx-backup
        
        # Compress the backup
        tar -czf "$BACKUP_DIR/influxdb-$TIMESTAMP.tar.gz" -C "$BACKUP_DIR" "influxdb-$TIMESTAMP"
        rm -rf "$BACKUP_DIR/influxdb-$TIMESTAMP"
        
        log_success "InfluxDB backup completed"
    else
        log_warning "InfluxDB container not running. Skipping InfluxDB backup."
    fi
}

# Backup configuration files
backup_configs() {
    log_info "Backing up configuration files..."
    
    local config_backup="$BACKUP_DIR/configs-$TIMESTAMP.tar.gz"
    
    # List of files/directories to backup
    local config_files=(
        ".env.production"
        "docker-compose.production.yml"
        "nginx/nginx.conf"
        "nginx/ssl"
        "monitoring/prometheus.yml"
        "monitoring/grafana"
    )
    
    # Create tar archive
    cd "$PROJECT_ROOT"
    tar -czf "$config_backup" "${config_files[@]}" 2>/dev/null || true
    
    log_success "Configuration backup completed: $config_backup"
}

# Backup Docker volumes
backup_docker_volumes() {
    log_info "Backing up Docker volumes..."
    
    # Get list of volumes
    local volumes=$(docker volume ls --format "{{.Name}}" | grep -E "apidirect|postgres|redis|influxdb|grafana|prometheus" || true)
    
    if [[ -n "$volumes" ]]; then
        for volume in $volumes; do
            log_info "Backing up volume: $volume"
            
            # Create temporary container to access volume
            docker run --rm -v "$volume:/data" -v "$BACKUP_DIR:/backup" alpine \
                tar -czf "/backup/volume-$volume-$TIMESTAMP.tar.gz" -C /data .
        done
        log_success "Docker volumes backup completed"
    else
        log_warning "No Docker volumes found to backup"
    fi
}

# Backup application logs
backup_logs() {
    log_info "Backing up application logs..."
    
    local logs_backup="$BACKUP_DIR/logs-$TIMESTAMP.tar.gz"
    
    if [[ -d "$PROJECT_ROOT/logs" ]]; then
        # Exclude current backup log
        tar -czf "$logs_backup" -C "$PROJECT_ROOT" logs --exclude="backup-$TIMESTAMP.log" 2>/dev/null || true
        log_success "Logs backup completed: $logs_backup"
    else
        log_warning "No logs directory found"
    fi
}

# Create backup metadata
create_metadata() {
    log_info "Creating backup metadata..."
    
    cat > "$BACKUP_DIR/metadata.json" <<EOF
{
    "timestamp": "$TIMESTAMP",
    "date": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "hostname": "$(hostname)",
    "backup_version": "1.0",
    "components": {
        "postgres": $(docker ps | grep -q apidirect-postgres && echo "true" || echo "false"),
        "redis": $(docker ps | grep -q apidirect-redis && echo "true" || echo "false"),
        "influxdb": $(docker ps | grep -q apidirect-influxdb && echo "true" || echo "false")
    },
    "docker_images": $(docker images --format '{"repository":"{{.Repository}}","tag":"{{.Tag}}","id":"{{.ID}}"}' | grep apidirect | jq -s .),
    "disk_usage": {
        "backup_size": "$(du -sh "$BACKUP_DIR" 2>/dev/null | cut -f1)",
        "free_space": "$(df -h "$BACKUP_ROOT" | tail -1 | awk '{print $4}')"
    }
}
EOF
    
    log_success "Metadata created"
}

# Compress full backup
compress_backup() {
    log_info "Compressing full backup..."
    
    cd "$BACKUP_ROOT"
    tar -czf "$TIMESTAMP.tar.gz" "$TIMESTAMP"
    
    # Calculate checksum
    sha256sum "$TIMESTAMP.tar.gz" > "$TIMESTAMP.tar.gz.sha256"
    
    log_success "Backup compressed: $BACKUP_ROOT/$TIMESTAMP.tar.gz"
}

# Upload to S3
upload_to_s3() {
    log_info "Uploading backup to S3..."
    
    if ! command -v aws &> /dev/null; then
        log_warning "AWS CLI not available. Skipping S3 upload."
        return
    fi
    
    if ! aws s3 ls "s3://$S3_BUCKET" &> /dev/null; then
        log_warning "Cannot access S3 bucket: $S3_BUCKET. Skipping upload."
        return
    fi
    
    # Upload compressed backup
    aws s3 cp "$BACKUP_ROOT/$TIMESTAMP.tar.gz" "s3://$S3_BUCKET/backups/$TIMESTAMP.tar.gz" \
        --storage-class STANDARD_IA
    
    # Upload checksum
    aws s3 cp "$BACKUP_ROOT/$TIMESTAMP.tar.gz.sha256" "s3://$S3_BUCKET/backups/$TIMESTAMP.tar.gz.sha256"
    
    # Upload metadata separately for easy access
    aws s3 cp "$BACKUP_DIR/metadata.json" "s3://$S3_BUCKET/backups/$TIMESTAMP-metadata.json"
    
    log_success "Backup uploaded to S3: s3://$S3_BUCKET/backups/$TIMESTAMP.tar.gz"
}

# Clean up old backups
cleanup_old_backups() {
    log_info "Cleaning up old backups..."
    
    # Local cleanup
    find "$BACKUP_ROOT" -name "*.tar.gz" -type f -mtime +$RETENTION_DAYS -delete
    find "$BACKUP_ROOT" -name "*.sha256" -type f -mtime +$RETENTION_DAYS -delete
    find "$BACKUP_ROOT" -type d -empty -delete
    
    # S3 cleanup (if available)
    if command -v aws &> /dev/null && aws s3 ls "s3://$S3_BUCKET" &> /dev/null; then
        # List and delete old backups
        aws s3 ls "s3://$S3_BUCKET/backups/" | while read -r line; do
            backup_date=$(echo "$line" | awk '{print $1}')
            backup_file=$(echo "$line" | awk '{print $4}')
            
            if [[ -n "$backup_date" ]] && [[ -n "$backup_file" ]]; then
                if [[ $(date -d "$backup_date" +%s 2>/dev/null) -lt $(date -d "$RETENTION_DAYS days ago" +%s) ]]; then
                    log_info "Deleting old S3 backup: $backup_file"
                    aws s3 rm "s3://$S3_BUCKET/backups/$backup_file"
                fi
            fi
        done
    fi
    
    log_success "Cleanup completed"
}

# Verify backup integrity
verify_backup() {
    log_info "Verifying backup integrity..."
    
    # Check if backup file exists and has size
    if [[ ! -f "$BACKUP_ROOT/$TIMESTAMP.tar.gz" ]]; then
        log_error "Backup file not found!"
        return 1
    fi
    
    local size=$(stat -f%z "$BACKUP_ROOT/$TIMESTAMP.tar.gz" 2>/dev/null || stat -c%s "$BACKUP_ROOT/$TIMESTAMP.tar.gz")
    if [[ $size -lt 1000 ]]; then
        log_error "Backup file too small: $size bytes"
        return 1
    fi
    
    # Verify checksum
    cd "$BACKUP_ROOT"
    if sha256sum -c "$TIMESTAMP.tar.gz.sha256" &> /dev/null; then
        log_success "Backup integrity verified"
        return 0
    else
        log_error "Backup integrity check failed!"
        return 1
    fi
}

# Main backup process
main() {
    log "Starting backup process..."
    
    # Set error trap
    trap 'handle_error $? $LINENO' ERR
    
    # Execute backup steps
    check_prerequisites
    backup_postgres
    backup_redis
    backup_influxdb
    backup_configs
    backup_docker_volumes
    backup_logs
    create_metadata
    compress_backup
    
    # Verify and upload
    if verify_backup; then
        upload_to_s3
        cleanup_old_backups
        
        # Remove uncompressed backup directory
        rm -rf "$BACKUP_DIR"
        
        log_success "Backup completed successfully!"
        send_notification "SUCCESS" "Backup completed: $BACKUP_ROOT/$TIMESTAMP.tar.gz"
    else
        log_error "Backup verification failed!"
        send_notification "FAILED" "Backup verification failed for $TIMESTAMP"
        exit 1
    fi
    
    log "Backup process finished at $(date)"
}

# Error handler
handle_error() {
    local exit_code=$1
    local line_number=$2
    log_error "Error occurred at line $line_number with exit code $exit_code"
    send_notification "FAILED" "Backup failed at line $line_number with exit code $exit_code"
    exit $exit_code
}

# Restore function (can be called with 'restore' argument)
restore_backup() {
    local backup_file="$1"
    
    if [[ -z "$backup_file" ]]; then
        echo "Usage: $0 restore <backup-file>"
        echo "Example: $0 restore /var/backups/apidirect/20240615-120000.tar.gz"
        exit 1
    fi
    
    if [[ ! -f "$backup_file" ]]; then
        log_error "Backup file not found: $backup_file"
        exit 1
    fi
    
    log_warning "This will restore the backup and overwrite current data. Are you sure? (yes/no)"
    read -r confirmation
    
    if [[ "$confirmation" != "yes" ]]; then
        log_info "Restore cancelled"
        exit 0
    fi
    
    log_info "Starting restore from: $backup_file"
    
    # Create restore directory
    local restore_dir="/tmp/apidirect-restore-$$"
    mkdir -p "$restore_dir"
    
    # Extract backup
    tar -xzf "$backup_file" -C "$restore_dir"
    
    # Find the backup directory (it should be named with timestamp)
    local backup_content_dir=$(ls -d "$restore_dir"/* | head -1)
    
    # Stop services
    log_info "Stopping services..."
    cd "$PROJECT_ROOT"
    docker-compose -f docker-compose.production.yml down
    
    # Restore PostgreSQL
    if [[ -f "$backup_content_dir"/postgres-*.sql.gz ]]; then
        log_info "Restoring PostgreSQL..."
        docker-compose -f docker-compose.production.yml up -d postgres
        sleep 10
        
        gunzip -c "$backup_content_dir"/postgres-*.sql.gz | docker exec -i apidirect-postgres psql -U apidirect
        log_success "PostgreSQL restored"
    fi
    
    # Restore Redis
    if [[ -f "$backup_content_dir"/redis-*.rdb ]]; then
        log_info "Restoring Redis..."
        docker-compose -f docker-compose.production.yml up -d redis
        docker exec apidirect-redis redis-cli SHUTDOWN
        docker cp "$backup_content_dir"/redis-*.rdb apidirect-redis:/data/dump.rdb
        docker-compose -f docker-compose.production.yml restart redis
        log_success "Redis restored"
    fi
    
    # Restore configs
    if [[ -f "$backup_content_dir"/configs-*.tar.gz ]]; then
        log_info "Restoring configurations..."
        cd "$PROJECT_ROOT"
        tar -xzf "$backup_content_dir"/configs-*.tar.gz
        log_success "Configurations restored"
    fi
    
    # Start all services
    log_info "Starting all services..."
    docker-compose -f docker-compose.production.yml up -d
    
    # Cleanup
    rm -rf "$restore_dir"
    
    log_success "Restore completed!"
}

# Script entry point
case "${1:-}" in
    "restore")
        restore_backup "$2"
        ;;
    "verify")
        check_prerequisites
        log_success "Backup system is ready"
        ;;
    *)
        main
        ;;
esac