# API-Direct Framework for Python (FastAPI-Compatible)

A drop-in replacement for FastAPI that adds built-in monetization, analytics, and deployment capabilities to your APIs.

## 🚀 Why API-Direct Framework?

### FastAPI vs API-Direct Framework

| Feature | FastAPI | API-Direct Framework |
|---------|---------|---------------------|
| **API Development** | ✅ Fast, modern Python API framework | ✅ 100% FastAPI-compatible syntax |
| **Documentation** | ✅ Automatic OpenAPI/Swagger docs | ✅ Enhanced docs with pricing info |
| **Performance** | ✅ High performance with async support | ✅ Same performance + monitoring |
| **Monetization** | ❌ Manual implementation required | ✅ Built-in with decorators |
| **API Keys** | ❌ Custom middleware needed | ✅ Automatic generation & validation |
| **Rate Limiting** | ❌ External service required | ✅ Built-in per-endpoint limits |
| **Analytics** | ❌ Custom tracking needed | ✅ Real-time usage analytics |
| **Billing** | ❌ Stripe integration from scratch | ✅ Automatic usage-based billing |
| **Deployment** | ❌ Manual Docker/K8s setup | ✅ One-command deployment |
| **Marketplace** | ❌ No distribution platform | ✅ Instant marketplace listing |

## 🎯 Key Advantages Over FastAPI

### 1. **Zero-Code Monetization**
```python
# FastAPI - No built-in monetization
@app.post("/premium-feature")
async def premium_feature():
    # You need to implement billing, API keys, rate limiting manually
    return {"data": "premium"}

# API-Direct - One decorator adds everything
@app.post("/premium-feature")
@app.monetize(free_calls=100, price_per_call=0.01)
@app.require_api_key()
async def premium_feature():
    return {"data": "premium"}
```

### 2. **Instant Analytics**
```python
# FastAPI - Manual tracking
@app.get("/data")
async def get_data():
    # You need to implement usage tracking
    track_usage("get_data")  # Custom function you'd need to write
    return {"data": "value"}

# API-Direct - Automatic tracking
@app.get("/data")
@app.track_usage()  # Automatic analytics
async def get_data():
    return {"data": "value"}
```

### 3. **Built-in API Key Management**
```python
# FastAPI - Custom middleware
from fastapi import HTTPException, Depends
from fastapi.security import HTTPBearer

security = HTTPBearer()

async def verify_api_key(token: str = Depends(security)):
    # You need to implement API key validation
    if not validate_key(token):  # Custom function
        raise HTTPException(status_code=401)
    return token

@app.get("/protected")
async def protected(api_key: str = Depends(verify_api_key)):
    return {"data": "protected"}

# API-Direct - One decorator
@app.get("/protected")
@app.require_api_key()  # Automatic validation
async def protected():
    return {"data": "protected"}
```

### 4. **One-Command Deployment**
```bash
# FastAPI - Manual deployment
docker build -t my-api .
docker push my-registry/my-api
kubectl apply -f k8s-manifests/
# Set up ingress, SSL, monitoring, etc.

# API-Direct - One command
apidirect deploy
```

## 📦 Installation

```bash
pip install -r requirements.txt
```

## 🏃‍♂️ Quick Start

### 1. Basic Migration from FastAPI

```python
# Before (FastAPI)
from fastapi import FastAPI

app = FastAPI()

@app.get("/")
async def root():
    return {"message": "Hello World"}

# After (API-Direct) - Same syntax!
from apidirect_framework import APIDirectFramework

app = APIDirectFramework()

@app.get("/")
async def root():
    return {"message": "Hello World"}
```

### 2. Add Monetization

```python
from apidirect_framework import APIDirectFramework, BaseModel

app = APIDirectFramework(
    title="My Monetized API",
    description="API with built-in billing"
)

class Item(BaseModel):
    name: str
    price: float

# Free endpoint
@app.get("/items")
async def list_items():
    return {"items": []}

# Monetized endpoint
@app.post("/items")
@app.monetize(free_calls=100, price_per_call=0.01)
@app.require_api_key()
async def create_item(item: Item):
    return {"message": "Item created", "item": item}

# Premium analytics
@app.get("/analytics")
@app.monetize(free_calls=10, price_per_call=0.10)
@app.rate_limit(calls_per_minute=30)
@app.require_api_key()
async def get_analytics():
    return {"total_revenue": 1000, "api_calls": 5000}

if __name__ == "__main__":
    app.run()
```

### 3. Run Your API

```bash
python your_api.py
```

Visit:
- `http://localhost:8000/docs` - API documentation
- `http://localhost:8000/_apidirect/stats` - Usage analytics
- `POST http://localhost:8000/_apidirect/api-keys` - Generate API keys

## 🎨 Framework Features

### Monetization Decorators

```python
# Basic monetization
@app.monetize(free_calls=1000, price_per_call=0.001)

# API key requirement
@app.require_api_key()

# Rate limiting
@app.rate_limit(calls_per_minute=100)

# Usage tracking
@app.track_usage()

# Combine multiple decorators
@app.post("/premium")
@app.monetize(free_calls=50, price_per_call=0.05)
@app.rate_limit(calls_per_minute=10)
@app.require_api_key()
async def premium_endpoint():
    return {"data": "premium"}
```

### Configuration

```python
from apidirect_framework import APIDirectFramework, APIDirectConfig

config = APIDirectConfig(
    enable_analytics=True,
    enable_billing=True,
    enable_rate_limiting=True,
    enable_api_keys=True,
    local_mode=True  # False for production
)

app = APIDirectFramework(
    title="My API",
    config=config
)
```

### Built-in Endpoints

The framework automatically adds these endpoints:

- `GET /_apidirect/health` - Health check
- `GET /_apidirect/stats` - Usage statistics
- `POST /_apidirect/api-keys` - Generate API keys

## 🔧 Advanced Usage

### Custom Pricing Tiers

```python
# Different pricing for different endpoints
@app.get("/basic-search")
@app.monetize(free_calls=1000, price_per_call=0.001)
async def basic_search():
    return {"results": []}

@app.get("/ai-search")
@app.monetize(free_calls=100, price_per_call=0.01)
async def ai_search():
    return {"ai_results": []}

@app.get("/premium-analytics")
@app.monetize(free_calls=10, price_per_call=0.10)
async def premium_analytics():
    return {"insights": []}
```

### Batch Operations

```python
@app.post("/batch-process")
@app.monetize(free_calls=5, price_per_call=0.25)  # Higher price for batch
@app.rate_limit(calls_per_minute=10)
async def batch_process(items: List[Item]):
    return {"processed": len(items)}
```

### Generate API-Direct Configuration

```python
# Automatically generate apidirect.yaml
app.save_apidirect_config("apidirect.yaml")
```

## 🚀 Deployment

### Local Development

```bash
python your_api.py
```

### Production Deployment

1. **Generate configuration:**
```python
app.save_apidirect_config()
```

2. **Deploy with API-Direct CLI:**
```bash
apidirect deploy
```

3. **Or use Docker:**
```dockerfile
FROM python:3.9-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt

COPY . .
EXPOSE 8000

CMD ["python", "your_api.py"]
```

## 📊 Monitoring & Analytics

### View Usage Statistics

```bash
curl http://localhost:8000/_apidirect/stats
```

Response:
```json
{
  "/items": {
    "calls": 150,
    "avg_response_time": 0.045,
    "status_codes": {"200": 148, "404": 2}
  },
  "/analytics": {
    "calls": 25,
    "avg_response_time": 0.120,
    "status_codes": {"200": 25}
  }
}
```

### Generate API Keys

```bash
curl -X POST http://localhost:8000/_apidirect/api-keys
```

Response:
```json
{
  "api_key": "apidirect_abc123...",
  "message": "API key created successfully"
}
```

## 🔄 Migration Guide

### From FastAPI to API-Direct

1. **Replace import:**
```python
# Before
from fastapi import FastAPI
app = FastAPI()

# After
from apidirect_framework import APIDirectFramework
app = APIDirectFramework()
```

2. **Add monetization (optional):**
```python
@app.post("/endpoint")
@app.monetize(free_calls=100, price_per_call=0.01)
async def endpoint():
    return {"data": "value"}
```

3. **Everything else stays the same!**
   - Same route decorators (`@app.get`, `@app.post`, etc.)
   - Same Pydantic models
   - Same dependency injection
   - Same middleware support

## 🆚 Comparison with Other Solutions

### vs. FastAPI + Custom Billing

| Task | FastAPI + Custom | API-Direct |
|------|------------------|------------|
| Setup billing | 2-3 weeks | 1 decorator |
| API key management | 1 week | Built-in |
| Rate limiting | 3-5 days | 1 decorator |
| Analytics | 1-2 weeks | Automatic |
| Deployment | 2-3 days | 1 command |
| **Total Time** | **6-8 weeks** | **< 1 day** |

### vs. AWS API Gateway

| Feature | AWS API Gateway | API-Direct |
|---------|----------------|------------|
| Vendor lock-in | ❌ AWS only | ✅ Deploy anywhere |
| Cold starts | ❌ Lambda latency | ✅ Always warm |
| Local development | ❌ Complex setup | ✅ Simple `python app.py` |
| Custom logic | ❌ Limited | ✅ Full Python power |
| Pricing | ❌ Per request + AWS costs | ✅ Keep 85% revenue |

## 🎯 Use Cases

### Perfect for:
- **SaaS APIs** - Built-in billing and user management
- **AI/ML APIs** - Usage-based pricing for model inference
- **Data APIs** - Monetize your datasets
- **Microservices** - Add billing to internal services
- **API Products** - Launch and monetize quickly

### Examples:
- Weather API with free/premium tiers
- Image processing with per-image pricing
- Text analysis with usage-based billing
- Financial data with subscription model
- AI chatbot with message-based pricing

## 🤝 Support

- **Documentation**: [docs.apidirect.com](https://docs.apidirect.com)
- **Examples**: See `example_app.py`
- **Issues**: [GitHub Issues](https://github.com/api-direct/framework)
- **Community**: [Discord](https://discord.gg/apidirect)

## 📄 License

MIT License - see LICENSE file for details.

---

**Ready to monetize your API in minutes instead of months?** 

Try the example:
```bash
python example_app.py
```

Then visit `http://localhost:8000/docs` to see your monetized API in action! 🚀
