"""
Tests for the Weather Information API handlers
"""
import json
import unittest
from unittest.mock import patch
import sys
import os

# Add the parent directory to the path so we can import main
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from main import health_check, get_cities, get_weather, get_forecast


class TestWeatherAPI(unittest.TestCase):
    
    def test_health_check(self):
        """Test the health check endpoint"""
        event = {}
        context = {}
        
        response = health_check(event, context)
        
        self.assertEqual(response['statusCode'], 200)
        self.assertEqual(response['headers']['Content-Type'], 'application/json')
        
        body = json.loads(response['body'])
        self.assertEqual(body['status'], 'healthy')
        self.assertEqual(body['service'], 'weather-info-api')
        self.assertIn('version', body)
        self.assertIn('timestamp', body)
    
    def test_get_cities_all(self):
        """Test getting all cities"""
        event = {}
        context = {}
        
        response = get_cities(event, context)
        
        self.assertEqual(response['statusCode'], 200)
        body = json.loads(response['body'])
        
        self.assertIn('cities', body)
        self.assertIn('total', body)
        self.assertEqual(body['total'], 6)  # We have 6 cities defined
        self.assertIsNone(body['filtered_by_country'])
        
        # Check that all expected cities are present
        city_ids = [city['id'] for city in body['cities']]
        expected_cities = ['new-york', 'london', 'tokyo', 'paris', 'sydney', 'san-francisco']
        for expected_city in expected_cities:
            self.assertIn(expected_city, city_ids)
    
    def test_get_cities_filtered_by_country(self):
        """Test getting cities filtered by country"""
        event = {
            'queryStringParameters': {
                'country': 'USA'
            }
        }
        context = {}
        
        response = get_cities(event, context)
        
        self.assertEqual(response['statusCode'], 200)
        body = json.loads(response['body'])
        
        self.assertEqual(body['total'], 2)  # New York and San Francisco
        self.assertEqual(body['filtered_by_country'], 'USA')
        
        city_names = [city['name'] for city in body['cities']]
        self.assertIn('New York', city_names)
        self.assertIn('San Francisco', city_names)
    
    def test_get_weather_valid_city(self):
        """Test getting weather for a valid city"""
        event = {
            'pathParameters': {
                'city': 'london'
            }
        }
        context = {}
        
        with patch('main.generate_mock_weather') as mock_weather:
            mock_weather.return_value = {
                'temperature': 15,
                'condition': 'sunny',
                'humidity': 60,
                'wind_speed': 10,
                'pressure': 1013,
                'visibility': 15,
                'uv_index': 5,
                'last_updated': '2024-01-15T10:00:00'
            }
            
            response = get_weather(event, context)
        
        self.assertEqual(response['statusCode'], 200)
        body = json.loads(response['body'])
        
        self.assertIn('city', body)
        self.assertIn('weather', body)
        self.assertEqual(body['city']['name'], 'London')
        self.assertEqual(body['weather']['temperature'], 15)
        self.assertEqual(body['weather']['temperature_unit'], '°C')
    
    def test_get_weather_fahrenheit(self):
        """Test getting weather with Fahrenheit units"""
        event = {
            'pathParameters': {
                'city': 'new-york'
            },
            'queryStringParameters': {
                'units': 'fahrenheit'
            }
        }
        context = {}
        
        with patch('main.generate_mock_weather') as mock_weather:
            mock_weather.return_value = {
                'temperature': 20,  # 20°C should become 68°F
                'condition': 'sunny',
                'humidity': 60,
                'wind_speed': 10,
                'pressure': 1013,
                'visibility': 15,
                'uv_index': 5,
                'last_updated': '2024-01-15T10:00:00'
            }
            
            response = get_weather(event, context)
        
        self.assertEqual(response['statusCode'], 200)
        body = json.loads(response['body'])
        
        self.assertEqual(body['weather']['temperature'], 68)  # 20°C = 68°F
        self.assertEqual(body['weather']['temperature_unit'], '°F')
        self.assertEqual(body['units'], 'fahrenheit')
    
    def test_get_weather_invalid_city(self):
        """Test getting weather for an invalid city"""
        event = {
            'pathParameters': {
                'city': 'invalid-city'
            }
        }
        context = {}
        
        response = get_weather(event, context)
        
        self.assertEqual(response['statusCode'], 404)
        body = json.loads(response['body'])
        
        self.assertIn('error', body)
        self.assertIn('available_cities', body)
        self.assertIn('invalid-city', body['error'])
    
    def test_get_weather_missing_city(self):
        """Test getting weather without city parameter"""
        event = {}
        context = {}
        
        response = get_weather(event, context)
        
        self.assertEqual(response['statusCode'], 400)
        body = json.loads(response['body'])
        
        self.assertIn('error', body)
        self.assertEqual(body['error'], 'City parameter is required')
    
    def test_get_forecast_valid_request(self):
        """Test getting forecast for a valid request"""
        event = {
            'body': json.dumps({
                'city': 'tokyo',
                'days': 3
            })
        }
        context = {}
        
        with patch('main.generate_mock_weather') as mock_weather:
            mock_weather.return_value = {
                'temperature': 25,
                'condition': 'sunny',
                'humidity': 70,
                'wind_speed': 8,
                'pressure': 1015,
                'visibility': 20,
                'uv_index': 7,
                'last_updated': '2024-01-15T10:00:00'
            }
            
            response = get_forecast(event, context)
        
        self.assertEqual(response['statusCode'], 200)
        body = json.loads(response['body'])
        
        self.assertIn('city', body)
        self.assertIn('forecast', body)
        self.assertEqual(body['days_requested'], 3)
        self.assertEqual(len(body['forecast']), 3)
        
        # Check forecast structure
        for day in body['forecast']:
            self.assertIn('date', day)
            self.assertIn('day_of_week', day)
            self.assertIn('weather', day)
    
    def test_get_forecast_default_days(self):
        """Test getting forecast with default number of days"""
        event = {
            'body': json.dumps({
                'city': 'paris'
            })
        }
        context = {}
        
        with patch('main.generate_mock_weather') as mock_weather:
            mock_weather.return_value = {
                'temperature': 18,
                'condition': 'cloudy',
                'humidity': 75,
                'wind_speed': 6,
                'pressure': 1010,
                'visibility': 12,
                'uv_index': 3,
                'last_updated': '2024-01-15T10:00:00'
            }
            
            response = get_forecast(event, context)
        
        self.assertEqual(response['statusCode'], 200)
        body = json.loads(response['body'])
        
        self.assertEqual(body['days_requested'], 5)  # Default is 5 days
        self.assertEqual(len(body['forecast']), 5)
    
    def test_get_forecast_invalid_city(self):
        """Test getting forecast for an invalid city"""
        event = {
            'body': json.dumps({
                'city': 'invalid-city',
                'days': 3
            })
        }
        context = {}
        
        response = get_forecast(event, context)
        
        self.assertEqual(response['statusCode'], 404)
        body = json.loads(response['body'])
        
        self.assertIn('error', body)
        self.assertIn('available_cities', body)
    
    def test_get_forecast_missing_city(self):
        """Test getting forecast without city"""
        event = {
            'body': json.dumps({
                'days': 3
            })
        }
        context = {}
        
        response = get_forecast(event, context)
        
        self.assertEqual(response['statusCode'], 400)
        body = json.loads(response['body'])
        
        self.assertEqual(body['error'], 'City is required in request body')
    
    def test_get_forecast_invalid_days(self):
        """Test getting forecast with invalid number of days"""
        event = {
            'body': json.dumps({
                'city': 'london',
                'days': 20  # Too many days
            })
        }
        context = {}
        
        response = get_forecast(event, context)
        
        self.assertEqual(response['statusCode'], 400)
        body = json.loads(response['body'])
        
        self.assertIn('Days must be an integer between 1 and 14', body['error'])
    
    def test_get_forecast_invalid_json(self):
        """Test getting forecast with invalid JSON"""
        event = {
            'body': 'invalid json'
        }
        context = {}
        
        response = get_forecast(event, context)
        
        self.assertEqual(response['statusCode'], 400)
        body = json.loads(response['body'])
        
        self.assertEqual(body['error'], 'Invalid JSON in request body')


if __name__ == '__main__':
    unittest.main()
