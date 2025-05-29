-- Migration: Enhanced Review System
-- Version: 004
-- Description: Add review system enhancements for Phase 2

-- Update api_reviews table with new fields
ALTER TABLE api_reviews 
ADD COLUMN IF NOT EXISTS helpful_votes INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_votes INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS creator_response TEXT,
ADD COLUMN IF NOT EXISTS response_date TIMESTAMP,
ADD COLUMN IF NOT EXISTS is_verified_purchase BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS title VARCHAR(255);

-- Review votes tracking table
CREATE TABLE IF NOT EXISTS review_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL REFERENCES api_reviews(id) ON DELETE CASCADE,
    consumer_id UUID NOT NULL REFERENCES consumers(id) ON DELETE CASCADE,
    is_helpful BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(review_id, consumer_id)
);

-- Create materialized view for API rating statistics
CREATE MATERIALIZED VIEW IF NOT EXISTS api_rating_stats AS
SELECT 
    api_id,
    COUNT(*) as total_reviews,
    AVG(rating)::NUMERIC(3,2) as average_rating,
    COUNT(CASE WHEN rating = 5 THEN 1 END) as five_star_count,
    COUNT(CASE WHEN rating = 4 THEN 1 END) as four_star_count,
    COUNT(CASE WHEN rating = 3 THEN 1 END) as three_star_count,
    COUNT(CASE WHEN rating = 2 THEN 1 END) as two_star_count,
    COUNT(CASE WHEN rating = 1 THEN 1 END) as one_star_count
FROM api_reviews
GROUP BY api_id;

-- Create index for materialized view
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_rating_stats_api_id ON api_rating_stats(api_id);

-- Function to refresh rating stats
CREATE OR REPLACE FUNCTION refresh_api_rating_stats()
RETURNS TRIGGER AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY api_rating_stats;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger to refresh stats on review changes
DROP TRIGGER IF EXISTS refresh_ratings_on_review ON api_reviews;
CREATE TRIGGER refresh_ratings_on_review
AFTER INSERT OR UPDATE OR DELETE ON api_reviews
FOR EACH STATEMENT EXECUTE FUNCTION refresh_api_rating_stats();

-- Update apis table to include rating stats (denormalized for performance)
ALTER TABLE apis 
ADD COLUMN IF NOT EXISTS average_rating NUMERIC(3,2),
ADD COLUMN IF NOT EXISTS total_reviews INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_subscriptions INTEGER DEFAULT 0;

-- Function to update API stats
CREATE OR REPLACE FUNCTION update_api_stats()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        UPDATE apis 
        SET 
            average_rating = (SELECT average_rating FROM api_rating_stats WHERE api_id = NEW.api_id),
            total_reviews = (SELECT total_reviews FROM api_rating_stats WHERE api_id = NEW.api_id)
        WHERE id = NEW.api_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE apis 
        SET 
            average_rating = (SELECT average_rating FROM api_rating_stats WHERE api_id = OLD.api_id),
            total_reviews = (SELECT total_reviews FROM api_rating_stats WHERE api_id = OLD.api_id)
        WHERE id = OLD.api_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update API stats on review changes
DROP TRIGGER IF EXISTS update_api_stats_on_review ON api_reviews;
CREATE TRIGGER update_api_stats_on_review
AFTER INSERT OR UPDATE OR DELETE ON api_reviews
FOR EACH ROW EXECUTE FUNCTION update_api_stats();

-- Function to update subscription count
CREATE OR REPLACE FUNCTION update_api_subscription_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE apis 
        SET total_subscriptions = total_subscriptions + 1
        WHERE id = NEW.api_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE apis 
        SET total_subscriptions = total_subscriptions - 1
        WHERE id = OLD.api_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update subscription count
DROP TRIGGER IF EXISTS update_api_subscription_count_trigger ON subscriptions;
CREATE TRIGGER update_api_subscription_count_trigger
AFTER INSERT OR DELETE ON subscriptions
FOR EACH ROW 
WHEN (NEW.status = 'active' OR OLD.status = 'active')
EXECUTE FUNCTION update_api_subscription_count();

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_api_reviews_api_id ON api_reviews(api_id);
CREATE INDEX IF NOT EXISTS idx_api_reviews_consumer_id ON api_reviews(consumer_id);
CREATE INDEX IF NOT EXISTS idx_api_reviews_created_at ON api_reviews(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_review_votes_review_id ON review_votes(review_id);

-- Initial refresh of materialized view
REFRESH MATERIALIZED VIEW api_rating_stats;
