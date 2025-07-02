"""
Database module for API-Direct backend
Provides database connection management, models, and operations
"""

import os
import asyncio
import json
from contextlib import asynccontextmanager
from datetime import datetime
from typing import Optional, Dict, Any, List
import logging
from sqlalchemy import (
    Column, String, Integer, Boolean, DateTime, ForeignKey, 
    Text, DECIMAL, JSON, Enum, UniqueConstraint, Index,
    text, create_engine
)
from sqlalchemy.dialects.postgresql import ARRAY as PG_ARRAY
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession, async_sessionmaker
from sqlalchemy.orm import declarative_base, relationship, sessionmaker
from sqlalchemy.pool import NullPool, QueuePool
from sqlalchemy.dialects.postgresql import UUID as pgUUID, INET
from sqlalchemy.exc import IntegrityError, OperationalError
import sqlalchemy.types as types
import uuid
import asyncpg

# Create a type that works for both PostgreSQL and SQLite
class UUID(types.TypeDecorator):
    """Platform-independent UUID type"""
    impl = types.CHAR(32)
    cache_ok = True

    def process_bind_param(self, value, dialect):
        if value is None:
            return value
        elif dialect.name == 'postgresql':
            return str(value)
        else:
            return str(value).replace('-', '')

    def process_result_value(self, value, dialect):
        if value is None:
            return value
        else:
            if '-' not in value:
                # Insert dashes for SQLite
                value = f"{value[:8]}-{value[8:12]}-{value[12:16]}-{value[16:20]}-{value[20:]}"
            return uuid.UUID(value)


class ARRAY(types.TypeDecorator):
    """Platform-independent ARRAY type that uses JSON for SQLite"""
    impl = types.JSON
    cache_ok = True

    def load_dialect_impl(self, dialect):
        if dialect.name == 'postgresql':
            return dialect.type_descriptor(PG_ARRAY(String))
        else:
            return dialect.type_descriptor(types.JSON())

    def process_bind_param(self, value, dialect):
        if value is None:
            return value
        if dialect.name == 'postgresql':
            return value
        else:
            return json.dumps(value) if not isinstance(value, str) else value

    def process_result_value(self, value, dialect):
        if value is None:
            return value
        if dialect.name == 'postgresql':
            return value
        else:
            return json.loads(value) if isinstance(value, str) else value

# Setup logging
logger = logging.getLogger(__name__)

# Base for all models
Base = declarative_base()

# Custom exceptions
class DeadlockError(Exception):
    """Raised when a database deadlock is detected"""
    pass


class User(Base):
    """User model"""
    __tablename__ = 'users'
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    email = Column(String(255), unique=True, nullable=False)
    username = Column(String(100), unique=True, nullable=False)
    password_hash = Column(String(255), nullable=False)
    name = Column(String(255))
    company = Column(String(255))
    phone = Column(String(50))
    bio = Column(Text)
    avatar_url = Column(String(500))
    
    # Account status
    email_verified = Column(Boolean, default=False)
    is_active = Column(Boolean, default=True)
    is_premium = Column(Boolean, default=False)
    
    # Preferences
    default_deployment_type = Column(String(20), default='hosted')
    timezone = Column(String(50), default='UTC')
    
    # Timestamps
    created_at = Column(DateTime(timezone=True), default=datetime.utcnow)
    updated_at = Column(DateTime(timezone=True), default=datetime.utcnow, onupdate=datetime.utcnow)
    last_login_at = Column(DateTime(timezone=True))
    
    # Relationships
    apis = relationship("API", back_populates="owner", cascade="all, delete-orphan", 
                       foreign_keys="API.owner_id")
    subscriptions = relationship("Subscription", back_populates="user", cascade="all, delete-orphan")


class API(Base):
    """API model"""
    __tablename__ = 'apis'
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), ForeignKey('users.id', ondelete='CASCADE'), nullable=False)
    owner_id = Column(UUID(as_uuid=True), ForeignKey('users.id', ondelete='CASCADE'), nullable=False)
    
    # API Identity
    name = Column(String(100), nullable=False)
    description = Column(Text)
    version = Column(String(20), default='1.0.0')
    
    # Deployment Configuration
    deployment_type = Column(String(20), nullable=False, default='hosted')
    status = Column(String(20), default='building')
    
    # Hosting Details
    endpoint_url = Column(String(500))
    custom_domain = Column(String(255))
    
    # Technical Configuration
    template_id = Column(String(50))
    runtime_config = Column(JSON, default={})
    scaling_config = Column(JSON, default={})
    
    # Business Configuration
    pricing_model = Column(String(20), default='per_request')
    pricing_type = Column(String(20), default='freemium')  # Added for test compatibility
    price_per_request = Column(DECIMAL(10, 6))
    
    # Marketplace
    is_public = Column(Boolean, default=False)
    marketplace_category = Column(String(50))
    
    # Timestamps
    created_at = Column(DateTime(timezone=True), default=datetime.utcnow)
    updated_at = Column(DateTime(timezone=True), default=datetime.utcnow, onupdate=datetime.utcnow)
    deployed_at = Column(DateTime(timezone=True))
    
    # Relationships
    owner = relationship("User", back_populates="apis", foreign_keys=[owner_id])
    subscriptions = relationship("Subscription", back_populates="api", cascade="all, delete-orphan")
    versions = relationship("APIVersion", back_populates="api", cascade="all, delete-orphan")
    
    # Constraints
    __table_args__ = (
        UniqueConstraint('user_id', 'name', name='_user_api_uc'),
    )


class APIVersion(Base):
    """API Version model"""
    __tablename__ = 'api_versions'
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    api_id = Column(UUID(as_uuid=True), ForeignKey('apis.id', ondelete='CASCADE'), nullable=False)
    
    # Version Info
    version_number = Column(String(20), nullable=False)
    version_type = Column(String(20), default='draft')  # draft, beta, stable
    is_active = Column(Boolean, default=False)
    
    # Release Notes
    release_notes = Column(Text)
    breaking_changes = Column(ARRAY(String), default=[])
    
    # Timestamps
    created_at = Column(DateTime(timezone=True), default=datetime.utcnow)
    published_at = Column(DateTime(timezone=True))
    deprecated_at = Column(DateTime(timezone=True))
    sunset_date = Column(DateTime(timezone=True))
    
    # API Specification
    openapi_spec = Column(JSON)
    endpoints_config = Column(JSON)
    
    # Relationships
    api = relationship("API", back_populates="versions")
    
    # Constraints
    __table_args__ = (
        UniqueConstraint('api_id', 'version_number', name='_api_version_uc'),
        Index('idx_api_version_active', 'api_id', 'is_active'),
    )


class Subscription(Base):
    """Subscription model"""
    __tablename__ = 'subscriptions'
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), ForeignKey('users.id', ondelete='CASCADE'), nullable=False)
    api_id = Column(UUID(as_uuid=True), ForeignKey('apis.id', ondelete='CASCADE'))
    
    # Subscription Details
    plan = Column(String(50), nullable=False)
    status = Column(String(20), default='active')
    
    # Stripe Integration
    stripe_subscription_id = Column(String(100), unique=True)
    stripe_customer_id = Column(String(100))
    
    # Billing
    amount = Column(DECIMAL(10, 2), nullable=False)
    currency = Column(String(3), default='USD')
    billing_interval = Column(String(20), default='month')
    
    # Timestamps
    current_period_start = Column(DateTime(timezone=True))
    current_period_end = Column(DateTime(timezone=True))
    created_at = Column(DateTime(timezone=True), default=datetime.utcnow)
    updated_at = Column(DateTime(timezone=True), default=datetime.utcnow, onupdate=datetime.utcnow)
    
    # Relationships
    user = relationship("User", back_populates="subscriptions")
    api = relationship("API", back_populates="subscriptions")


class Transaction(Base):
    """Transaction/Billing event model"""
    __tablename__ = 'billing_events'
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), ForeignKey('users.id', ondelete='CASCADE'), nullable=False)
    api_id = Column(UUID(as_uuid=True), ForeignKey('apis.id', ondelete='SET NULL'))
    
    # Event Details
    event_type = Column(String(50), nullable=False)
    amount = Column(DECIMAL(10, 2), nullable=False)
    currency = Column(String(3), default='USD')
    
    # Stripe Integration
    stripe_charge_id = Column(String(100))
    stripe_payout_id = Column(String(100))
    
    # Transaction Details
    description = Column(Text)
    transaction_metadata = Column('metadata', JSON, default={})
    
    # Status
    status = Column(String(20), default='pending')
    
    # Platform Commission
    platform_fee = Column(DECIMAL(10, 2), default=0)
    net_amount = Column(DECIMAL(10, 2))
    
    # Timestamps
    created_at = Column(DateTime(timezone=True), default=datetime.utcnow)
    processed_at = Column(DateTime(timezone=True))


class Webhook(Base):
    """Webhook subscription model"""
    __tablename__ = 'webhooks'
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), ForeignKey('users.id', ondelete='CASCADE'), nullable=False)
    
    # Webhook Configuration
    url = Column(String(500), nullable=False)
    events = Column(ARRAY(String), nullable=False)
    description = Column(Text)
    secret = Column(String(100), nullable=False)
    
    # Headers and settings
    headers = Column(JSON, default={})
    retry_enabled = Column(Boolean, default=True)
    max_retries = Column(Integer, default=3)
    timeout_seconds = Column(Integer, default=30)
    
    # Status
    status = Column(String(20), default='active')
    failure_count = Column(Integer, default=0)
    success_count = Column(Integer, default=0)
    last_triggered_at = Column(DateTime(timezone=True))
    
    # Timestamps
    created_at = Column(DateTime(timezone=True), default=datetime.utcnow)
    updated_at = Column(DateTime(timezone=True), default=datetime.utcnow, onupdate=datetime.utcnow)
    
    # Relationships
    deliveries = relationship("WebhookDelivery", back_populates="webhook", cascade="all, delete-orphan")
    
    # Indexes
    __table_args__ = (
        Index('idx_webhook_user_status', 'user_id', 'status'),
    )


class WebhookEvent(Base):
    """Webhook event record"""
    __tablename__ = 'webhook_events'
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    event_type = Column(String(50), nullable=False)
    user_id = Column(UUID(as_uuid=True), ForeignKey('users.id', ondelete='CASCADE'), nullable=False)
    api_id = Column(UUID(as_uuid=True), ForeignKey('apis.id', ondelete='CASCADE'))
    
    # Event data
    payload = Column(JSON, nullable=False)
    
    # Timestamps
    created_at = Column(DateTime(timezone=True), default=datetime.utcnow)
    
    # Relationships
    deliveries = relationship("WebhookDelivery", back_populates="event", cascade="all, delete-orphan")
    
    # Indexes
    __table_args__ = (
        Index('idx_webhook_event_user_type', 'user_id', 'event_type'),
        Index('idx_webhook_event_created', 'created_at'),
    )


class WebhookDelivery(Base):
    """Webhook delivery attempt record"""
    __tablename__ = 'webhook_deliveries'
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    webhook_id = Column(UUID(as_uuid=True), ForeignKey('webhooks.id', ondelete='CASCADE'), nullable=False)
    event_id = Column(UUID(as_uuid=True), ForeignKey('webhook_events.id', ondelete='CASCADE'), nullable=False)
    
    # Delivery details
    event_type = Column(String(50), nullable=False)
    payload = Column(JSON, nullable=False)
    status = Column(String(20), default='pending')
    attempt_count = Column(Integer, default=0)
    
    # Response data
    response_status = Column(Integer)
    response_body = Column(Text)
    error_message = Column(Text)
    
    # Timestamps
    created_at = Column(DateTime(timezone=True), default=datetime.utcnow)
    delivered_at = Column(DateTime(timezone=True))
    next_retry_at = Column(DateTime(timezone=True))
    
    # Relationships
    webhook = relationship("Webhook", back_populates="deliveries")
    event = relationship("WebhookEvent", back_populates="deliveries")
    
    # Indexes
    __table_args__ = (
        Index('idx_delivery_webhook_status', 'webhook_id', 'status'),
        Index('idx_delivery_retry', 'status', 'next_retry_at'),
    )


class DatabaseManager:
    """Database connection and session manager"""
    
    def __init__(self, connection_string: str = None, pool_size: int = 20, max_overflow: int = 40):
        self.connection_string = connection_string or os.getenv(
            'DATABASE_URL', 
            'postgresql://user:password@localhost:5432/api_marketplace'
        )
        self.pool_size = pool_size
        self.max_overflow = max_overflow
        self.engine = None
        self.async_session_maker = None
        
    async def initialize(self):
        """Initialize the database engine and session maker"""
        # For SQLite testing, use NullPool
        if 'sqlite' in self.connection_string:
            self.engine = create_async_engine(
                self.connection_string,
                poolclass=NullPool,
                echo=False
            )
        else:
            # For PostgreSQL, use QueuePool
            self.engine = create_async_engine(
                self.connection_string,
                pool_size=self.pool_size,
                max_overflow=self.max_overflow,
                pool_pre_ping=True,
                pool_recycle=3600,  # Recycle connections after 1 hour
                echo=False
            )
        
        self.async_session_maker = async_sessionmaker(
            self.engine,
            class_=AsyncSession,
            expire_on_commit=False
        )
    
    async def create_tables(self):
        """Create all tables (for testing)"""
        if self.engine:
            async with self.engine.begin() as conn:
                await conn.run_sync(Base.metadata.create_all)
    
    async def close(self):
        """Close the database engine"""
        if self.engine:
            await self.engine.dispose()
    
    @asynccontextmanager
    async def get_session(self) -> AsyncSession:
        """Get a database session"""
        if not self.async_session_maker:
            await self.initialize()
            
        async with self.async_session_maker() as session:
            try:
                yield session
                await session.commit()
            except Exception:
                await session.rollback()
                raise
            finally:
                await session.close()
    
    async def _create_connection(self):
        """Create a raw database connection (for PostgreSQL specific operations)"""
        if 'postgresql' in self.connection_string:
            return await asyncpg.connect(self.connection_string)
        return None
    
    async def connect_with_retry(self, max_retries: int = 3, retry_delay: float = 1.0):
        """Connect to database with retry logic"""
        retry_count = 0
        last_error = None
        
        while retry_count < max_retries:
            try:
                connection = await self._create_connection()
                return connection
            except (asyncpg.PostgresConnectionError, OperationalError) as e:
                retry_count += 1
                last_error = e
                logger.warning(f"Database connection attempt {retry_count} failed: {e}")
                if retry_count < max_retries:
                    await asyncio.sleep(retry_delay * retry_count)
        
        raise last_error
    
    async def health_check(self) -> Dict[str, Any]:
        """Check database health"""
        start_time = asyncio.get_event_loop().time()
        
        try:
            async with self.get_session() as session:
                # Simple query to check connection
                result = await session.execute(text("SELECT 1"))
                result.scalar()
            
            end_time = asyncio.get_event_loop().time()
            response_time_ms = (end_time - start_time) * 1000
            
            pool_status = self.get_pool_metrics()
            
            return {
                'status': 'healthy',
                'response_time_ms': response_time_ms,
                'connection_count': pool_status.get('total', 0),
                'pool_size': pool_status.get('size', 0),
                **pool_status
            }
        except Exception as e:
            logger.error(f"Database health check failed: {e}")
            return {
                'status': 'unhealthy',
                'error': str(e),
                'response_time_ms': -1
            }
    
    def get_pool_metrics(self) -> Dict[str, int]:
        """Get connection pool metrics"""
        if not self.engine or not hasattr(self.engine.pool, 'size'):
            return {
                'size': 0,
                'checked_in': 0,
                'checked_out': 0,
                'overflow': 0,
                'total': 0
            }
        
        pool = self.engine.pool
        return {
            'size': pool.size() if hasattr(pool, 'size') else 0,
            'checked_in': pool.checkedin() if hasattr(pool, 'checkedin') else 0,
            'checked_out': pool.checkedout() if hasattr(pool, 'checkedout') else 0,
            'overflow': pool.overflow() if hasattr(pool, 'overflow') else 0,
            'total': pool.total() if hasattr(pool, 'total') else 0
        }


# Global database manager instance
_db_manager = None


def get_db_manager() -> DatabaseManager:
    """Get the global database manager instance"""
    global _db_manager
    if _db_manager is None:
        _db_manager = DatabaseManager()
    return _db_manager


async def get_db() -> AsyncSession:
    """Dependency to get database session"""
    db_manager = get_db_manager()
    if not db_manager.engine:
        await db_manager.initialize()
    
    async with db_manager.get_session() as session:
        yield session