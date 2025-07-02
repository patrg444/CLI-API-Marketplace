# 🚀 AWS Quick Start for API Direct Marketplace

## Prerequisites Checklist

- [ ] AWS Account created
- [ ] AWS CLI installed ([Download](https://aws.amazon.com/cli/))
- [ ] Terraform installed ([Download](https://www.terraform.io/downloads))
- [ ] Domain name ready (or use a subdomain)

## Step 1: Configure AWS CLI

```bash
aws configure
```

Enter:
- **AWS Access Key ID**: (from AWS Console → IAM → Users → Your User → Security Credentials)
- **AWS Secret Access Key**: (same location)
- **Default region**: us-east-1 (or your preferred region)
- **Default output format**: json

## Step 2: Run Setup Script

```bash
./setup-aws.sh
```

This will:
1. Check prerequisites ✓
2. Ask for your configuration (domain, email)
3. Generate secure passwords
4. Create all AWS infrastructure
5. Output connection details

## Step 3: What Gets Created

### Core Infrastructure ($80-120/month for dev)
- **VPC** with public/private subnets
- **RDS PostgreSQL** database
- **ElastiCache Redis** for caching
- **ECS Fargate** for containers
- **Application Load Balancer**

### Authentication & Storage
- **Cognito User Pool** for user auth
- **S3 Buckets** for file storage
- **CloudWatch** for logs

### Networking & Security
- **Security Groups** properly configured
- **IAM Roles** with least privilege
- **SSL/TLS** ready

## Step 4: After Setup

1. **Update DNS** - Point your domain to the Load Balancer
2. **Deploy Services** - Run the deployment scripts
3. **Verify** - Check all services are running

## Cost Optimization Tips

### Development Environment
- Use t3.micro instances (~$15-30/month each)
- Stop services when not in use
- Use AWS Free Tier where possible

### Production Environment
- Use Reserved Instances (save 30-70%)
- Enable auto-scaling
- Monitor with CloudWatch

## Common Issues

### "Access Denied" Error
```bash
# Check your IAM permissions
aws iam get-user
```

### "Invalid credentials"
```bash
# Re-run configuration
aws configure
```

### High Costs
- Check CloudWatch for unused resources
- Use AWS Cost Explorer
- Set up billing alerts

## Manual Setup Alternative

If you prefer to set up services manually through AWS Console:

1. **Create VPC** (Services → VPC → Create VPC)
2. **Create RDS Database** (Services → RDS → Create database)
3. **Create Cognito User Pool** (Services → Cognito → Create user pool)
4. **Create S3 Buckets** (Services → S3 → Create bucket)
5. **Create ECS Cluster** (Services → ECS → Create cluster)

## Support

Need help? Check:
- [AWS Documentation](https://docs.aws.amazon.com/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- GitHub Issues

Ready? Run `./setup-aws.sh` to get started! 🎉