# 🚀 API-Direct Production Setup Guide

## Current Status
- ✅ Platform architecture complete (90%)
- ✅ Console monetization UI implemented
- ✅ All subdomains deployed
- ✅ Testing infrastructure ready
- 🔄 Production configuration needed
- 🔄 Third-party services setup required

## Quick Start (Next Steps)

### 1. Generate Secure Secrets
```bash
./scripts/generate-secrets.sh
# Copy the generated values to .env.production
```

### 2. Set Up Stripe (Required for Payments)

#### Create Stripe Account
1. Go to https://stripe.com and create an account
2. Complete business verification
3. Navigate to Developers > API Keys

#### Get API Keys
```bash
# Live keys (for production)
STRIPE_SECRET_KEY=sk_live_...
STRIPE_PUBLISHABLE_KEY=pk_live_...
```

#### Configure Webhooks
1. Go to Developers > Webhooks
2. Add endpoint: `https://api.apidirect.dev/webhooks/stripe`
3. Select events:
   - `payment_intent.succeeded`
   - `customer.subscription.created`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
   - `invoice.payment_succeeded`
4. Copy webhook secret: `whsec_...`

#### Set Up Stripe Connect (for Creator Payouts)
1. Go to Settings > Connect
2. Enable Express accounts
3. Get Connect client ID: `ca_...`
4. Configure OAuth redirect: `https://console.apidirect.dev/settings/stripe-callback`

### 3. Configure Email Service

#### Option A: SendGrid (Recommended)
```bash
# Sign up at https://sendgrid.com
# Verify your domain
# Create API key with "Mail Send" permission

SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=SG.your-api-key-here
EMAIL_FROM=noreply@apidirect.dev
```

#### Option B: AWS SES
```bash
# In AWS Console:
# 1. Verify domain in SES
# 2. Create SMTP credentials
# 3. Move out of sandbox for production

SMTP_HOST=email-smtp.us-east-1.amazonaws.com
SMTP_PORT=587
SMTP_USER=your-ses-smtp-username
SMTP_PASSWORD=your-ses-smtp-password
EMAIL_FROM=noreply@apidirect.dev
```

### 4. Set Up AWS Services

#### Create IAM User
```bash
# In AWS Console:
# 1. IAM > Users > Add User
# 2. Name: apidirect-production
# 3. Access type: Programmatic access
# 4. Attach policies:
#    - AmazonS3FullAccess (scope down later)
#    - AmazonCognitoPowerUser (if using Cognito)
```

#### Create S3 Buckets
```bash
# Run this AWS CLI command or create in console:
aws s3 mb s3://apidirect-assets --region us-east-1
aws s3 mb s3://apidirect-code-storage --region us-east-1
aws s3 mb s3://apidirect-artifacts --region us-east-1
aws s3 mb s3://apidirect-backups --region us-east-1

# Configure public access for assets bucket
aws s3api put-bucket-policy --bucket apidirect-assets --policy '{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "PublicRead",
    "Effect": "Allow",
    "Principal": "*",
    "Action": "s3:GetObject",
    "Resource": "arn:aws:s3:::apidirect-assets/*"
  }]
}'
```

#### (Optional) Set Up Cognito
```bash
# If using AWS authentication:
# 1. Create user pool in Cognito
# 2. Configure app client
# 3. Set up hosted UI domain
# 4. Update .env.production with pool ID and client ID
```

### 5. Configure Production Server

#### Server Requirements
- Ubuntu 20.04/22.04 LTS
- 4GB+ RAM (8GB recommended)
- 2+ CPU cores
- 50GB+ SSD storage
- Open ports: 22, 80, 443

#### Initial Setup
```bash
# SSH into your server
ssh user@your-server-ip

# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo apt install docker-compose -y

# Install certbot for SSL
sudo apt install certbot python3-certbot-nginx -y

# Clone repository
git clone https://github.com/yourusername/CLI-API-Marketplace.git
cd CLI-API-Marketplace

# Copy production environment
cp .env.production.example .env.production
# Edit with your values
nano .env.production
```

### 6. Set Up SSL Certificates

```bash
# Generate certificates
sudo certbot certonly --standalone \
  -d apidirect.dev \
  -d www.apidirect.dev \
  -d console.apidirect.dev \
  -d marketplace.apidirect.dev \
  -d api.apidirect.dev \
  --email admin@apidirect.dev \
  --agree-tos \
  --no-eff-email

# Certificates will be in /etc/letsencrypt/live/apidirect.dev/
```

### 7. Deploy to Production

```bash
# Run deployment script
./scripts/deploy-production.sh

# Or manually:
docker-compose -f docker-compose.production.yml up -d

# Check status
docker-compose -f docker-compose.production.yml ps

# View logs
docker-compose -f docker-compose.production.yml logs -f
```

### 8. Verify Deployment

```bash
# Health check
curl https://api.apidirect.dev/health

# Test each subdomain
curl -I https://apidirect.dev
curl -I https://console.apidirect.dev
curl -I https://marketplace.apidirect.dev

# Check SSL
openssl s_client -connect apidirect.dev:443 -servername apidirect.dev
```

### 9. Configure DNS

Add these records to your domain:
```
A     @              YOUR_SERVER_IP
A     www            YOUR_SERVER_IP  
A     console        YOUR_SERVER_IP
A     marketplace    YOUR_SERVER_IP
A     api            YOUR_SERVER_IP
```

### 10. Post-Deployment Tasks

#### Set Up Monitoring
```bash
# Access Grafana
https://your-domain.com:3000
# Default: admin/admin (change immediately)

# Configure alerts
# Set up dashboards for:
# - API response times
# - Error rates
# - Resource usage
# - Transaction volume
```

#### Configure Backups
```bash
# Test backup script
./scripts/backup-database.sh

# Set up cron job
crontab -e
# Add: 0 2 * * * /path/to/backup-database.sh
```

#### Security Hardening
```bash
# Configure firewall
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable

# Set up fail2ban
sudo apt install fail2ban -y
sudo systemctl enable fail2ban
```

## Troubleshooting

### Docker Issues
```bash
# If Docker won't start
sudo systemctl status docker
sudo systemctl restart docker

# Permission issues
sudo usermod -aG docker $USER
newgrp docker
```

### SSL Issues
```bash
# Renew certificates
sudo certbot renew --dry-run
sudo certbot renew

# Check nginx config
sudo nginx -t
```

### Database Issues
```bash
# Connect to database
docker-compose -f docker-compose.production.yml exec postgres psql -U apidirect

# Run migrations manually
docker-compose -f docker-compose.production.yml exec backend python -m alembic upgrade head
```

## Production Checklist

- [ ] All environment variables configured
- [ ] Stripe account set up with webhooks
- [ ] Email service configured and tested
- [ ] AWS services created (S3, IAM)
- [ ] SSL certificates installed
- [ ] DNS records configured
- [ ] Firewall configured
- [ ] Monitoring set up
- [ ] Backups configured
- [ ] Health checks passing
- [ ] Test user registration flow
- [ ] Test API deployment flow
- [ ] Test payment flow
- [ ] Test payout flow

## Support Resources

- Documentation: https://docs.apidirect.dev
- GitHub Issues: https://github.com/yourusername/CLI-API-Marketplace/issues
- Stripe Support: https://support.stripe.com
- AWS Support: https://aws.amazon.com/support

## Next Steps After Launch

1. **Monitor First 48 Hours**
   - Watch error logs
   - Monitor resource usage
   - Track user signups
   - Check payment processing

2. **Optimize Performance**
   - Enable CDN (CloudFlare)
   - Optimize database queries
   - Configure caching

3. **Marketing Launch**
   - Announce on social media
   - Submit to directories
   - Reach out to beta users
   - Create demo videos

Good luck with your launch! 🚀