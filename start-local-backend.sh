#!/bin/bash

echo "🚀 Starting API-Direct Local Development Environment"
echo "=================================================="

# Check if docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker Desktop."
    exit 1
fi

# Use the mock auth version of main.py
echo "Setting up mock authentication..."
cp backend/api/main_with_mock.py backend/api/main.py

# Start services with docker-compose
echo "Starting PostgreSQL and Redis..."
docker-compose -f docker-compose.local.yml up -d postgres redis

# Wait for services to be healthy
echo "Waiting for services to be ready..."
sleep 5

# Install Python dependencies locally if needed
if [ ! -d "backend/venv" ]; then
    echo "Creating Python virtual environment..."
    cd backend
    python3 -m venv venv
    source venv/bin/activate
    pip install -r requirements.txt
    cd ..
fi

# Start the FastAPI backend
echo "Starting FastAPI backend on http://localhost:8000"
echo "API Documentation: http://localhost:8000/docs"
echo ""
echo "📝 Mock Login Credentials:"
echo "Email: demo@apidirect.dev"
echo "Password: secret"
echo ""

# Run the backend
cd backend
source venv/bin/activate
export DATABASE_URL="postgresql://apidirect:localpassword@localhost:5432/apidirect"
export REDIS_URL="redis://localhost:6379"
export JWT_SECRET="local-development-secret"
export USE_MOCK_AUTH="true"
export CORS_ORIGINS="http://localhost:3000,http://localhost:3001,http://localhost:8080,https://console.apidirect.dev,https://marketplace.apidirect.dev"

uvicorn api.main:app --host 0.0.0.0 --port 8000 --reload