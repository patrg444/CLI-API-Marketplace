"""
Application Performance Monitoring (APM) with OpenTelemetry

Provides distributed tracing, metrics, and logging for the API backend.
Supports multiple APM backends: Jaeger, Datadog, New Relic, etc.
"""

import os
import logging
from typing import Optional, Dict, Any
from contextlib import contextmanager

# OpenTelemetry imports
from opentelemetry import trace, metrics, baggage
from opentelemetry.trace import Status, StatusCode
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.resources import Resource, SERVICE_NAME, SERVICE_VERSION
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.asyncpg import AsyncPGInstrumentor
from opentelemetry.instrumentation.redis import RedisInstrumentor
from opentelemetry.instrumentation.requests import RequestsInstrumentor
from opentelemetry.instrumentation.logging import LoggingInstrumentor
from opentelemetry.propagate import set_global_textmap

# Exporters
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.exporter.otlp.proto.grpc.metric_exporter import OTLPMetricExporter
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.exporter.prometheus import PrometheusMetricReader

# Context propagation
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator

logger = logging.getLogger(__name__)


class APMManager:
    """Manages APM configuration and instrumentation"""
    
    def __init__(self, service_name: str = "api-direct-backend", 
                 service_version: str = "1.0.0",
                 environment: str = None):
        self.service_name = service_name
        self.service_version = service_version
        self.environment = environment or os.getenv("ENVIRONMENT", "development")
        self.tracer = None
        self.meter = None
        self.initialized = False
        
        # Custom metrics
        self.request_counter = None
        self.error_counter = None
        self.business_metrics = {}
    
    def initialize(self, 
                   apm_backend: str = None,
                   otlp_endpoint: str = None,
                   jaeger_endpoint: str = None,
                   datadog_agent_url: str = None):
        """Initialize APM with the specified backend"""
        
        if self.initialized:
            logger.warning("APM already initialized")
            return
        
        # Determine APM backend
        apm_backend = apm_backend or os.getenv("APM_BACKEND", "otlp")
        
        # Create resource
        resource = Resource.create({
            SERVICE_NAME: self.service_name,
            SERVICE_VERSION: self.service_version,
            "service.environment": self.environment,
            "service.instance.id": os.getenv("HOSTNAME", "local"),
            "telemetry.sdk.language": "python",
            "telemetry.sdk.name": "opentelemetry",
        })
        
        # Initialize tracing
        self._initialize_tracing(apm_backend, resource, otlp_endpoint, jaeger_endpoint)
        
        # Initialize metrics
        self._initialize_metrics(resource)
        
        # Set up propagation
        set_global_textmap(TraceContextTextMapPropagator())
        
        # Auto-instrument libraries
        self._auto_instrument()
        
        self.initialized = True
        logger.info(f"APM initialized with {apm_backend} backend")
    
    def _initialize_tracing(self, backend: str, resource: Resource, 
                           otlp_endpoint: str = None, 
                           jaeger_endpoint: str = None):
        """Initialize tracing with the specified backend"""
        
        # Create tracer provider
        provider = TracerProvider(resource=resource)
        
        # Configure exporter based on backend
        if backend == "otlp":
            endpoint = otlp_endpoint or os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
            exporter = OTLPSpanExporter(endpoint=endpoint, insecure=True)
        elif backend == "jaeger":
            endpoint = jaeger_endpoint or os.getenv("JAEGER_ENDPOINT", "localhost:14268")
            exporter = JaegerExporter(
                agent_host_name=endpoint.split(":")[0],
                agent_port=int(endpoint.split(":")[1]) if ":" in endpoint else 6831,
                collector_endpoint=f"http://{endpoint}/api/traces" if ":" in endpoint else None
            )
        else:
            logger.warning(f"Unknown APM backend: {backend}, using OTLP")
            exporter = OTLPSpanExporter(insecure=True)
        
        # Add span processor
        provider.add_span_processor(BatchSpanProcessor(exporter))
        
        # Set as global tracer provider
        trace.set_tracer_provider(provider)
        
        # Get tracer
        self.tracer = trace.get_tracer(self.service_name, self.service_version)
    
    def _initialize_metrics(self, resource: Resource):
        """Initialize metrics collection"""
        
        # Create metric readers
        readers = []
        
        # Prometheus reader for /metrics endpoint
        prometheus_reader = PrometheusMetricReader()
        readers.append(prometheus_reader)
        
        # OTLP metric exporter for APM backend
        if os.getenv("OTEL_METRICS_EXPORTER_ENABLED", "true").lower() == "true":
            otlp_exporter = OTLPMetricExporter(
                endpoint=os.getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "localhost:4317"),
                insecure=True
            )
            readers.append(otlp_exporter)
        
        # Create meter provider
        provider = MeterProvider(resource=resource, metric_readers=readers)
        metrics.set_meter_provider(provider)
        
        # Get meter
        self.meter = metrics.get_meter(self.service_name, self.service_version)
        
        # Create custom metrics
        self._create_custom_metrics()
    
    def _create_custom_metrics(self):
        """Create custom business metrics"""
        
        # Request metrics
        self.request_counter = self.meter.create_counter(
            "api.requests.total",
            description="Total API requests",
            unit="requests"
        )
        
        self.error_counter = self.meter.create_counter(
            "api.errors.total",
            description="Total API errors",
            unit="errors"
        )
        
        # Business metrics
        self.business_metrics = {
            "api_calls": self.meter.create_counter(
                "business.api_calls.total",
                description="Total API calls through platform",
                unit="calls"
            ),
            "revenue": self.meter.create_counter(
                "business.revenue.total",
                description="Total revenue generated",
                unit="USD"
            ),
            "active_users": self.meter.create_up_down_counter(
                "business.active_users",
                description="Currently active users",
                unit="users"
            ),
            "trial_activations": self.meter.create_counter(
                "business.trial_activations.total",
                description="Total trial activations",
                unit="activations"
            )
        }
    
    def _auto_instrument(self):
        """Auto-instrument common libraries"""
        
        # Instrument FastAPI
        FastAPIInstrumentor.instrument(
            tracer_provider=trace.get_tracer_provider(),
            excluded_urls="/health,/metrics"
        )
        
        # Instrument database
        AsyncPGInstrumentor().instrument()
        
        # Instrument Redis
        RedisInstrumentor().instrument()
        
        # Instrument HTTP requests
        RequestsInstrumentor().instrument()
        
        # Instrument logging
        LoggingInstrumentor().instrument(set_logging_format=True)
    
    @contextmanager
    def trace_operation(self, operation_name: str, attributes: Dict[str, Any] = None):
        """Context manager for tracing custom operations"""
        
        if not self.tracer:
            yield None
            return
        
        with self.tracer.start_as_current_span(operation_name) as span:
            # Add attributes
            if attributes:
                for key, value in attributes.items():
                    span.set_attribute(key, str(value))
            
            # Add baggage for context propagation
            if attributes:
                for key, value in attributes.items():
                    baggage.set_baggage(key, str(value))
            
            try:
                yield span
            except Exception as e:
                # Record exception
                span.record_exception(e)
                span.set_status(Status(StatusCode.ERROR, str(e)))
                raise
            finally:
                # Clear baggage
                baggage.clear()
    
    async def trace_async_operation(self, operation_name: str, 
                                   attributes: Dict[str, Any] = None):
        """Decorator for tracing async operations"""
        def decorator(func):
            async def wrapper(*args, **kwargs):
                with self.trace_operation(operation_name, attributes):
                    return await func(*args, **kwargs)
            return wrapper
        return decorator
    
    def record_api_call(self, api_id: str, user_id: str, 
                       status_code: int, response_time: float,
                       revenue: float = 0):
        """Record API call metrics"""
        
        if not self.meter:
            return
        
        # Record request
        self.request_counter.add(1, {
            "api_id": api_id,
            "user_id": user_id,
            "status_code": str(status_code),
            "status_class": f"{status_code // 100}xx"
        })
        
        # Record error if applicable
        if status_code >= 400:
            self.error_counter.add(1, {
                "api_id": api_id,
                "status_code": str(status_code),
                "error_type": "client" if status_code < 500 else "server"
            })
        
        # Record business metrics
        self.business_metrics["api_calls"].add(1, {
            "api_id": api_id,
            "user_id": user_id
        })
        
        if revenue > 0:
            self.business_metrics["revenue"].add(revenue, {
                "api_id": api_id,
                "currency": "USD"
            })
        
        # Add custom span for tracking
        if self.tracer:
            with self.tracer.start_as_current_span("api_call_recorded") as span:
                span.set_attributes({
                    "api_id": api_id,
                    "user_id": user_id,
                    "status_code": status_code,
                    "response_time_ms": response_time * 1000,
                    "revenue": revenue
                })
    
    def record_trial_activation(self, api_id: str, user_id: str):
        """Record trial activation"""
        
        if self.business_metrics.get("trial_activations"):
            self.business_metrics["trial_activations"].add(1, {
                "api_id": api_id,
                "user_id": user_id
            })
    
    def update_active_users(self, delta: int):
        """Update active users count"""
        
        if self.business_metrics.get("active_users"):
            self.business_metrics["active_users"].add(delta)
    
    def create_child_span(self, name: str, attributes: Dict[str, Any] = None):
        """Create a child span under the current span"""
        
        if not self.tracer:
            return None
        
        span = self.tracer.start_span(name)
        if attributes:
            for key, value in attributes.items():
                span.set_attribute(key, str(value))
        
        return span
    
    def add_span_event(self, name: str, attributes: Dict[str, Any] = None):
        """Add an event to the current span"""
        
        current_span = trace.get_current_span()
        if current_span:
            current_span.add_event(name, attributes=attributes or {})
    
    def set_span_attribute(self, key: str, value: Any):
        """Set an attribute on the current span"""
        
        current_span = trace.get_current_span()
        if current_span:
            current_span.set_attribute(key, str(value))
    
    def get_trace_id(self) -> Optional[str]:
        """Get the current trace ID"""
        
        current_span = trace.get_current_span()
        if current_span:
            context = current_span.get_span_context()
            return format(context.trace_id, '032x') if context.trace_id else None
        return None


# Global APM instance
apm = APMManager()


# Middleware for FastAPI
async def apm_middleware(request, call_next):
    """APM middleware for FastAPI"""
    
    # Skip if APM not initialized
    if not apm.initialized:
        return await call_next(request)
    
    # Get or create trace context
    span = trace.get_current_span()
    
    # Add request attributes
    if span:
        span.set_attributes({
            "http.method": request.method,
            "http.url": str(request.url),
            "http.scheme": request.url.scheme,
            "http.host": request.url.hostname,
            "http.target": request.url.path,
            "http.user_agent": request.headers.get("user-agent", ""),
            "http.remote_addr": request.client.host if request.client else ""
        })
    
    # Process request
    response = await call_next(request)
    
    # Add response attributes
    if span:
        span.set_attribute("http.status_code", response.status_code)
    
    # Add trace ID to response headers
    trace_id = apm.get_trace_id()
    if trace_id:
        response.headers["X-Trace-ID"] = trace_id
    
    return response