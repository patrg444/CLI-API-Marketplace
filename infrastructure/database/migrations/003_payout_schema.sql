-- Migration: Add payout tracking tables for creator earnings
-- This migration adds support for Stripe Connect and payout management

-- Creator payment accounts (Stripe Connect)
CREATE TABLE IF NOT EXISTS creator_payment_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    creator_id UUID NOT NULL UNIQUE,
    stripe_account_id VARCHAR(255) UNIQUE,
    account_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    -- Status: pending, onboarding, active, restricted, disabled
    
    -- Onboarding details
    details_submitted BOOLEAN DEFAULT FALSE,
    charges_enabled BOOLEAN DEFAULT FALSE,
    payouts_enabled BOOLEAN DEFAULT FALSE,
    
    -- Account settings
    default_currency VARCHAR(3) DEFAULT 'USD',
    country VARCHAR(2),
    business_type VARCHAR(50),
    
    -- Metadata
    onboarding_completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_creator_payment_creator
        FOREIGN KEY (creator_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE
);

-- Payout records
CREATE TABLE IF NOT EXISTS payouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    creator_id UUID NOT NULL,
    stripe_payout_id VARCHAR(255) UNIQUE,
    
    -- Financial details
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    platform_fee DECIMAL(10, 2) NOT NULL, -- 20% commission
    net_amount DECIMAL(10, 2) NOT NULL, -- Amount after commission
    
    -- Period covered
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    
    -- Status tracking
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    -- Status: pending, processing, paid, failed, cancelled
    
    -- Stripe details
    arrival_date DATE,
    stripe_failure_code VARCHAR(100),
    stripe_failure_message TEXT,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    paid_at TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_payout_creator
        FOREIGN KEY (creator_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE
);

-- Payout line items (detailed breakdown)
CREATE TABLE IF NOT EXISTS payout_line_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payout_id UUID NOT NULL,
    api_id UUID NOT NULL,
    
    -- Revenue details
    gross_revenue DECIMAL(10, 2) NOT NULL,
    platform_fee DECIMAL(10, 2) NOT NULL,
    net_revenue DECIMAL(10, 2) NOT NULL,
    
    -- Usage summary
    total_subscriptions INTEGER DEFAULT 0,
    total_api_calls BIGINT DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_payout_item_payout
        FOREIGN KEY (payout_id) 
        REFERENCES payouts(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_payout_item_api
        FOREIGN KEY (api_id) 
        REFERENCES apis(id) 
        ON DELETE CASCADE
);

-- Creator earnings summary (real-time view)
CREATE TABLE IF NOT EXISTS creator_earnings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    creator_id UUID NOT NULL,
    api_id UUID NOT NULL,
    
    -- Current month earnings
    current_month_gross DECIMAL(10, 2) DEFAULT 0,
    current_month_net DECIMAL(10, 2) DEFAULT 0,
    
    -- Lifetime earnings
    lifetime_gross DECIMAL(10, 2) DEFAULT 0,
    lifetime_net DECIMAL(10, 2) DEFAULT 0,
    lifetime_payouts DECIMAL(10, 2) DEFAULT 0,
    
    -- Pending balance
    pending_payout DECIMAL(10, 2) DEFAULT 0,
    
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(creator_id, api_id),
    
    CONSTRAINT fk_earnings_creator
        FOREIGN KEY (creator_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_earnings_api
        FOREIGN KEY (api_id) 
        REFERENCES apis(id) 
        ON DELETE CASCADE
);

-- Platform revenue tracking
CREATE TABLE IF NOT EXISTS platform_revenue (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    month DATE NOT NULL UNIQUE,
    
    -- Revenue metrics
    total_gross_revenue DECIMAL(12, 2) DEFAULT 0,
    total_platform_fees DECIMAL(12, 2) DEFAULT 0,
    total_creator_payouts DECIMAL(12, 2) DEFAULT 0,
    
    -- Transaction counts
    total_transactions BIGINT DEFAULT 0,
    total_api_calls BIGINT DEFAULT 0,
    active_subscriptions INTEGER DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_creator_payment_accounts_creator ON creator_payment_accounts(creator_id);
CREATE INDEX idx_creator_payment_accounts_status ON creator_payment_accounts(account_status);
CREATE INDEX idx_payouts_creator ON payouts(creator_id);
CREATE INDEX idx_payouts_status ON payouts(status);
CREATE INDEX idx_payouts_period ON payouts(period_start, period_end);
CREATE INDEX idx_payout_items_payout ON payout_line_items(payout_id);
CREATE INDEX idx_creator_earnings_creator ON creator_earnings(creator_id);
CREATE INDEX idx_platform_revenue_month ON platform_revenue(month);

-- Function to update timestamps
CREATE OR REPLACE FUNCTION update_payout_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
CREATE TRIGGER update_creator_payment_accounts_updated_at
    BEFORE UPDATE ON creator_payment_accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payouts_updated_at
    BEFORE UPDATE ON payouts
    FOR EACH ROW
    EXECUTE FUNCTION update_payout_updated_at();

-- View for creator dashboard
CREATE OR REPLACE VIEW creator_revenue_dashboard AS
SELECT 
    ce.creator_id,
    ce.api_id,
    a.name as api_name,
    ce.current_month_gross,
    ce.current_month_net,
    ce.lifetime_gross,
    ce.lifetime_net,
    ce.pending_payout,
    ce.last_updated
FROM creator_earnings ce
JOIN apis a ON ce.api_id = a.id;

-- View for platform analytics
CREATE OR REPLACE VIEW platform_revenue_analytics AS
SELECT 
    month,
    total_gross_revenue,
    total_platform_fees,
    total_creator_payouts,
    total_transactions,
    total_api_calls,
    active_subscriptions,
    ROUND((total_platform_fees / NULLIF(total_gross_revenue, 0)) * 100, 2) as platform_fee_percentage
FROM platform_revenue
ORDER BY month DESC;
