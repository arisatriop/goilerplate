-- Migration: create_user_permissions_table
-- Created at: 2025-10-25T12:00:08+07:00

CREATE TABLE user_permissions (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    is_granted BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255) NOT NULL,
    PRIMARY KEY (user_id, permission_id)
);

-- Indexes for user_permissions
CREATE INDEX idx_user_permissions_user_id ON user_permissions(user_id);
CREATE INDEX idx_user_permissions_permission_id ON user_permissions(permission_id);
CREATE INDEX idx_user_permissions_is_granted ON user_permissions(is_granted);

-- Comments for user_permissions
COMMENT ON TABLE user_permissions IS 'User-specific permission overrides (grants or revokes)';
COMMENT ON COLUMN user_permissions.user_id IS 'Reference to user';
COMMENT ON COLUMN user_permissions.permission_id IS 'Reference to permission';
COMMENT ON COLUMN user_permissions.is_granted IS 'TRUE = explicitly granted, FALSE = explicitly revoked';
