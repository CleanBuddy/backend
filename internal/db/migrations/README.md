# Database Migrations

This directory contains SQL migrations managed by [golang-migrate](https://github.com/golang-migrate/migrate).

## Migration Naming Convention

Migrations are named with a sequential number followed by a descriptive name:

```
000001_initial_schema.sql
000002_add_user_table.sql
000003_add_booking_status.sql
```

Each migration consists of two files:
- `NNNNNN_name.up.sql` - Forward migration (apply changes)
- `NNNNNN_name.down.sql` - Rollback migration (undo changes)

## Common Commands

### Apply all pending migrations
```bash
make db-migrate
# or
make db-migrate-up
```

### Rollback last migration
```bash
make db-migrate-down
```

### Check current migration version
```bash
make db-migrate-version
```

### Create new migration
```bash
make db-migrate-create
# Then enter migration name when prompted
```

### Force migration to specific version (use with caution!)
```bash
make db-migrate-force
# Then enter version number when prompted
```

## Manual Commands (without Makefile)

If you need to run migrations manually:

```bash
# Apply all migrations
migrate -path backend/internal/db/migrations \
  -database "postgresql://cleanbuddy:devpassword@localhost:5432/cleanbuddy?sslmode=disable" \
  up

# Rollback one migration
migrate -path backend/internal/db/migrations \
  -database "postgresql://cleanbuddy:devpassword@localhost:5432/cleanbuddy?sslmode=disable" \
  down 1

# Check version
migrate -path backend/internal/db/migrations \
  -database "postgresql://cleanbuddy:devpassword@localhost:5432/cleanbuddy?sslmode=disable" \
  version
```

## Writing Migrations

### Best Practices

1. **Always create both UP and DOWN migrations**
   - UP: Apply the change
   - DOWN: Undo the change completely

2. **Make migrations atomic**
   - Each migration should represent one logical change
   - Use transactions when possible

3. **Be reversible**
   - DOWN migrations should restore the database to the previous state
   - Never drop data in DOWN migrations if possible (backup first)

4. **Test migrations**
   - Test UP migration
   - Test DOWN migration (rollback)
   - Test UP again (re-apply)

5. **Add comments**
   - Explain WHY the migration is needed
   - Document any tricky logic

### Example Migration

**000023_add_user_preferences.up.sql**:
```sql
-- Migration: Add user preferences table
-- Reason: Store user notification and privacy preferences

CREATE TABLE user_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email_notifications BOOLEAN DEFAULT TRUE,
    sms_notifications BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id)
);

CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);

-- Add update trigger
CREATE TRIGGER update_user_preferences_updated_at
    BEFORE UPDATE ON user_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

**000023_add_user_preferences.down.sql**:
```sql
-- Rollback: Drop user preferences table

DROP TRIGGER IF EXISTS update_user_preferences_updated_at ON user_preferences;
DROP INDEX IF EXISTS idx_user_preferences_user_id;
DROP TABLE IF EXISTS user_preferences;
```

## Production Deployment

### Pre-deployment Checklist

- [ ] All migrations tested locally
- [ ] DOWN migrations tested (rollback works)
- [ ] Database backup created
- [ ] Migration plan documented
- [ ] Rollback plan documented
- [ ] Downtime window scheduled (if needed)

### Running Migrations in Production

1. **Backup the database**
   ```bash
   pg_dump -h <host> -U <user> -d cleanbuddy > backup_$(date +%Y%m%d_%H%M%S).sql
   ```

2. **Apply migrations**
   ```bash
   migrate -path backend/internal/db/migrations \
     -database "postgresql://<user>:<pass>@<host>:5432/cleanbuddy?sslmode=require" \
     up
   ```

3. **Verify migration**
   ```bash
   # Check version
   migrate -path backend/internal/db/migrations \
     -database "postgresql://<user>:<pass>@<host>:5432/cleanbuddy?sslmode=require" \
     version

   # Run smoke tests
   # Verify critical queries work
   ```

4. **If migration fails**
   ```bash
   # Rollback
   migrate -path backend/internal/db/migrations \
     -database "postgresql://<user>:<pass>@<host>:5432/cleanbuddy?sslmode=require" \
     down 1

   # Restore from backup (if needed)
   psql -h <host> -U <user> -d cleanbuddy < backup_YYYYMMDD_HHMMSS.sql
   ```

## Troubleshooting

### "Dirty database version"
This happens when a migration fails mid-execution. To fix:

1. Check the schema_migrations table:
   ```sql
   SELECT * FROM schema_migrations;
   ```

2. Fix the database manually (complete or rollback the partial migration)

3. Force the version:
   ```bash
   make db-migrate-force
   # Enter the correct version number
   ```

### "File does not exist"
Make sure you're running commands from the project root directory.

### "Connection refused"
Ensure PostgreSQL is running:
```bash
make docker-up
```

## CI/CD Integration

Migrations should be run automatically in CI/CD pipeline:

```yaml
# Example GitHub Actions
- name: Run database migrations
  run: |
    make db-migrate
  env:
    DB_URL: ${{ secrets.DATABASE_URL }}
```

## Schema Version Control

The current schema version is stored in the `schema_migrations` table:

```sql
SELECT * FROM schema_migrations;
```

This table is automatically managed by golang-migrate.
