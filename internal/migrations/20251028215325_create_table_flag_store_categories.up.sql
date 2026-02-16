-- Migration: create_table_flag_store_categories
-- Created at: 2025-10-28T21:53:25+07:00

-- Add your up migration here

CREATE TABLE flag_store_categories (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    code VARCHAR(25) NOT NULL UNIQUE,
    name TEXT NOT NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL DEFAULT 'system',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    updated_by VARCHAR(255) NOT NULL DEFAULT 'system',
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    deleted_by VARCHAR(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create indexes for better performance
CREATE INDEX idx_flag_store_categories_code ON flag_store_categories(code);
CREATE INDEX idx_flag_store_categories_is_active ON flag_store_categories(is_active);
CREATE INDEX idx_flag_store_categories_created_by ON flag_store_categories(created_by);
CREATE INDEX idx_flag_store_categories_deleted_at ON flag_store_categories(deleted_at);

-- Composite index for active records (MySQL doesn't support partial indexes)
CREATE INDEX idx_flag_store_categories_active ON flag_store_categories(code, is_active, deleted_at);