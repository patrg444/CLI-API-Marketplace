#!/bin/bash

# Development environment setup for API-Direct

echo "🔧 Setting up API-Direct development environment..."
echo ""

# Check Python version
echo "Checking Python version..."
python3 --version

# Install backend dependencies
echo ""
echo "Installing backend dependencies..."
pip3 install fastapi uvicorn[standard] pydantic python-jose[cryptography] passlib[bcrypt] python-multipart asyncpg redis aioredis influxdb-client stripe pytest httpx

# Create necessary directories
echo ""
echo "Creating necessary directories..."
mkdir -p logs
mkdir -p web/console/static/js
mkdir -p web/console/static/css

echo ""
echo "✅ Setup complete!"
echo ""
echo "To start the platform locally, run:"
echo "  ./start-local.sh"
echo ""
echo "Or start services individually:"
echo "  Backend:  cd backend && python3 -m uvicorn api.main:app --reload"
echo "  Frontend: cd web/console && python3 -m http.server 8080"
echo "  Landing:  cd web/landing && python3 -m http.server 3003"