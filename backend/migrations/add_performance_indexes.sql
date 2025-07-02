-- Performance indexes for API-Direct database
-- Run this migration to improve query performance

-- API Keys table indexes
CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON user_api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON user_api_keys(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_key_prefix ON user_api_keys(SUBSTRING(key_hash, 1, 8));
CREATE INDEX IF NOT EXISTS idx_api_keys_last_used ON user_api_keys(last_used_at);

-- APIs table indexes
CREATE INDEX IF NOT EXISTS idx_apis_creator_id ON apis(creator_id);
CREATE INDEX IF NOT EXISTS idx_apis_status ON apis(status);
CREATE INDEX IF NOT EXISTS idx_apis_created_at ON apis(created_at);
CREATE INDEX IF NOT EXISTS idx_apis_creator_status ON apis(creator_id, status);
CREATE INDEX IF NOT EXISTS idx_apis_visibility ON apis(visibility) WHERE visibility = 'public';

-- Users table indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active) WHERE is_active = TRUE;

-- Subscriptions table indexes
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_api_id ON subscriptions(api_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_expires_at ON subscriptions(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_status ON subscriptions(user_id, status);

-- Usage metrics table indexes
CREATE INDEX IF NOT EXISTS idx_usage_api_id ON api_usage_metrics(api_id);
CREATE INDEX IF NOT EXISTS idx_usage_timestamp ON api_usage_metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_usage_api_timestamp ON api_usage_metrics(api_id, timestamp);

-- Billing records table indexes
CREATE INDEX IF NOT EXISTS idx_billing_user_id ON billing_records(user_id);
CREATE INDEX IF NOT EXISTS idx_billing_api_id ON billing_records(api_id);
CREATE INDEX IF NOT EXISTS idx_billing_period ON billing_records(billing_period);
CREATE INDEX IF NOT EXISTS idx_billing_status ON billing_records(status);

-- API documentation table indexes (for caching)
CREATE INDEX IF NOT EXISTS idx_api_docs_api_id ON api_documentation(api_id);
CREATE INDEX IF NOT EXISTS idx_api_docs_version ON api_documentation(version);
CREATE INDEX IF NOT EXISTS idx_api_docs_updated ON api_documentation(updated_at);

-- Marketplace search optimization
CREATE INDEX IF NOT EXISTS idx_apis_search ON apis USING gin(to_tsvector('english', name || ' ' || description));
CREATE INDEX IF NOT EXISTS idx_apis_tags ON apis USING gin(tags) WHERE tags IS NOT NULL;

-- Add composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_api_stats ON api_usage_metrics(api_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_user_apis_active ON apis(creator_id, status, created_at DESC) WHERE status = 'active';

-- Function to analyze index usage
CREATE OR REPLACE FUNCTION analyze_index_usage() RETURNS TABLE (
    schemaname text,
    tablename text,
    indexname text,
    index_size text,
    idx_scan bigint,
    idx_tup_read bigint,
    idx_tup_fetch bigint
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        s.schemaname,
        s.tablename,
        s.indexname,
        pg_size_pretty(pg_relation_size(s.indexrelid)) as index_size,
        s.idx_scan,
        s.idx_tup_read,
        s.idx_tup_fetch
    FROM pg_stat_user_indexes s
    ORDER BY s.idx_scan DESC;
END;
$$ LANGUAGE plpgsql;

-- Query to identify missing indexes (run periodically)
CREATE OR REPLACE VIEW missing_indexes AS
SELECT 
    schemaname,
    tablename,
    attname,
    n_distinct,
    correlation
FROM pg_stats
WHERE schemaname = 'public'
AND n_distinct > 100
AND correlation < 0.1
ORDER BY n_distinct DESC;