"""
Simple API test to verify basic setup
"""

import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

def test_import():
    """Test that we can import the main module"""
    try:
        from main import app
        assert app is not None
        print("Successfully imported app")
    except Exception as e:
        print(f"Import failed: {e}")
        raise

def test_health_endpoint():
    """Test the health check endpoint"""
    from main import app
    from fastapi.testclient import TestClient
    
    client = TestClient(app)
    response = client.get("/health")
    
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "healthy"
    assert "timestamp" in data
    assert "version" in data
    print(f"Health check response: {data}")

if __name__ == "__main__":
    test_import()
    test_health_endpoint()
    print("All tests passed!")