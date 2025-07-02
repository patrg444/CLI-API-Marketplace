#!/bin/bash
# Post-deployment test script

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}API Direct Marketplace - Deployment Test${NC}"
echo "========================================"

# Load environment variables
if [[ -f ".env.production" ]]; then
    source .env.production
else
    echo -e "${RED}Error: .env.production not found${NC}"
    echo "Run setup-aws.sh first"
    exit 1
fi

TESTS_PASSED=0
TESTS_FAILED=0

# Test function
test_service() {
    local name=$1
    local command=$2
    
    echo -n "Testing $name... "
    
    if eval $command &> /dev/null; then
        echo -e "${GREEN}✓ Passed${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗ Failed${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

echo -e "\n${YELLOW}1. Infrastructure Tests${NC}"
echo "======================"

# Test VPC
test_service "VPC" "aws ec2 describe-vpcs --filters 'Name=tag:Project,Values=apidirect' --query 'Vpcs[0].VpcId'"

# Test RDS
test_service "RDS Database" "aws rds describe-db-instances --query 'DBInstances[?contains(DBInstanceIdentifier, `apidirect`)]' | grep -q DBInstanceIdentifier"

# Test ElastiCache
test_service "Redis Cache" "aws elasticache describe-cache-clusters --query 'CacheClusters[?contains(CacheClusterId, `apidirect`)]' | grep -q CacheClusterId"

# Test Cognito
test_service "Cognito User Pool" "aws cognito-idp list-user-pools --max-results 60 --query 'UserPools[?contains(Name, `apidirect`)]' | grep -q Name"

# Test S3 Buckets
test_service "S3 Buckets" "aws s3 ls | grep -q apidirect"

# Test ECS Cluster
test_service "ECS Cluster" "aws ecs list-clusters | grep -q apidirect"

# Test ALB
test_service "Load Balancer" "aws elbv2 describe-load-balancers --query 'LoadBalancers[?contains(LoadBalancerName, `apidirect`)]' | grep -q LoadBalancerName"

echo -e "\n${YELLOW}2. Connectivity Tests${NC}"
echo "===================="

# Test database connection
if [[ -n "$DATABASE_URL" ]]; then
    DB_HOST=$(echo $DATABASE_URL | sed -n 's/.*@\([^:]*\):.*/\1/p')
    echo -n "Testing database connectivity... "
    if nc -zv $DB_HOST 5432 &> /dev/null; then
        echo -e "${GREEN}✓ Passed${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗ Failed${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
fi

# Test Redis connection
if [[ -n "$REDIS_URL" ]]; then
    REDIS_HOST=$(echo $REDIS_URL | sed -n 's/.*@\([^:]*\):.*/\1/p')
    echo -n "Testing Redis connectivity... "
    if nc -zv $REDIS_HOST 6379 &> /dev/null; then
        echo -e "${GREEN}✓ Passed${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗ Failed${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
fi

echo -e "\n${YELLOW}3. Service Health Checks${NC}"
echo "======================="

# Get ALB DNS name
ALB_DNS=$(aws elbv2 describe-load-balancers --query "LoadBalancers[?contains(LoadBalancerName, 'apidirect')].DNSName" --output text 2>/dev/null || echo "")

if [[ -n "$ALB_DNS" ]]; then
    echo "Load Balancer DNS: $ALB_DNS"
    
    # Test ALB health
    echo -n "Testing load balancer health... "
    if curl -s -o /dev/null -w "%{http_code}" http://$ALB_DNS/health | grep -q "200"; then
        echo -e "${GREEN}✓ Passed${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${YELLOW}⚠ Not ready yet${NC}"
    fi
fi

# Test API Gateway
API_GW_URL=$(aws apigatewayv2 get-apis --query "Items[?contains(Name, 'apidirect')].ApiEndpoint" --output text 2>/dev/null || echo "")
if [[ -n "$API_GW_URL" ]]; then
    echo "API Gateway URL: $API_GW_URL"
fi

echo -e "\n${YELLOW}4. Security Tests${NC}"
echo "================="

# Check SSL certificate
test_service "SSL Certificate" "aws acm list-certificates --query 'CertificateSummaryList[?contains(DomainName, `$DOMAIN`)]' | grep -q DomainName"

# Check security groups
echo -n "Checking security group rules... "
SG_COUNT=$(aws ec2 describe-security-groups --filters "Name=tag:Project,Values=apidirect" --query 'length(SecurityGroups)' --output text)
if [[ $SG_COUNT -gt 0 ]]; then
    echo -e "${GREEN}✓ $SG_COUNT security groups found${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ No security groups found${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

echo -e "\n${YELLOW}5. Cost Estimation${NC}"
echo "================="

# Get instance types
RDS_INSTANCE=$(aws rds describe-db-instances --query "DBInstances[?contains(DBInstanceIdentifier, 'apidirect')].DBInstanceClass" --output text 2>/dev/null || echo "Not found")
REDIS_NODE=$(aws elasticache describe-cache-clusters --query "CacheClusters[?contains(CacheClusterId, 'apidirect')].CacheNodeType" --output text 2>/dev/null || echo "Not found")

echo "RDS Instance Type: $RDS_INSTANCE"
echo "Redis Node Type: $REDIS_NODE"
echo ""
echo "Estimated monthly costs:"
echo "  - RDS: ~$15-30"
echo "  - Redis: ~$13-26"
echo "  - ECS: ~$20-50"
echo "  - ALB: ~$25"
echo "  - Total: ~$80-150/month"

echo -e "\n${YELLOW}6. Next Steps${NC}"
echo "============="

if [[ $TESTS_FAILED -eq 0 ]]; then
    echo -e "${GREEN}✅ All infrastructure tests passed!${NC}"
    echo ""
    echo "1. Update DNS records:"
    echo "   - Point $DOMAIN to $ALB_DNS"
    echo ""
    echo "2. Deploy application:"
    echo "   - Build and push Docker images"
    echo "   - Deploy ECS services"
    echo ""
    echo "3. Configure Stripe:"
    echo "   - Add API keys to Secrets Manager"
    echo "   - Set up webhooks"
else
    echo -e "${RED}❌ Some tests failed ($TESTS_FAILED failures)${NC}"
    echo ""
    echo "Check the failed tests above and:"
    echo "1. Review CloudFormation/Terraform outputs"
    echo "2. Check CloudWatch logs"
    echo "3. Verify IAM permissions"
fi

echo -e "\n${YELLOW}Summary${NC}"
echo "======="
echo "Tests passed: $TESTS_PASSED"
echo "Tests failed: $TESTS_FAILED"

# Generate report
cat > deployment-test-report.txt <<EOF
API Direct Marketplace - Deployment Test Report
Generated: $(date)

Infrastructure Status:
- VPC: $(aws ec2 describe-vpcs --filters 'Name=tag:Project,Values=apidirect' --query 'Vpcs[0].State' --output text 2>/dev/null || echo "Not found")
- RDS: $(aws rds describe-db-instances --query "DBInstances[?contains(DBInstanceIdentifier, 'apidirect')].DBInstanceStatus" --output text 2>/dev/null || echo "Not found")
- Redis: $(aws elasticache describe-cache-clusters --query "CacheClusters[?contains(CacheClusterId, 'apidirect')].CacheClusterStatus" --output text 2>/dev/null || echo "Not found")
- ECS: $(aws ecs describe-clusters --clusters apidirect-$ENVIRONMENT --query 'clusters[0].status' --output text 2>/dev/null || echo "Not found")

Tests Passed: $TESTS_PASSED
Tests Failed: $TESTS_FAILED

Next Steps:
$(if [[ $TESTS_FAILED -eq 0 ]]; then echo "- Ready for application deployment"; else echo "- Fix failed tests before proceeding"; fi)
EOF

echo ""
echo "Test report saved to: deployment-test-report.txt"

exit $TESTS_FAILED