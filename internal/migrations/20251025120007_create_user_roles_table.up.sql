-- Migration: create_user_roles_table
-- Created at: 2025-10-25T12:00:07+07:00

CREATE TABLE user_roles (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

-- Comments
COMMENT ON COLUMN user_roles.user_id IS 'Reference to user';
COMMENT ON COLUMN user_roles.role_id IS 'Reference to role';
COMMENT ON COLUMN user_roles.created_at IS 'Timestamp when relationship was created';
COMMENT ON COLUMN user_roles.created_by IS 'User who created this relationship';
COMMENT ON TABLE user_roles IS 'Many-to-many relationship between users and roles';

-- Indexes for user_roles
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
