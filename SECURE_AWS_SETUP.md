# 🔐 Secure AWS Setup Guide

## ⚠️ Security First

Since AWS credentials were exposed, follow these steps:

### 1. **Create New IAM User**

```bash
# Create a new IAM user with limited permissions
aws iam create-user --user-name apidirect-dev

# Create access key
aws iam create-access-key --user-name apidirect-dev

# Attach necessary policies
aws iam attach-user-policy --user-name apidirect-dev --policy-arn arn:aws:iam::aws:policy/AmazonS3FullAccess
aws iam attach-user-policy --user-name apidirect-dev --policy-arn arn:aws:iam::aws:policy/AmazonCognitoPowerUser
```

### 2. **Update .env File**

```bash
# Edit .env file
nano .env

# Add your NEW credentials:
AWS_ACCESS_KEY_ID=your-new-key
AWS_SECRET_ACCESS_KEY=your-new-secret
```

### 3. **Set Up AWS Resources**

#### Create Cognito User Pool
```bash
# Create user pool
aws cognito-idp create-user-pool \
  --pool-name apidirect-users \
  --auto-verified-attributes email \
  --username-attributes email \
  --schema Name=email,Required=true,Mutable=false

# Note the UserPoolId from output

# Create app client
aws cognito-idp create-user-pool-client \
  --user-pool-id YOUR_POOL_ID \
  --client-name apidirect-app \
  --explicit-auth-flows ALLOW_USER_PASSWORD_AUTH ALLOW_REFRESH_TOKEN_AUTH
```

#### Create S3 Buckets
```bash
# Create unique bucket names
TIMESTAMP=$(date +%s)
aws s3 mb s3://apidirect-code-storage-$TIMESTAMP
aws s3 mb s3://apidirect-artifacts-$TIMESTAMP

# Update .env with bucket names
```

### 4. **Configure Environment**

Update your `.env` file:

```env
# AWS Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-new-access-key-id
AWS_SECRET_ACCESS_KEY=your-new-secret-access-key

# Cognito (from step 3)
COGNITO_USER_POOL_ID=us-east-1_xxxxxxxxx
COGNITO_CLIENT_ID=xxxxxxxxxxxxxxxxxxxxxxxxxx
COGNITO_REGION=us-east-1

# S3 Buckets (from step 3)
CODE_STORAGE_BUCKET=apidirect-code-storage-1234567890
ARTIFACTS_BUCKET=apidirect-artifacts-1234567890

# Switch to real auth
USE_MOCK_AUTH=false
```

### 5. **Test AWS Connection**

```bash
# Test AWS credentials
aws sts get-caller-identity

# Test S3 access
aws s3 ls

# Test Cognito
aws cognito-idp describe-user-pool --user-pool-id YOUR_POOL_ID
```

## 🚀 Start Platform with AWS

```bash
# Load environment variables
source .env

# Start all services
./start-platform.sh

# The platform will now use:
# - Real AWS Cognito authentication
# - S3 for storage
# - AWS services for deployment
```

## 🔒 Security Best Practices

1. **Use IAM Roles in Production**
   - Don't use access keys in production
   - Use EC2/ECS instance roles

2. **Rotate Credentials Regularly**
   ```bash
   aws iam create-access-key --user-name apidirect-dev
   aws iam delete-access-key --access-key-id OLD_KEY --user-name apidirect-dev
   ```

3. **Use AWS Secrets Manager**
   ```bash
   aws secretsmanager create-secret \
     --name apidirect/dev/credentials \
     --secret-string '{"access_key":"xxx","secret_key":"yyy"}'
   ```

4. **Enable MFA**
   ```bash
   aws iam enable-mfa-device \
     --user-name apidirect-dev \
     --serial-number arn:aws:iam::ACCOUNT:mfa/DEVICE \
     --authentication-code1 123456 \
     --authentication-code2 789012
   ```

## 📊 Monitor AWS Usage

```bash
# Check for any unauthorized usage
aws cloudtrail lookup-events --start-time 2024-01-01

# Monitor costs
aws ce get-cost-and-usage \
  --time-period Start=2024-01-01,End=2024-01-31 \
  --granularity DAILY \
  --metrics "UnblendedCost"
```

## ✅ Verification Checklist

- [ ] Old AWS credentials deleted
- [ ] New IAM user created with minimal permissions
- [ ] .env updated with new credentials
- [ ] Cognito User Pool created
- [ ] S3 buckets created
- [ ] Platform starts successfully
- [ ] CLI can authenticate with Cognito
- [ ] No credentials in Git history

## 🎯 Next Steps

1. Test the CLI with real Cognito:
   ```bash
   export APIDIRECT_COGNITO_POOL=your-pool-id
   export APIDIRECT_COGNITO_CLIENT=your-client-id
   ./cli/apidirect login
   ```

2. Deploy a test API:
   ```bash
   cd examples/hello-world
   apidirect deploy
   ```

3. Monitor in AWS Console:
   - CloudWatch logs
   - S3 bucket contents
   - Cognito user pool

Remember: **Never share AWS credentials in any communication channel!**