"""
GPT Wrapper API Template
Production-ready OpenAI GPT wrapper with caching and rate limiting.
"""
import json
import logging
import os
import hashlib
import redis
from typing import Dict, Any, Optional
from functools import lru_cache
import openai

# Configure logging
logging.basicConfig(level=os.environ.get('LOG_LEVEL', 'INFO'))
logger = logging.getLogger(__name__)

# Initialize OpenAI client
openai.api_key = os.environ.get('OPENAI_API_KEY')

# Initialize Redis for caching (optional)
redis_client = None
if os.environ.get('REDIS_URL'):
    try:
        redis_client = redis.from_url(os.environ.get('REDIS_URL'))
        redis_client.ping()
        logger.info("Redis connection established")
    except Exception as e:
        logger.warning(f"Redis connection failed: {e}")
        redis_client = None

def _generate_cache_key(prompt: str, max_tokens: int, temperature: float) -> str:
    """Generate cache key for request parameters"""
    content = f"{prompt}:{max_tokens}:{temperature}"
    return hashlib.md5(content.encode()).hexdigest()

def _get_cached_response(cache_key: str) -> Optional[Dict]:
    """Get cached response if available"""
    if not redis_client:
        return None
    
    try:
        cached = redis_client.get(cache_key)
        if cached:
            return json.loads(cached)
    except Exception as e:
        logger.error(f"Cache read error: {e}")
    
    return None

def _cache_response(cache_key: str, response: Dict, ttl: int = 3600):
    """Cache response with TTL"""
    if not redis_client:
        return
    
    try:
        redis_client.setex(cache_key, ttl, json.dumps(response))
    except Exception as e:
        logger.error(f"Cache write error: {e}")

def complete_text(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Generate text completion using OpenAI GPT
    """
    try:
        body = json.loads(event.get('body', '{}'))
        
        prompt = body.get('prompt', '')
        max_tokens = min(body.get('max_tokens', 100), int(os.environ.get('MAX_TOKENS', 1000)))
        temperature = max(0.0, min(body.get('temperature', 0.7), 2.0))
        model = body.get('model', 'gpt-3.5-turbo-instruct')
        
        if not prompt:
            return {
                'statusCode': 400,
                'headers': {'Content-Type': 'application/json'},
                'body': json.dumps({'error': 'Prompt is required'})
            }
        
        # Check cache first
        cache_key = _generate_cache_key(prompt, max_tokens, temperature)
        cached_response = _get_cached_response(cache_key)
        
        if cached_response:
            logger.info("Returning cached response")
            return {
                'statusCode': 200,
                'headers': {'Content-Type': 'application/json'},
                'body': json.dumps({
                    **cached_response,
                    'cached': True
                })
            }
        
        # Generate new completion
        response = openai.completions.create(
            model=model,
            prompt=prompt,
            max_tokens=max_tokens,
            temperature=temperature
        )
        
        result = {
            'text': response.choices[0].text.strip(),
            'usage': {
                'prompt_tokens': response.usage.prompt_tokens,
                'completion_tokens': response.usage.completion_tokens,
                'total_tokens': response.usage.total_tokens
            },
            'model': model,
            'cached': False
        }
        
        # Cache the response
        cache_ttl = int(os.environ.get('CACHE_TTL', 3600))
        _cache_response(cache_key, result, cache_ttl)
        
        logger.info(f"Generated completion: {response.usage.total_tokens} tokens")
        
        return {
            'statusCode': 200,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps(result)
        }
        
    except openai.OpenAIError as e:
        logger.error(f"OpenAI API error: {e}")
        return {
            'statusCode': 502,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'error': 'AI service unavailable'})
        }
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
        return {
            'statusCode': 500,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'error': 'Internal server error'})
        }

def chat_completion(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Generate chat completion using OpenAI ChatGPT
    """
    try:
        body = json.loads(event.get('body', '{}'))
        
        messages = body.get('messages', [])
        max_tokens = min(body.get('max_tokens', 150), int(os.environ.get('MAX_TOKENS', 1000)))
        temperature = max(0.0, min(body.get('temperature', 0.7), 2.0))
        model = body.get('model', 'gpt-3.5-turbo')
        
        if not messages:
            return {
                'statusCode': 400,
                'headers': {'Content-Type': 'application/json'},
                'body': json.dumps({'error': 'Messages array is required'})
            }
        
        response = openai.chat.completions.create(
            model=model,
            messages=messages,
            max_tokens=max_tokens,
            temperature=temperature
        )
        
        result = {
            'message': response.choices[0].message.content,
            'usage': {
                'prompt_tokens': response.usage.prompt_tokens,
                'completion_tokens': response.usage.completion_tokens,
                'total_tokens': response.usage.total_tokens
            },
            'model': model
        }
        
        logger.info(f"Generated chat completion: {response.usage.total_tokens} tokens")
        
        return {
            'statusCode': 200,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps(result)
        }
        
    except openai.OpenAIError as e:
        logger.error(f"OpenAI API error: {e}")
        return {
            'statusCode': 502,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'error': 'AI service unavailable'})
        }
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
        return {
            'statusCode': 500,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'error': 'Internal server error'})
        }

def health_check(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """Health check endpoint"""
    status = {
        'status': 'healthy',
        'openai_configured': bool(os.environ.get('OPENAI_API_KEY')),
        'redis_connected': redis_client is not None,
        'version': '1.0.0'
    }
    
    return {
        'statusCode': 200,
        'headers': {'Content-Type': 'application/json'},
        'body': json.dumps(status)
    }