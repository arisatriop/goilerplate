-- Migration: add_new_email_to_one_time_tokens
-- Created at: 2026-10-06T07:58:24Z

SET LOCAL lock_timeout = '5s';

-- An email change is confirmed with a code sent to the new address, so the token has to carry
-- that address until it is redeemed. The CHECK keeps it on email_change tokens only, and
-- required there: a change token with nowhere to change to is a bug the database should refuse.
ALTER TABLE one_time_tokens ADD COLUMN new_email TEXT NULL;
ALTER TABLE one_time_tokens ADD CONSTRAINT one_time_tokens_new_email_check
    CHECK ((token_type = 'email_change') = (new_email IS NOT NULL));

COMMENT ON COLUMN one_time_tokens.new_email IS 'email_change only: the address the account moves to once the code sent there is confirmed';

-- #99 added password_reset to this comment by editing an applied migration, which databases
-- migrated before it never saw. Restated here so every database ends up with the same text.
COMMENT ON COLUMN user_sessions.revoked_reason IS 'logout, logout_all, password_change, password_reset, reuse_detected, or admin';
