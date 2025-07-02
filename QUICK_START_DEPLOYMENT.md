# 🚀 Quick Start Deployment Guide

This guide will help you deploy API Direct Marketplace to production in under 30 minutes.

## Prerequisites

- Ubuntu 20.04/22.04 server with at least 4GB RAM
- Domain name pointed to your server
- SSH access to the server
- Basic knowledge of Linux commands

## Step 1: Initial Server Setup (5 minutes)

Connect to your server and run:

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install required packages
sudo apt install -y git curl wget nginx certbot python3-certbot-nginx

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo apt install -y docker-compose

# Add your user to docker group
sudo usermod -aG docker $USER

# Log out and back in for group changes to take effect
exit
# Then SSH back in
```

## Step 2: Clone and Configure (5 minutes)

```bash
# Clone the repository
cd /opt
sudo git clone https://github.com/yourusername/CLI-API-Marketplace.git apidirect
sudo chown -R $USER:$USER /opt/apidirect
cd /opt/apidirect

# Copy environment template
cp .env.production.example .env.production

# Generate secure passwords
echo "Generating secure passwords..."
POSTGRES_PASSWORD=$(openssl rand -base64 32)
REDIS_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 64 | tr -d '\n')
GRAFANA_PASSWORD=$(openssl rand -base64 16)

# Update .env.production with generated passwords
sed -i "s/CHANGE_ME_SECURE_PASSWORD_32_CHARS/$POSTGRES_PASSWORD/g" .env.production
sed -i "s/CHANGE_ME_SECURE_PASSWORD_32_CHARS/$REDIS_PASSWORD/g" .env.production
sed -i "s/CHANGE_ME_MINIMUM_64_CHARACTERS/$JWT_SECRET/g" .env.production
sed -i "s/CHANGE_ME_SECURE_PASSWORD/$GRAFANA_PASSWORD/g" .env.production
```

## Step 3: Configure Required Services (10 minutes)

### Edit .env.production with your values:

```bash
nano .env.production
```

Update these critical values:
- `DOMAIN=yourdomain.com` - Your actual domain
- `STRIPE_SECRET_KEY` - From Stripe dashboard
- `STRIPE_PUBLISHABLE_KEY` - From Stripe dashboard
- `AWS_ACCESS_KEY_ID` - Your AWS credentials
- `AWS_SECRET_ACCESS_KEY` - Your AWS credentials
- `EMAIL_FROM` - Your email address
- SMTP settings for your email provider

## Step 4: SSL Certificate (3 minutes)

```bash
# Get SSL certificate
sudo certbot certonly --standalone -d yourdomain.com -d www.yourdomain.com -d api.yourdomain.com -d console.yourdomain.com

# Create SSL directory
mkdir -p nginx/ssl
sudo ln -s /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/
sudo ln -s /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/
```

## Step 5: Deploy Services (5 minutes)

```bash
# Run the deployment script
./deploy-production.sh

# Or manually:
# Build and start all services
docker-compose -f docker-compose.production.yml up -d --build

# Check status
docker-compose -f docker-compose.production.yml ps
```

## Step 6: Verify Installation (2 minutes)

```bash
# Check all services are running
docker ps

# Test endpoints
curl http://localhost:8000/health
curl http://localhost:3001/api/health

# Check logs for errors
docker-compose -f docker-compose.production.yml logs --tail=50
```

## Step 7: Configure DNS

Add these DNS records to your domain:

```
A     @              YOUR_SERVER_IP
A     www            YOUR_SERVER_IP  
A     api            YOUR_SERVER_IP
A     console        YOUR_SERVER_IP
```

## Step 8: Access Your Services

After DNS propagation (5-30 minutes):

- **Marketplace**: https://yourdomain.com
- **API Gateway**: https://api.yourdomain.com
- **Console**: https://console.yourdomain.com
- **Grafana**: http://YOUR_SERVER_IP:3000 (admin/[generated password])
- **Prometheus**: http://YOUR_SERVER_IP:9090

## Post-Deployment Tasks

### 1. Security Hardening (Required)

```bash
# Setup firewall
sudo ./security/setup-firewall.sh

# Install fail2ban
sudo apt install -y fail2ban
sudo cp security/fail2ban-jail.conf /etc/fail2ban/jail.local
sudo systemctl restart fail2ban
```

### 2. Configure Backups (Recommended)

```bash
# Setup automated backups
crontab -e
# Add: 0 2 * * * /opt/apidirect/scripts/backup-automation.sh

# Create backup directories
sudo mkdir -p /var/backups/apidirect
sudo chown $USER:$USER /var/backups/apidirect
```

### 3. Setup Monitoring (Recommended)

```bash
# Start monitoring stack
cd monitoring
./setup-monitoring.sh
./start-monitoring.sh
```

## Troubleshooting

### Services won't start
```bash
# Check logs
docker-compose -f docker-compose.production.yml logs [service-name]

# Restart specific service
docker-compose -f docker-compose.production.yml restart [service-name]
```

### Database connection issues
```bash
# Check PostgreSQL is running
docker exec apidirect-postgres pg_isready

# Check environment variables
docker-compose -f docker-compose.production.yml config
```

### SSL issues
```bash
# Renew certificates
sudo certbot renew --dry-run

# Check nginx config
docker exec apidirect-nginx nginx -t
```

## Quick Commands Reference

```bash
# View all logs
docker-compose -f docker-compose.production.yml logs -f

# Restart all services
docker-compose -f docker-compose.production.yml restart

# Stop all services
docker-compose -f docker-compose.production.yml down

# Update and redeploy
git pull
docker-compose -f docker-compose.production.yml up -d --build

# Database backup
docker exec apidirect-postgres pg_dump -U apidirect apidirect > backup.sql

# Check resource usage
docker stats
```

## Health Checks

Run these commands to verify everything is working:

```bash
# API Health
curl https://api.yourdomain.com/health

# Database connectivity
docker exec apidirect-postgres psql -U apidirect -c "SELECT 1"

# Redis connectivity
docker exec apidirect-redis redis-cli ping

# Service mesh health
for service in backend marketplace gateway apikey billing metering; do
  echo "Checking $service..."
  docker exec apidirect-$service curl -f http://localhost:8000/health || echo "Failed"
done
```

## Support

If you encounter issues:

1. Check the logs: `docker-compose logs [service-name]`
2. Review the [troubleshooting guide](./docs/TROUBLESHOOTING.md)
3. Check [GitHub Issues](https://github.com/yourusername/CLI-API-Marketplace/issues)
4. Contact support at support@yourdomain.com

## Next Steps

1. ✅ Create your first API
2. ✅ Configure payment settings in Stripe
3. ✅ Set up monitoring alerts
4. ✅ Review security checklist
5. ✅ Plan your marketing launch

Congratulations! Your API marketplace is now live! 🎉