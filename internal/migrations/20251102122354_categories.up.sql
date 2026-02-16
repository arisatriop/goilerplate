-- Migration: categories
-- Created at: 2025-11-02T12:23:54+07:00

-- Add your up migration here
CREATE TABLE categories (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    name TEXT NOT NULL COMMENT 'Category name',
    is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'Whether the record is active',
    store_id CHAR(36) NOT NULL COMMENT 'Foreign key to stores table',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when record was created',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Timestamp when record was last updated',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT 'Timestamp when record was soft deleted',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this record',
    updated_by VARCHAR(255) NOT NULL COMMENT 'User who last updated this record',
    deleted_by VARCHAR(255) DEFAULT NULL COMMENT 'User who deleted this record',
    FOREIGN KEY (store_id) REFERENCES stores(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='categories table for storing example records';

-- Create indexes for better performance
CREATE INDEX idx_categories_is_active ON categories(is_active);
CREATE INDEX idx_categories_deleted_at ON categories(deleted_at);
CREATE INDEX idx_categories_store_id ON categories(store_id);

-- Composite index for active records (MySQL doesn't support partial indexes)
CREATE INDEX idx_categories_active ON categories(store_id, is_active, deleted_at);
