#!/usr/bin/env python3
"""
Test the deployment functionality locally
"""
import asyncio
import requests
import os

# Set environment to use mock auth
os.environ['USE_MOCK_AUTH'] = 'true'

BASE_URL = "http://localhost:8000"

async def test_deployment():
    # First register a user
    register_data = {
        "name": "Test User",
        "email": "test@example.com",
        "password": "testpass123"
    }
    
    print("1. Registering user...")
    response = requests.post(f"{BASE_URL}/auth/register", json=register_data)
    if response.status_code == 200:
        auth_data = response.json()
        token = auth_data['access_token']
        print(f"✓ User registered, token: {token[:20]}...")
    else:
        print(f"✗ Registration failed: {response.text}")
        return
    
    # Create API key
    headers = {"Authorization": f"Bearer {token}"}
    
    print("\n2. Creating API key...")
    api_key_data = {"name": "Test API Key"}
    response = requests.post(f"{BASE_URL}/api-keys", json=api_key_data, headers=headers)
    if response.status_code == 200:
        api_key = response.json()['key']
        print(f"✓ API key created: {api_key}")
    else:
        print(f"✗ API key creation failed: {response.text}")
        return
    
    # Test deployment endpoint
    print("\n3. Testing deployment endpoint...")
    deployment_data = {
        "api_name": "test-api",
        "runtime": "python3.9"
    }
    
    headers = {"X-API-Key": api_key}
    response = requests.post(f"{BASE_URL}/api/deploy", json=deployment_data, headers=headers)
    
    if response.status_code == 200:
        deployment = response.json()
        print(f"✓ Deployment initiated:")
        print(f"  - Deployment ID: {deployment['deployment_id']}")
        print(f"  - Status: {deployment['status']}")
        print(f"  - Endpoint: {deployment['endpoint']}")
    else:
        print(f"✗ Deployment failed: {response.text}")
        return
    
    # List deployments
    print("\n4. Listing deployments...")
    response = requests.get(f"{BASE_URL}/api/deployments", headers=headers)
    if response.status_code == 200:
        deployments = response.json()['deployments']
        print(f"✓ Found {len(deployments)} deployments:")
        for d in deployments:
            print(f"  - {d['api_id']}: {d['status']} ({d['id']})")
    else:
        print(f"✗ Failed to list deployments: {response.text}")

if __name__ == "__main__":
    print("Testing API Direct Deployment Functionality")
    print("=" * 50)
    asyncio.run(test_deployment())