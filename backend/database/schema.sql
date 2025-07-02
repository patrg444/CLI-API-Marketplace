-- API-Direct Database Schema
-- PostgreSQL 14+

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Users and Authentication
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    company VARCHAR(255),
    phone VARCHAR(50),
    bio TEXT,
    avatar_url VARCHAR(500),
    
    -- Account status
    email_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    is_premium BOOLEAN DEFAULT FALSE,
    
    -- Preferences
    default_deployment_type VARCHAR(20) DEFAULT 'hosted', -- 'hosted' or 'byoa'
    timezone VARCHAR(50) DEFAULT 'UTC',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- User API Keys for CLI authentication
CREATE TABLE user_api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) NOT NULL, -- bcrypt hash of the actual key
    name VARCHAR(100) NOT NULL,
    scopes JSONB DEFAULT '[]', -- Array of permitted operations
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Password Reset Tokens
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(100) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id)
);

-- Deployed APIs
CREATE TABLE apis (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- API Identity
    name VARCHAR(100) NOT NULL,
    description TEXT,
    version VARCHAR(20) DEFAULT '1.0.0',
    
    -- Deployment Configuration
    deployment_type VARCHAR(20) NOT NULL, -- 'hosted' or 'byoa'
    status VARCHAR(20) DEFAULT 'building', -- 'building', 'running', 'error', 'stopped'
    
    -- Hosting Details
    endpoint_url VARCHAR(500),
    custom_domain VARCHAR(255),
    
    -- Technical Configuration
    template_id VARCHAR(50), -- Reference to template used
    runtime_config JSONB DEFAULT '{}', -- Runtime settings, env vars, etc.
    scaling_config JSONB DEFAULT '{}', -- Auto-scaling settings
    
    -- Business Configuration
    pricing_model VARCHAR(20) DEFAULT 'per_request', -- 'per_request', 'subscription', 'free'
    price_per_request DECIMAL(10,6), -- Price in USD
    
    -- Marketplace
    is_public BOOLEAN DEFAULT FALSE,
    marketplace_category VARCHAR(50),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deployed_at TIMESTAMP WITH TIME ZONE,
    
    UNIQUE(user_id, name)
);

-- Deployment History
CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    api_id UUID NOT NULL REFERENCES apis(id) ON DELETE CASCADE,
    
    -- Deployment Details
    version VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL, -- 'pending', 'building', 'success', 'failed'
    deployment_method VARCHAR(20) NOT NULL, -- 'cli', 'web', 'api'
    
    -- Configuration at time of deployment
    config_snapshot JSONB,
    
    -- Build Information
    build_logs TEXT,
    build_duration_seconds INTEGER,
    
    -- Timestamps
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- API Call Logs (for billing and basic analytics)
CREATE TABLE api_calls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    api_id UUID NOT NULL REFERENCES apis(id) ON DELETE CASCADE,
    
    -- Request Details
    method VARCHAR(10) NOT NULL,
    path VARCHAR(500) NOT NULL,
    status_code INTEGER NOT NULL,
    
    -- Performance
    response_time_ms INTEGER NOT NULL,
    request_size_bytes INTEGER,
    response_size_bytes INTEGER,
    
    -- Client Information
    user_agent TEXT,
    ip_address INET,
    country_code VARCHAR(2),
    
    -- Billing
    billable BOOLEAN DEFAULT TRUE,
    amount_charged DECIMAL(10,6), -- Amount in USD
    
    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Partition api_calls by month for performance
CREATE TABLE api_calls_y2025m01 PARTITION OF api_calls
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE api_calls_y2025m02 PARTITION OF api_calls
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
-- Additional partitions would be created monthly

-- Billing and Revenue
CREATE TABLE billing_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_id UUID REFERENCES apis(id) ON DELETE SET NULL,
    
    -- Event Details
    event_type VARCHAR(50) NOT NULL, -- 'api_charge', 'subscription_fee', 'payout', 'refund'
    amount DECIMAL(10,2) NOT NULL, -- Amount in USD
    currency VARCHAR(3) DEFAULT 'USD',
    
    -- Stripe Integration
    stripe_charge_id VARCHAR(100),
    stripe_payout_id VARCHAR(100),
    
    -- Transaction Details
    description TEXT,
    metadata JSONB DEFAULT '{}',
    
    -- Status
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'completed', 'failed', 'cancelled'
    
    -- Platform Commission (for BYOA)
    platform_fee DECIMAL(10,2) DEFAULT 0,
    net_amount DECIMAL(10,2), -- Amount after fees
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE
);

-- User Subscriptions
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Subscription Details
    plan VARCHAR(50) NOT NULL, -- 'free', 'starter', 'pro', 'enterprise'
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'cancelled', 'past_due'
    
    -- Stripe Integration
    stripe_subscription_id VARCHAR(100) UNIQUE,
    stripe_customer_id VARCHAR(100),
    
    -- Billing
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    billing_interval VARCHAR(20) DEFAULT 'month', -- 'month', 'year'
    
    -- Timestamps
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Payout Settings
CREATE TABLE payout_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    
    -- Stripe Connect
    stripe_account_id VARCHAR(100) NOT NULL,
    account_verified BOOLEAN DEFAULT FALSE,
    
    -- Payout Schedule
    schedule VARCHAR(20) DEFAULT 'weekly', -- 'daily', 'weekly', 'monthly'
    minimum_amount DECIMAL(10,2) DEFAULT 50.00,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Marketplace Listings
CREATE TABLE marketplace_listings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    api_id UUID NOT NULL REFERENCES apis(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Listing Details
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(50) NOT NULL,
    tags TEXT DEFAULT '', -- Comma-separated tags
    features TEXT DEFAULT '', -- Comma-separated features
    use_cases TEXT DEFAULT '',
    
    -- Pricing
    pricing_model VARCHAR(20) DEFAULT 'per-call', -- 'free', 'freemium', 'per-call', 'subscription'
    price_per_call DECIMAL(10,6),
    
    -- Media
    logo_url VARCHAR(500),
    screenshots JSONB DEFAULT '[]',
    
    -- Metrics
    view_count INTEGER DEFAULT 0,
    user_count INTEGER DEFAULT 0,
    
    -- Status
    featured BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'pending', 'paused', 'rejected'
    
    -- Timestamps
    published_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(api_id)
);

-- Marketplace Reviews
CREATE TABLE marketplace_reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    listing_id UUID NOT NULL REFERENCES marketplace_listings(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Review Content
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(200),
    content TEXT,
    
    -- Moderation
    is_verified BOOLEAN DEFAULT FALSE,
    is_visible BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(listing_id, user_id)
);

-- System Configuration
CREATE TABLE system_config (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default system configuration
INSERT INTO system_config (key, value, description) VALUES
('platform_commission_rate', '0.20', 'Commission rate for BYOA deployments'),
('hosted_request_limits', '{"free": 1000, "starter": 100000, "pro": 1000000}', 'Monthly request limits by plan'),
('stripe_webhook_secret', '""', 'Stripe webhook endpoint secret'),
('marketplace_categories', '["AI/ML", "Data Processing", "Authentication", "Utilities", "Enterprise"]', 'Available marketplace categories');

-- Indexes for Performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_apis_user_id ON apis(user_id);
CREATE INDEX idx_apis_status ON apis(status);
CREATE INDEX idx_api_calls_api_id ON api_calls(api_id);
CREATE INDEX idx_api_calls_created_at ON api_calls(created_at);
CREATE INDEX idx_billing_events_user_id ON billing_events(user_id);
CREATE INDEX idx_billing_events_created_at ON billing_events(created_at);
CREATE INDEX idx_marketplace_listings_category ON marketplace_listings(category);
CREATE INDEX idx_marketplace_reviews_listing_id ON marketplace_reviews(listing_id);

-- Update timestamps trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply update triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_apis_updated_at BEFORE UPDATE ON apis
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_marketplace_listings_updated_at BEFORE UPDATE ON marketplace_listings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Views for Common Queries
CREATE VIEW user_api_summary AS
SELECT 
    u.id as user_id,
    u.email,
    u.name,
    COUNT(a.id) as total_apis,
    COUNT(CASE WHEN a.status = 'running' THEN 1 END) as running_apis,
    COUNT(CASE WHEN a.deployment_type = 'hosted' THEN 1 END) as hosted_apis,
    COUNT(CASE WHEN a.deployment_type = 'byoa' THEN 1 END) as byoa_apis
FROM users u
LEFT JOIN apis a ON u.id = a.user_id
GROUP BY u.id, u.email, u.name;

CREATE VIEW api_revenue_summary AS
SELECT 
    a.id as api_id,
    a.name,
    a.user_id,
    COUNT(ac.id) as total_calls,
    SUM(ac.amount_charged) as total_revenue,
    AVG(ac.response_time_ms) as avg_response_time,
    COUNT(CASE WHEN ac.status_code >= 400 THEN 1 END) as error_count
FROM apis a
LEFT JOIN api_calls ac ON a.id = ac.api_id
WHERE ac.created_at >= NOW() - INTERVAL '30 days'
GROUP BY a.id, a.name, a.user_id;