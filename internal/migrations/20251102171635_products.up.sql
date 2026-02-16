-- Migration: products
-- Created at: 2025-11-02T17:16:35+07:00

-- Add your up migration here
CREATE TABLE products (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    name TEXT NOT NULL COMMENT 'Category name',
    description TEXT NULL COMMENT 'Category description',
    price DECIMAL(10,2) NOT NULL COMMENT 'Product price',
    images TEXT NULL COMMENT 'List of image URLs in JSON format',
    is_available TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'Whether the record is available',
    is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'Whether the record is active',
    store_id CHAR(36) NOT NULL COMMENT 'Foreign key to stores table',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when record was created',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Timestamp when record was last updated',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT 'Timestamp when record was soft deleted',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this record',
    updated_by VARCHAR(255) NOT NULL COMMENT 'User who last updated this record',
    deleted_by VARCHAR(255) DEFAULT NULL COMMENT 'User who deleted this record',
    FOREIGN KEY (store_id) REFERENCES stores(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='products table for storing example records';

-- Create indexes for better performance
CREATE INDEX idx_products_price ON products(price);
CREATE INDEX idx_products_is_active ON products(is_active);
CREATE INDEX idx_products_is_available ON products(is_available);
CREATE INDEX idx_products_deleted_at ON products(deleted_at);
CREATE INDEX idx_products_store_id ON products(store_id);

-- Composite index for active records (MySQL doesn't support partial indexes)
CREATE INDEX idx_products_active ON products(price, store_id, is_active, is_available, deleted_at);
