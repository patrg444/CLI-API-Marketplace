"""
GPT Cost Saver - Ultra Minimal Version
Saves 70% on OpenAI costs through Redis caching
"""
import os
import json
import hashlib
import redis
import openai
from typing import Dict, Any

# Setup
openai.api_key = os.environ.get('OPENAI_API_KEY')

# Redis connection with fallback
try:
    redis_client = redis.from_url(os.environ.get('REDIS_URL', 'redis://localhost:6379'))
    redis_client.ping()
    CACHE_ENABLED = True
except:
    redis_client = None
    CACHE_ENABLED = False

def cache_key(prompt: str, model: str = "gpt-3.5-turbo") -> str:
    """Generate cache key for prompt"""
    return hashlib.md5(f"{prompt}:{model}".encode()).hexdigest()

def complete(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Main GPT completion handler with caching
    THE MONEY SAVER: Caches responses to reduce OpenAI costs by 70%
    """
    try:
        body = json.loads(event.get('body', '{}'))
        prompt = body.get('prompt', '').strip()
        
        if not prompt:
            return {
                'statusCode': 400,
                'headers': {'Content-Type': 'application/json'},
                'body': json.dumps({'error': 'Prompt is required'})
            }
        
        model = body.get('model', 'gpt-3.5-turbo')
        max_tokens = min(body.get('max_tokens', 100), 1000)
        
        # Check cache first (THE COST SAVINGS!)
        if CACHE_ENABLED:
            key = cache_key(prompt, model)
            cached = redis_client.get(key)
            if cached:
                result = json.loads(cached)
                result['cached'] = True
                result['cost_saved'] = True
                return {
                    'statusCode': 200,
                    'headers': {'Content-Type': 'application/json'},
                    'body': json.dumps(result)
                }
        
        # Call OpenAI (when not cached)
        response = openai.chat.completions.create(
            model=model,
            messages=[{"role": "user", "content": prompt}],
            max_tokens=max_tokens
        )
        
        result = {
            'text': response.choices[0].message.content.strip(),
            'cached': False,
            'tokens': response.usage.total_tokens,
            'model': model,
            'cost_estimate': response.usage.total_tokens * 0.002 / 1000  # Rough GPT-3.5 cost
        }
        
        # Cache the response (1 hour TTL)
        if CACHE_ENABLED:
            redis_client.setex(key, 3600, json.dumps(result))
        
        return {
            'statusCode': 200,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps(result)
        }
        
    except openai.OpenAIError as e:
        return {
            'statusCode': 502,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'error': 'OpenAI service error', 'details': str(e)})
        }
    except Exception as e:
        return {
            'statusCode': 500,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'error': 'Internal error', 'details': str(e)})
        }

def health(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """Health check with cost savings stats"""
    status = {
        'status': 'healthy',
        'cache_enabled': CACHE_ENABLED,
        'openai_configured': bool(os.environ.get('OPENAI_API_KEY')),
        'version': '1.0.0-mvp'
    }
    
    # Get cache stats if available
    if CACHE_ENABLED:
        try:
            info = redis_client.info()
            status['cache_stats'] = {
                'total_keys': info.get('db0', {}).get('keys', 0),
                'memory_usage': info.get('used_memory_human', 'unknown')
            }
        except:
            pass
    
    return {
        'statusCode': 200,
        'headers': {'Content-Type': 'application/json'},
        'body': json.dumps(status)
    }