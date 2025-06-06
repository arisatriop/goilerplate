CREATE TABLE menu_permission_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    menu_permission_id UUID NOT NULL,
    role_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by VARCHAR(100) NOT NULL,
    updated_at TIMESTAMP NULL,
    updated_by VARCHAR(100) NULL,
    deleted_at TIMESTAMP NULL,
    deleted_by VARCHAR(100) NULL,

    -- Foreign key constraints
    CONSTRAINT fk_menu_permission_roles_menu_id FOREIGN KEY (menu_permission_id) REFERENCES menu_permissions(id),
    CONSTRAINT fk_menu_permission_roles_role_id FOREIGN KEY (role_id) REFERENCES roles(id)
);

CREATE INDEX idx_menu_permission_roles_menu_id ON menu_permission_roles(menu_permission_id);
CREATE INDEX idx_menu_permission_roles_role_id ON menu_permission_roles(role_id);