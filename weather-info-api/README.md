# Weather Information API

A simple weather information API built with API-Direct that provides mock weather data for various cities around the world.

## Features

- **Health Check**: Monitor API status
- **City Listings**: Get available cities with optional country filtering
- **Current Weather**: Get current weather conditions for any supported city
- **Weather Forecasts**: Get multi-day weather forecasts

## API Endpoints

### Health Check
```
GET /health
```
Returns API health status and version information.

### List Cities
```
GET /cities
GET /cities?country=USA
```
Returns a list of available cities. Optionally filter by country.

### Current Weather
```
GET /weather/{city}
GET /weather/{city}?units=fahrenheit
```
Get current weather for a specific city. Supports both Celsius (default) and Fahrenheit units.

**Supported Cities:**
- `new-york` - New York, USA
- `london` - London, UK
- `tokyo` - Tokyo, Japan
- `paris` - Paris, France
- `sydney` - Sydney, Australia
- `san-francisco` - San Francisco, USA

### Weather Forecast
```
POST /forecast
Content-Type: application/json

{
  "city": "london",
  "days": 7
}
```
Get weather forecast for a specified city and number of days (1-14).

## Example Responses

### Health Check
```json
{
  "status": "healthy",
  "service": "weather-info-api",
  "version": "1.0.0",
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

### Current Weather
```json
{
  "city": {
    "name": "London",
    "country": "UK",
    "timezone": "Europe/London"
  },
  "weather": {
    "temperature": 15,
    "temperature_unit": "°C",
    "condition": "partly-cloudy",
    "humidity": 65,
    "wind_speed": 12,
    "pressure": 1013,
    "visibility": 15,
    "uv_index": 4,
    "last_updated": "2024-01-15T10:30:00.000Z"
  },
  "units": "celsius"
}
```

### Weather Forecast
```json
{
  "city": {
    "name": "London",
    "country": "UK",
    "timezone": "Europe/London"
  },
  "forecast": [
    {
      "date": "2024-01-15",
      "day_of_week": "Monday",
      "weather": {
        "temperature": 15,
        "condition": "partly-cloudy",
        "humidity": 65,
        "wind_speed": 12,
        "pressure": 1013,
        "visibility": 15,
        "uv_index": 4,
        "last_updated": "2024-01-15T10:30:00.000Z"
      }
    }
  ],
  "days_requested": 7,
  "generated_at": "2024-01-15T10:30:00.000Z"
}
```

## Getting Started

### Prerequisites
- Python 3.9+
- API-Direct CLI

### Local Development

1. **Clone or create the project**:
   ```bash
   apidirect init weather-info-api --runtime python3.9
   cd weather-info-api
   ```

2. **Test locally** (when local testing is available):
   ```bash
   apidirect run
   ```

3. **Deploy to API-Direct**:
   ```bash
   apidirect deploy
   ```

4. **Publish to marketplace**:
   ```bash
   apidirect publish weather-info-api \
     --description "Simple weather information API with mock data" \
     --category "Weather" \
     --tags "weather,forecast,cities,demo"
   ```

## Testing the API

Once deployed, you can test the API endpoints:

```bash
# Health check
curl https://your-api-endpoint/health

# List all cities
curl https://your-api-endpoint/cities

# Get weather for London
curl https://your-api-endpoint/weather/london

# Get weather in Fahrenheit
curl https://your-api-endpoint/weather/new-york?units=fahrenheit

# Get 3-day forecast for Tokyo
curl -X POST https://your-api-endpoint/forecast \
  -H "Content-Type: application/json" \
  -d '{"city": "tokyo", "days": 3}'
```

## Error Handling

The API returns appropriate HTTP status codes:

- `200` - Success
- `400` - Bad Request (invalid parameters)
- `404` - Not Found (city not supported)
- `500` - Internal Server Error

Error responses include descriptive messages:

```json
{
  "error": "City \"invalid-city\" not found",
  "available_cities": ["new-york", "london", "tokyo", "paris", "sydney", "san-francisco"]
}
```

## Data Notes

This API uses mock weather data for demonstration purposes. The weather conditions are randomly generated and do not represent real weather information.

## Environment Variables

- `LOG_LEVEL` - Logging level (default: INFO)
- `API_VERSION` - API version (default: 1.0.0)

## Support

For questions or issues:
- API-Direct Documentation: https://docs.api-direct.io
- Support: support@api-direct.io

## License

MIT License
