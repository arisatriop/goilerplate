-- Migration: product_images
-- Created at: 2025-11-12T14:30:52+07:00

-- Add your up migration here
CREATE TABLE product_images (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    product_id CHAR(36) NOT NULL COMMENT 'Foreign key to products table',
    file_type VARCHAR(100) NOT NULL COMMENT 'MIME type of the product image',
    file_storage VARCHAR(50) NOT NULL COMMENT 'Storage type of the product image',
    file_name VARCHAR(255) NOT NULL COMMENT 'File name of the product image',
    file_path TEXT NOT NULL COMMENT 'File path of the product image',
    file_url TEXT NOT NULL COMMENT 'File URL of the product image',
    is_primary TINYINT(1) NOT NULL DEFAULT 0 COMMENT 'Whether the image is primary',
    is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'Whether the record is active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when record was created',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Timestamp when record was last updated',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this record',
    updated_by VARCHAR(255) NOT NULL COMMENT 'User who last updated this record',
    deleted_at TIMESTAMP NULL COMMENT 'Timestamp when record was deleted',
    deleted_by VARCHAR(255) DEFAULT NULL COMMENT 'User who deleted this record',
    FOREIGN KEY (product_id) REFERENCES products(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='product_images table for storing example records';

-- Create indexes for better performance
CREATE INDEX idx_product_images_product_id ON product_images(product_id);
CREATE INDEX idx_product_images_is_active ON product_images(is_active);
CREATE INDEX idx_product_images_deleted_at ON product_images(deleted_at);

-- Composite index for active records (MySQL doesn't support partial indexes)
CREATE INDEX idx_product_images_active ON product_images(product_id, is_active, deleted_at);

