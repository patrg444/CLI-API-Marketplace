-- API-Direct Hosted Platform Database Initialization
-- Creates multi-tenant database structure for hosted APIs

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Hosted deployments table
CREATE TABLE hosted_deployments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    api_name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    image_tag VARCHAR(255),
    database_name VARCHAR(255),
    database_user VARCHAR(255),
    database_password VARCHAR(255),
    resource_limits JSONB,
    environment_vars JSONB,
    custom_domain VARCHAR(255),
    ssl_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deployed_at TIMESTAMP,
    UNIQUE(user_id, api_name)
);

-- User quotas and billing
CREATE TABLE user_quotas (
    user_id VARCHAR(255) PRIMARY KEY,
    plan VARCHAR(50) NOT NULL DEFAULT 'free',
    max_apis INTEGER NOT NULL DEFAULT 1,
    max_requests_per_month BIGINT NOT NULL DEFAULT 1000,
    max_cpu_cores DECIMAL(3,2) NOT NULL DEFAULT 0.25,
    max_memory_gb DECIMAL(4,2) NOT NULL DEFAULT 0.5,
    current_apis INTEGER NOT NULL DEFAULT 0,
    current_requests_this_month BIGINT NOT NULL DEFAULT 0,
    billing_cycle_start DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- API usage tracking for billing
CREATE TABLE api_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    deployment_id UUID NOT NULL REFERENCES hosted_deployments(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code INTEGER NOT NULL,
    response_time_ms INTEGER,
    request_size_bytes INTEGER,
    response_size_bytes INTEGER,
    billable BOOLEAN DEFAULT true
);

-- Container images registry
CREATE TABLE container_images (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    deployment_id UUID NOT NULL REFERENCES hosted_deployments(id) ON DELETE CASCADE,
    image_tag VARCHAR(255) NOT NULL,
    size_bytes BIGINT,
    build_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    build_log TEXT,
    dockerfile TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Database schemas for user APIs (metadata only, actual DBs created dynamically)
CREATE TABLE user_databases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    deployment_id UUID NOT NULL REFERENCES hosted_deployments(id) ON DELETE CASCADE,
    database_name VARCHAR(255) NOT NULL UNIQUE,
    database_user VARCHAR(255) NOT NULL,
    size_mb DECIMAL(10,2) DEFAULT 0,
    connection_limit INTEGER DEFAULT 20,
    backup_enabled BOOLEAN DEFAULT true,
    last_backup TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- SSL certificates management
CREATE TABLE ssl_certificates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    domain VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    certificate_pem TEXT,
    private_key_pem TEXT,
    expires_at TIMESTAMP,
    auto_renew BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_hosted_deployments_user_id ON hosted_deployments(user_id);
CREATE INDEX idx_hosted_deployments_subdomain ON hosted_deployments(subdomain);
CREATE INDEX idx_hosted_deployments_status ON hosted_deployments(status);
CREATE INDEX idx_api_usage_deployment_id ON api_usage(deployment_id);
CREATE INDEX idx_api_usage_timestamp ON api_usage(timestamp);
CREATE INDEX idx_api_usage_user_billing ON api_usage(user_id, timestamp) WHERE billable = true;

-- Functions for quota management
CREATE OR REPLACE FUNCTION check_user_quota(p_user_id VARCHAR, p_plan VARCHAR DEFAULT 'free')
RETURNS BOOLEAN AS $$
DECLARE
    current_count INTEGER;
    max_allowed INTEGER;
BEGIN
    -- Get current API count
    SELECT COUNT(*) INTO current_count
    FROM hosted_deployments 
    WHERE user_id = p_user_id AND status = 'running';
    
    -- Get max allowed based on plan
    SELECT max_apis INTO max_allowed
    FROM user_quotas 
    WHERE user_id = p_user_id;
    
    -- If no quota record exists, create one
    IF max_allowed IS NULL THEN
        INSERT INTO user_quotas (user_id, plan) VALUES (p_user_id, p_plan);
        max_allowed := CASE p_plan 
            WHEN 'free' THEN 1
            WHEN 'starter' THEN 5
            WHEN 'pro' THEN 100
            ELSE 1000
        END;
    END IF;
    
    RETURN current_count < max_allowed;
END;
$$ LANGUAGE plpgsql;

-- Function to generate unique subdomain
CREATE OR REPLACE FUNCTION generate_subdomain(p_api_name VARCHAR, p_user_id VARCHAR)
RETURNS VARCHAR AS $$
DECLARE
    base_name VARCHAR := lower(regexp_replace(p_api_name, '[^a-zA-Z0-9]', '-', 'g'));
    user_suffix VARCHAR := lower(substring(p_user_id from 1 for 6));
    subdomain VARCHAR;
    counter INTEGER := 0;
BEGIN
    LOOP
        IF counter = 0 THEN
            subdomain := base_name || '-' || user_suffix;
        ELSE
            subdomain := base_name || '-' || user_suffix || '-' || counter::TEXT;
        END IF;
        
        -- Check if subdomain is available
        IF NOT EXISTS (SELECT 1 FROM hosted_deployments WHERE subdomain = subdomain) THEN
            RETURN subdomain;
        END IF;
        
        counter := counter + 1;
        
        -- Prevent infinite loop
        IF counter > 100 THEN
            RETURN base_name || '-' || user_suffix || '-' || extract(epoch from now())::INTEGER::TEXT;
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update timestamps
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_hosted_deployments_timestamp
    BEFORE UPDATE ON hosted_deployments
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER update_user_quotas_timestamp
    BEFORE UPDATE ON user_quotas
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- Insert default quota plans
INSERT INTO user_quotas (user_id, plan, max_apis, max_requests_per_month, max_cpu_cores, max_memory_gb) VALUES
('demo-user', 'free', 1, 1000, 0.25, 0.5),
('test-user', 'starter', 5, 100000, 1.0, 2.0),
('pro-user', 'pro', 100, 1000000, 4.0, 8.0)
ON CONFLICT (user_id) DO NOTHING;

-- Create sample deployment for testing
INSERT INTO hosted_deployments (
    user_id, 
    api_name, 
    subdomain, 
    status, 
    database_name,
    database_user,
    resource_limits,
    environment_vars
) VALUES (
    'demo-user',
    'sentiment-analyzer',
    'sentiment-analyzer-demo01',
    'running',
    'api_demo_user_sentiment_analyzer',
    'api_demo_user_sentiment',
    '{"cpu": "250m", "memory": "512Mi"}',
    '{"MODEL_NAME": "cardiffnlp/twitter-roberta-base-sentiment-latest", "LOG_LEVEL": "INFO"}'
) ON CONFLICT (user_id, api_name) DO NOTHING;