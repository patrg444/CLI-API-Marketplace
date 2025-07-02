from fastapi import FastAPI

app = FastAPI(title="Test API")

@app.get("/")
def read_root():
    return {"message": "Hello from hosted deployment\!"}

@app.get("/health")
def health_check():
    return {"status": "healthy"}

@app.get("/api/users")
def get_users():
    return {"users": [{"id": 1, "name": "Test User"}]}
EOF < /dev/null