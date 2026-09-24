-- Migration: make_bars_code_unique_among_live_rows
-- Created at: 2026-09-24T07:45:19Z

-- bars is soft-deleted, but code was UNIQUE across every row. A deleted bar therefore kept its
-- code forever: creating a new bar with it passed the application's check (which reads live rows
-- only) and then failed on the constraint with a 500. Uniqueness now covers live rows only, and
-- the repository maps a violation of this index to 409, so the index — not a racy pre-check — is
-- what guarantees one live bar per code.
ALTER TABLE bars DROP CONSTRAINT bars_code_key;
CREATE UNIQUE INDEX uq_bars_code_live ON bars (code) WHERE deleted_at IS NULL;

-- Redundant or unused: idx_bars_code duplicated the index the UNIQUE constraint already had;
-- is_active and deleted_at alone are too low-cardinality for the planner to choose; and the
-- (code, is_active, deleted_at) composite is covered by uq_bars_code_live for every query the
-- repository runs (lookups by code among live rows).
DROP INDEX idx_bars_code;
DROP INDEX idx_bars_is_active;
DROP INDEX idx_bars_deleted_at;
DROP INDEX idx_bars_active;
