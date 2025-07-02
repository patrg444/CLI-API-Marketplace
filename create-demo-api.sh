#!/bin/bash
# Create a demo API to showcase on the platform

echo "🚀 Creating Demo Weather API"
echo "==========================="

# Create demo directory
mkdir -p demo-weather-api
cd demo-weather-api

# Create the API file
cat > main.py << 'EOF'
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from datetime import datetime
import random

app = FastAPI(
    title="Weather Forecast API",
    description="Get real-time weather forecasts for any location",
    version="1.0.0"
)

class WeatherResponse(BaseModel):
    location: str
    temperature: float
    condition: str
    humidity: int
    wind_speed: float
    forecast: str
    timestamp: str

@app.get("/")
async def root():
    return {
        "api": "Weather Forecast API",
        "version": "1.0.0",
        "endpoints": ["/weather/{city}", "/forecast/{city}"]
    }

@app.get("/weather/{city}", response_model=WeatherResponse)
async def get_weather(city: str):
    """Get current weather for a city"""
    
    # Simulate weather data
    conditions = ["Sunny", "Cloudy", "Rainy", "Partly Cloudy", "Stormy"]
    forecasts = ["Clear skies expected", "Rain likely", "Storms approaching", "Mild conditions"]
    
    return WeatherResponse(
        location=city.title(),
        temperature=round(random.uniform(50, 95), 1),
        condition=random.choice(conditions),
        humidity=random.randint(30, 80),
        wind_speed=round(random.uniform(5, 25), 1),
        forecast=random.choice(forecasts),
        timestamp=datetime.now().isoformat()
    )

@app.get("/forecast/{city}")
async def get_forecast(city: str, days: int = 5):
    """Get weather forecast for multiple days"""
    
    forecasts = []
    for i in range(days):
        conditions = ["Sunny", "Cloudy", "Rainy", "Partly Cloudy"]
        forecasts.append({
            "day": i + 1,
            "high": round(random.uniform(70, 95), 1),
            "low": round(random.uniform(50, 70), 1),
            "condition": random.choice(conditions),
            "precipitation": f"{random.randint(0, 80)}%"
        })
    
    return {
        "location": city.title(),
        "days": days,
        "forecast": forecasts
    }

@app.get("/health")
async def health():
    return {"status": "healthy"}
EOF

# Create requirements file
cat > requirements.txt << 'EOF'
fastapi==0.115.0
uvicorn==0.30.0
pydantic==2.0.0
EOF

# Create API manifest for API Direct
cat > apidirect.yaml << 'EOF'
name: weather-forecast-api
version: 1.0.0
description: Real-time weather forecasts for any location
category: Data & Analytics

pricing:
  - name: Free Tier
    price: 0
    requests: 1000
    description: Perfect for testing
  - name: Starter
    price: 9.99
    requests: 10000
    description: For small applications
  - name: Pro
    price: 49.99
    requests: 100000
    description: For production use

endpoints:
  - path: /weather/{city}
    method: GET
    description: Get current weather for a city
    parameters:
      - name: city
        type: string
        required: true
        description: City name
  - path: /forecast/{city}
    method: GET
    description: Get multi-day forecast
    parameters:
      - name: city
        type: string
        required: true
        description: City name
      - name: days
        type: integer
        required: false
        default: 5
        description: Number of days to forecast

tags:
  - weather
  - forecast
  - climate
  - data
EOF

echo "✅ Demo API created in ./demo-weather-api/"
echo ""
echo "To deploy this API:"
echo "1. cd demo-weather-api"
echo "2. apidirect init"
echo "3. apidirect deploy"
echo ""
echo "To test locally:"
echo "uvicorn main:app --reload"