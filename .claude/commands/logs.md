# /logs — Logs, Traces & Observability

Tail, filter, and analyse logs for the HOF backend.
Arguments: $ARGUMENTS — optional filter (e.g. "error", "webhook", "migration", "user-id:xxx").

---

## Local development logs

```bash
# Run server with full structured logging
APP_ENV=dev ./bin/server 2>&1 | tee server.log

# Tail in real time
tail -f server.log

# Filter errors only
tail -f server.log | grep '"level":"error"\|FATAL\|ERROR'

# Filter by request path
tail -f server.log | grep '/subscription/webhook\|/session/sign_in'

# Filter by user ID (find all requests for a user)
tail -f server.log | grep '<user-uuid>'
```

---

## Structured log fields (zap)

Every log line in dev mode is human-readable. In production, it's JSON.

Key fields:
| Field | Description |
|---|---|
| `level` | debug / info / warn / error / fatal |
| `ts` | timestamp |
| `caller` | file:line |
| `msg` | log message |
| `error` | error string (on error logs) |
| `event` | Paystack event type (webhook logs) |
| `version` | migration version applied |
| `addr` | server bind address |

Common log messages to watch:
```
"migration applied"          ← good on startup
"all migrations up to date"  ← good on startup  
"server starting"            ← server ready
"paystack webhook received"  ← incoming webhook
"paystack webhook: invalid signature"  ← bad sig (could be attack or misconfigured secret)
"failed to update paystack customer info"  ← WARN — non-fatal
"S3 unavailable"             ← WARN — file uploads disabled
"shutdown signal received"   ← graceful shutdown
```

---

## Docker logs

```bash
# Follow all container logs
make logs

# App only
docker compose logs -f app

# Postgres only
docker compose logs -f db

# Filter errors from app
docker compose logs app | grep -E "ERROR|FATAL|error"

# Last 100 lines
docker compose logs --tail=100 app
```

---

## Heroku logs (staging / production)

```bash
# Tail live logs
heroku logs --tail --app <app-name>

# Last 200 lines
heroku logs -n 200 --app <app-name>

# Filter by dyno
heroku logs --tail --app <app-name> --dyno web

# Filter errors
heroku logs -n 500 --app <app-name> | grep -E '"level":"error"|FATAL'

# Find a specific request by path
heroku logs -n 500 --app <app-name> | grep "POST /subscription/webhook"

# Find all 500 errors
heroku logs -n 500 --app <app-name> | grep '"status":500\|500 '
```

---

## Request tracing

Chi's logger middleware logs every request in this format:
```
2026/04/22 07:33:05 [REQUEST-ID] "METHOD http://host/path HTTP/1.1" from ip - STATUS BYTESb in DURATIONs
```

To trace a full request lifecycle:
```bash
# 1. Make the request with verbose output
curl -v -X POST http://localhost:8080/subscription/webhook \
  -H "Content-Type: application/json" \
  -d '{"event":"charge.success","data":{}}' 2>&1

# 2. Check server logs for the same request ID or timestamp
tail -50 server.log | grep "subscription/webhook"
```

---

## Database query debugging

To temporarily enable GORM SQL logging (shows every query):
```go
// In internal/infrastructure/database/gorm.go, change:
logger.Silent → logger.Info
```

Then rebuild and run. All SQL queries will be logged.
**Always revert to `logger.Silent` before committing.**

```bash
# Or use postgres query log directly
docker compose exec db psql -U hofuser hofdb -c \
  "SELECT query, calls, mean_exec_time FROM pg_stat_statements ORDER BY calls DESC LIMIT 20;"
```

---

## Common error patterns and their causes

| Log message | Cause | Fix |
|---|---|---|
| `cached plan ... does not match` (`SQLSTATE 0A000`) | ALTER TABLE ran while connections were open — pool has stale plans | Restart the server |
| `relation "xxx" does not exist` | Migration not applied | Check `schema_migrations` table; run migration |
| `invalid input: Key: ... failed on ... required` | Validation failure — check request body | Verify field names match DTO struct tags |
| `an unexpected error occurred` | Unclassified error → 500 | Check server logs for the real error from zap |
| `paystack webhook: invalid signature` | Wrong `PAYSTACK_SECRET` env var OR Paystack test vs live key mismatch | Verify `PAYSTACK_SECRET` matches the Paystack dashboard key |
| `S3 unavailable` | AWS credentials missing or wrong | Check `AWS_ENDPOINT`, `AWS_SECRET`, `AWS_BUCKET` |
| `failed to upgrade legacy password hash` | Write to DB after legacy login failed | Check DB write permissions; non-fatal |

---

## Checking the database directly

```bash
# Local (docker)
make db-shell

# Heroku
heroku pg:psql --app <app-name>

# Useful queries
\dt                              -- list all tables
SELECT * FROM schema_migrations ORDER BY applied_at DESC LIMIT 10;
SELECT * FROM users ORDER BY date_added DESC LIMIT 5;
SELECT * FROM subscriptions WHERE deleted_at IS NULL ORDER BY date_added DESC LIMIT 10;
SELECT * FROM global_parameters;
SELECT sub_code, status, next_payment_date FROM subscriptions WHERE sub_code IS NOT NULL;
```
