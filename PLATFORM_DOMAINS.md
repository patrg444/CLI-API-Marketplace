# API-Direct Platform Domains

## 🌐 Official Domain Structure

### Production Domains (Live)
- **Main Website**: https://apidirect.dev
- **Developer Console**: https://console.apidirect.dev  
- **API Marketplace**: https://marketplace.apidirect.dev
- **Documentation**: https://docs.apidirect.dev

### Future Domains (To Be Configured)
- **API Gateway**: https://api.apidirect.dev (for deployed APIs)
- **CLI Downloads**: https://cli.apidirect.dev (for CLI installation)

## 📁 Directory Mapping

```
/web/
├── landing/     → apidirect.dev
├── console/     → console.apidirect.dev
├── marketplace/ → marketplace.apidirect.dev
└── docs/        → docs.apidirect.dev
```

## 🚀 Available Features & Components

### 1. **Example APIs Ready to Deploy**
- **Weather API** - Complete weather service with forecasts
- **GPT Wrapper** - Cost-optimized OpenAI integration
- **Sentiment Analysis** - Advanced emotion detection API

### 2. **ML/AI Templates**
```bash
# Create a new AI API instantly
./cli/apidirect init my-ai-api --template gpt-wrapper
./cli/apidirect init my-vision-api --template image-classification
./cli/apidirect init my-sentiment-api --template sentiment-analysis
```

### 3. **API-Direct Framework**
```python
# Zero-code monetization for Python APIs
from apidirect import FastAPI, monetize

app = FastAPI()

@app.get("/predict")
@monetize(tier="premium", cost_per_call=0.01)
async def predict(text: str):
    return {"prediction": "result"}
```

### 4. **CLI Commands to Try**
```bash
# Initialize a new API project
./cli/apidirect init weather-api --template weather

# Import existing project
cd your-existing-api
./cli/apidirect import

# Deploy to platform
./cli/apidirect deploy

# Check analytics
./cli/apidirect analytics

# Publish to marketplace
./cli/apidirect publish --pricing freemium
```

### 5. **Platform Scripts**
```bash
# Health monitoring
./scripts/health-check.sh

# Platform testing
./test-platform.py

# E2E testing
./run-e2e-tests.sh
```

## 🎯 Quick Actions

### Deploy Your First API
```bash
# 1. Choose a template
cd demo-ml-api

# 2. Deploy it
../cli/apidirect deploy

# 3. View in console
open https://console.apidirect.dev
```

### Test the Marketplace
```bash
# Search for APIs
./cli/apidirect search weather

# Browse categories
./cli/apidirect browse ai-ml
```

### Monitor Platform Health
```bash
# Check all services
./scripts/health-check.sh

# View real-time logs
docker-compose logs -f
```

## 📊 Platform Capabilities

- **830+ E2E Tests** - Comprehensive test coverage
- **15+ API Templates** - Ready-to-deploy solutions
- **ML Framework** - Built-in AI/ML support
- **Auto-scaling** - Handle any load
- **Monetization** - Built-in billing & payments
- **Analytics** - Real-time usage tracking
- **Review System** - Community feedback
- **API Playground** - Test APIs directly

The platform is fully built with production-ready features!