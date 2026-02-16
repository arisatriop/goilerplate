-- Migration: create_roles_table
-- Created at: 2025-10-25T12:00:01+07:00

CREATE TABLE roles (
    id CHAR(36) PRIMARY KEY DEFAULT (gen_random_uuid()) COMMENT 'Unique identifier for the role',
    name VARCHAR(100) NOT NULL COMMENT 'Display name of the role',
    slug VARCHAR(100) UNIQUE NOT NULL COMMENT 'Unique slug identifier for the role',
    description TEXT COMMENT 'Description of the role',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when role was created',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this role',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Timestamp when role was last updated',
    updated_by VARCHAR(255) NOT NULL COMMENT 'User who last updated this role',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT 'Timestamp when role was soft deleted',
    deleted_by VARCHAR(255) DEFAULT NULL COMMENT 'User who deleted this role'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Roles table for RBAC system';

-- Indexes for roles
CREATE INDEX idx_roles_slug ON roles(slug);
CREATE INDEX idx_roles_deleted_at ON roles(deleted_at);
