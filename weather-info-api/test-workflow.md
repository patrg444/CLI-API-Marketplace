# Weather Info API - Complete Workflow Test

This document demonstrates the complete workflow for creating, deploying, and publishing an API using the API-Direct platform.

## Prerequisites

1. **API-Direct CLI installed and configured**
2. **Authentication set up** (AWS Cognito credentials)
3. **Platform services running** (Storage, Deployment, Marketplace, etc.)

## Step-by-Step Workflow

### 1. Initialize the API Project

```bash
# This would normally be done with:
# apidirect init weather-info-api --runtime python3.9

# But since we've already created it manually, we can skip this step
cd weather-info-api
```

### 2. Review the Project Structure

```bash
# Check the project files
ls -la

# Review the API configuration
cat apidirect.yaml

# Review the main implementation
head -20 main.py
```

### 3. Run Tests Locally

```bash
# Run the unit tests to ensure everything works
python -m pytest tests/ -v

# Expected output: All 13 tests should pass
```

### 4. Deploy the API

```bash
# Deploy the API to the API-Direct platform
apidirect deploy

# This will:
# - Package the code
# - Upload to storage service
# - Create Kubernetes deployment
# - Set up routing through the gateway
# - Return the API endpoint URL
```

### 5. Test the Deployed API

Once deployed, test the endpoints:

```bash
# Replace YOUR_API_ENDPOINT with the actual endpoint from deployment
API_ENDPOINT="https://your-api-endpoint"

# Test health check
curl $API_ENDPOINT/health

# Test cities endpoint
curl $API_ENDPOINT/cities

# Test cities with country filter
curl "$API_ENDPOINT/cities?country=USA"

# Test weather for a specific city
curl $API_ENDPOINT/weather/london

# Test weather with Fahrenheit units
curl "$API_ENDPOINT/weather/new-york?units=fahrenheit"

# Test forecast endpoint
curl -X POST $API_ENDPOINT/forecast \
  -H "Content-Type: application/json" \
  -d '{"city": "tokyo", "days": 3}'
```

### 6. Publish to Marketplace

```bash
# Publish the API to the marketplace with metadata
apidirect publish weather-info-api \
  --description "Simple weather information API with mock data for testing and demonstration" \
  --category "Weather" \
  --tags "weather,forecast,cities,demo,testing"

# This will:
# - Update the marketplace database
# - Index the API in Elasticsearch for search
# - Make it discoverable in the marketplace
# - Return the marketplace URL
```

### 7. Verify Marketplace Listing

```bash
# Check marketplace listings
apidirect marketplace list

# Search for the API
apidirect marketplace search --query "weather"

# Get details about the published API
apidirect marketplace get weather-info-api
```

### 8. Manage the API

```bash
# View API logs
apidirect logs weather-info-api

# Update API settings
apidirect env set weather-info-api LOG_LEVEL=DEBUG

# Unpublish from marketplace (if needed)
apidirect unpublish weather-info-api

# Re-publish with updated information
apidirect publish weather-info-api \
  --description "Updated weather API with enhanced features" \
  --category "Weather" \
  --tags "weather,forecast,cities,demo,v2"
```

## Expected Results

### Health Check Response
```json
{
  "status": "healthy",
  "service": "weather-info-api",
  "version": "1.0.0",
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

### Cities Response
```json
{
  "cities": [
    {
      "id": "new-york",
      "name": "New York",
      "country": "USA",
      "timezone": "America/New_York"
    },
    {
      "id": "london",
      "name": "London",
      "country": "UK",
      "timezone": "Europe/London"
    }
  ],
  "total": 6,
  "filtered_by_country": null
}
```

### Weather Response
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

### Forecast Response
```json
{
  "city": {
    "name": "Tokyo",
    "country": "Japan",
    "timezone": "Asia/Tokyo"
  },
  "forecast": [
    {
      "date": "2024-01-15",
      "day_of_week": "Monday",
      "weather": {
        "temperature": 25,
        "condition": "sunny",
        "humidity": 70,
        "wind_speed": 8,
        "pressure": 1015,
        "visibility": 20,
        "uv_index": 7,
        "last_updated": "2024-01-15T10:00:00.000Z"
      }
    }
  ],
  "days_requested": 3,
  "generated_at": "2024-01-15T10:30:00.000Z"
}
```

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   ```bash
   # Re-authenticate with Cognito
   apidirect auth login
   ```

2. **Deployment Failures**
   ```bash
   # Check deployment logs
   apidirect logs weather-info-api --deployment
   
   # Verify configuration
   cat apidirect.yaml
   ```

3. **API Not Responding**
   ```bash
   # Check API status
   apidirect status weather-info-api
   
   # View runtime logs
   apidirect logs weather-info-api --runtime
   ```

4. **Marketplace Issues**
   ```bash
   # Check marketplace service status
   apidirect marketplace status
   
   # Re-index the API
   apidirect publish weather-info-api --force
   ```

## Performance Testing

```bash
# Load test the health endpoint
for i in {1..10}; do
  curl -s $API_ENDPOINT/health | jq '.status'
done

# Test multiple cities
for city in london tokyo paris; do
  echo "Testing $city:"
  curl -s $API_ENDPOINT/weather/$city | jq '.city.name, .weather.condition'
done
```

## Cleanup

```bash
# Unpublish from marketplace
apidirect unpublish weather-info-api

# Delete the deployment
apidirect delete weather-info-api

# Remove local files (if desired)
cd ..
rm -rf weather-info-api
```

## Summary

This workflow demonstrates:

1. ✅ **API Creation**: Complete project structure with configuration, code, tests, and documentation
2. ✅ **Local Testing**: Unit tests covering all endpoints and error cases
3. ✅ **Deployment**: Integration with storage, deployment, and gateway services
4. ✅ **API Testing**: Comprehensive endpoint testing with various parameters
5. ✅ **Marketplace Publishing**: Publishing with metadata, search indexing, and discoverability
6. ✅ **Management**: Logging, monitoring, and lifecycle management

The Weather Info API serves as a perfect example of the complete API-Direct workflow, showcasing all the platform's capabilities from development to marketplace publication.
