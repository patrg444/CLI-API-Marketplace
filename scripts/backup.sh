#!/bin/bash

# API-Direct Backup Script
set -e

# Configuration
BACKUP_DIR="/var/backups/apidirect"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
RETENTION_DAYS=30

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

print_status() {
    echo -e "${GREEN}[BACKUP]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Load environment variables
if [ -f .env.production ]; then
    source .env.production
else
    print_error ".env.production not found"
    exit 1
fi

# Create backup directory
mkdir -p "$BACKUP_DIR"

print_status "Starting backup process..."

# Backup PostgreSQL database
print_status "Backing up PostgreSQL database..."
docker-compose -f docker-compose.production.yml exec -T postgres pg_dump -U apidirect apidirect | gzip > "$BACKUP_DIR/postgres_$TIMESTAMP.sql.gz"

# Backup InfluxDB
print_status "Backing up InfluxDB..."
docker-compose -f docker-compose.production.yml exec -T influxdb influx backup /tmp/influxdb_backup_$TIMESTAMP
docker cp "$(docker-compose -f docker-compose.production.yml ps -q influxdb):/tmp/influxdb_backup_$TIMESTAMP" "$BACKUP_DIR/"
tar -czf "$BACKUP_DIR/influxdb_$TIMESTAMP.tar.gz" -C "$BACKUP_DIR" "influxdb_backup_$TIMESTAMP"
rm -rf "$BACKUP_DIR/influxdb_backup_$TIMESTAMP"

# Backup application data
print_status "Backing up application data..."
tar -czf "$BACKUP_DIR/app_data_$TIMESTAMP.tar.gz" \
    --exclude='logs' \
    --exclude='node_modules' \
    --exclude='.git' \
    --exclude='*.log' \
    .

# Upload to S3 if configured
if [ ! -z "$AWS_ACCESS_KEY_ID" ] && [ ! -z "$BACKUP_S3_BUCKET" ]; then
    print_status "Uploading backups to S3..."
    aws s3 cp "$BACKUP_DIR/postgres_$TIMESTAMP.sql.gz" "s3://$BACKUP_S3_BUCKET/postgres/"
    aws s3 cp "$BACKUP_DIR/influxdb_$TIMESTAMP.tar.gz" "s3://$BACKUP_S3_BUCKET/influxdb/"
    aws s3 cp "$BACKUP_DIR/app_data_$TIMESTAMP.tar.gz" "s3://$BACKUP_S3_BUCKET/app_data/"
fi

# Clean up old backups
print_status "Cleaning up old backups (older than $RETENTION_DAYS days)..."
find "$BACKUP_DIR" -name "*.gz" -mtime +$RETENTION_DAYS -delete

# Log backup completion
print_status "Backup completed successfully!"
print_status "Backup files:"
ls -lah "$BACKUP_DIR" | grep "$TIMESTAMP"

# Optional: Send notification (implement based on your notification system)
# curl -X POST "https://api.your-notification-service.com/notify" \
#      -H "Content-Type: application/json" \
#      -d '{"message": "API-Direct backup completed successfully", "timestamp": "'$TIMESTAMP'"}'