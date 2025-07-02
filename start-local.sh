#!/bin/bash

# Simple local startup script for API-Direct

echo "🚀 Starting API-Direct Platform Locally..."

# Create logs directory
mkdir -p logs

# Kill any existing processes
echo "Cleaning up existing processes..."
pkill -f "python.*8000" || true
pkill -f "python.*8080" || true
pkill -f "python.*3003" || true
sleep 2

# Start Backend API (in background)
echo "Starting Backend API on http://localhost:8000..."
cd backend
python3 -m uvicorn api.main:app --host 0.0.0.0 --port 8000 --reload > ../logs/backend.log 2>&1 &
cd ..

# Wait for backend to start
sleep 3

# Start Frontend (in background)
echo "Starting Creator Portal on http://localhost:8080..."
cd web/console
python3 -m http.server 8080 > ../../logs/frontend.log 2>&1 &
cd ../..

# Start Landing Page (in background)
echo "Starting Landing Page on http://localhost:3003..."
cd web/landing
python3 -m http.server 3003 > ../../logs/landing.log 2>&1 &
cd ../..

sleep 2

echo ""
echo "✅ All services started!"
echo ""
echo "🌐 Access the platform:"
echo "   Landing Page:    http://localhost:3003"
echo "   Creator Portal:  http://localhost:8080/login.html"
echo "   API Docs:        http://localhost:8000/docs"
echo ""
echo "📝 Logs are available in the logs/ directory"
echo ""
echo "Press Ctrl+C to stop all services"

# Wait for Ctrl+C
trap 'echo "Stopping services..."; pkill -f "python.*8000"; pkill -f "python.*8080"; pkill -f "python.*3003"; exit' INT

while true; do
    sleep 1
done