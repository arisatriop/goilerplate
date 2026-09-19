-- Migration: seed_owner_role
-- Created at: 2026-09-18T16:00:00Z

-- Registration assigns every new account the 'owner' role. Without this row,
-- POST /api/v1/auth/register fails with a 500 on a freshly migrated database
-- ("failed to get role: record not found"), so a new developer cannot create the
-- first account at all. That makes the row part of the baseline schema rather
-- than sample data, which is why it lives in a migration and not in a seeder.
--
-- The role grants nothing on its own: what it permits is decided by
-- role_permissions, which starts empty.
INSERT INTO roles (id, name, slug, description, created_by, updated_by)
VALUES (
    gen_random_uuid(),
    'Owner',
    'owner',
    'Default role assigned at registration',
    'migration',
    'migration'
)
ON CONFLICT (slug) DO NOTHING;
