#!/bin/bash

echo "🚀 Starting Docker and Testing API-Direct"
echo "========================================"
echo ""

# Start Docker Desktop
echo "1. Starting Docker Desktop..."
open -a Docker

echo "   Waiting 30 seconds for Docker to start..."
echo "   (Docker Desktop needs time to initialize)"

# Show countdown
for i in {30..1}; do
    printf "\r   %2d seconds remaining..." "$i"
    sleep 1
done
echo ""

# Check if Docker is running
echo ""
echo "2. Verifying Docker is running..."
if docker ps &> /dev/null; then
    echo "✅ Docker is running!"
else
    echo "❌ Docker failed to start. Please start Docker Desktop manually."
    exit 1
fi

echo ""
echo "3. Starting API-Direct services..."
echo "   This may take 2-3 minutes on first run as images are downloaded..."
echo ""

# Start services
docker-compose -f docker-compose.local.yml up -d

echo ""
echo "4. Services are starting. Waiting 20 seconds for initialization..."
sleep 20

echo ""
echo "5. Checking service status..."
docker-compose -f docker-compose.local.yml ps

echo ""
echo "✅ Setup complete!"
echo ""
echo "📝 Next Steps:"
echo "1. Set up Stripe webhook forwarding:"
echo "   stripe listen --forward-to localhost:8000/api/webhooks/stripe"
echo ""
echo "2. Update .env with the webhook secret shown above"
echo ""
echo "3. Access the services:"
echo "   - Console: http://localhost:3001"
echo "   - Marketplace: http://localhost:3000"
echo "   - API Docs: http://localhost:8000/docs"
echo ""
echo "4. Create a test account and try deploying an API!"
echo ""
echo "To view logs: docker-compose -f docker-compose.local.yml logs -f"
echo ""