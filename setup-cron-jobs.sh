#!/bin/bash
# Production Cron Jobs Setup for API Direct

set -e

echo "🕐 Setting up production cron jobs for API Direct"
echo "================================================"
echo ""

# Get the absolute path to the project
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Create logs directory if it doesn't exist
mkdir -p "$PROJECT_ROOT/logs/cron"

# Function to add cron job if it doesn't exist
add_cron_job() {
    local schedule="$1"
    local command="$2"
    local description="$3"
    
    # Check if cron job already exists
    if crontab -l 2>/dev/null | grep -q "$command"; then
        echo "✅ Already exists: $description"
    else
        # Add the cron job
        (crontab -l 2>/dev/null; echo "# $description"; echo "$schedule $command") | crontab -
        echo "✅ Added: $description"
    fi
}

echo "Adding cron jobs..."
echo ""

# 1. Daily backup at 2 AM
add_cron_job \
    "0 2 * * *" \
    "$PROJECT_ROOT/scripts/backup-automation.sh >> $PROJECT_ROOT/logs/cron/backup.log 2>&1" \
    "Daily backup (2 AM)"

# 2. Database cleanup weekly (Sunday 3 AM)
add_cron_job \
    "0 3 * * 0" \
    "cd $PROJECT_ROOT && docker-compose -f docker-compose.production.yml exec -T postgres psql -U apidirect -c 'VACUUM ANALYZE;' >> $PROJECT_ROOT/logs/cron/db-maintenance.log 2>&1" \
    "Weekly database maintenance"

# 3. Log rotation monthly (1st day of month, 4 AM)
add_cron_job \
    "0 4 1 * *" \
    "find $PROJECT_ROOT/logs -name '*.log' -mtime +30 -delete >> $PROJECT_ROOT/logs/cron/log-rotation.log 2>&1" \
    "Monthly log rotation"

# 4. SSL certificate renewal check (daily at 1 AM)
add_cron_job \
    "0 1 * * *" \
    "$PROJECT_ROOT/scripts/check-ssl-renewal.sh >> $PROJECT_ROOT/logs/cron/ssl-renewal.log 2>&1" \
    "Daily SSL certificate check"

# 5. Health check every 5 minutes
add_cron_job \
    "*/5 * * * *" \
    "curl -s -o /dev/null -w '%{http_code}' https://api.apidirect.dev/health || echo 'Health check failed at $(date)' >> $PROJECT_ROOT/logs/cron/health-check.log" \
    "Health check every 5 minutes"

echo ""
echo "Creating SSL renewal check script..."

# Create SSL renewal check script
cat > "$PROJECT_ROOT/scripts/check-ssl-renewal.sh" << 'EOF'
#!/bin/bash
# Check SSL certificates and renew if needed

DOMAIN="apidirect.dev"
DAYS_BEFORE_EXPIRY=30

# Check certificate expiration
EXPIRY=$(echo | openssl s_client -servername $DOMAIN -connect $DOMAIN:443 2>/dev/null | openssl x509 -noout -dates | grep notAfter | cut -d= -f2)
EXPIRY_EPOCH=$(date -d "$EXPIRY" +%s)
CURRENT_EPOCH=$(date +%s)
DAYS_LEFT=$(( ($EXPIRY_EPOCH - $CURRENT_EPOCH) / 86400 ))

echo "[$(date)] Certificate expires in $DAYS_LEFT days"

if [ $DAYS_LEFT -lt $DAYS_BEFORE_EXPIRY ]; then
    echo "[$(date)] Certificate renewal needed!"
    # Add your Let's Encrypt renewal command here
    # certbot renew --quiet
else
    echo "[$(date)] Certificate is still valid"
fi
EOF

chmod +x "$PROJECT_ROOT/scripts/check-ssl-renewal.sh"

echo ""
echo "Current crontab:"
echo "================"
crontab -l 2>/dev/null || echo "No crontab configured yet"
echo ""
echo "✅ Cron jobs setup complete!"
echo ""
echo "Note: To view logs, check the $PROJECT_ROOT/logs/cron/ directory"
echo "To edit cron jobs manually, run: crontab -e"
echo "To remove all cron jobs, run: crontab -r"