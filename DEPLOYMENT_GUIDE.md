# 🚀 API-Direct Production Deployment Guide

This guide covers deploying API-Direct to production with full monitoring, SSL certificates, and automated backups.

## 📋 Prerequisites

- Ubuntu 20.04+ or similar Linux distribution
- Minimum 4GB RAM, 2 CPU cores, 50GB storage
- Domain names configured:
  - `api-direct.io` (main website)
  - `console.api-direct.io` (creator portal)
  - `api.api-direct.io` (API endpoints)
- Email address for SSL certificates
- Stripe account for payments
- AWS account for backups (optional)

## 🔧 Quick Deployment

### 1. Clone and Configure

```bash
git clone https://github.com/your-org/api-direct.git
cd api-direct

# Copy and configure environment
cp .env.production.example .env.production
nano .env.production  # Configure all required variables
```

### 2. Run Deployment Script

**For Production:**
```bash
chmod +x scripts/deploy-production.sh
./scripts/deploy-production.sh
```

**For Development:**
```bash
chmod +x scripts/deploy-development.sh
./scripts/deploy-development.sh
```

The production script will:
- Install Docker and Docker Compose
- Generate SSL certificates with Let's Encrypt
- Build and start all services (including marketplace frontend)
- Set up automated backups
- Configure monitoring

### 3. Verify Deployment

```bash
./scripts/health-check.sh
```

## 🔐 Environment Configuration

### Required Variables

```env
# Database
POSTGRES_PASSWORD=your_secure_postgres_password

# Redis
REDIS_PASSWORD=your_secure_redis_password

# InfluxDB
INFLUXDB_PASSWORD=your_secure_influxdb_password
INFLUXDB_TOKEN=your_influxdb_admin_token

# JWT Security
JWT_SECRET=your_very_secure_jwt_secret_key_minimum_32_characters

# Stripe Integration
STRIPE_SECRET_KEY=sk_live_your_stripe_secret_key
STRIPE_PUBLISHABLE_KEY=pk_live_your_stripe_publishable_key
STRIPE_WEBHOOK_SECRET=whsec_your_stripe_webhook_secret

# Authentication (NextAuth)
NEXTAUTH_SECRET=your_very_secure_nextauth_secret_key
NEXTAUTH_URL=https://your-domain.com

# AWS Cognito (if using)
COGNITO_USER_POOL_ID=your_cognito_pool_id
COGNITO_WEB_CLIENT_ID=your_cognito_client_id
AWS_REGION=us-east-1

# Marketplace Frontend URLs
NEXT_PUBLIC_API_URL=https://api.your-domain.com
NEXT_PUBLIC_APP_URL=https://your-domain.com

# Email Configuration
SMTP_HOST=smtp.your-email-provider.com
SMTP_PORT=587
SMTP_USER=your-smtp-username
SMTP_PASSWORD=your-smtp-password
```

### Optional Variables

```env
# AWS Backups
BACKUP_S3_BUCKET=apidirect-backups
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key

# Monitoring
GRAFANA_PASSWORD=your_secure_grafana_password
```

## 🏗️ Architecture Overview

```
Internet
    │
    ▼
┌─────────────┐
│    Nginx    │ ← SSL Termination, Load Balancing
│ (Port 80/443)│
└─────────────┘
    │
    ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   FastAPI   │    │ PostgreSQL  │    │   Redis     │
│  Backend    │────│  Database   │    │   Cache     │
│ (Port 8000) │    │ (Port 5432) │    │ (Port 6379) │
└─────────────┘    └─────────────┘    └─────────────┘
    │
    ▼
┌─────────────┐    ┌─────────────┐
│  InfluxDB   │    │  Grafana    │
│ Analytics   │    │ Monitoring  │
│ (Port 8086) │    │ (Port 3000) │
└─────────────┘    └─────────────┘
```

## 🌐 Service URLs

After deployment, the following services will be available:

- **Marketplace Frontend**: https://api-direct.io (port 3001)
- **Creator Portal**: https://console.api-direct.io (port 3000)
- **API Endpoints**: https://api.api-direct.io
- **Monitoring Dashboard**: https://api-direct.io:3000
- **Prometheus Metrics**: http://localhost:9090
- **Health Check**: https://api-direct.io/api/health

## 📊 Monitoring & Logging

### Prometheus Metrics

Available at `http://localhost:9090`:
- API request rates and response times
- Database connection pools
- System resources (CPU, memory, disk)
- Custom business metrics

### Grafana Dashboard

Available at `https://api-direct.io:3000`:
- Real-time system overview
- API performance metrics
- User activity tracking
- Revenue and billing analytics

### Log Files

```bash
# View all service logs
docker-compose -f docker-compose.production.yml logs

# View specific service logs
docker-compose -f docker-compose.production.yml logs backend
docker-compose -f docker-compose.production.yml logs nginx

# Follow logs in real-time
docker-compose -f docker-compose.production.yml logs -f
```

## 🔄 Backup & Recovery

### Automated Backups

Daily backups are automatically configured for:
- PostgreSQL database
- InfluxDB analytics data
- Application configuration

Backups are stored locally and optionally uploaded to S3.

### Manual Backup

```bash
./scripts/backup.sh
```

### Restore from Backup

```bash
# Restore database
gunzip -c /var/backups/apidirect/postgres_YYYYMMDD_HHMMSS.sql.gz | \
docker-compose -f docker-compose.production.yml exec -T postgres psql -U apidirect apidirect

# Restore InfluxDB
docker-compose -f docker-compose.production.yml exec influxdb influx restore /path/to/backup
```

## 🔧 Maintenance Commands

### Update Application

```bash
git pull origin main
docker-compose -f docker-compose.production.yml build
docker-compose -f docker-compose.production.yml up -d
```

### Scale Services

```bash
# Scale backend to 4 instances
docker-compose -f docker-compose.production.yml up -d --scale backend=4
```

### SSL Certificate Renewal

SSL certificates are automatically renewed via cron job. Manual renewal:

```bash
sudo certbot renew
docker-compose -f docker-compose.production.yml restart nginx
```

## 🛡️ Security Considerations

### Firewall Configuration

```bash
# Allow HTTP, HTTPS, and SSH only
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 22/tcp
sudo ufw enable
```

### Database Security

- Databases are not exposed to public internet
- Strong passwords enforced
- Regular security updates applied

### API Rate Limiting

- Console: 5 requests/second per IP
- API: 10 requests/second per IP
- Configurable in `nginx/nginx.conf`

## 🔍 Troubleshooting

### Common Issues

**Services won't start:**
```bash
# Check logs
docker-compose -f docker-compose.production.yml logs

# Check system resources
./scripts/health-check.sh
```

**SSL certificate issues:**
```bash
# Check certificate status
sudo certbot certificates

# Regenerate certificates
sudo certbot certonly --standalone -d api-direct.io
```

**Database connection errors:**
```bash
# Test database connection
docker-compose -f docker-compose.production.yml exec postgres pg_isready -U apidirect
```

### Performance Tuning

**Increase backend workers:**
Edit `docker-compose.production.yml`:
```yaml
command: ["uvicorn", "api.main:app", "--host", "0.0.0.0", "--port", "8000", "--workers", "8"]
```

**Database optimization:**
```sql
-- Run in PostgreSQL
ANALYZE;
REINDEX DATABASE apidirect;
```

## 📞 Support

For deployment support:
- Check the health monitoring dashboard
- Review application logs
- Run the health check script
- Contact the development team with specific error messages

## 🔄 Updates & Migration

When updating API-Direct:

1. **Backup** current deployment
2. **Pull** latest changes
3. **Review** changelog for breaking changes
4. **Update** environment variables if needed
5. **Deploy** using the deployment script
6. **Verify** with health checks

Remember to test updates in a staging environment first!