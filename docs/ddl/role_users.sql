CREATE TABLE role_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    roles_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by VARCHAR(100) NOT NULL,
    updated_at TIMESTAMP NULL,
    updated_by VARCHAR(100) NULL,
    deleted_at TIMESTAMP NULL,
    deleted_by VARCHAR(100) NULL,

    -- Foreign key constraints
    CONSTRAINT fk_role_users_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_role_users_roles_id FOREIGN KEY (roles_id) REFERENCES roles(id) ON DELETE CASCADE
);


CREATE INDEX idx_role_users_user_id ON role_users(user_id);
CREATE INDEX idx_role_users_roles_id ON role_users(roles_id);