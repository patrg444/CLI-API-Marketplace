package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetMLTemplates returns AI/ML specific templates
func GetMLTemplates() []APITemplate {
	return []APITemplate{
		{
			ID:          "gpt-wrapper",
			Name:        "GPT Wrapper API",
			Description: "Production-ready OpenAI GPT wrapper with caching and rate limiting",
			Runtime:     "python3.9",
			Category:    "AI/ML",
			Features: []string{
				"Response caching with Redis",
				"Rate limiting protection",
				"Cost optimization",
				"Error handling and retry logic",
				"Usage analytics",
			},
		},
		{
			ID:          "image-classifier",
			Name:        "Image Classification API",
			Description: "Computer vision API using pre-trained Vision Transformer models",
			Runtime:     "python3.9",
			Category:    "AI/ML",
			Features: []string{
				"Vision Transformer models",
				"Multi-format image support",
				"Batch processing capability",
				"GPU optimization",
				"Confidence scoring",
			},
		},
		{
			ID:          "sentiment-analyzer",
			Name:        "Sentiment Analysis API",
			Description: "Advanced sentiment analysis with emotion detection using transformers",
			Runtime:     "python3.9",
			Category:    "AI/ML",
			Features: []string{
				"Multi-language support",
				"Emotion detection",
				"Confidence scores",
				"Batch text processing",
				"Custom model support",
			},
		},
		{
			ID:          "embeddings-api",
			Name:        "Text Embeddings API",
			Description: "Generate semantic embeddings for text using sentence transformers",
			Runtime:     "python3.9",
			Category:    "AI/ML",
			Features: []string{
				"Sentence transformer models",
				"Vector similarity search",
				"Batch embedding generation",
				"Multiple embedding models",
				"Dimensionality options",
			},
		},
		{
			ID:          "time-series-predictor",
			Name:        "Time Series Prediction API",
			Description: "Forecast time series data using Prophet and LSTM models",
			Runtime:     "python3.9",
			Category:    "AI/ML",
			Features: []string{
				"Prophet forecasting",
				"LSTM neural networks",
				"Seasonal decomposition",
				"Confidence intervals",
				"Multi-step predictions",
			},
		},
		{
			ID:          "document-qa",
			Name:        "Document Q&A API",
			Description: "Question answering over documents using BERT and retrieval",
			Runtime:     "python3.9",
			Category:    "AI/ML",
			Features: []string{
				"Document ingestion",
				"Question answering",
				"Context retrieval",
				"Multiple document formats",
				"Relevance scoring",
			},
		},
	}
}

// ML Template Configuration Generators
func getMLTemplateConfig(apiName, runtime string, template APITemplate) string {
	switch template.ID {
	case "gpt-wrapper":
		return fmt.Sprintf(`# API-Direct Configuration
name: %s
runtime: %s

# API Endpoints
endpoints:
  - path: /complete
    method: POST
    handler: main.complete_text
  
  - path: /chat
    method: POST
    handler: main.chat_completion
  
  - path: /health
    method: GET
    handler: main.health_check

# Environment Variables
environment:
  OPENAI_API_KEY: ${OPENAI_API_KEY}
  REDIS_URL: ${REDIS_URL}
  MAX_TOKENS: 1000
  CACHE_TTL: 3600
  LOG_LEVEL: INFO

# AWS Configuration (Optimized for AI workloads)
aws:
  cpu: 1024
  memory: 2048
  instance_type: "t3.large"
  min_capacity: 1
  max_capacity: 10
  
# Pricing Suggestions
pricing:
  free_tier: 100
  tiers:
    - name: "Starter"
      price_per_1k: 0.50
      features: ["Basic GPT-3.5", "Rate limiting"]
    - name: "Pro" 
      price_per_1k: 1.00
      features: ["GPT-4 access", "Priority processing", "Analytics"]
`, apiName, runtime)

	case "image-classifier":
		return fmt.Sprintf(`# API-Direct Configuration
name: %s
runtime: %s

# API Endpoints
endpoints:
  - path: /classify
    method: POST
    handler: main.classify_image
  
  - path: /classify/batch
    method: POST
    handler: main.classify_batch
  
  - path: /models
    method: GET
    handler: main.list_models
  
  - path: /health
    method: GET
    handler: main.health_check

# Environment Variables
environment:
  MODEL_NAME: "google/vit-base-patch16-224"
  BATCH_SIZE: 8
  CONFIDENCE_THRESHOLD: 0.5
  LOG_LEVEL: INFO

# AWS Configuration (GPU-enabled for image processing)
aws:
  cpu: 4096
  memory: 16384
  instance_type: "g4dn.xlarge" 
  gpu: 1
  min_capacity: 1
  max_capacity: 5

# Pricing Suggestions  
pricing:
  free_tier: 50
  tiers:
    - name: "Basic"
      price_per_image: 0.01
      features: ["Standard models", "Single image"]
    - name: "Advanced"
      price_per_image: 0.02 
      features: ["Premium models", "Batch processing", "Custom models"]
`, apiName, runtime)

	case "sentiment-analyzer":
		return fmt.Sprintf(`# API-Direct Configuration
name: %s
runtime: %s

# API Endpoints
endpoints:
  - path: /analyze
    method: POST
    handler: main.analyze_sentiment
  
  - path: /analyze/batch
    method: POST
    handler: main.analyze_batch
  
  - path: /emotions
    method: POST
    handler: main.detect_emotions
  
  - path: /health
    method: GET
    handler: main.health_check

# Environment Variables
environment:
  MODEL_NAME: "cardiffnlp/twitter-roberta-base-sentiment-latest"
  EMOTION_MODEL: "j-hartmann/emotion-english-distilroberta-base"
  BATCH_SIZE: 16
  LOG_LEVEL: INFO

# AWS Configuration
aws:
  cpu: 2048
  memory: 8192
  instance_type: "t3.xlarge"
  min_capacity: 1
  max_capacity: 8

# Pricing Suggestions
pricing:
  free_tier: 1000
  tiers:
    - name: "Starter"
      price_per_1k: 0.25
      features: ["Basic sentiment", "English only"]
    - name: "Professional"
      price_per_1k: 0.50
      features: ["Multi-language", "Emotion detection", "Batch processing"]
`, apiName, runtime)

	case "embeddings-api":
		return fmt.Sprintf(`# API-Direct Configuration
name: %s
runtime: %s

# API Endpoints
endpoints:
  - path: /embed
    method: POST
    handler: main.generate_embeddings
  
  - path: /embed/batch
    method: POST
    handler: main.batch_embeddings
  
  - path: /similarity
    method: POST
    handler: main.compute_similarity
  
  - path: /search
    method: POST
    handler: main.semantic_search
  
  - path: /health
    method: GET
    handler: main.health_check

# Environment Variables
environment:
  MODEL_NAME: "all-MiniLM-L6-v2"
  VECTOR_DIM: 384
  SIMILARITY_THRESHOLD: 0.7
  LOG_LEVEL: INFO

# AWS Configuration
aws:
  cpu: 2048
  memory: 4096
  instance_type: "t3.xlarge"
  min_capacity: 1
  max_capacity: 6

# Pricing Suggestions
pricing:
  free_tier: 5000
  tiers:
    - name: "Basic"
      price_per_1k: 0.10
      features: ["Standard embeddings", "Similarity search"]
    - name: "Premium"
      price_per_1k: 0.20
      features: ["Multiple models", "Vector search", "Custom dimensions"]
`, apiName, runtime)

	case "time-series-predictor":
		return fmt.Sprintf(`# API-Direct Configuration
name: %s
runtime: %s

# API Endpoints
endpoints:
  - path: /forecast
    method: POST
    handler: main.forecast_timeseries
  
  - path: /analyze
    method: POST
    handler: main.analyze_trends
  
  - path: /detect-anomalies
    method: POST
    handler: main.detect_anomalies
  
  - path: /health
    method: GET
    handler: main.health_check

# Environment Variables
environment:
  DEFAULT_PERIODS: 30
  CONFIDENCE_INTERVAL: 0.95
  MIN_HISTORY_POINTS: 50
  LOG_LEVEL: INFO

# AWS Configuration
aws:
  cpu: 2048
  memory: 4096
  instance_type: "t3.xlarge"
  min_capacity: 1
  max_capacity: 4

# Pricing Suggestions
pricing:
  free_tier: 100
  tiers:
    - name: "Basic"
      price_per_forecast: 0.05
      features: ["Prophet forecasting", "30-day horizon"]
    - name: "Advanced"
      price_per_forecast: 0.15
      features: ["Multiple models", "Custom horizons", "Anomaly detection"]
`, apiName, runtime)

	case "document-qa":
		return fmt.Sprintf(`# API-Direct Configuration
name: %s
runtime: %s

# API Endpoints
endpoints:
  - path: /upload
    method: POST
    handler: main.upload_document
  
  - path: /ask
    method: POST
    handler: main.answer_question
  
  - path: /documents
    method: GET
    handler: main.list_documents
  
  - path: /documents/{id}
    method: DELETE
    handler: main.delete_document
  
  - path: /health
    method: GET
    handler: main.health_check

# Environment Variables
environment:
  QA_MODEL: "deepset/roberta-base-squad2"
  CHUNK_SIZE: 500
  CHUNK_OVERLAP: 50
  MAX_DOCUMENT_SIZE: 10485760  # 10MB
  LOG_LEVEL: INFO

# AWS Configuration
aws:
  cpu: 4096
  memory: 8192
  instance_type: "t3.2xlarge"
  min_capacity: 1
  max_capacity: 3

# Pricing Suggestions
pricing:
  free_tier: 25
  tiers:
    - name: "Starter"
      price_per_query: 0.02
      features: ["5 documents", "Basic Q&A"]
    - name: "Business"
      price_per_query: 0.05
      features: ["Unlimited documents", "Advanced retrieval", "Multi-format support"]
`, apiName, runtime)

	default:
		return getPythonConfigTemplate(apiName, runtime)
	}
}

// ML Template Main File Generators
func getMLTemplateMain(template APITemplate) string {
	switch template.ID {
	case "gpt-wrapper":
		return `"""
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
`

	case "image-classifier":
		return `"""
Image Classification API Template
Computer vision API using pre-trained Vision Transformer models.
"""
import json
import logging
import os
import io
import base64
from typing import Dict, Any, List
from PIL import Image
from transformers import pipeline, AutoImageProcessor, AutoModelForImageClassification
import torch

# Configure logging
logging.basicConfig(level=os.environ.get('LOG_LEVEL', 'INFO'))
logger = logging.getLogger(__name__)

# Load model on startup
MODEL_NAME = os.environ.get('MODEL_NAME', 'google/vit-base-patch16-224')
logger.info(f"Loading model: {MODEL_NAME}")

try:
    classifier = pipeline("image-classification", model=MODEL_NAME)
    logger.info("Image classification model loaded successfully")
except Exception as e:
    logger.error(f"Failed to load model: {e}")
    classifier = None

def _decode_image(image_data: str) -> Image.Image:
    """Decode base64 image data"""
    # Remove data URL prefix if present
    if ',' in image_data:
        image_data = image_data.split(',')[1]
    
    image_bytes = base64.b64decode(image_data)
    image = Image.open(io.BytesIO(image_bytes))
    
    # Convert to RGB if necessary
    if image.mode != 'RGB':
        image = image.convert('RGB')
    
    return image

def classify_image(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Classify a single image
    """
    try:
        if not classifier:
            return {
                'statusCode': 503,
                'headers': {'Content-Type': 'application/json'},
                'body': json.dumps({'error': 'Model not available'})
            }
        
        body = json.loads(event.get('body', '{}'))
        
        if 'image' not in body:
            return {
                'statusCode': 400,
                'headers': {'Content-Type': 'application/json'},
                'body': json.dumps({'error': 'Image data is required'})
            }
        
        # Decode and process image
        image = _decode_image(body['image'])
        top_k = min(body.get('top_k', 5), 10)
        threshold = body.get('confidence_threshold', float(os.environ.get('CONFIDENCE_THRESHOLD', 0.0)))
        
        # Classify image
        results = classifier(image)
        
        # Filter by confidence threshold and limit results
        filtered_results = [
            r for r in results 
            if r['score'] >= threshold
        ][:top_k]
        
        response_data = {
            'predictions': filtered_results,
            'model': MODEL_NAME,
            'image_size': image.size,
            'total_predictions': len(results)
        }
        
        logger.info(f"Classified image: {len(filtered_results)} predictions above threshold")
        
        return {
            'statusCode': 200,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps(response_data)
        }
        
    except Exception as e:
        logger.error(f"Classification error: {e}")
        return {
            'statusCode': 500,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'error': 'Classification failed'})
        }

def classify_batch(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Classify multiple images in batch
    """
    try:
        if not classifier:
            return {
                'statusCode': 503,
                'headers': {'Content-Type': 'application/json'},
                'body': json.dumps({'error': 'Model not available'})
            }
        
        body = json.loads(event.get('body', '{}'))
        
        if 'images' not in body or not isinstance(body['images'], list):
            return {
                'statusCode': 400,
                'headers': {'Content-Type': 'application/json'},
                'body': json.dumps({'error': 'Images array is required'})
            }
        
        images_data = body['images']
        max_batch_size = int(os.environ.get('BATCH_SIZE', 8))
        
        if len(images_data) > max_batch_size:
            return {
                'statusCode': 400,
                'headers': {'Content-Type': 'application/json'},
                'body': json.dumps({
                    'error': f'Batch size too large. Maximum: {max_batch_size}'
                })
            }
        
        # Process images
        images = []
        for i, img_data in enumerate(images_data):
            try:
                images.append(_decode_image(img_data))
            except Exception as e:
                logger.error(f"Failed to decode image {i}: {e}")
                return {
                    'statusCode': 400,
                    'headers': {'Content-Type': 'application/json'},
                    'body': json.dumps({
                        'error': f'Invalid image data at index {i}'
                    })
                }
        
        # Classify all images
        top_k = min(body.get('top_k', 5), 10)
        threshold = body.get('confidence_threshold', float(os.environ.get('CONFIDENCE_THRESHOLD', 0.0)))
        
        results = []
        for i, image in enumerate(images):
            try:
                predictions = classifier(image)
                filtered_predictions = [
                    p for p in predictions 
                    if p['score'] >= threshold
                ][:top_k]
                
                results.append({
                    'image_index': i,
                    'predictions': filtered_predictions,
                    'image_size': image.size
                })
            except Exception as e:
                logger.error(f"Classification failed for image {i}: {e}")
                results.append({
                    'image_index': i,
                    'error': 'Classification failed',
                    'predictions': []
                })
        
        response_data = {
            'results': results,
            'model': MODEL_NAME,
            'batch_size': len(images)
        }
        
        logger.info(f"Batch classification completed: {len(images)} images")
        
        return {
            'statusCode': 200,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps(response_data)
        }
        
    except Exception as e:
        logger.error(f"Batch classification error: {e}")
        return {
            'statusCode': 500,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'error': 'Batch classification failed'})
        }

def list_models(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """List available models"""
    available_models = [
        {
            'name': 'google/vit-base-patch16-224',
            'description': 'Vision Transformer base model',
            'type': 'general classification'
        },
        {
            'name': 'microsoft/resnet-50',
            'description': 'ResNet-50 model for image classification',
            'type': 'general classification'
        }
    ]
    
    return {
        'statusCode': 200,
        'headers': {'Content-Type': 'application/json'},
        'body': json.dumps({
            'current_model': MODEL_NAME,
            'available_models': available_models
        })
    }

def health_check(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """Health check endpoint"""
    status = {
        'status': 'healthy' if classifier else 'degraded',
        'model_loaded': classifier is not None,
        'model_name': MODEL_NAME,
        'version': '1.0.0'
    }
    
    return {
        'statusCode': 200,
        'headers': {'Content-Type': 'application/json'},
        'body': json.dumps(status)
    }
`

	case "sentiment-analyzer":
		return `"""
Sentiment Analysis API Template
Advanced sentiment analysis with emotion detection using transformers.
"""
import json
import logging
import os
from typing import Dict, Any, List
from transformers import pipeline
import torch

# Configure logging
logging.basicConfig(level=os.environ.get('LOG_LEVEL', 'INFO'))
logger = logging.getLogger(__name__)

# Load models on startup
SENTIMENT_MODEL = os.environ.get('MODEL_NAME', 'cardiffnlp/twitter-roberta-base-sentiment-latest')
EMOTION_MODEL = os.environ.get('EMOTION_MODEL', 'j-hartmann/emotion-english-distilroberta-base')

logger.info(f"Loading sentiment model: {SENTIMENT_MODEL}")
logger.info(f"Loading emotion model: {EMOTION_MODEL}")

try:
    sentiment_analyzer = pipeline("sentiment-analysis", model=SENTIMENT_MODEL)
    emotion_analyzer = pipeline("text-classification", model=EMOTION_MODEL)
    logger.info("Models loaded successfully")
except Exception as e:
    logger.error(f"Failed to load models: {e}")
    sentiment_analyzer = None
    emotion_analyzer = None

def analyze_sentiment(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Analyze sentiment of a single text
    """
    try:
        if not sentiment_analyzer:
            return {
                'statusCode': 503,
                'headers': {'Content-Type': 'application/json'},
                'body': json.dumps({'error': 'Model not available'})
            }
        
        body = json.loads(event.get('body', '{}'))
        
        text = body.get('text', '').strip()
        if not text:
            return {
                'statusCode': 400,
                'headers': {'Content-Type': 'application/json'},
                'body': json.dumps({'error': 'Text is required'})
            }
        
        # Analyze sentiment
        sentiment_result = sentiment_analyzer(text)[0]
        
        # Normalize sentiment labels (different models use different labels)
        label = sentiment_result['label'].lower()
        if label in ['label_0', 'negative']:
            sentiment = 'negative'
        elif label in ['label_1', 'neutral']:
            sentiment = 'neutral'
        elif label in ['label_2', 'positive']:
            sentiment = 'positive'
        else:
            sentiment = label
        
        response_data = {
            'text': text,
            'sentiment': sentiment,
            'confidence': sentiment_result['score'],
            'model': SENTIMENT_MODEL
        }
        
        logger.info(f"Analyzed sentiment: {sentiment} ({sentiment_result['score']:.3f})")
        
        return {
            'statusCode': 200,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps(response_data)
        }
        
    except Exception as e:
        logger.error(f"Sentiment analysis error: {e}")
        return {
            'statusCode': 500,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'error': 'Sentiment analysis failed'})
        }

def analyze_batch(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Analyze sentiment for multiple texts
    """
    try:
        if not sentiment_analyzer:
            return {
                'statusCode': 503,
                'headers': {'Content-Type': 'application/json'},
                'body': json.dumps({'error': 'Model not available'})
            }
        
        body = json.loads(event.get('body', '{}'))
        
        texts = body.get('texts', [])
        if not texts or not isinstance(texts, list):
            return {
                'statusCode': 400,
                'headers': {'Content-Type': 'application/json'},
                'body': json.dumps({'error': 'Texts array is required'})
            }
        
        max_batch_size = int(os.environ.get('BATCH_SIZE', 16))
        if len(texts) > max_batch_size:
            return {
                'statusCode': 400,
                'headers': {'Content-Type': 'application/json'},
                'body': json.dumps({
                    'error': f'Batch size too large. Maximum: {max_batch_size}'
                })
            }
        
        # Analyze all texts
        sentiment_results = sentiment_analyzer(texts)
        
        results = []
        for i, (text, result) in enumerate(zip(texts, sentiment_results)):
            # Normalize sentiment labels
            label = result['label'].lower()
            if label in ['label_0', 'negative']:
                sentiment = 'negative'
            elif label in ['label_1', 'neutral']:
                sentiment = 'neutral'
            elif label in ['label_2', 'positive']:
                sentiment = 'positive'
            else:
                sentiment = label
            
            results.append({
                'text_index': i,
                'text': text[:100] + '...' if len(text) > 100 else text,
                'sentiment': sentiment,
                'confidence': result['score']
            })
        
        response_data = {
            'results': results,
            'model': SENTIMENT_MODEL,
            'batch_size': len(texts)
        }
        
        logger.info(f"Batch sentiment analysis completed: {len(texts)} texts")
        
        return {
            'statusCode': 200,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps(response_data)
        }
        
    except Exception as e:
        logger.error(f"Batch sentiment analysis error: {e}")
        return {
            'statusCode': 500,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'error': 'Batch sentiment analysis failed'})
        }

def detect_emotions(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Detect emotions in text
    """
    try:
        if not emotion_analyzer:
            return {
                'statusCode': 503,
                'headers': {'Content-Type': 'application/json'},
                'body': json.dumps({'error': 'Emotion model not available'})
            }
        
        body = json.loads(event.get('body', '{}'))
        
        text = body.get('text', '').strip()
        if not text:
            return {
                'statusCode': 400,
                'headers': {'Content-Type': 'application/json'},
                'body': json.dumps({'error': 'Text is required'})
            }
        
        # Analyze emotions
        emotion_results = emotion_analyzer(text)
        
        # Get top emotions
        top_k = min(body.get('top_k', 3), len(emotion_results))
        top_emotions = sorted(emotion_results, key=lambda x: x['score'], reverse=True)[:top_k]
        
        response_data = {
            'text': text,
            'emotions': [
                {
                    'emotion': result['label'],
                    'confidence': result['score']
                }
                for result in top_emotions
            ],
            'dominant_emotion': top_emotions[0]['label'] if top_emotions else None,
            'model': EMOTION_MODEL
        }
        
        logger.info(f"Detected emotions: {[e['emotion'] for e in response_data['emotions']]}")
        
        return {
            'statusCode': 200,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps(response_data)
        }
        
    except Exception as e:
        logger.error(f"Emotion detection error: {e}")
        return {
            'statusCode': 500,
            'headers': {'Content-Type': 'application/json'},
            'body': json.dumps({'error': 'Emotion detection failed'})
        }

def health_check(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """Health check endpoint"""
    status = {
        'status': 'healthy' if (sentiment_analyzer and emotion_analyzer) else 'degraded',
        'sentiment_model_loaded': sentiment_analyzer is not None,
        'emotion_model_loaded': emotion_analyzer is not None,
        'sentiment_model': SENTIMENT_MODEL,
        'emotion_model': EMOTION_MODEL,
        'version': '1.0.0'
    }
    
    return {
        'statusCode': 200,
        'headers': {'Content-Type': 'application/json'},
        'body': json.dumps(status)
    }
`

	default:
		return getPythonMainTemplate()
	}
}

// ML Template Requirements Generator
func getMLTemplateRequirements(template APITemplate) string {
	switch template.ID {
	case "gpt-wrapper":
		return `# Core AI dependencies
openai==1.3.0
redis==5.0.1

# Caching and optimization
python-multipart==0.0.6

# Core utilities
pydantic==2.5.0
requests==2.31.0

# Development dependencies
pytest==7.4.3
pytest-asyncio==0.21.1
`

	case "image-classifier":
		return `# Computer Vision dependencies
transformers==4.35.0
torch==2.1.0+cpu
torchvision==0.16.0+cpu
Pillow==10.1.0

# Image processing
opencv-python-headless==4.8.1.78

# Utilities
numpy==1.24.3
python-multipart==0.0.6

# Development dependencies
pytest==7.4.3
`

	case "sentiment-analyzer":
		return `# NLP dependencies
transformers==4.35.0
torch==2.1.0+cpu

# Text processing
nltk==3.8.1
spacy==3.7.2

# Utilities
numpy==1.24.3
python-multipart==0.0.6

# Development dependencies
pytest==7.4.3
`

	case "embeddings-api":
		return `# Embedding dependencies
sentence-transformers==2.2.2
transformers==4.35.0
torch==2.1.0+cpu

# Vector operations
numpy==1.24.3
faiss-cpu==1.7.4
scipy==1.11.4

# Utilities
python-multipart==0.0.6

# Development dependencies
pytest==7.4.3
`

	case "time-series-predictor":
		return `# Time series dependencies
prophet==1.1.5
pandas==2.1.3
numpy==1.24.3
scikit-learn==1.3.2

# Statistical analysis
statsmodels==0.14.0
scipy==1.11.4

# Plotting (optional)
matplotlib==3.8.0
plotly==5.17.0

# Utilities
python-multipart==0.0.6

# Development dependencies
pytest==7.4.3
`

	case "document-qa":
		return `# Document processing dependencies
transformers==4.35.0
torch==2.1.0+cpu

# Document parsing
PyPDF2==3.0.1
python-docx==1.1.0
markdown==3.5.1

# Text processing
sentence-transformers==2.2.2
faiss-cpu==1.7.4

# Utilities
numpy==1.24.3
python-multipart==0.0.6

# Development dependencies
pytest==7.4.3
`

	default:
		return getPythonRequirementsTemplate()
	}
}

// InitMLProject initializes an ML project with specific template
func InitMLProject(apiName, runtime string, template APITemplate) error {
	projectPath := apiName
	
	// Create project structure
	dirs := []string{
		"",
		"tests",
		"models", // For ML model files
		"data",   // For sample data
	}
	
	for _, dir := range dirs {
		fullPath := filepath.Join(projectPath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
		}
	}
	
	// Create files
	files := map[string]string{
		"apidirect.yaml":    getMLTemplateConfig(apiName, runtime, template),
		"main.py":           getMLTemplateMain(template),
		"requirements.txt":  getMLTemplateRequirements(template),
		".gitignore":        getPythonGitignoreTemplate(),
		"README.md":         getMLTemplateReadme(apiName, template),
		"tests/__init__.py": "",
		"tests/test_main.py": getMLTemplateTests(template),
		"data/.gitkeep":     "", // Keep data directory in git
		"models/.gitkeep":   "", // Keep models directory in git
	}
	
	for filename, content := range files {
		fullPath := filepath.Join(projectPath, filename)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", fullPath, err)
		}
	}
	
	return nil
}

func getMLTemplateReadme(apiName string, template APITemplate) string {
	return fmt.Sprintf(`# %s

%s

**Template:** %s  
**Category:** %s  
**Runtime:** Python 3.9

## 🚀 Features

%s

## 🔧 Quick Start

1. **Install dependencies**:
   ` + "```bash" + `
   pip install -r requirements.txt
   ` + "```" + `

2. **Set environment variables**:
   ` + "```bash" + `
   export OPENAI_API_KEY="your-key-here"  # If using OpenAI
   ` + "```" + `

3. **Test locally**:
   ` + "```bash" + `
   apidirect run
   ` + "```" + `

4. **Deploy to production**:
   ` + "```bash" + `
   apidirect deploy
   ` + "```" + `

5. **Publish to marketplace**:
   ` + "```bash" + `
   apidirect publish %s
   ` + "```" + `

## 📊 API Endpoints

See ` + "`apidirect.yaml`" + ` for complete endpoint documentation.

## 💰 Pricing Suggestions

This template includes optimized pricing recommendations based on:
- Model computational requirements
- Market analysis of similar APIs
- Cost optimization strategies

## 🔗 Resources

- [API-Direct Documentation](https://docs.api-direct.io)
- [AI/ML Best Practices](https://docs.api-direct.io/ai-ml)
- [Pricing Guide](https://docs.api-direct.io/pricing)

## 🆘 Support

- Documentation: https://docs.api-direct.io
- Support: support@api-direct.io
- Community: https://discord.gg/api-direct
`, 
		apiName, 
		template.Description,
		template.Name,
		template.Category,
		strings.Join(template.Features, "\n- "),
		apiName)
}

func getMLTemplateTests(template APITemplate) string {
	switch template.ID {
	case "gpt-wrapper":
		return `"""
Tests for GPT Wrapper API
"""
import json
import unittest
from unittest.mock import patch, MagicMock
from main import complete_text, chat_completion, health_check


class TestGPTWrapperAPI(unittest.TestCase):
    
    @patch('main.openai')
    def test_complete_text_success(self, mock_openai):
        # Mock OpenAI response
        mock_response = MagicMock()
        mock_response.choices[0].text = "This is a test completion"
        mock_response.usage.prompt_tokens = 10
        mock_response.usage.completion_tokens = 5
        mock_response.usage.total_tokens = 15
        mock_openai.completions.create.return_value = mock_response
        
        event = {
            'body': json.dumps({
                'prompt': 'Test prompt',
                'max_tokens': 100,
                'temperature': 0.7
            })
        }
        context = {}
        
        response = complete_text(event, context)
        
        self.assertEqual(response['statusCode'], 200)
        body = json.loads(response['body'])
        self.assertIn('text', body)
        self.assertIn('usage', body)
        self.assertEqual(body['text'], 'This is a test completion')
    
    def test_complete_text_missing_prompt(self):
        event = {
            'body': json.dumps({})
        }
        context = {}
        
        response = complete_text(event, context)
        
        self.assertEqual(response['statusCode'], 400)
        body = json.loads(response['body'])
        self.assertIn('error', body)
    
    def test_health_check(self):
        event = {}
        context = {}
        
        response = health_check(event, context)
        
        self.assertEqual(response['statusCode'], 200)
        body = json.loads(response['body'])
        self.assertIn('status', body)


if __name__ == '__main__':
    unittest.main()
`

	case "image-classifier":
		return `"""
Tests for Image Classification API
"""
import json
import unittest
import base64
from unittest.mock import patch, MagicMock
from PIL import Image
import io
from main import classify_image, health_check


class TestImageClassifierAPI(unittest.TestCase):
    
    def create_test_image_data(self):
        """Create a test image as base64 data"""
        img = Image.new('RGB', (224, 224), color='red')
        buffer = io.BytesIO()
        img.save(buffer, format='JPEG')
        img_data = base64.b64encode(buffer.getvalue()).decode()
        return img_data
    
    @patch('main.classifier')
    def test_classify_image_success(self, mock_classifier):
        # Mock classifier response
        mock_classifier.return_value = [
            {'label': 'cat', 'score': 0.9},
            {'label': 'dog', 'score': 0.1}
        ]
        
        event = {
            'body': json.dumps({
                'image': self.create_test_image_data(),
                'top_k': 2
            })
        }
        context = {}
        
        response = classify_image(event, context)
        
        self.assertEqual(response['statusCode'], 200)
        body = json.loads(response['body'])
        self.assertIn('predictions', body)
        self.assertIn('model', body)
        self.assertEqual(len(body['predictions']), 2)
    
    def test_classify_image_missing_data(self):
        event = {
            'body': json.dumps({})
        }
        context = {}
        
        response = classify_image(event, context)
        
        self.assertEqual(response['statusCode'], 400)
        body = json.loads(response['body'])
        self.assertIn('error', body)
    
    def test_health_check(self):
        event = {}
        context = {}
        
        response = health_check(event, context)
        
        self.assertEqual(response['statusCode'], 200)
        body = json.loads(response['body'])
        self.assertIn('status', body)


if __name__ == '__main__':
    unittest.main()
`

	default:
		return getPythonTestTemplate()
	}
}