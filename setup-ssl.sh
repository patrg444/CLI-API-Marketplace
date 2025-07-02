#!/bin/bash
# SSL Setup Script for API Direct

echo "🔒 Setting up SSL for api.apidirect.dev"
echo "========================================"

# Check if DNS has propagated
echo "Checking DNS propagation..."
CURRENT_IP=$(dig +short api.apidirect.dev | head -1)
EXPECTED_IP="34.194.31.245"

if [ "$CURRENT_IP" != "$EXPECTED_IP" ]; then
    echo "❌ DNS not propagated yet"
    echo "   Current: $CURRENT_IP"
    echo "   Expected: $EXPECTED_IP"
    echo ""
    echo "Please wait for DNS propagation (5-30 minutes)"
    echo "You can check status at: https://dnschecker.org/#A/api.apidirect.dev"
    exit 1
fi

echo "✅ DNS is pointing to the correct IP!"
echo ""

# SSH into server and setup SSL
echo "Setting up SSL certificate..."
ssh -i ~/.ssh/api-direct-key.pem -t ubuntu@34.194.31.245 << 'REMOTE_SCRIPT'
# Ensure certbot is installed
sudo apt-get update
sudo apt-get install -y certbot python3-certbot-nginx

# Get SSL certificate
sudo certbot --nginx -d api.apidirect.dev \
    --non-interactive \
    --agree-tos \
    --email admin@apidirect.dev \
    --redirect

# Restart nginx
sudo systemctl reload nginx

echo ""
echo "✅ SSL certificate installed!"
echo ""
echo "Testing HTTPS..."
curl -s -o /dev/null -w "%{http_code}" https://api.apidirect.dev/health
REMOTE_SCRIPT

echo ""
echo "🎉 SSL setup complete!"
echo ""
echo "Your API is now available at:"
echo "- https://api.apidirect.dev"
echo "- https://api.apidirect.dev/health"
echo "- https://api.apidirect.dev/docs"