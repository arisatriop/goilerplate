-- Migration: create_permissions_table
-- Created at: 2025-10-25T12:00:02+07:00

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(150) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255) NOT NULL,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    deleted_by VARCHAR(255) DEFAULT NULL
);

-- Comments
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
COMMENT ON TABLE permissions IS 'Permissions table for RBAC system';

-- Indexes for permissions
CREATE INDEX idx_permissions_slug ON permissions(slug);
CREATE INDEX idx_permissions_deleted_at ON permissions(deleted_at);
