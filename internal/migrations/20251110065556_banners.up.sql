-- Migration: banners
-- Created at: 2025-11-10T06:55:56+07:00

-- Add your up migration here
CREATE TABLE banners (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    store_id CHAR(36) NOT NULL COMMENT 'ID of the store associated with the banner',
    file_type VARCHAR(100) NOT NULL COMMENT 'MIME type of the banner image',
    file_storage VARCHAR(50) NOT NULL COMMENT 'Storage type of the banner image',
    file_name VARCHAR(255) NOT NULL COMMENT 'File name of the banner image',
    file_path TEXT NOT NULL COMMENT 'File path of the banner image',
    file_url TEXT COMMENT 'File URL of the banner image',
    is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'Whether the record is active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when record was created',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Timestamp when record was last updated',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT 'Timestamp when record was soft deleted',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this record',
    updated_by VARCHAR(255) NOT NULL COMMENT 'User who last updated this record',
    deleted_by VARCHAR(255) DEFAULT NULL COMMENT 'User who deleted this record',
    FOREIGN KEY (store_id) REFERENCES stores(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Banners table for storing banner records';

-- Create indexes for better performance
CREATE INDEX idx_banners_is_active ON banners(is_active);
CREATE INDEX idx_banners_deleted_at ON banners(deleted_at);
CREATE INDEX idx_banners_store_id ON banners(store_id);

-- Composite index for active records (MySQL doesn't support partial indexes)
CREATE INDEX idx_banners_active ON banners(store_id, is_active, deleted_at);