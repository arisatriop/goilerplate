-- Migration: plan_types
-- Created at: 2025-10-31T13:56:58+07:00

-- Add your up migration here
CREATE TABLE plan_types (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    code VARCHAR(255) NOT NULL UNIQUE COMMENT 'Unique code identifier',
    name VARCHAR(100) NOT NULL COMMENT 'Plan name',
    is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'Whether the record is active',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this record',
    updated_by VARCHAR(255) NOT NULL COMMENT 'User who last updated this record',
    deleted_by VARCHAR(255) DEFAULT NULL COMMENT 'User who deleted this record',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when record was created',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Timestamp when record was last updated',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT 'Timestamp when record was soft deleted'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='plan_types table for storing example records';

-- Create indexes for better performance
CREATE INDEX idx_plan_types_code ON plan_types(code);
CREATE INDEX idx_plan_types_name ON plan_types(name);
CREATE INDEX idx_plan_types_is_active ON plan_types(is_active);
CREATE INDEX idx_plan_types_created_by ON plan_types(created_by);
CREATE INDEX idx_plan_types_deleted_at ON plan_types(deleted_at);

-- Composite index for active records (MySQL doesn't support partial indexes)
CREATE INDEX idx_plan_types_active ON plan_types(code, name, is_active, deleted_at);