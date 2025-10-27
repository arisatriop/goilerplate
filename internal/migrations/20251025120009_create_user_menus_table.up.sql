-- Migration: create_user_menus_table
-- Created at: 2025-10-25T12:00:09+07:00

CREATE TABLE user_menus (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    menu_id UUID NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    PRIMARY KEY (user_id, menu_id)
);

-- Indexes for user_menus
CREATE INDEX idx_user_menus_user_id ON user_menus(user_id);
CREATE INDEX idx_user_menus_menu_id ON user_menus(menu_id);

-- Comments for user_menus
COMMENT ON TABLE user_menus IS 'User-specific menu access overrides';
COMMENT ON COLUMN user_menus.user_id IS 'Reference to user';
COMMENT ON COLUMN user_menus.menu_id IS 'Reference to menu';
