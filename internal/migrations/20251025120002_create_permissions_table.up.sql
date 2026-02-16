-- Migration: create_permissions_table
-- Created at: 2025-10-25T12:00:02+07:00

CREATE TABLE permissions (
    id CHAR(36) PRIMARY KEY DEFAULT (gen_random_uuid()) COMMENT 'Unique identifier for the permission',
    name VARCHAR(100) NOT NULL COMMENT 'Display name of the permission',
    slug VARCHAR(150) UNIQUE NOT NULL COMMENT 'Unique slug identifier for the permission (e.g., create:example, update:example)',
    description TEXT COMMENT 'Description of what this permission grants',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when permission was created',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this permission',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Timestamp when permission was last updated',
    updated_by VARCHAR(255) NOT NULL COMMENT 'User who last updated this permission',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT 'Timestamp when permission was soft deleted',
    deleted_by VARCHAR(255) DEFAULT NULL COMMENT 'User who deleted this permission'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Permissions table for RBAC system';

-- Indexes for permissions
CREATE INDEX idx_permissions_slug ON permissions(slug);
CREATE INDEX idx_permissions_deleted_at ON permissions(deleted_at);
