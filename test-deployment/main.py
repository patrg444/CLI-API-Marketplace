from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List
import os

app = FastAPI(
    title="E2E Test API",
    description="API for testing BYOA deployment",
    version="1.0.0"
)

# Models
class User(BaseModel):
    id: int
    name: str
    email: str

# In-memory database
users_db = [
    User(id=1, name="Test User", email="test@example.com"),
]

@app.get("/")
def read_root():
    return {
        "message": "Hello from BYOA E2E Test!",
        "environment": os.getenv("ENVIRONMENT", "development")
    }

@app.get("/health")
def health_check():
    return {"status": "healthy"}

@app.get("/api/users", response_model=List[User])
def get_users():
    return users_db

@app.post("/api/users", response_model=User)
def create_user(user: User):
    users_db.append(user)
    return user