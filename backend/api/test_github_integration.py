#!/usr/bin/env python3
"""
Test script for GitHub integration
"""
import asyncio
import os
from github_integration import github_integration

async def test_detector():
    """Test the detector on a sample repository"""
    print("Testing GitHub integration with detector...")
    
    # Test with a public repository
    test_repo = "https://github.com/tiangolo/fastapi-example.git"
    
    try:
        print(f"\nAnalyzing repository: {test_repo}")
        analysis = await github_integration.clone_and_analyze_repo(test_repo)
        
        print("\n=== Analysis Results ===")
        print(f"Language: {analysis['language']}")
        print(f"Runtime: {analysis['runtime']}")
        print(f"Framework: {analysis['framework']}")
        print(f"Main file: {analysis['main_file']}")
        print(f"Start command: {analysis['start_command']}")
        print(f"Port: {analysis['port']}")
        print(f"Health check: {analysis['health_check']}")
        
        print("\nEndpoints detected:")
        for endpoint in analysis['endpoints']:
            print(f"  {endpoint['method']} {endpoint['path']}")
        
        print("\nEnvironment variables:")
        print(f"  Required: {analysis['environment']['required']}")
        print(f"  Optional: {analysis['environment']['optional']}")
        
    except Exception as e:
        print(f"Error: {e}")

async def test_basic_detection():
    """Test basic repository detection"""
    print("\nTesting basic repository detection...")
    
    # Create a temporary test directory
    import tempfile
    import json
    
    with tempfile.TemporaryDirectory() as temp_dir:
        # Create a simple Python FastAPI project
        main_py = os.path.join(temp_dir, "main.py")
        with open(main_py, "w") as f:
            f.write('''from fastapi import FastAPI

app = FastAPI()

@app.get("/")
def read_root():
    return {"Hello": "World"}

@app.get("/items/{item_id}")
def read_item(item_id: int, q: str = None):
    return {"item_id": item_id, "q": q}

@app.post("/items/")
def create_item(item: dict):
    return item

@app.get("/health")
def health_check():
    return {"status": "healthy"}
''')
        
        # Create requirements.txt
        req_txt = os.path.join(temp_dir, "requirements.txt")
        with open(req_txt, "w") as f:
            f.write("fastapi==0.104.1\nuvicorn==0.24.0\n")
        
        # Create .env.example
        env_example = os.path.join(temp_dir, ".env.example")
        with open(env_example, "w") as f:
            f.write("API_KEY=your-api-key-here\nPORT=8080\nDATABASE_URL=\n")
        
        # Run detector
        analysis = await github_integration.run_detector_analysis(temp_dir)
        
        print("\n=== Basic Detection Results ===")
        print(json.dumps(analysis, indent=2))

if __name__ == "__main__":
    print("GitHub Integration Test Script")
    print("==============================")
    
    # Run tests
    asyncio.run(test_basic_detection())
    
    # Uncomment to test with real GitHub repo
    # asyncio.run(test_detector())