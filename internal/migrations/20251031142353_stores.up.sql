-- Migration: stores
-- Created at: 2025-10-31T14:23:53+07:00

CREATE TABLE stores (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()) COMMENT 'Unique identifier for the store',
    user_id CHAR(36) NOT NULL COMMENT 'Reference to user who owns the store',
    name VARCHAR(255) NOT NULL COMMENT 'Store name',
    `desc` TEXT COMMENT 'Store description',
    address TEXT COMMENT 'Store physical address',
    phone VARCHAR(50) COMMENT 'Store phone number',
    email VARCHAR(100) COMMENT 'Store email address',
    web_url TEXT NOT NULL COMMENT 'Store website URL',
    is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'Whether store is active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when store was created',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this store',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Timestamp when store was last updated',
    updated_by VARCHAR(255) NOT NULL COMMENT 'User who last updated this store',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT 'Timestamp when store was soft deleted',
    deleted_by VARCHAR(255) DEFAULT NULL COMMENT 'User who deleted this store',
    FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Stores table for managing stores';

-- Indexes for stores
CREATE INDEX idx_stores_user_id ON stores(user_id);
CREATE INDEX idx_stores_is_active ON stores(is_active);
CREATE INDEX idx_stores_deleted_at ON stores(deleted_at);

-- Composite index for active stores lookup
CREATE INDEX idx_stores_active ON stores(user_id, is_active, deleted_at);
