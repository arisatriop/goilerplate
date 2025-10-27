# Database Migrations

This boilerplate includes a robust migration system that supports both PostgreSQL and MySQL.

## Migration Files Structure

```
internal/migrations/
├── 001_create_examples_table.up.sql    # Forward migration
├── 001_create_examples_table.down.sql  # Rollback migration
├── 002_add_users_table.up.sql
├── 002_add_users_table.down.sql
└── ...
```

## Migration Commands

### Using Make (Recommended)

```bash
# Run all pending migrations
make migrate-up

# Rollback the last migration
make migrate-down

# Check migration status
make migrate-status

# Create a new migration
make migrate-create name=add_users_table

# Reset database (rollback all + migrate up)
make db-reset
```

### Using Go directly

```bash
# Run migrations
go run cmd/migrate/main.go -action=up

# Rollback migration
go run cmd/migrate/main.go -action=down

# Check status
go run cmd/migrate/main.go -action=status

# Create new migration
go run cmd/migrate/main.go -action=create -name=add_users_table
```

## Migration File Naming Convention

Migration files should follow this pattern:

- `{timestamp}_{name}.up.sql` - Forward migration
- `{timestamp}_{name}.down.sql` - Rollback migration

Example:

- `20241011153000_create_examples_table.up.sql`
- `20241011153000_create_examples_table.down.sql`

## Best Practices

### 1. Always Create Both Up and Down Migrations

```sql
-- 001_create_users.up.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL
);

-- 001_create_users.down.sql
DROP TABLE IF EXISTS users;
```

### 2. Use Transactions for Complex Migrations

```sql
-- 002_complex_migration.up.sql
BEGIN;

ALTER TABLE users ADD COLUMN phone VARCHAR(20);
UPDATE users SET phone = 'unknown' WHERE phone IS NULL;
ALTER TABLE users ALTER COLUMN phone SET NOT NULL;

COMMIT;
```

### 3. Create Indexes for Performance

```sql
-- Always create indexes for frequently queried columns
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);
```

### 4. Handle Data Migration Safely

```sql
-- Use conditional logic for data migrations
UPDATE users
SET status = 'active'
WHERE status IS NULL AND created_at > '2024-01-01';
```

## Migration Status

The system tracks executed migrations in a `migrations` table:

```sql
CREATE TABLE migrations (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Error Handling

- Migrations run in transactions (automatic rollback on error)
- Failed migrations don't get recorded in the migrations table
- You can safely re-run failed migrations after fixing them

## Example Migration Workflow

1. **Create a new migration:**

   ```bash
   make migrate-create name=add_user_profiles
   ```

2. **Edit the generated files:**

   ```sql
   -- internal/migrations/20241011153000_add_user_profiles.up.sql
   CREATE TABLE user_profiles (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       user_id UUID NOT NULL REFERENCES users(id),
       bio TEXT,
       avatar_url VARCHAR(255),
       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
   );
   ```

3. **Run the migration:**

   ```bash
   make migrate-up
   ```

4. **Check status:**

   ```bash
   make migrate-status
   ```

5. **If needed, rollback:**
   ```bash
   make migrate-down
   ```

## Database Support

The migration system supports:

- ✅ PostgreSQL (recommended)
- ✅ MySQL
- ✅ Any database supported by GORM

## Tips

- Keep migrations small and focused
- Test migrations on a copy of production data
- Always review the down migration before deploying
- Use descriptive migration names
- Document complex migrations with comments
