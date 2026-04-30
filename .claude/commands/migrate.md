# /migrate — Database Migration

Create and apply a database migration for the HOF backend.
Arguments: $ARGUMENTS — description of the schema change (e.g. "add phone to users").

---

## Migration file naming

```
NNN_description_with_underscores.sql
```
- `NNN` — zero-padded 3-digit sequence number
- Description — lowercase, underscores, no spaces
- Example: `026_add_phone_to_users.sql`

```bash
# Find next number
ls migrations/ | sort | tail -1
# → 025_add_plan_offering_fields.sql  ← so next is 026
```

---

## Template

```sql
-- Write your migrate up statements here

-- Adding a column:
ALTER TABLE <table>
    ADD COLUMN IF NOT EXISTS <column> <type> [DEFAULT <value>] [NOT NULL];

-- Adding an index:
CREATE INDEX IF NOT EXISTS idx_<table>_<column> ON <table> (<column>);

-- Creating a table:
CREATE TABLE IF NOT EXISTS <table> (
    id          UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    <col>       <type>      NOT NULL,
    date_added  TIMESTAMP   NOT NULL DEFAULT NOW(),
    last_updated TIMESTAMP  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMP
);

---- create above / drop below ----

-- Drop operations (reversal):
ALTER TABLE <table> DROP COLUMN IF EXISTS <column>;
DROP TABLE IF EXISTS <table>;
DROP INDEX IF EXISTS idx_<table>_<column>;
```

---

## Rules

1. **Always use `IF NOT EXISTS` / `IF EXISTS`** — migrations must be idempotent.
2. **Never modify an already-applied migration** — the runner skips applied versions. Create a new file.
3. **Match the entity** — after adding a column, update the GORM entity struct with the correct tag.
4. **Standard timestamp columns** — all tables should have `date_added`, `last_updated`, `deleted_at`.
5. **UUID primary keys** — use `gen_random_uuid()` as default.
6. **Soft deletes** — use `deleted_at TIMESTAMP` (GORM convention for the `DeletedAt *time.Time` field).

---

## GORM column tag cross-reference

| Go field | GORM default column | Override needed? |
|---|---|---|
| `CreatedAt` | `created_at` | Yes — `gorm:"column:date_added"` |
| `UpdatedAt` | `updated_at` | Yes — `gorm:"column:last_updated"` |
| `DeletedAt` | `deleted_at` | No (matches) |
| `PlanID` | `plan_id` | Yes if column is `subscription_plan_id` |
| `Frequency` | `frequency` | Yes — `gorm:"column:freq"` |

---

## Applying migrations

Migrations run automatically at server startup:
```go
database.RunMigrations(db, "./migrations", zapLog)
```

To apply manually (dev):
```bash
# Restart the server — it will pick up and apply new migrations
make build && ./bin/server &
# Watch for: INFO "migration applied" {"version":"026_add_phone_to_users"}
```

To verify applied:
```bash
# Local
make db-shell
SELECT * FROM schema_migrations ORDER BY applied_at DESC LIMIT 5;

# Heroku
heroku pg:psql --app <app-name>
SELECT version, applied_at FROM schema_migrations ORDER BY applied_at DESC LIMIT 10;
```

---

## Common patterns

### Add a nullable field
```sql
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS phone VARCHAR(20) DEFAULT NULL;
```
```go
// In entity.go
Phone *string `gorm:"column:phone;type:varchar(20)"`
```

### Add a non-nullable field with default
```sql
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS provider VARCHAR(50) NOT NULL DEFAULT 'paystack';
```
```go
Provider string `gorm:"column:provider;type:varchar(50);default:'paystack'"`
```

### Add a foreign key reference
```sql
ALTER TABLE audio_messages
    ADD COLUMN IF NOT EXISTS meditation_id UUID REFERENCES meditations(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_audio_messages_meditation_id ON audio_messages (meditation_id);
```

### Add a soft-delete to an existing table
```sql
ALTER TABLE favourites
    ADD COLUMN IF NOT EXISTS date_added  TIMESTAMP NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS last_updated TIMESTAMP NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS deleted_at  TIMESTAMP;
```

---

## After creating the migration

1. Update the GORM entity struct in `internal/domain/<domain>/entity.go`
2. If a new table, add a `TableName() string` method
3. If a new column that's needed in queries, update the relevant repository methods
4. Restart server and confirm `INFO "migration applied" {"version":"NNN_..."}`
5. Run `/deskcheck` to verify the affected endpoints
