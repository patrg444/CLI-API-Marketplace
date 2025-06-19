"""
Example API using API-Direct Framework (FastAPI-compatible)
This demonstrates how to migrate from FastAPI to API-Direct with minimal changes
"""

from apidirect_framework import APIDirectFramework, APIDirectConfig, Query, Path, Body, BaseModel
from typing import Optional, List
import asyncio

# Create API-Direct app (FastAPI-compatible)
app = APIDirectFramework(
    title="E-commerce API",
    description="A sample e-commerce API built with API-Direct Framework",
    version="1.0.0",
    config=APIDirectConfig(
        enable_analytics=True,
        enable_billing=True,
        enable_rate_limiting=True,
        enable_api_keys=True,
        local_mode=True  # Set to False for production deployment
    )
)

# Pydantic models (same as FastAPI)
class Item(BaseModel):
    id: Optional[int] = None
    name: str
    description: Optional[str] = None
    price: float
    tax: Optional[float] = None

class User(BaseModel):
    id: Optional[int] = None
    username: str
    email: str
    full_name: Optional[str] = None

# In-memory storage for demo
items_db = []
users_db = []

# Basic endpoints (same syntax as FastAPI)
@app.get("/")
async def root():
    return {"message": "Welcome to API-Direct E-commerce API", "framework": "API-Direct"}

@app.get("/health")
async def health_check():
    return {"status": "healthy", "service": "ecommerce-api"}

# Free endpoints
@app.get("/items")
async def list_items(skip: int = Query(0), limit: int = Query(10)):
    """List all items (free endpoint)"""
    return {"items": items_db[skip:skip + limit], "total": len(items_db)}

@app.get("/items/{item_id}")
async def get_item(item_id: int = Path(..., description="The ID of the item")):
    """Get a specific item (free endpoint)"""
    for item in items_db:
        if item.get("id") == item_id:
            return item
    return {"error": "Item not found"}, 404

# Monetized endpoints with API-Direct decorators
@app.post("/items")
@app.monetize(free_calls=100, price_per_call=0.01)  # $0.01 per call after 100 free
@app.require_api_key()
async def create_item(item: Item):
    """Create a new item (monetized endpoint)"""
    item_dict = item.dict()
    item_dict["id"] = len(items_db) + 1
    items_db.append(item_dict)
    return {"message": "Item created", "item": item_dict}

@app.put("/items/{item_id}")
@app.monetize(free_calls=50, price_per_call=0.02)  # $0.02 per call after 50 free
@app.require_api_key()
async def update_item(item_id: int, item: Item):
    """Update an item (monetized endpoint)"""
    for i, existing_item in enumerate(items_db):
        if existing_item.get("id") == item_id:
            item_dict = item.dict()
            item_dict["id"] = item_id
            items_db[i] = item_dict
            return {"message": "Item updated", "item": item_dict}
    return {"error": "Item not found"}, 404

@app.delete("/items/{item_id}")
@app.monetize(free_calls=25, price_per_call=0.05)  # $0.05 per call after 25 free
@app.require_api_key()
async def delete_item(item_id: int):
    """Delete an item (monetized endpoint)"""
    for i, item in enumerate(items_db):
        if item.get("id") == item_id:
            deleted_item = items_db.pop(i)
            return {"message": "Item deleted", "item": deleted_item}
    return {"error": "Item not found"}, 404

# Premium analytics endpoint
@app.get("/analytics/items")
@app.monetize(free_calls=10, price_per_call=0.10)  # $0.10 per call after 10 free
@app.rate_limit(calls_per_minute=30)
@app.require_api_key()
async def get_item_analytics():
    """Get item analytics (premium endpoint)"""
    total_items = len(items_db)
    avg_price = sum(item.get("price", 0) for item in items_db) / max(total_items, 1)
    
    return {
        "total_items": total_items,
        "average_price": round(avg_price, 2),
        "price_ranges": {
            "under_10": len([i for i in items_db if i.get("price", 0) < 10]),
            "10_to_50": len([i for i in items_db if 10 <= i.get("price", 0) < 50]),
            "over_50": len([i for i in items_db if i.get("price", 0) >= 50])
        }
    }

# User management endpoints
@app.post("/users")
@app.monetize(free_calls=200, price_per_call=0.005)
async def create_user(user: User):
    """Create a new user"""
    user_dict = user.dict()
    user_dict["id"] = len(users_db) + 1
    users_db.append(user_dict)
    return {"message": "User created", "user": user_dict}

@app.get("/users/{user_id}")
async def get_user(user_id: int):
    """Get user information (free endpoint)"""
    for user in users_db:
        if user.get("id") == user_id:
            return user
    return {"error": "User not found"}, 404

# Batch operations (higher pricing)
@app.post("/items/batch")
@app.monetize(free_calls=5, price_per_call=0.25)  # $0.25 per batch operation
@app.rate_limit(calls_per_minute=10)
@app.require_api_key()
async def create_items_batch(items: List[Item]):
    """Create multiple items in batch (premium endpoint)"""
    created_items = []
    for item in items:
        item_dict = item.dict()
        item_dict["id"] = len(items_db) + 1
        items_db.append(item_dict)
        created_items.append(item_dict)
    
    return {
        "message": f"Created {len(created_items)} items",
        "items": created_items
    }

# Search endpoint with different pricing tiers
@app.get("/search/items")
@app.monetize(free_calls=500, price_per_call=0.002)  # Cheap search
async def search_items(q: str = Query(..., description="Search query")):
    """Search items by name or description"""
    results = []
    query_lower = q.lower()
    
    for item in items_db:
        if (query_lower in item.get("name", "").lower() or 
            query_lower in item.get("description", "").lower()):
            results.append(item)
    
    return {"query": q, "results": results, "count": len(results)}

# Add some sample data on startup
@app.get("/setup")
async def setup_sample_data():
    """Setup sample data for testing"""
    global items_db, users_db
    
    # Sample items
    sample_items = [
        {"id": 1, "name": "Laptop", "description": "High-performance laptop", "price": 999.99, "tax": 99.99},
        {"id": 2, "name": "Mouse", "description": "Wireless mouse", "price": 29.99, "tax": 3.00},
        {"id": 3, "name": "Keyboard", "description": "Mechanical keyboard", "price": 79.99, "tax": 8.00},
        {"id": 4, "name": "Monitor", "description": "4K monitor", "price": 299.99, "tax": 30.00},
        {"id": 5, "name": "Headphones", "description": "Noise-canceling headphones", "price": 199.99, "tax": 20.00}
    ]
    
    # Sample users
    sample_users = [
        {"id": 1, "username": "john_doe", "email": "john@example.com", "full_name": "John Doe"},
        {"id": 2, "username": "jane_smith", "email": "jane@example.com", "full_name": "Jane Smith"}
    ]
    
    items_db.extend(sample_items)
    users_db.extend(sample_users)
    
    return {
        "message": "Sample data created",
        "items_count": len(items_db),
        "users_count": len(users_db)
    }

if __name__ == "__main__":
    # Generate API-Direct configuration
    app.save_apidirect_config("apidirect.yaml")
    
    # Run the development server
    print("\n" + "="*60)
    print("🚀 API-Direct Framework Demo")
    print("="*60)
    print("This API demonstrates FastAPI compatibility with API-Direct features:")
    print("• 💰 Built-in monetization with usage-based pricing")
    print("• 🔑 API key management and validation")
    print("• 📊 Real-time analytics and usage tracking")
    print("• ⚡ Rate limiting and performance monitoring")
    print("• 🎯 Easy deployment to API-Direct marketplace")
    print("\nTry these endpoints:")
    print("• GET  /setup - Create sample data")
    print("• GET  /items - List items (free)")
    print("• POST /items - Create item (monetized)")
    print("• GET  /analytics/items - Analytics (premium)")
    print("• POST /_apidirect/api-keys - Generate API key")
    print("• GET  /_apidirect/stats - View usage statistics")
    print("="*60)
    
    app.run(host="0.0.0.0", port=8000, reload=True)
