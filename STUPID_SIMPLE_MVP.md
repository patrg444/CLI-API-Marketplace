# ⚡ STUPID SIMPLE MVP: GPT Cost Saver

## 🎯 2-Hour Sprint Goal
ONE command that saves developers money. Nothing else matters.

```bash
npx create-apidirect-gpt my-api
cd my-api && apidirect deploy
# ✅ Live API with 70% cost savings in 3 minutes
```

## 📦 What Gets Built (ONLY)

### 1. Ultra-Minimal CLI (30 minutes)
```bash
# Single command, zero choices
npx create-apidirect-gpt my-gpt-api
```

**Generated Project (4 files only):**
```
my-gpt-api/
├── main.py          # 50 lines max - GPT + Redis caching
├── requirements.txt # 3 dependencies only
├── apidirect.yaml   # Minimal config
└── README.md        # How to save money
```

### 2. Core GPT Wrapper (60 minutes)
**main.py** - Absolute minimum:
```python
import os
import json
import hashlib
import redis
import openai
from functools import lru_cache

# Setup
openai.api_key = os.environ.get('OPENAI_API_KEY')
redis_client = redis.from_url(os.environ.get('REDIS_URL', 'redis://localhost:6379'))

def cache_key(prompt, model="gpt-3.5-turbo"):
    return hashlib.md5(f"{prompt}:{model}".encode()).hexdigest()

def complete(event, context):
    body = json.loads(event.get('body', '{}'))
    prompt = body.get('prompt')
    
    # Check cache first (THE MONEY SAVER)
    key = cache_key(prompt)
    cached = redis_client.get(key)
    if cached:
        result = json.loads(cached)
        result['cached'] = True
        result['cost_saved'] = True
        return {'statusCode': 200, 'body': json.dumps(result)}
    
    # Call OpenAI
    response = openai.chat.completions.create(
        model="gpt-3.5-turbo",
        messages=[{"role": "user", "content": prompt}],
        max_tokens=100
    )
    
    result = {
        'text': response.choices[0].message.content,
        'cached': False,
        'tokens': response.usage.total_tokens
    }
    
    # Cache for 1 hour
    redis_client.setex(key, 3600, json.dumps(result))
    
    return {'statusCode': 200, 'body': json.dumps(result)}
```

### 3. One-Click Deploy (30 minutes)
```yaml
# apidirect.yaml - MINIMAL
name: my-gpt-api
runtime: python3.9
endpoints:
  - path: /complete
    method: POST
    handler: main.complete
aws:
  cpu: 512
  memory: 1024
```

```bash
# Deploy command that works
apidirect deploy
# ✅ Returns: https://abc123.api-direct.io/complete
```

## 🚀 What We DON'T Build (Critical)

❌ No template choices  
❌ No interactive wizard  
❌ No additional features  
❌ No other AI models  
❌ No enterprise features  
❌ No documentation beyond README  
❌ No tests (yet)  
❌ No monitoring dashboard  
❌ No pricing tiers  
❌ No marketplace  

## 📊 Success Metrics for Morning

**By Noon:**
- [ ] `npx create-apidirect-gpt` works
- [ ] Generated API deploys successfully  
- [ ] Can make GPT call through deployed API
- [ ] Redis caching works (measure cache hits)
- [ ] Cost savings are measurable

**Demo Ready:**
```bash
# Record this working:
curl -X POST https://abc123.api-direct.io/complete \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello world"}'

# Should return:
{
  "text": "Hello! How can I help you today?",
  "cached": false,
  "tokens": 15
}

# Second call returns:
{
  "text": "Hello! How can I help you today?", 
  "cached": true,
  "cost_saved": true
}
```

## 🎯 The ONLY Value Prop

**"Your $300 OpenAI bill becomes $90"**

That's it. Nothing else. Pure cost arbitrage through caching.

## ⏰ Afternoon (After Noon)

- Landing page with cost calculator
- "Save $X/month" as the only headline
- One demo showing cache hits saving money
- Email signup for beta access

**NO MORE CODING AFTER NOON TODAY!**

Ready to execute this ultra-focused 2-hour sprint?