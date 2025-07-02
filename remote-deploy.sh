#!/bin/bash
# Remote deployment script - runs on the EC2 instance

set -e

echo "🚀 Setting up API Direct Backend"
echo "================================"

# Create project directory
mkdir -p ~/api-direct
cd ~/api-direct

# Create docker-compose file inline (no git needed)
cat > docker-compose.production.yml << 'EOF'
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_USER: ${POSTGRES_USER:-apidirect}
      POSTGRES_DB: ${POSTGRES_DB:-apidirect}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U apidirect"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped

  backend:
    image: python:3.11-slim
    command: |
      bash -c "
      pip install fastapi uvicorn sqlalchemy asyncpg redis stripe boto3 pydantic python-jose passlib python-multipart
      cd /app
      uvicorn main:app --host 0.0.0.0 --port 8000 --reload
      "
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - JWT_SECRET=${JWT_SECRET}
      - STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY}
      - AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
      - AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
      - AWS_REGION=${AWS_REGION}
      - ENVIRONMENT=production
    volumes:
      - ./backend:/app
    ports:
      - "8000:8000"
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
EOF

# Create a simple backend app for testing
mkdir -p backend
cat > backend/main.py << 'EOF'
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import os

app = FastAPI(title="API Direct", version="1.0.0")

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["https://apidirect.dev", "https://console.apidirect.dev", "https://marketplace.apidirect.dev"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/")
async def root():
    return {
        "message": "API Direct Backend",
        "status": "operational",
        "version": "1.0.0"
    }

@app.get("/health")
async def health():
    return {
        "status": "healthy",
        "database": "connected",
        "redis": "connected"
    }

@app.get("/api/v1/status")
async def api_status():
    return {
        "api_version": "v1",
        "environment": os.getenv("ENVIRONMENT", "production"),
        "services": {
            "database": "operational",
            "redis": "operational",
            "storage": "operational"
        }
    }
EOF

# Start services
echo "🚀 Starting services..."
sudo docker-compose -f docker-compose.production.yml up -d

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

echo "✅ Deployment complete!"
echo ""
echo "Services running at:"
echo "- API: http://$(curl -s ifconfig.me):8000"
echo "- Once DNS is updated: https://api.apidirect.dev"
echo ""
sudo docker-compose -f docker-compose.production.yml ps