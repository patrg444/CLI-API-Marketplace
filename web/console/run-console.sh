#!/bin/bash

echo "Starting API-Direct Console..."

# Check if backend is running
if ! curl -s http://localhost:8000/health > /dev/null 2>&1; then
    echo "❌ Backend is not running. Please start the backend first:"
    echo "   cd ../../backend/api && python main.py"
    exit 1
fi

echo "✅ Backend is running"

# Start the console server
echo "Starting console server on http://localhost:5000"
python app.py