-- Migration: add_new_email_to_one_time_tokens (down)

ALTER TABLE one_time_tokens DROP CONSTRAINT IF EXISTS one_time_tokens_new_email_check;
ALTER TABLE one_time_tokens DROP COLUMN IF EXISTS new_email;
