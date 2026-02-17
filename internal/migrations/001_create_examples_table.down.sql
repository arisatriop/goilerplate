-- Rollback: create_examples_table
-- Created at: 2025-10-11T15:30:00Z

-- Drop indexes first
DROP INDEX IF EXISTS idx_examples_active;
DROP INDEX IF EXISTS idx_examples_deleted_at;
DROP INDEX IF EXISTS idx_examples_created_by;
DROP INDEX IF EXISTS idx_examples_is_active;
DROP INDEX IF EXISTS idx_examples_code;

-- Drop the table
DROP TABLE IF EXISTS examples;