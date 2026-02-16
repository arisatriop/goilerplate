-- Migration: create_subscriptions
-- Created at: 2025-10-31T14:30:00+07:00

CREATE TABLE subscriptions (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()) COMMENT 'Unique identifier for the subscription',
    store_id CHAR(36) NOT NULL COMMENT 'Reference to store',
    plan_id CHAR(36) NOT NULL COMMENT 'Reference to plan',
    start_date TIMESTAMP COMMENT 'Subscription start date',
    end_date TIMESTAMP COMMENT 'Subscription end date',
    price DECIMAL(10,2) NOT NULL COMMENT 'Subscription price',
    status VARCHAR(50) NOT NULL COMMENT 'Subscription status (active, expired, cancelled, etc.)',
    is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'Whether subscription is active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when subscription was created',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this subscription',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Timestamp when subscription was last updated',
    updated_by VARCHAR(255) NOT NULL COMMENT 'User who last updated this subscription',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT 'Timestamp when subscription was soft deleted',
    deleted_by VARCHAR(255) DEFAULT NULL COMMENT 'User who deleted this subscription',
    FOREIGN KEY (store_id) REFERENCES stores(id),
    FOREIGN KEY (plan_id) REFERENCES plans(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Subscriptions table for managing store subscription plans';

-- Indexes for subscriptions
CREATE INDEX idx_subscriptions_store_id ON subscriptions(store_id);
CREATE INDEX idx_subscriptions_plan_id ON subscriptions(plan_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_is_active ON subscriptions(is_active);
CREATE INDEX idx_subscriptions_deleted_at ON subscriptions(deleted_at);
CREATE INDEX idx_subscriptions_dates ON subscriptions(start_date, end_date);

-- Composite index for active subscriptions lookup
CREATE INDEX idx_subscriptions_active ON subscriptions(store_id, plan_id, is_active, deleted_at);
