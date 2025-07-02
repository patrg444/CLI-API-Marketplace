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
