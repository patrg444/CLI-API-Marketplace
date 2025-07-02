# AWS Infrastructure Verification Checklist

## Pre-Deployment Verification

### 1. AWS Account Setup ✓
```bash
# Verify correct account
aws sts get-caller-identity
# Should show Account: 0121-7803-6894

# Check IAM permissions
aws iam get-user
aws iam list-attached-user-policies --user-name $(aws iam get-user --query 'User.UserName' --output text)

# Verify you have admin or sufficient permissions
aws ec2 describe-regions --output table
```

### 2. Prerequisites Check
```bash
# Check all tools are installed
aws --version          # Should be 2.x
terraform --version    # Should be 1.x
docker --version       # Should be 20.x+
node --version         # Should be 16.x+
```

### 3. Dry Run Infrastructure
```bash
cd infrastructure/terraform/aws
terraform init
terraform validate     # Check configuration syntax
terraform plan        # Preview what will be created
```

## Post-Deployment Verification

### 1. Infrastructure Health Checks

#### VPC and Networking
```bash
# Check VPC
aws ec2 describe-vpcs --filters "Name=tag:Project,Values=apidirect"

# Check subnets
aws ec2 describe-subnets --filters "Name=tag:Project,Values=apidirect"

# Check security groups
aws ec2 describe-security-groups --filters "Name=tag:Project,Values=apidirect"
```

#### Database
```bash
# Check RDS instance
aws rds describe-db-instances --db-instance-identifier apidirect-$ENVIRONMENT

# Test connection (after getting endpoint)
PGPASSWORD=$DB_PASSWORD psql -h <rds-endpoint> -U apidirect -d apidirect -c "SELECT 1;"
```

#### Redis
```bash
# Check ElastiCache cluster
aws elasticache describe-cache-clusters --show-cache-node-info

# Test connection
redis-cli -h <redis-endpoint> -a $REDIS_TOKEN ping
```

#### Cognito
```bash
# Check user pool
aws cognito-idp list-user-pools --max-results 10

# Get user pool details
aws cognito-idp describe-user-pool --user-pool-id <pool-id>
```

### 2. Application Deployment Tests

#### Build and Push Images
```bash
# Test ECR login
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com

# Build test image
docker build -t test-app .
docker tag test-app:latest $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/apidirect-marketplace:test
docker push $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/apidirect-marketplace:test
```

#### ECS Service Health
```bash
# Check ECS cluster
aws ecs list-clusters
aws ecs describe-clusters --clusters apidirect-$ENVIRONMENT

# Check services
aws ecs list-services --cluster apidirect-$ENVIRONMENT
aws ecs describe-services --cluster apidirect-$ENVIRONMENT --services marketplace
```

### 3. End-to-End Tests

#### API Gateway
```bash
# Test API endpoint
curl -X GET https://<api-gateway-url>/health

# Test with API key
curl -X GET https://<api-gateway-url>/api/marketplace/apis \
  -H "x-api-key: <test-api-key>"
```

#### Load Balancer
```bash
# Check ALB health
aws elbv2 describe-load-balancers --names apidirect-$ENVIRONMENT-alb

# Test ALB endpoint
curl -I http://<alb-dns-name>
```

#### S3 Buckets
```bash
# List buckets
aws s3 ls | grep apidirect

# Test upload
echo "test" > test.txt
aws s3 cp test.txt s3://apidirect-assets-$ENVIRONMENT/test.txt
aws s3 rm s3://apidirect-assets-$ENVIRONMENT/test.txt
```

### 4. Security Verification

```bash
# Check SSL certificate
aws acm list-certificates

# Verify security groups are restrictive
aws ec2 describe-security-groups --group-ids <sg-id> --query 'SecurityGroups[*].IpPermissions'

# Check IAM roles
aws iam list-roles | grep apidirect
```

### 5. Monitoring Setup

```bash
# Check CloudWatch log groups
aws logs describe-log-groups --log-group-name-prefix /ecs/apidirect

# Check alarms
aws cloudwatch describe-alarms --alarm-name-prefix apidirect
```

## Automated Test Script

Create `test-infrastructure.sh`:

```bash
#!/bin/bash
set -e

echo "🔍 Testing AWS Infrastructure..."

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# Test function
test_command() {
    if $1 &> /dev/null; then
        echo -e "${GREEN}✓${NC} $2"
    else
        echo -e "${RED}✗${NC} $2"
        return 1
    fi
}

# Run tests
test_command "aws sts get-caller-identity" "AWS credentials"
test_command "aws ec2 describe-vpcs --filters Name=tag:Project,Values=apidirect" "VPC created"
test_command "aws rds describe-db-instances" "RDS database"
test_command "aws elasticache describe-cache-clusters" "Redis cache"
test_command "aws cognito-idp list-user-pools --max-results 10" "Cognito user pool"
test_command "aws s3 ls | grep apidirect" "S3 buckets"
test_command "aws ecs list-clusters | grep apidirect" "ECS cluster"

echo "✅ Infrastructure tests complete!"
```

## Load Testing

Once deployed, run load tests:

```bash
# Install k6
brew install k6

# Run load test
k6 run --vus 10 --duration 30s load-test.js
```

## Rollback Plan

If something goes wrong:

```bash
# Destroy infrastructure (careful!)
cd infrastructure/terraform/aws
terraform destroy

# Or rollback to previous version
terraform apply -target=<specific-resource>
```

## Success Criteria

- [ ] All AWS resources created successfully
- [ ] Database connection works
- [ ] Redis connection works
- [ ] Can create Cognito user
- [ ] S3 upload/download works
- [ ] ECS services are healthy
- [ ] Load balancer responds
- [ ] API Gateway accepts requests
- [ ] CloudWatch logs are flowing
- [ ] Estimated costs match expectations

## Troubleshooting

### Common Issues:

1. **"Access Denied"**
   - Check IAM permissions
   - Ensure using correct AWS profile

2. **"Resource limit exceeded"**
   - Check AWS service quotas
   - Request limit increase if needed

3. **"Connection timeout"**
   - Check security groups
   - Verify VPC networking

4. **High costs**
   - Check for unused resources
   - Enable cost alerts
   - Use AWS Cost Explorer