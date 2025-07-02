"""
Prometheus metrics for API monitoring

Provides custom metrics for the API backend including:
- Request counts and latencies
- Business metrics (API calls, revenue, users)
- System health metrics
"""

from prometheus_client import Counter, Histogram, Gauge, Info, generate_latest
from prometheus_client.core import CollectorRegistry
from functools import wraps
from time import time
from typing import Callable, Any
import asyncio
import logging

logger = logging.getLogger(__name__)

# Create a custom registry
registry = CollectorRegistry()

# Request metrics
http_requests_total = Counter(
    'http_requests_total',
    'Total HTTP requests',
    ['method', 'endpoint', 'status'],
    registry=registry
)

http_request_duration_seconds = Histogram(
    'http_request_duration_seconds',
    'HTTP request latency',
    ['method', 'endpoint'],
    registry=registry
)

http_request_size_bytes = Histogram(
    'http_request_size_bytes',
    'HTTP request size in bytes',
    ['method', 'endpoint'],
    registry=registry
)

http_response_size_bytes = Histogram(
    'http_response_size_bytes',
    'HTTP response size in bytes',
    ['method', 'endpoint'],
    registry=registry
)

# Business metrics
active_users_total = Gauge(
    'active_users_total',
    'Total active users',
    registry=registry
)

api_calls_total = Counter(
    'api_calls_total',
    'Total API calls made through the platform',
    ['api_id', 'user_id', 'status'],
    registry=registry
)

api_revenue_total = Counter(
    'api_revenue_total',
    'Total revenue from API calls',
    ['api_id', 'currency'],
    registry=registry
)

active_apis_total = Gauge(
    'active_apis_total',
    'Total active APIs on the platform',
    registry=registry
)

trial_activations_total = Counter(
    'trial_activations_total',
    'Total trial activations',
    ['api_id'],
    registry=registry
)

# Authentication metrics
auth_attempts_total = Counter(
    'auth_attempts_total',
    'Total authentication attempts',
    ['method', 'result'],
    registry=registry
)

api_key_operations_total = Counter(
    'api_key_operations_total',
    'API key operations',
    ['operation', 'result'],
    registry=registry
)

# WebSocket metrics
websocket_connections_active = Gauge(
    'websocket_connections_active',
    'Active WebSocket connections',
    registry=registry
)

websocket_messages_total = Counter(
    'websocket_messages_total',
    'Total WebSocket messages',
    ['direction', 'type'],
    registry=registry
)

# Database metrics
db_connections_active = Gauge(
    'db_connections_active',
    'Active database connections',
    registry=registry
)

db_query_duration_seconds = Histogram(
    'db_query_duration_seconds',
    'Database query duration',
    ['query_type'],
    registry=registry
)

# Cache metrics
cache_hits_total = Counter(
    'cache_hits_total',
    'Total cache hits',
    ['cache_type'],
    registry=registry
)

cache_misses_total = Counter(
    'cache_misses_total',
    'Total cache misses',
    ['cache_type'],
    registry=registry
)

# Error metrics
unhandled_exceptions_total = Counter(
    'unhandled_exceptions_total',
    'Total unhandled exceptions',
    ['exception_type'],
    registry=registry
)

# Version info
api_info = Info(
    'api_info',
    'API version information',
    registry=registry
)

# Set version info
api_info.info({
    'version': '1.0.0',
    'environment': 'production'
})


def track_request_metrics(func: Callable) -> Callable:
    """Decorator to track HTTP request metrics"""
    @wraps(func)
    async def wrapper(request, *args, **kwargs):
        start_time = time()
        endpoint = request.url.path
        method = request.method
        
        # Track request size
        request_size = int(request.headers.get('content-length', 0))
        http_request_size_bytes.labels(method=method, endpoint=endpoint).observe(request_size)
        
        try:
            # Call the actual endpoint
            response = await func(request, *args, **kwargs)
            
            # Track successful request
            status = getattr(response, 'status_code', 200)
            http_requests_total.labels(method=method, endpoint=endpoint, status=status).inc()
            
            # Track response size
            response_size = len(str(response.body)) if hasattr(response, 'body') else 0
            http_response_size_bytes.labels(method=method, endpoint=endpoint).observe(response_size)
            
            return response
            
        except Exception as e:
            # Track failed request
            http_requests_total.labels(method=method, endpoint=endpoint, status=500).inc()
            unhandled_exceptions_total.labels(exception_type=type(e).__name__).inc()
            raise
            
        finally:
            # Track request duration
            duration = time() - start_time
            http_request_duration_seconds.labels(method=method, endpoint=endpoint).observe(duration)
    
    return wrapper


def track_async_operation(operation_name: str, labels: dict = None):
    """Decorator to track async operation metrics"""
    def decorator(func: Callable) -> Callable:
        @wraps(func)
        async def wrapper(*args, **kwargs):
            histogram = Histogram(
                f'{operation_name}_duration_seconds',
                f'Duration of {operation_name} operation',
                list(labels.keys()) if labels else [],
                registry=registry
            )
            
            counter = Counter(
                f'{operation_name}_total',
                f'Total {operation_name} operations',
                list(labels.keys()) + ['status'] if labels else ['status'],
                registry=registry
            )
            
            start_time = time()
            try:
                result = await func(*args, **kwargs)
                
                # Track success
                label_values = {**(labels or {}), 'status': 'success'}
                counter.labels(**label_values).inc()
                
                return result
                
            except Exception as e:
                # Track failure
                label_values = {**(labels or {}), 'status': 'failure'}
                counter.labels(**label_values).inc()
                raise
                
            finally:
                # Track duration
                duration = time() - start_time
                if labels:
                    histogram.labels(**labels).observe(duration)
                else:
                    histogram.observe(duration)
        
        return wrapper
    return decorator


class MetricsCollector:
    """Collects and updates business metrics periodically"""
    
    def __init__(self, db_pool, redis_client):
        self.db_pool = db_pool
        self.redis_client = redis_client
        self.running = False
    
    async def start(self):
        """Start the metrics collection loop"""
        self.running = True
        while self.running:
            try:
                await self.collect_metrics()
            except Exception as e:
                logger.error(f"Error collecting metrics: {e}")
            
            # Collect metrics every 30 seconds
            await asyncio.sleep(30)
    
    async def stop(self):
        """Stop the metrics collection loop"""
        self.running = False
    
    async def collect_metrics(self):
        """Collect current metrics from the database"""
        async with self.db_pool.acquire() as conn:
            # Update active users
            active_users = await conn.fetchval("""
                SELECT COUNT(DISTINCT user_id)
                FROM api_calls
                WHERE created_at > NOW() - INTERVAL '24 hours'
            """)
            active_users_total.set(active_users or 0)
            
            # Update active APIs
            active_apis = await conn.fetchval("""
                SELECT COUNT(*)
                FROM apis
                WHERE status = 'active'
            """)
            active_apis_total.set(active_apis or 0)
            
            # Update database connections
            db_stats = await conn.fetchrow("""
                SELECT 
                    numbackends as active_connections,
                    (SELECT count(*) FROM pg_stat_activity) as total_connections
                FROM pg_stat_database 
                WHERE datname = current_database()
            """)
            if db_stats:
                db_connections_active.set(db_stats['active_connections'])
    
    async def track_api_call(self, api_id: str, user_id: str, status: int, revenue: float = 0):
        """Track an API call"""
        api_calls_total.labels(api_id=api_id, user_id=user_id, status=str(status)).inc()
        
        if revenue > 0:
            api_revenue_total.labels(api_id=api_id, currency='USD').inc(revenue)
    
    async def track_auth_attempt(self, method: str, success: bool):
        """Track authentication attempt"""
        result = 'success' if success else 'failure'
        auth_attempts_total.labels(method=method, result=result).inc()
    
    async def track_cache_access(self, cache_type: str, hit: bool):
        """Track cache access"""
        if hit:
            cache_hits_total.labels(cache_type=cache_type).inc()
        else:
            cache_misses_total.labels(cache_type=cache_type).inc()
    
    async def track_websocket_connection(self, delta: int):
        """Track WebSocket connection change"""
        websocket_connections_active.inc(delta)
    
    async def track_websocket_message(self, direction: str, message_type: str):
        """Track WebSocket message"""
        websocket_messages_total.labels(direction=direction, type=message_type).inc()


def get_metrics():
    """Generate Prometheus metrics in text format"""
    return generate_latest(registry)


# Create middleware for FastAPI
async def metrics_middleware(request, call_next):
    """Middleware to track all HTTP requests"""
    start_time = time()
    endpoint = request.url.path
    method = request.method
    
    # Skip metrics endpoint to avoid recursion
    if endpoint == '/metrics':
        return await call_next(request)
    
    # Track request size
    request_size = int(request.headers.get('content-length', 0))
    http_request_size_bytes.labels(method=method, endpoint=endpoint).observe(request_size)
    
    try:
        # Process request
        response = await call_next(request)
        
        # Track metrics
        http_requests_total.labels(method=method, endpoint=endpoint, status=response.status_code).inc()
        
        # Estimate response size
        response_size = int(response.headers.get('content-length', 0))
        if response_size == 0 and hasattr(response, 'body'):
            response_size = len(response.body)
        http_response_size_bytes.labels(method=method, endpoint=endpoint).observe(response_size)
        
        return response
        
    except Exception as e:
        # Track failed request
        http_requests_total.labels(method=method, endpoint=endpoint, status=500).inc()
        unhandled_exceptions_total.labels(exception_type=type(e).__name__).inc()
        raise
        
    finally:
        # Track request duration
        duration = time() - start_time
        http_request_duration_seconds.labels(method=method, endpoint=endpoint).observe(duration)