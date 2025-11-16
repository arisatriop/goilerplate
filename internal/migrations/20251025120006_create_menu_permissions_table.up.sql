-- Migration: create_menu_permissions_table
-- Created at: 2025-10-25T12:00:06+07:00

CREATE TABLE menu_permissions (
    menu_id UUID NOT NULL, -- Reference to menu
    permission_id UUID NOT NULL, -- Reference to permission
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp when relationship was created
    created_by VARCHAR(255) NOT NULL, -- User who created this relationship
    PRIMARY KEY (menu_id, permission_id),
    FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

-- Add table and column comments
COMMENT ON TABLE menu_permissions IS 'Many-to-many relationship between menus and permissions';
COMMENT ON COLUMN menu_permissions.menu_id IS 'Reference to menu';
COMMENT ON COLUMN menu_permissions.permission_id IS 'Reference to permission';
COMMENT ON COLUMN menu_permissions.created_at IS 'Timestamp when relationship was created';
COMMENT ON COLUMN menu_permissions.created_by IS 'User who created this relationship';

-- Indexes for menu_permissions
CREATE INDEX idx_menu_permissions_menu_id ON menu_permissions(menu_id);
CREATE INDEX idx_menu_permissions_permission_id ON menu_permissions(permission_id);
