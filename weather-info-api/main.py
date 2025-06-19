"""
Weather Information API
A simple weather API demonstrating API-Direct functionality
"""
import json
import logging
import os
import random
from datetime import datetime, timedelta
from typing import Dict, Any, List

# Configure logging
logging.basicConfig(level=os.environ.get('LOG_LEVEL', 'INFO'))
logger = logging.getLogger(__name__)

# Mock weather data
CITIES = {
    "new-york": {
        "name": "New York",
        "country": "USA",
        "timezone": "America/New_York"
    },
    "london": {
        "name": "London", 
        "country": "UK",
        "timezone": "Europe/London"
    },
    "tokyo": {
        "name": "Tokyo",
        "country": "Japan", 
        "timezone": "Asia/Tokyo"
    },
    "paris": {
        "name": "Paris",
        "country": "France",
        "timezone": "Europe/Paris"
    },
    "sydney": {
        "name": "Sydney",
        "country": "Australia",
        "timezone": "Australia/Sydney"
    },
    "san-francisco": {
        "name": "San Francisco",
        "country": "USA",
        "timezone": "America/Los_Angeles"
    }
}

WEATHER_CONDITIONS = [
    "sunny", "cloudy", "partly-cloudy", "rainy", "stormy", "snowy", "foggy"
]


def generate_mock_weather(city_name: str) -> Dict[str, Any]:
    """Generate mock weather data for a city"""
    return {
        "temperature": random.randint(-10, 35),
        "condition": random.choice(WEATHER_CONDITIONS),
        "humidity": random.randint(30, 90),
        "wind_speed": random.randint(0, 25),
        "pressure": random.randint(980, 1030),
        "visibility": random.randint(5, 20),
        "uv_index": random.randint(1, 11),
        "last_updated": datetime.now().isoformat()
    }


def health_check(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Health check endpoint
    """
    logger.info("Health check endpoint called")
    
    return {
        'statusCode': 200,
        'headers': {
            'Content-Type': 'application/json'
        },
        'body': json.dumps({
            'status': 'healthy',
            'service': 'weather-info-api',
            'version': os.environ.get('API_VERSION', '1.0.0'),
            'timestamp': datetime.now().isoformat()
        })
    }


def get_cities(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Get list of available cities
    """
    logger.info("Get cities endpoint called")
    
    # Check for query parameters
    query_params = event.get('queryStringParameters') or {}
    country_filter = query_params.get('country')
    
    cities_list = []
    for city_id, city_info in CITIES.items():
        if country_filter and city_info['country'].lower() != country_filter.lower():
            continue
            
        cities_list.append({
            'id': city_id,
            'name': city_info['name'],
            'country': city_info['country'],
            'timezone': city_info['timezone']
        })
    
    return {
        'statusCode': 200,
        'headers': {
            'Content-Type': 'application/json'
        },
        'body': json.dumps({
            'cities': cities_list,
            'total': len(cities_list),
            'filtered_by_country': country_filter
        })
    }


def get_weather(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Get current weather for a specific city
    """
    path_params = event.get('pathParameters', {})
    city_id = path_params.get('city', '').lower()
    
    logger.info(f"Get weather endpoint called for city: {city_id}")
    
    if not city_id:
        return {
            'statusCode': 400,
            'headers': {
                'Content-Type': 'application/json'
            },
            'body': json.dumps({
                'error': 'City parameter is required'
            })
        }
    
    if city_id not in CITIES:
        return {
            'statusCode': 404,
            'headers': {
                'Content-Type': 'application/json'
            },
            'body': json.dumps({
                'error': f'City "{city_id}" not found',
                'available_cities': list(CITIES.keys())
            })
        }
    
    city_info = CITIES[city_id]
    weather_data = generate_mock_weather(city_info['name'])
    
    # Check for units parameter
    query_params = event.get('queryStringParameters') or {}
    units = query_params.get('units', 'celsius')
    
    if units.lower() == 'fahrenheit':
        weather_data['temperature'] = (weather_data['temperature'] * 9/5) + 32
        weather_data['temperature_unit'] = '°F'
    else:
        weather_data['temperature_unit'] = '°C'
    
    response_data = {
        'city': city_info,
        'weather': weather_data,
        'units': units
    }
    
    return {
        'statusCode': 200,
        'headers': {
            'Content-Type': 'application/json'
        },
        'body': json.dumps(response_data)
    }


def get_forecast(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Get weather forecast for a city
    """
    try:
        # Parse request body
        body = json.loads(event.get('body', '{}'))
        logger.info(f"Get forecast endpoint called with data: {body}")
        
        city_id = body.get('city', '').lower()
        days = body.get('days', 5)
        
        if not city_id:
            return {
                'statusCode': 400,
                'headers': {
                    'Content-Type': 'application/json'
                },
                'body': json.dumps({
                    'error': 'City is required in request body'
                })
            }
        
        if city_id not in CITIES:
            return {
                'statusCode': 404,
                'headers': {
                    'Content-Type': 'application/json'
                },
                'body': json.dumps({
                    'error': f'City "{city_id}" not found',
                    'available_cities': list(CITIES.keys())
                })
            }
        
        if not isinstance(days, int) or days < 1 or days > 14:
            return {
                'statusCode': 400,
                'headers': {
                    'Content-Type': 'application/json'
                },
                'body': json.dumps({
                    'error': 'Days must be an integer between 1 and 14'
                })
            }
        
        city_info = CITIES[city_id]
        
        # Generate forecast data
        forecast = []
        for i in range(days):
            date = datetime.now() + timedelta(days=i)
            weather = generate_mock_weather(city_info['name'])
            
            forecast.append({
                'date': date.strftime('%Y-%m-%d'),
                'day_of_week': date.strftime('%A'),
                'weather': weather
            })
        
        response_data = {
            'city': city_info,
            'forecast': forecast,
            'days_requested': days,
            'generated_at': datetime.now().isoformat()
        }
        
        return {
            'statusCode': 200,
            'headers': {
                'Content-Type': 'application/json'
            },
            'body': json.dumps(response_data)
        }
        
    except json.JSONDecodeError as e:
        logger.error(f"JSON decode error: {e}")
        return {
            'statusCode': 400,
            'headers': {
                'Content-Type': 'application/json'
            },
            'body': json.dumps({
                'error': 'Invalid JSON in request body'
            })
        }
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
        return {
            'statusCode': 500,
            'headers': {
                'Content-Type': 'application/json'
            },
            'body': json.dumps({
                'error': 'Internal server error'
            })
        }
