-- Rollback: make_bars_code_unique_among_live_rows
-- Created at: 2026-09-24T07:45:19Z

-- Restoring the table-wide UNIQUE fails if a code has since been reused after a soft delete,
-- which is exactly what the up migration allows. Resolve those rows before rolling back.
CREATE INDEX idx_bars_active ON bars (code, is_active, deleted_at);
CREATE INDEX idx_bars_deleted_at ON bars (deleted_at);
CREATE INDEX idx_bars_is_active ON bars (is_active);
CREATE INDEX idx_bars_code ON bars (code);

DROP INDEX uq_bars_code_live;
ALTER TABLE bars ADD CONSTRAINT bars_code_key UNIQUE (code);
