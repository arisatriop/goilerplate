-- Migration: product_categories
-- Created at: 2025-11-02T17:17:52+07:00

-- Add your up migration here
CREATE TABLE product_categories (
    product_id CHAR(36) NOT NULL COMMENT 'Foreign key to products table',
    category_id CHAR(36) NOT NULL COMMENT 'Foreign key to categories table',
    is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'Whether the record is active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when record was created',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Timestamp when record was last updated',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this record',
    updated_by VARCHAR(255) NOT NULL COMMENT 'User who last updated this record',
    PRIMARY KEY (product_id, category_id),
    FOREIGN KEY (product_id) REFERENCES products(id),
    FOREIGN KEY (category_id) REFERENCES categories(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='product_categories table for storing example records';

-- Create indexes for better performance
CREATE INDEX idx_product_categories_is_active ON product_categories(is_active);
CREATE INDEX idx_product_categories_category_id ON product_categories(category_id);
CREATE INDEX idx_product_categories_product_id ON product_categories(product_id);

-- Composite index for active records (MySQL doesn't support partial indexes)
CREATE INDEX idx_product_categories_active ON product_categories(product_id, category_id, is_active);

