-- Rollback: create_user_tokens_table
-- Created at: 2025-10-13T19:41:18+07:00

-- Drop indexes
DROP INDEX IF EXISTS idx_user_tokens_type_expires;
DROP INDEX IF EXISTS idx_user_tokens_user_type;
DROP INDEX IF EXISTS idx_user_tokens_used_at;
DROP INDEX IF EXISTS idx_user_tokens_expires_at;
DROP INDEX IF EXISTS idx_user_tokens_type;
DROP INDEX IF EXISTS idx_user_tokens_token_hash;
DROP INDEX IF EXISTS idx_user_tokens_user_id;

-- Drop the table
DROP TABLE IF EXISTS user_tokens;
