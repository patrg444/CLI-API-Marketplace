# 🚀 API-Direct Deployment Options

API-Direct offers **two deployment modes** to match your needs - whether you want zero-hassle managed hosting or full control with your own AWS account.

## 📊 Deployment Comparison

| Feature | Hosted (Default) | BYOA |
|---------|-----------------|------|
| **AWS Account Required** | ❌ No | ✅ Yes |
| **Setup Time** | ⚡ 2 minutes | ⏱️ 5 minutes |
| **Infrastructure Management** | 🤖 Fully managed | 👤 Self-managed |
| **Data Location** | ☁️ API-Direct cloud | 🏢 Your AWS account |
| **Pricing** | 💵 Usage-based | 💰 Direct AWS costs |
| **Compliance** | ✅ SOC2, GDPR | 🎯 Your compliance |
| **Scaling** | 🚀 Automatic | ⚙️ Configurable |
| **SSL/TLS** | 🔒 Included | 🔐 You manage |
| **Monitoring** | 📊 Built-in | 📈 CloudWatch |
| **Updates** | 🔄 Automatic | 🔧 Manual |

## 🌟 Hosted Deployment (Recommended for Most Users)

### How to Deploy
```bash
# Simple deployment (hosted is default)
apidirect deploy my-api

# Or explicitly
apidirect deploy my-api --hosted

# With options
apidirect deploy my-api --replicas 3 --version v1.2.0
```

### What Happens
1. **📦 Code Packaging**: Your code is packaged into a container
2. **🐳 Image Building**: Docker image built on API-Direct infrastructure
3. **🚀 Deployment**: Deployed to managed Kubernetes/ECS cluster
4. **🌐 Instant Access**: Get HTTPS endpoint immediately
5. **📊 Auto-scaling**: Scales based on traffic automatically

### Example Output
```
🚀 Deploying 'my-api' to hosted infrastructure
☁️  Using API-Direct hosted infrastructure...
📋 Configuration: python3.9 runtime, port 8000
🐳 Building container image...
⬆️  Uploading code and building image...
🚀 Deploying to platform...
⏳ Waiting for deployment to be ready... ✓

✅ Deployment successful!
🌐 API URL: https://my-api-abc123.api-direct.io
🆔 Deployment ID: dep_8n2k4m6p
📊 Dashboard: https://console.api-direct.io/apis/dep_8n2k4m6p

📍 Available endpoints:
   https://my-api-abc123.api-direct.io/
   https://my-api-abc123.api-direct.io/health
   https://my-api-abc123.api-direct.io/api/users

🧪 Test your API:
   curl https://my-api-abc123.api-direct.io/health

📝 Next steps:
   View logs:  apidirect logs my-api
   Update:     apidirect deploy
   Scale:      apidirect scale my-api --replicas 3
```

### Pricing (Hosted)
- **Free Tier**: 100K requests/month
- **Starter**: $9/month + $0.50 per million requests
- **Pro**: $49/month + $0.30 per million requests
- **Enterprise**: Custom pricing

### Best For
- ✅ Rapid prototyping
- ✅ Small to medium APIs
- ✅ Teams without DevOps expertise
- ✅ Cost-predictable workloads
- ✅ Quick time-to-market

## 🏢 BYOA Deployment (Bring Your Own AWS)

### How to Deploy
```bash
# Deploy to your AWS account
apidirect deploy my-api --hosted=false

# Skip confirmations
apidirect deploy my-api --hosted=false --yes
```

### What Happens
1. **🔐 AWS Verification**: Checks your AWS credentials
2. **📋 Plan Generation**: Creates Terraform plan
3. **🏗️ Infrastructure Creation**: Builds complete AWS infrastructure
4. **🐳 Container Deployment**: Pushes to your ECR, deploys to ECS
5. **🔧 Configuration**: Sets up monitoring, scaling, security

### Example Output
```
🚀 Deploying 'my-api' to your AWS account
🔐 AWS Account: 123456789012
👤 AWS User: arn:aws:iam::123456789012:user/developer
🔧 Preparing deployment environment...
📋 Creating deployment plan...

⚠️  This will create AWS resources in account 123456789012 (region: us-east-1)
   Estimated cost: ~$50-300/month depending on usage

Do you want to continue? [y/N]: y

🚀 Deploying infrastructure...
✅ BYOA Deployment successful!
🌐 API URL: https://alb-my-api-123456.us-east-1.elb.amazonaws.com
🆔 Deployment ID: my-api-prod-123456789012
☁️  AWS Account: 123456789012
📍 AWS Region: us-east-1

📝 Next steps:
   1. Update DNS: Point your domain to alb-my-api-123456.us-east-1.elb.amazonaws.com
   2. Configure SSL: Add certificate to ALB
   3. Set environment variables in AWS Systems Manager
   4. Monitor: Check CloudWatch logs and metrics

💡 Manage your deployment:
   View status:  apidirect status my-api
   View logs:    apidirect logs my-api
   Update:       apidirect deploy
   Destroy:      apidirect destroy my-api
```

### Infrastructure Created (BYOA)
- **VPC**: Multi-AZ with public/private subnets
- **Load Balancer**: Application Load Balancer with health checks
- **Compute**: ECS Fargate with auto-scaling
- **Database**: RDS PostgreSQL (optional)
- **Storage**: S3 buckets for artifacts
- **Security**: IAM roles, security groups, KMS encryption
- **Monitoring**: CloudWatch logs, metrics, dashboards

### Cost Breakdown (BYOA)
Direct AWS pricing:
- **ALB**: ~$25/month + data transfer
- **ECS Fargate**: ~$0.04/vCPU/hour
- **RDS** (if enabled): ~$50/month (db.t3.micro)
- **Data Transfer**: ~$0.09/GB
- **Total**: ~$50-300/month depending on usage

### Best For
- ✅ Enterprise deployments
- ✅ Compliance requirements (HIPAA, PCI)
- ✅ Large-scale APIs
- ✅ Custom infrastructure needs
- ✅ Complete data control

## 🤔 How to Choose?

### Choose **Hosted** if you:
- Want to deploy in < 2 minutes
- Don't have AWS expertise
- Prefer predictable pricing
- Need automatic updates
- Want built-in SSL and monitoring
- Are building MVPs or prototypes

### Choose **BYOA** if you:
- Need complete infrastructure control
- Have compliance requirements
- Want to minimize vendor lock-in
- Have DevOps expertise
- Need custom networking/security
- Want direct AWS pricing

## 🔄 Switching Between Modes

You can start with Hosted and switch to BYOA later:

```bash
# Start with hosted (quick MVP)
apidirect deploy my-api

# Later, migrate to BYOA
apidirect export my-api
apidirect deploy my-api --hosted=false
```

## 📈 Real-World Scenarios

### Scenario 1: Startup MVP
```bash
# Quick deployment for user testing
apidirect init weather-api --template python-fastapi
apidirect deploy weather-api
# Live in 2 minutes! ✅
```

### Scenario 2: Enterprise API
```bash
# Full control for production workload
apidirect import ./existing-api
apidirect deploy api-prod --hosted=false
# Complete infrastructure with compliance ✅
```

### Scenario 3: Development → Production
```bash
# Development (hosted)
apidirect deploy my-api --version dev

# Staging (hosted with more resources)
apidirect deploy my-api-staging --replicas 2

# Production (BYOA for full control)
apidirect deploy my-api-prod --hosted=false
```

## 🎯 Summary

**API-Direct gives you the best of both worlds:**
- **Hosted**: Zero-friction deployment for 90% of use cases
- **BYOA**: Enterprise-grade control when you need it

Start with Hosted to get live quickly, then graduate to BYOA as your needs grow. Either way, you get the same powerful CLI, monitoring, and marketplace features!