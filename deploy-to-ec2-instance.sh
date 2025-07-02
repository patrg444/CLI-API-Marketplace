#!/bin/bash
# Deploy API Direct to EC2 Instance

set -e

# Configuration from our deployment
INSTANCE_IP="34.194.31.245"
KEY_PATH="$HOME/.ssh/api-direct-key.pem"
REMOTE_USER="ubuntu"

echo "🚀 Deploying API Direct to EC2 Instance"
echo "======================================"
echo "Instance IP: $INSTANCE_IP"
echo ""

# Wait for instance to be ready for SSH
echo "⏳ Waiting for instance to be ready..."
while ! ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 -i "$KEY_PATH" "$REMOTE_USER@$INSTANCE_IP" "echo 'Instance ready'" 2>/dev/null; do
    echo -n "."
    sleep 5
done
echo ""
echo "✅ Instance is ready!"

# Copy necessary files
echo "📦 Copying deployment files..."
scp -i "$KEY_PATH" -o StrictHostKeyChecking=no .env.production "$REMOTE_USER@$INSTANCE_IP:~/"
scp -i "$KEY_PATH" -o StrictHostKeyChecking=no docker-compose.production.yml "$REMOTE_USER@$INSTANCE_IP:~/"
scp -i "$KEY_PATH" -o StrictHostKeyChecking=no deploy-production.sh "$REMOTE_USER@$INSTANCE_IP:~/"
scp -i "$KEY_PATH" -o StrictHostKeyChecking=no nginx.conf "$REMOTE_USER@$INSTANCE_IP:~/" 2>/dev/null || true

# Create deployment script on remote
ssh -i "$KEY_PATH" "$REMOTE_USER@$INSTANCE_IP" << 'REMOTE_SCRIPT'
#!/bin/bash
set -e

echo "🔧 Setting up API Direct on EC2..."

# Update system
sudo apt-get update -y
sudo apt-get upgrade -y

# Install required packages
sudo apt-get install -y docker.io docker-compose git nginx certbot python3-certbot-nginx

# Enable and start Docker
sudo systemctl enable docker
sudo systemctl start docker
sudo usermod -aG docker ubuntu

# Clone repository if not exists
if [ ! -d "CLI-API-Marketplace" ]; then
    git clone https://github.com/yourusername/CLI-API-Marketplace.git
fi

# Copy production files
cp ~/docker-compose.production.yml CLI-API-Marketplace/
cp ~/.env.production CLI-API-Marketplace/
cp ~/deploy-production.sh CLI-API-Marketplace/
chmod +x CLI-API-Marketplace/deploy-production.sh

# Create necessary directories
cd CLI-API-Marketplace
mkdir -p logs backups

# Pull Docker images
sudo docker-compose -f docker-compose.production.yml pull

# Start services
echo "🚀 Starting services..."
sudo docker-compose -f docker-compose.production.yml up -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be healthy..."
sleep 30

# Check service status
sudo docker-compose -f docker-compose.production.yml ps

# Configure Nginx
sudo tee /etc/nginx/sites-available/api-direct << 'EOF'
server {
    listen 80;
    server_name api.apidirect.dev;
    
    location / {
        proxy_pass http://localhost:8000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

# Enable site
sudo ln -sf /etc/nginx/sites-available/api-direct /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl reload nginx

echo "✅ Services deployed successfully!"
echo ""
echo "📊 Service Status:"
sudo docker-compose -f docker-compose.production.yml ps
echo ""
echo "🔗 API endpoint will be available at: http://$HOSTNAME:8000"
echo "   Once DNS is updated: https://api.apidirect.dev"
REMOTE_SCRIPT

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📋 Next Steps:"
echo "1. Update DNS A record for api.apidirect.dev to: $INSTANCE_IP"
echo "2. Once DNS propagates, set up SSL:"
echo "   ssh -i $KEY_PATH $REMOTE_USER@$INSTANCE_IP"
echo "   sudo certbot --nginx -d api.apidirect.dev"
echo ""
echo "📊 To check service status:"
echo "ssh -i $KEY_PATH $REMOTE_USER@$INSTANCE_IP 'cd CLI-API-Marketplace && sudo docker-compose -f docker-compose.production.yml ps'"
echo ""
echo "🔍 To view logs:"
echo "ssh -i $KEY_PATH $REMOTE_USER@$INSTANCE_IP 'cd CLI-API-Marketplace && sudo docker-compose -f docker-compose.production.yml logs -f'"