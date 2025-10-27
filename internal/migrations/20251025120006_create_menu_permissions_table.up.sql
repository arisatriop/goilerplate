-- Migration: create_menu_permissions_table
-- Created at: 2025-10-25T12:00:06+07:00

CREATE TABLE menu_permissions (
    menu_id UUID NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    PRIMARY KEY (menu_id, permission_id)
);

-- Indexes for menu_permissions
CREATE INDEX idx_menu_permissions_menu_id ON menu_permissions(menu_id);
CREATE INDEX idx_menu_permissions_permission_id ON menu_permissions(permission_id);

-- Comments for menu_permissions
COMMENT ON TABLE menu_permissions IS 'Many-to-many relationship between menus and permissions';
COMMENT ON COLUMN menu_permissions.menu_id IS 'Reference to menu';
COMMENT ON COLUMN menu_permissions.permission_id IS 'Reference to permission';
