"""
Retry utilities with exponential backoff and circuit breaker pattern
For handling transient failures in external service calls
"""

import asyncio
import time
from typing import TypeVar, Callable, Optional, Union, Dict, Any
from functools import wraps
import random
import logging
from dataclasses import dataclass
from enum import Enum
from datetime import datetime, timedelta

logger = logging.getLogger(__name__)

T = TypeVar('T')


class CircuitState(Enum):
    CLOSED = "closed"  # Normal operation
    OPEN = "open"      # Failing, reject calls
    HALF_OPEN = "half_open"  # Testing if service recovered


@dataclass
class RetryConfig:
    """Configuration for retry behavior"""
    max_attempts: int = 3
    initial_delay: float = 1.0
    max_delay: float = 60.0
    exponential_base: float = 2.0
    jitter: bool = True
    retry_on: tuple = (Exception,)  # Exceptions to retry on


class CircuitBreaker:
    """
    Circuit breaker pattern implementation
    Prevents cascading failures by stopping calls to failing services
    """
    
    def __init__(
        self,
        failure_threshold: int = 5,
        recovery_timeout: int = 60,
        expected_exception: type = Exception
    ):
        self.failure_threshold = failure_threshold
        self.recovery_timeout = recovery_timeout
        self.expected_exception = expected_exception
        
        self._failure_count = 0
        self._last_failure_time: Optional[datetime] = None
        self._state = CircuitState.CLOSED
    
    @property
    def state(self) -> CircuitState:
        """Get current circuit state"""
        if self._state == CircuitState.OPEN:
            if self._last_failure_time and \
               datetime.utcnow() - self._last_failure_time > timedelta(seconds=self.recovery_timeout):
                self._state = CircuitState.HALF_OPEN
        return self._state
    
    def call_succeeded(self):
        """Record successful call"""
        self._failure_count = 0
        self._state = CircuitState.CLOSED
    
    def call_failed(self):
        """Record failed call"""
        self._failure_count += 1
        self._last_failure_time = datetime.utcnow()
        
        if self._failure_count >= self.failure_threshold:
            self._state = CircuitState.OPEN
            logger.warning(f"Circuit breaker opened after {self._failure_count} failures")
    
    def can_execute(self) -> bool:
        """Check if call can be executed"""
        return self.state != CircuitState.OPEN
    
    async def call(self, func: Callable, *args, **kwargs):
        """Execute function with circuit breaker protection"""
        if not self.can_execute():
            raise Exception("Circuit breaker is OPEN")
        
        try:
            result = await func(*args, **kwargs)
            self.call_succeeded()
            return result
        except self.expected_exception as e:
            self.call_failed()
            raise


async def retry_with_backoff(
    func: Callable[..., T],
    config: RetryConfig = RetryConfig(),
    *args,
    **kwargs
) -> T:
    """
    Retry a function with exponential backoff
    
    Args:
        func: Async function to retry
        config: Retry configuration
        *args, **kwargs: Arguments for the function
        
    Returns:
        Function result
        
    Raises:
        Last exception if all retries fail
    """
    last_exception = None
    
    for attempt in range(config.max_attempts):
        try:
            return await func(*args, **kwargs)
        except config.retry_on as e:
            last_exception = e
            
            if attempt == config.max_attempts - 1:
                logger.error(f"All {config.max_attempts} attempts failed for {func.__name__}")
                raise
            
            # Calculate delay with exponential backoff
            delay = min(
                config.initial_delay * (config.exponential_base ** attempt),
                config.max_delay
            )
            
            # Add jitter to prevent thundering herd
            if config.jitter:
                delay *= (0.5 + random.random())
            
            logger.warning(
                f"Attempt {attempt + 1}/{config.max_attempts} failed for {func.__name__}, "
                f"retrying in {delay:.2f}s: {str(e)}"
            )
            
            await asyncio.sleep(delay)
    
    raise last_exception


def with_retry(
    max_attempts: int = 3,
    initial_delay: float = 1.0,
    max_delay: float = 60.0,
    retry_on: tuple = (Exception,)
):
    """
    Decorator for adding retry logic to async functions
    
    Usage:
        @with_retry(max_attempts=5, initial_delay=2.0)
        async def call_external_api():
            return await httpx.get("https://api.example.com")
    """
    config = RetryConfig(
        max_attempts=max_attempts,
        initial_delay=initial_delay,
        max_delay=max_delay,
        retry_on=retry_on
    )
    
    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            return await retry_with_backoff(func, config, *args, **kwargs)
        return wrapper
    return decorator


class RetryableHTTPClient:
    """
    HTTP client with built-in retry logic and circuit breaker
    """
    
    def __init__(
        self,
        base_url: Optional[str] = None,
        retry_config: RetryConfig = RetryConfig(),
        circuit_breaker: Optional[CircuitBreaker] = None
    ):
        self.base_url = base_url
        self.retry_config = retry_config
        self.circuit_breaker = circuit_breaker or CircuitBreaker()
        
        # Initialize httpx client with timeout
        import httpx
        self.client = httpx.AsyncClient(
            base_url=base_url,
            timeout=httpx.Timeout(30.0, connect=5.0)
        )
    
    async def _request_with_retry(
        self,
        method: str,
        url: str,
        **kwargs
    ) -> Any:
        """Execute HTTP request with retry logic"""
        async def make_request():
            response = await self.client.request(method, url, **kwargs)
            response.raise_for_status()
            return response
        
        if self.circuit_breaker:
            return await self.circuit_breaker.call(
                lambda: retry_with_backoff(make_request, self.retry_config)
            )
        else:
            return await retry_with_backoff(make_request, self.retry_config)
    
    async def get(self, url: str, **kwargs):
        """GET request with retry"""
        return await self._request_with_retry("GET", url, **kwargs)
    
    async def post(self, url: str, **kwargs):
        """POST request with retry"""
        return await self._request_with_retry("POST", url, **kwargs)
    
    async def put(self, url: str, **kwargs):
        """PUT request with retry"""
        return await self._request_with_retry("PUT", url, **kwargs)
    
    async def delete(self, url: str, **kwargs):
        """DELETE request with retry"""
        return await self._request_with_retry("DELETE", url, **kwargs)
    
    async def close(self):
        """Close the HTTP client"""
        await self.client.aclose()


class ExternalServiceClient:
    """
    Base class for external service clients with retry logic
    """
    
    def __init__(
        self,
        service_name: str,
        base_url: str,
        api_key: Optional[str] = None,
        retry_config: RetryConfig = RetryConfig()
    ):
        self.service_name = service_name
        self.base_url = base_url
        self.api_key = api_key
        
        # Create HTTP client with retry logic
        self.http = RetryableHTTPClient(
            base_url=base_url,
            retry_config=retry_config,
            circuit_breaker=CircuitBreaker(
                failure_threshold=5,
                recovery_timeout=60
            )
        )
        
        # Set default headers
        self.headers = {}
        if api_key:
            self.headers["Authorization"] = f"Bearer {api_key}"
    
    async def health_check(self) -> bool:
        """Check if service is healthy"""
        try:
            response = await self.http.get("/health", headers=self.headers)
            return response.status_code == 200
        except Exception as e:
            logger.error(f"{self.service_name} health check failed: {e}")
            return False


class StripeClient(ExternalServiceClient):
    """Stripe API client with retry logic"""
    
    def __init__(self, secret_key: str):
        super().__init__(
            service_name="Stripe",
            base_url="https://api.stripe.com/v1",
            api_key=secret_key,
            retry_config=RetryConfig(
                max_attempts=3,
                retry_on=(Exception,)
            )
        )
    
    @with_retry(max_attempts=3)
    async def create_customer(self, email: str, name: str) -> Dict[str, Any]:
        """Create Stripe customer with retry"""
        response = await self.http.post(
            "/customers",
            headers=self.headers,
            data={"email": email, "name": name}
        )
        return response.json()
    
    @with_retry(max_attempts=3)
    async def create_subscription(
        self,
        customer_id: str,
        price_id: str
    ) -> Dict[str, Any]:
        """Create subscription with retry"""
        response = await self.http.post(
            "/subscriptions",
            headers=self.headers,
            data={
                "customer": customer_id,
                "items": [{"price": price_id}]
            }
        )
        return response.json()


class AWSClient(ExternalServiceClient):
    """AWS service client with retry logic"""
    
    def __init__(self, service: str, region: str):
        # This is a simplified example
        # In production, use boto3 with its built-in retry logic
        super().__init__(
            service_name=f"AWS-{service}",
            base_url=f"https://{service}.{region}.amazonaws.com",
            retry_config=RetryConfig(
                max_attempts=3,
                initial_delay=1.0,
                retry_on=(Exception,)
            )
        )


# Global clients (initialized in main.py)
stripe_client: Optional[StripeClient] = None
aws_s3_client: Optional[AWSClient] = None


def init_external_clients(config: Dict[str, str]) -> None:
    """Initialize external service clients"""
    global stripe_client, aws_s3_client
    
    if config.get("STRIPE_SECRET_KEY"):
        stripe_client = StripeClient(config["STRIPE_SECRET_KEY"])
        logger.info("Stripe client initialized with retry logic")
    
    if config.get("AWS_REGION"):
        aws_s3_client = AWSClient("s3", config["AWS_REGION"])
        logger.info("AWS S3 client initialized with retry logic")