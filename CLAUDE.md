# HOF Backend — Claude Project Context

Heritage of Faith Church backend — audio content platform with subscriptions, user management,
Paystack payments, S3/Cloudinary storage, and Brevo SMTP email. Go 1.26, Chi, GORM, PostgreSQL, zap.

---

## Architecture — DDD (Clean Architecture)

```
cmd/main.go                    ← wiring only; no logic here
internal/
  domain/                      ← Layer 1: entities + interfaces (no imports from other layers)
    shared/errors.go           ← typed errors used everywhere
    user/entity.go             ← User, DeviceRecord, Role, AppVersion
    content/entity.go          ← AudioMessage (with AccessLevel), AudioSeries, Meditation (with Text)
    subscription/entity.go     ← Plan, Offering, PlanOffering, Subscription, GlobalParameters
    */repository.go            ← interfaces only (no GORM, no SQL)
  application/                 ← Layer 2: use cases / business logic (imports domain only)
    auth/service.go            ← Login, AdminLogin, Authenticate → returns SessionResponse
    user/service.go            ← SignUp, UpdateProfile, devices, favourites, etc.
    content/service.go         ← CRUD for messages, series, meditations
    subscription/service.go    ← plans, offerings, verify, webhook dispatch
    */dto.go                   ← request/response structs for each service
  infrastructure/              ← Layer 3: technical implementations (imports domain + application)
    config/config.go           ← all env vars (caarlos0/env)
    database/gorm.go           ← GORM connect + RunMigrations
    persistence/               ← repository implementations (GORM queries)
    security/                  ← JWT (JWTService), bcrypt, MD5
    payment/paystack/          ← Paystack HTTP client + service
    storage/                   ← S3 (primary) + Cloudinary (fallback); factory in storage/factory.go
    mailer/                    ← Brevo SMTP
    logger/                    ← zap
  interfaces/http/             ← Layer 4: HTTP (imports application only)
    handler/                   ← one handler file per domain
    middleware/auth.go         ← UserIDFromContext, Authenticate(jwtSvc)
    response/response.go       ← JSON, JSONList, Error, BadRequest, Unauthorized
    router.go                  ← all routes wired here
    server.go                  ← HTTP server + graceful shutdown
migrations/                    ← sequential SQL files (NNN_description.sql)
templates/                     ← email HTML templates
```

### The golden rule
**Every new feature follows the same path:**
`domain entity/interface → application service → infrastructure impl → HTTP handler → router`

---

## Critical Conventions

### Error handling — THIS IS THE MOST IMPORTANT RULE

All errors that map to specific HTTP codes **MUST** use the typed errors from `domain/shared/errors.go`.
Plain `errors.New(...)` or `fmt.Errorf(...)` without wrapping a typed error maps to **HTTP 500**.

| Typed error | HTTP status |
|---|---|
| `shared.ErrNotFound{Resource, ID}` | 404 |
| `shared.ErrAlreadyExists{Resource, Field, Value}` | 409 |
| `shared.ErrInvalidInput{Field, Message}` | 400 |
| `shared.ErrUnauthorized{Message}` | 401 |
| `shared.ErrForbidden{Message}` | 403 |
| Anything else | 500 |

```go
// ✅ Correct — returns 400
return shared.ErrInvalidInput{Message: "passwords do not match"}

// ❌ Wrong — returns 500
return errors.New("passwords do not match")
```

### GORM column naming
GORM snake_cases struct field names unless you override with `gorm:"column:..."`.
Always check entity tags when fields have non-obvious names (e.g. `Frequency` → `gorm:"column:freq"`).

### Adding a new DDD layer
When adding a feature, always add the repository method to **both**:
1. `internal/domain/<domain>/repository.go` (interface)
2. `internal/infrastructure/persistence/<domain>_repository.go` (implementation)

### JWT / Auth context
- `middleware.Authenticate(jwtSvc)` — sets `userIDKey` in context
- `middleware.UserIDFromContext(ctx)` → `(uuid.UUID, bool)` — reads it back
- `jwtSvc.Middleware` (global) — attaches JWT claims without enforcing (no 401)
- `jwtSvc.PathTokenMiddleware` — extracts token from last URL path segment (email verify links)

### Session response
Every `sign_in`, `sign_in/admin`, and `authenticate` response includes:
- `user` (id, name, email, is_verified, roles)
- `subscription` (status, next_payment_date, plan_name)
- `global_parameters` (activate_subscription)
- `token`, `refresh_token`

### Paystack webhook
`POST /subscription/webhook` is public but verified via HMAC-SHA512 on `X-Paystack-Signature`.
The handler always returns 200 to prevent Paystack retries.
Events handled: `charge.success`, `invoice.update`, `subscription.create`, `subscription.not_renew`, `invoice.payment_failed`.

### Audio message access control
`AudioMessage` has an `access_level` column (`varchar(50)`, default `"members"`).
Valid values and their hierarchy (broadest → narrowest): `"members"` → `"stewards"` → `"leaders"`.
- `"members"` — visible to all authenticated users
- `"stewards"` — stewards and leaders only
- `"leaders"` — leaders only

Set via `access` field in create/update requests (preferred). Legacy `allow_steward: true` maps to `"stewards"`.
`GetMessage` enforces access with `ErrForbidden` when the viewer's role is insufficient.
`ListMessages` accepts `access` query param (viewer role) and filters via `AccessIn`.

### Email queue (async delivery + retry)
All emails go through `mailer.EmailQueue` — never call `*Mailer` directly from application code.

**Flow:** `enqueue()` writes a row to `email_jobs` (status `pending`) then pushes to a buffered channel (200 slots). The background worker drains the channel, calls `mailer.send()`, and updates the row. On failure it reschedules: attempt 1→2 waits 1 min, 2→3 waits 5 min; after 3 failures the row is marked `failed` and logged as an error. A 5-minute DB poll recovers jobs missed by the channel (restarts, full buffer).

**Adding a new email type:**
1. Add a method on `EmailQueue` (e.g. `SendWelcome`) that calls `q.enqueue(to, subject, templateFile, data)`
2. Add the same method signature to the `EmailSender` interface in `mailer.go`
3. Add the template file under `templates/`
4. Call it via the `EmailSender` interface in the application service — never the concrete type

**Key files:**
- `internal/infrastructure/mailer/queue.go` — `EmailQueue`, worker, retry logic
- `internal/infrastructure/mailer/mailer.go` — `EmailSender` interface + `Mailer` (SMTP)
- `migrations/030_create_email_jobs.sql` — `email_jobs` table schema

**Debugging failed emails:** query `SELECT * FROM email_jobs WHERE status = 'failed'`. To manually requeue: `UPDATE email_jobs SET status = 'pending', attempts = 0, scheduled_at = NOW() WHERE id = '...'`.

### Storage backends
`storage/factory.go` selects the backend at startup:
1. **S3** — used when `AWS_SECRET` is set
2. **Cloudinary** — fallback when `CLOUDINARY_CLOUD_NAME` + `CLOUDINARY_API_KEY` + `CLOUDINARY_API_SECRET` are set

Cloudinary does **not** support presigned URLs — calling `GeneratePresignedURL` returns `ErrForbidden`.
Both implement the `Storage` interface (`Upload`, `GeneratePresignedURL`, `GetMaxFileSize`).

### Date format for `date_released`
All `date_released` fields accept **DD/MM/YYYY** format (e.g. `"10/11/2022"`), not RFC3339.
Invalid format returns `ErrInvalidInput` with a clear message.

### Admin-only write routes
All mutating content, subscription, and admin-user routes live under `/admin` and require admin auth:
- `POST /admin/user/create` — create a new admin user (replaces the old public `/session/sign_up/admin`)
- `POST/PUT/DELETE /admin/audio_message/` — create/update/delete audio messages
- `POST/PUT/DELETE /admin/audio_series/` — create/update/delete series
- `POST/PUT/DELETE /admin/audio_message/meditation` — create/update/delete meditations
- `POST/DELETE /admin/subscription/plan/` — plan management
- `POST/DELETE /admin/subscription/offering/` — offering management

Public routes (`/audio_message/`, `/audio_series/`, `/audio_message/meditations`) are read-only.

### Admin bootstrapping
`POST /session/sign_up/admin` **no longer exists** — it was publicly accessible and has been removed.

- **First admin**: run `make seed-admin EMAIL=… FIRST=… LAST=… PASS=…` (see README). The seed tool (`cmd/seed/main.go`) refuses to run if any `church_admin` already exists.
- **Subsequent admins**: existing admin calls `POST /admin/user/create` with a valid admin JWT.
- The seed script reuses existing infrastructure (config, DB, persistence, security) and lives in `cmd/seed/main.go`.

### `errors.AsType` generic helper
Use `errors.AsType[T](err)` instead of `var e T; errors.As(err, &e)` for typed error checks.
This is the pattern used throughout `shared/errors.go` and `response/response.go`.

### Migrations
Files are `NNN_description.sql`, ordered alphabetically. The runner tracks applied versions in
`schema_migrations`. Use the separator `---- create above / drop below ----` to split up/down.
Always `ADD COLUMN IF NOT EXISTS` / `CREATE TABLE IF NOT EXISTS`.
**Never modify an already-applied migration** — create a new numbered file.

### Password versions
Legacy users have MD5 hashed passwords (`password_version = "md5"`).
On successful login with MD5, the hash is auto-upgraded to bcrypt.
New users always get bcrypt (`password_version = "bcrypt"`).

---

## Running Locally

```bash
make env          # create .env from .env.example (first time only)
make run          # go run ./cmd/main.go  (hot-reloads .env)
make build        # compile → bin/server
./bin/server      # run compiled binary
```

Environment variables are loaded from `.env` via godotenv.
Minimum required: `DATABASE_URL`, `JWT_SIGNING_KEY`.
Storage requires either AWS or Cloudinary env vars (see Storage backends above). Paystack degrades gracefully when unconfigured.

```bash
make up           # docker-compose: postgres + app
make down         # stop everything
make db-shell     # psql into postgres container
```

## Tests

```bash
make test                                         # all tests
go test ./... -v -race                            # verbose + race detector
go test ./... -coverprofile=coverage.out          # with coverage
go tool cover -html=coverage.out                  # view in browser

# Integration tests (require DATABASE_URL pointing to a real DB)
go test ./internal/interfaces/http/... -tags integration -v
```

Integration tests use a separate `hofdb_test` database, run migrations, and test via HTTP.
Set `TEST_DATABASE_URL` to point at the test DB; falls back to `DATABASE_URL`.

## Linting & Security

```bash
make lint                    # golangci-lint run ./...
govulncheck ./...            # check for known vulnerabilities
go vet ./...                 # standard vet
```

## Swagger Docs

```bash
make swagger                 # regenerate from annotations
# then visit http://localhost:8080/docs (Scalar UI)
# or  http://localhost:8080/swagger/index.html (Swagger UI)
```

---

## Project-Specific Gotchas

1. **Missing `time` import** — `service.go` files that use `parsePaystackDate` need `"time"` imported.
2. **Cached plan error (`SQLSTATE 0A000`)** — restart the server after `ALTER TABLE` to clear GORM's connection pool plan cache.
3. **`getDeviceRecord` returns 404 when no devices** — this is correct; no device record is created until the user adds a device.
4. **`UpsertDeviceRecord` appends** — calling it twice with the same identifier creates duplicates. The `RegisterDevice` service method already handles deduplication via `Identifier`.
5. **`req.Email` in `InitializeTransaction`** — email comes from the JSON body only; the old `r.URL.Query().Get("email")` override was removed.
6. **`GlobalParameters` seeded on first read** — `GetGlobalParameters` inserts a default row if none exists. Safe to call at startup.
7. **Old Bitbucket pipeline** — `bitbucket-pipelines.yml` deploys to Heroku. GitHub Actions (`.github/workflows/ci.yml`) is the primary CI for the GitHub mirror.
8. **Module path** — `bitbucket.org/hofng/hofApp` (kept as-is to avoid import path churn).
9. **`GOOS=linux CGO_ENABLED=0`** required for Docker builds.
10. **`swag` version** — use `swag@latest`; older versions silently drop some annotations.

---

## Key File Map

| What you want to change | File to edit |
|---|---|
| Add a new route | `internal/interfaces/http/router.go` |
| Add a new handler | `internal/interfaces/http/handler/<domain>.go` |
| Add a use case | `internal/application/<domain>/service.go` + `dto.go` |
| Add a repo method | `internal/domain/<domain>/repository.go` (interface) + `internal/infrastructure/persistence/<domain>_repository.go` (impl) |
| Add a new domain entity | `internal/domain/<domain>/entity.go` + migration |
| Change session response | `internal/application/auth/dto.go` + `service.go:buildSession` |
| Change error mapping | `internal/interfaces/http/response/response.go:classify` |
| Add env var | `internal/infrastructure/config/config.go` + `.env.example` |
| Add email template | `templates/<name>.html` |
| Add a new email type | `internal/infrastructure/mailer/queue.go` + `mailer.go` (interface) + template |
| Change storage backend | `internal/infrastructure/storage/factory.go` |
