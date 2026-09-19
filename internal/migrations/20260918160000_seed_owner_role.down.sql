-- Rollback: seed_owner_role
-- Created at: 2026-09-18T16:00:00Z

-- Removed only when no account still holds it. user_roles.role_id cascades on
-- delete, so an unconditional DELETE would strip the role from every existing
-- user as a side effect of rolling back one migration.
DELETE FROM roles
 WHERE slug = 'owner'
   AND NOT EXISTS (
       SELECT 1 FROM user_roles WHERE user_roles.role_id = roles.id
   );
