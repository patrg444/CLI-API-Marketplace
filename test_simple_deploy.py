#!/usr/bin/env python3
"""Test simple deployment"""
import os
import requests

os.environ['USE_MOCK_AUTH'] = 'true'

# Get an API key first
response = requests.post("http://localhost:8000/auth/register", json={
    "name": "Test",
    "email": f"test{os.getpid()}@example.com",  # Unique email per run
    "password": "test123"
})

if response.status_code != 200:
    print(f"Registration failed: {response.text}")
    exit(1)
    
token = response.json()['access_token']

# Create API key
response = requests.post("http://localhost:8000/api-keys", 
                        json={"name": "test"}, 
                        headers={"Authorization": f"Bearer {token}"})

api_key = response.json()['key']

# Try to deploy
response = requests.post("http://localhost:8000/api/deploy",
                        json={"api_name": "test-api", "runtime": "python3.9"},
                        headers={"X-API-Key": api_key})

print(f"Status: {response.status_code}")
print(f"Response: {response.text}")