"""
Simple auth test to check actual responses
"""

import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from fastapi.testclient import TestClient
from main import app

client = TestClient(app)

def test_register_endpoint():
    """Test what the register endpoint actually returns"""
    response = client.post("/auth/register", json={
        "email": "test@example.com",
        "password": "SecurePassword123!",
        "name": "Test User",
        "company": "Test Company"
    })
    
    print(f"Status code: {response.status_code}")
    print(f"Response: {response.json()}")
    
    # Check what we got
    if response.status_code == 200:
        data = response.json()
        print(f"Response keys: {list(data.keys())}")

if __name__ == "__main__":
    test_register_endpoint()