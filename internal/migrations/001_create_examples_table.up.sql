-- Migration: create_examples_table
-- Created at: 2025-10-11T15:30:00Z

CREATE TABLE examples (
    id CHAR(36) PRIMARY KEY DEFAULT (gen_random_uuid()),
    code VARCHAR(255) NOT NULL UNIQUE COMMENT 'Unique code identifier',
    example TEXT NOT NULL COMMENT 'Example text content',
    is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'Whether the record is active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when record was created',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Timestamp when record was last updated',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT 'Timestamp when record was soft deleted',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this record',
    updated_by VARCHAR(255) NOT NULL COMMENT 'User who last updated this record',
    deleted_by VARCHAR(255) DEFAULT NULL COMMENT 'User who deleted this record'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Examples table for storing example records';

-- Create indexes for better performance
CREATE INDEX idx_examples_code ON examples(code);
CREATE INDEX idx_examples_is_active ON examples(is_active);
CREATE INDEX idx_examples_deleted_at ON examples(deleted_at);

-- Composite index for active records (MySQL doesn't support partial indexes)
CREATE INDEX idx_examples_active ON examples(code, is_active, deleted_at);