-- Rollback: create_user_sessions_table
-- Created at: 2025-10-13T19:40:35+07:00

-- Drop indexes
DROP INDEX IF EXISTS idx_user_sessions_user_device;
DROP INDEX IF EXISTS idx_user_sessions_user_active;
DROP INDEX IF EXISTS idx_user_sessions_ip_address;
DROP INDEX IF EXISTS idx_user_sessions_last_used;
DROP INDEX IF EXISTS idx_user_sessions_is_active;
DROP INDEX IF EXISTS idx_user_sessions_expires_at;
DROP INDEX IF EXISTS idx_user_sessions_device_id;
DROP INDEX IF EXISTS idx_user_sessions_refresh_token;
DROP INDEX IF EXISTS idx_user_sessions_user_id;

-- Drop the table
DROP TABLE IF EXISTS user_sessions;
