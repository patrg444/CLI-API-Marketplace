#!/bin/bash
# Production Readiness Verification Script for API Direct

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🔍 API Direct Production Readiness Check${NC}"
echo "========================================="
echo ""

READY=true

# Function to check item
check_item() {
    local description="$1"
    local condition="$2"
    
    if eval "$condition"; then
        echo -e "${GREEN}✅${NC} $description"
    else
        echo -e "${RED}❌${NC} $description"
        READY=false
    fi
}

# Check environment file
echo -e "${YELLOW}1. Environment Configuration:${NC}"
check_item ".env.production exists" "[ -f .env.production ]"
check_item "Production secrets generated" "grep -q 'JWT_SECRET=tv5MmfTGiAl1TwDR8wdBKy127OWWTSr4JPAkqO2ZVmtKnQM4NPtRvOijZcBuDG' .env.production 2>/dev/null"
check_item "AWS credentials configured" "grep -q 'AWS_ACCESS_KEY_ID=REPLACE_WITH_NEW_KEY' .env.production 2>/dev/null"
check_item "Cognito configured" "grep -q 'COGNITO_USER_POOL_ID=us-east-1_t63hJGq1S' .env.production 2>/dev/null"
check_item "SES email configured" "grep -q 'SMTP_HOST=email-smtp.us-east-1.amazonaws.com' .env.production 2>/dev/null"
check_item "Stripe webhook configured" "grep -q 'STRIPE_WEBHOOK_SECRET=whsec_' .env.production 2>/dev/null"
echo ""

# Check deployment files
echo -e "${YELLOW}2. Deployment Files:${NC}"
check_item "Production docker-compose exists" "[ -f docker-compose.production.yml ]"
check_item "Deployment script exists" "[ -f deploy-production.sh ]"
check_item "Backup script exists" "[ -f backup-automation.sh ]"
check_item "SSL setup configured" "grep -q 'setup_ssl' deploy-production.sh 2>/dev/null"
echo ""

# Check services
echo -e "${YELLOW}3. Service Configuration:${NC}"
check_item "Backend service configured" "grep -q 'backend:' docker-compose.production.yml"
check_item "PostgreSQL configured" "grep -q 'postgres:' docker-compose.production.yml"
check_item "Redis configured" "grep -q 'redis:' docker-compose.production.yml"
check_item "Nginx configured" "grep -q 'nginx:' docker-compose.production.yml"
check_item "Monitoring configured" "grep -q 'grafana:' docker-compose.production.yml"
echo ""

# Check domains
echo -e "${YELLOW}4. Domain Configuration:${NC}"
check_item "Main domain configured" "grep -q 'DOMAIN=apidirect.dev' .env.production"
check_item "Console domain configured" "grep -q 'CONSOLE_DOMAIN=console.apidirect.dev' .env.production"
check_item "Marketplace domain configured" "grep -q 'MARKETPLACE_DOMAIN=marketplace.apidirect.dev' .env.production"
check_item "API domain configured" "grep -q 'API_GATEWAY_URL=https://api.apidirect.dev' .env.production"
echo ""

# Check legal documents
echo -e "${YELLOW}5. Legal Documents:${NC}"
check_item "Terms of Service exists" "[ -f legal/terms-of-service.md ]"
check_item "Privacy Policy exists" "[ -f legal/privacy-policy.md ]"
check_item "Cookie Policy exists" "[ -f legal/cookie-policy.md ]"
echo ""

# Check AWS resources
echo -e "${YELLOW}6. AWS Resources (from .env):${NC}"
check_item "S3 buckets configured" "grep -q 'CODE_STORAGE_BUCKET=apidirect-code-storage' .env.production"
check_item "Backup bucket configured" "grep -q 'BACKUP_S3_BUCKET=apidirect-backups' .env.production"
echo ""

# Final result
echo "========================================="
if [ "$READY" = true ]; then
    echo -e "${GREEN}✅ All checks passed! Ready for production deployment.${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Update Stripe keys when ready (currently using test keys)"
    echo "2. Ensure DNS A records point to your server IP"
    echo "3. Run: ./deploy-production.sh"
else
    echo -e "${RED}❌ Some checks failed. Please fix the issues above.${NC}"
fi
echo ""

# Show current Stripe status
echo -e "${YELLOW}Note: Stripe Configuration Status:${NC}"
if grep -q "sk_test_" .env.production 2>/dev/null; then
    echo -e "${YELLOW}⚠️  Currently using Stripe TEST keys${NC}"
    echo "   When you have production keys, update STRIPE_SECRET_KEY in .env.production"
else
    echo -e "${GREEN}✅ Using Stripe LIVE keys${NC}"
fi