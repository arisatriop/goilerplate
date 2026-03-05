-- Rollback: create_bars_table
-- Created at: 2025-10-11T15:30:00Z

-- Drop indexes first
DROP INDEX IF EXISTS idx_bars_active;
DROP INDEX IF EXISTS idx_bars_deleted_at;
DROP INDEX IF EXISTS idx_bars_created_by;
DROP INDEX IF EXISTS idx_bars_is_active;
DROP INDEX IF EXISTS idx_bars_code;

-- Drop the table
DROP TABLE IF EXISTS bars;