-- Migration: create_subscription_payments
-- Created at: 2025-10-31T14:31:00+07:00

CREATE TABLE subscription_payments (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()) COMMENT 'Unique identifier for the payment',
    subscription_id CHAR(36) NOT NULL COMMENT 'Reference to subscription',
    amount DECIMAL(10,2) NOT NULL COMMENT 'Payment amount',
    payment_date TIMESTAMP NOT NULL COMMENT 'Date when payment was made',
    payment_method VARCHAR(50) NOT NULL COMMENT 'Payment method (credit_card, bank_transfer, etc.)',
    transaction_id VARCHAR(255) DEFAULT NULL COMMENT 'External transaction ID from payment gateway',
    status VARCHAR(50) NOT NULL COMMENT 'Payment status (pending, completed, failed, refunded)',
    notes TEXT DEFAULT NULL COMMENT 'Additional notes about the payment',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when payment record was created',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this payment record',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Timestamp when payment was last updated',
    updated_by VARCHAR(255) NOT NULL COMMENT 'User who last updated this payment',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT 'Timestamp when payment was soft deleted',
    deleted_by VARCHAR(255) DEFAULT NULL COMMENT 'User who deleted this payment',
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Payment records for subscriptions';

-- Indexes for subscription_payments
CREATE INDEX idx_subscription_payments_subscription_id ON subscription_payments(subscription_id);
CREATE INDEX idx_subscription_payments_payment_date ON subscription_payments(payment_date);
CREATE INDEX idx_subscription_payments_status ON subscription_payments(status);
CREATE INDEX idx_subscription_payments_transaction_id ON subscription_payments(transaction_id);
CREATE INDEX idx_subscription_payments_deleted_at ON subscription_payments(deleted_at);

-- Composite index for payment history queries
CREATE INDEX idx_subscription_payments_history ON subscription_payments(subscription_id, payment_date, status);
