CREATE TABLE menu_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    menu_id UUID NOT NULL,
    permission VARCHAR(100) NOT NULL,
    "order" INT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by VARCHAR(100) NOT NULL,
    updated_at TIMESTAMP NULL,
    updated_by VARCHAR(100) NULL,
    deleted_at TIMESTAMP NULL,
    deleted_by VARCHAR(100) NULL,

    -- Foreign key constraints
    CONSTRAINT fk_menu_permissions_menu_id FOREIGN KEY (menu_id) REFERENCES menus(id)
);

CREATE INDEX idx_menus_permission_menu_id ON menu_permissions(menu_id);
CREATE INDEX idx_menus_permission_permission ON menu_permissions(permission);