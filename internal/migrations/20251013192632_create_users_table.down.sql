-- Rollback: create_users_table
-- Created at: 2025-10-13T19:26:32+07:00

-- Add your down migration here

-- Drop indexes first (PostgreSQL will automatically drop them when table is dropped, but explicit is better)
DROP INDEX IF EXISTS idx_users_email_active;
DROP INDEX IF EXISTS idx_users_active_created;
DROP INDEX IF EXISTS idx_users_name;
DROP INDEX IF EXISTS idx_users_created_by;
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_updated_at;
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_email;

-- Drop the table
DROP TABLE IF EXISTS users;