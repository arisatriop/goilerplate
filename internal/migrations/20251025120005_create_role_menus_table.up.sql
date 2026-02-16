-- Migration: create_role_menus_table
-- Created at: 2025-10-25T12:00:05+07:00

CREATE TABLE role_menus (
    role_id UUID NOT NULL,
    menu_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    PRIMARY KEY (role_id, menu_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE
);

-- Comments
COMMENT ON COLUMN role_menus.role_id IS 'Reference to role';
COMMENT ON COLUMN role_menus.menu_id IS 'Reference to menu';
COMMENT ON COLUMN role_menus.created_at IS 'Timestamp when relationship was created';
COMMENT ON COLUMN role_menus.created_by IS 'User who created this relationship';
COMMENT ON TABLE role_menus IS 'Many-to-many relationship between roles and menus';

-- Indexes for role_menus
CREATE INDEX idx_role_menus_role_id ON role_menus(role_id);
CREATE INDEX idx_role_menus_menu_id ON role_menus(menu_id);
