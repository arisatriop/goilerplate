-- Migration: create_user_permissions_table
-- Created at: 2025-10-25T12:00:08+07:00

CREATE TABLE user_permissions (
    user_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    is_granted BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255) NOT NULL,
    PRIMARY KEY (user_id, permission_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

-- Comments
COMMENT ON COLUMN user_permissions.user_id IS 'Reference to user';
COMMENT ON COLUMN user_permissions.permission_id IS 'Reference to permission';
COMMENT ON COLUMN user_permissions.is_granted IS 'TRUE = explicitly granted, FALSE = explicitly revoked';
COMMENT ON COLUMN user_permissions.created_at IS 'Timestamp when relationship was created';
COMMENT ON COLUMN user_permissions.created_by IS 'User who created this relationship';
COMMENT ON COLUMN user_permissions.updated_at IS 'Timestamp when relationship was last updated';
COMMENT ON COLUMN user_permissions.updated_by IS 'User who last updated this relationship';
COMMENT ON TABLE user_permissions IS 'User-specific permission overrides (grants or revokes)';

-- Indexes for user_permissions
CREATE INDEX idx_user_permissions_user_id ON user_permissions(user_id);
CREATE INDEX idx_user_permissions_permission_id ON user_permissions(permission_id);
CREATE INDEX idx_user_permissions_is_granted ON user_permissions(is_granted);
