#!/usr/bin/env python3
"""
Demo Weather API using API-Direct Framework
Shows how easy it is to create a monetized API
"""

from datetime import datetime
import random

# In production, you'd use: from apidirect_framework import APIDirectFramework
# For demo, we'll use FastAPI
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List, Optional

app = FastAPI(
    title="Weather Forecast API",
    description="Real-time weather data and forecasts - Powered by API-Direct",
    version="1.0.0"
)

# Data models
class WeatherResponse(BaseModel):
    location: str
    temperature: float
    feels_like: float
    humidity: int
    description: str
    wind_speed: float
    timestamp: datetime

class ForecastDay(BaseModel):
    date: str
    high: float
    low: float
    description: str
    precipitation_chance: int

class ForecastResponse(BaseModel):
    location: str
    days: List[ForecastDay]

# Mock data
CITIES = {
    "new-york": {"lat": 40.7128, "lon": -74.0060, "name": "New York, NY"},
    "san-francisco": {"lat": 37.7749, "lon": -122.4194, "name": "San Francisco, CA"},
    "london": {"lat": 51.5074, "lon": -0.1278, "name": "London, UK"},
    "tokyo": {"lat": 35.6762, "lon": 139.6503, "name": "Tokyo, Japan"},
    "sydney": {"lat": -33.8688, "lon": 151.2093, "name": "Sydney, Australia"}
}

WEATHER_CONDITIONS = [
    "Clear", "Partly Cloudy", "Cloudy", "Light Rain", 
    "Rain", "Thunderstorm", "Snow", "Fog"
]

@app.get("/")
async def root():
    """Welcome endpoint"""
    return {
        "message": "Weather Forecast API",
        "endpoints": {
            "/weather/current/{city}": "Get current weather",
            "/weather/forecast/{city}": "Get 7-day forecast",
            "/locations": "List available cities",
            "/docs": "API documentation"
        },
        "pricing": {
            "current_weather": "Free (1000 calls/day)",
            "forecast": "$0.001 per call after 100 free calls"
        }
    }

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {"status": "healthy", "timestamp": datetime.utcnow()}

@app.get("/locations")
async def list_locations():
    """List available cities"""
    return {
        "locations": [
            {"id": key, "name": value["name"], "coordinates": {"lat": value["lat"], "lon": value["lon"]}}
            for key, value in CITIES.items()
        ]
    }

@app.get("/weather/current/{city}", response_model=WeatherResponse)
# In production: @app.monetize(free_calls=1000, price_per_call=0)
async def get_current_weather(city: str):
    """
    Get current weather for a city
    
    This endpoint is FREE up to 1000 calls per day
    """
    if city not in CITIES:
        raise HTTPException(status_code=404, detail=f"City '{city}' not found")
    
    city_data = CITIES[city]
    
    # Generate mock weather data
    temp = random.uniform(0, 35)
    return WeatherResponse(
        location=city_data["name"],
        temperature=round(temp, 1),
        feels_like=round(temp + random.uniform(-3, 3), 1),
        humidity=random.randint(30, 90),
        description=random.choice(WEATHER_CONDITIONS),
        wind_speed=round(random.uniform(0, 30), 1),
        timestamp=datetime.utcnow()
    )

@app.get("/weather/forecast/{city}", response_model=ForecastResponse)
# In production: @app.monetize(free_calls=100, price_per_call=0.001)
# In production: @app.require_api_key()
async def get_weather_forecast(city: str, days: int = 7):
    """
    Get weather forecast for a city
    
    Premium endpoint: 100 free calls, then $0.001 per call
    Requires API key for access
    """
    if city not in CITIES:
        raise HTTPException(status_code=404, detail=f"City '{city}' not found")
    
    if days < 1 or days > 14:
        raise HTTPException(status_code=400, detail="Days must be between 1 and 14")
    
    city_data = CITIES[city]
    
    # Generate mock forecast
    forecast_days = []
    base_temp = random.uniform(10, 25)
    
    for i in range(days):
        date = f"2024-01-{20+i:02d}"
        high = round(base_temp + random.uniform(5, 10) + random.uniform(-3, 3), 1)
        low = round(base_temp - random.uniform(5, 10) + random.uniform(-3, 3), 1)
        
        forecast_days.append(ForecastDay(
            date=date,
            high=high,
            low=low,
            description=random.choice(WEATHER_CONDITIONS),
            precipitation_chance=random.randint(0, 100)
        ))
    
    return ForecastResponse(
        location=city_data["name"],
        days=forecast_days
    )

# API-Direct specific endpoints (simulated)
@app.get("/_apidirect/stats")
async def get_api_stats():
    """Get API usage statistics"""
    return {
        "total_calls": 15420,
        "endpoints": {
            "/weather/current/{city}": {
                "calls": 12500,
                "avg_response_time": 0.045,
                "errors": 12
            },
            "/weather/forecast/{city}": {
                "calls": 2920,
                "avg_response_time": 0.089,
                "errors": 3
            }
        },
        "revenue": {
            "today": 2.82,
            "this_month": 84.50,
            "total": 420.75
        }
    }

@app.post("/_apidirect/api-keys")
async def generate_api_key():
    """Generate a new API key"""
    import secrets
    return {
        "api_key": f"apidirect_{secrets.token_urlsafe(32)}",
        "created_at": datetime.utcnow(),
        "message": "API key created successfully. Use this in your X-API-Key header."
    }

if __name__ == "__main__":
    import uvicorn
    print("🌤️  Weather API Demo")
    print("====================")
    print("This demo shows how API-Direct makes it easy to:")
    print("✅ Create monetized APIs with simple decorators")
    print("✅ Automatic API key management")
    print("✅ Built-in usage tracking and billing")
    print("✅ One-command deployment")
    print("")
    print("In production, you would use @app.monetize() decorators")
    print("to automatically handle billing and API keys.")
    print("")
    print("Starting server at http://localhost:8002")
    print("Documentation at http://localhost:8002/docs")
    print("")
    uvicorn.run(app, host="0.0.0.0", port=8002)