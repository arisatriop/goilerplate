-- Migration: create_permissions_table
-- Created at: 2025-10-25T12:00:02+07:00

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Unique identifier for the permission
    name VARCHAR(100) NOT NULL, -- Display name of the permission
    slug VARCHAR(150) UNIQUE NOT NULL, -- Unique slug identifier for the permission (e.g., create:example, update:example)
    description TEXT, -- Description of what this permission grants
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp when permission was created
    created_by VARCHAR(255) NOT NULL, -- User who created this permission
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp when permission was last updated
    updated_by VARCHAR(255) NOT NULL, -- User who last updated this permission
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL, -- Timestamp when permission was soft deleted
    deleted_by VARCHAR(255) DEFAULT NULL -- User who deleted this permission
);

-- Add table and column comments
COMMENT ON TABLE permissions IS 'Permissions table for RBAC system';
COMMENT ON COLUMN permissions.id IS 'Unique identifier for the permission';
COMMENT ON COLUMN permissions.name IS 'Display name of the permission';
COMMENT ON COLUMN permissions.slug IS 'Unique slug identifier for the permission (e.g., create:example, update:example)';
COMMENT ON COLUMN permissions.description IS 'Description of what this permission grants';
COMMENT ON COLUMN permissions.created_at IS 'Timestamp when permission was created';
COMMENT ON COLUMN permissions.created_by IS 'User who created this permission';
COMMENT ON COLUMN permissions.updated_at IS 'Timestamp when permission was last updated';
COMMENT ON COLUMN permissions.updated_by IS 'User who last updated this permission';
COMMENT ON COLUMN permissions.deleted_at IS 'Timestamp when permission was soft deleted';
COMMENT ON COLUMN permissions.deleted_by IS 'User who deleted this permission';

-- Indexes for permissions
CREATE INDEX idx_permissions_slug ON permissions(slug);
CREATE INDEX idx_permissions_deleted_at ON permissions(deleted_at);
