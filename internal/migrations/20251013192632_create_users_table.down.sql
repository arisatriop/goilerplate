-- Rollback: create_users_table
-- Created at: 2025-10-13T19:26:32+07:00

-- Drop trigger
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes first (PostgreSQL will automatically drop them when table is dropped, but explicit is better)
DROP INDEX IF EXISTS idx_users_name;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_email_verified;
DROP INDEX IF EXISTS idx_users_password_changed_at;
DROP INDEX IF EXISTS idx_users_last_login_at;
DROP INDEX IF EXISTS idx_users_failed_attempts;
DROP INDEX IF EXISTS idx_users_locked_until;

-- Drop the table
DROP TABLE IF EXISTS users;