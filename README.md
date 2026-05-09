# HOF Backend API

REST API for the **Heritage of Faith Church** mobile application — audio content, subscriptions, and user management.

> **Before contributing, read [CONTRIBUTING.md](CONTRIBUTING.md).**

- **Language:** Go 1.26
- **Framework:** Chi v5
- **ORM:** GORM v2 (PostgreSQL)
- **Auth:** JWT (48 h access / 30 d refresh)
- **Payments:** Paystack
- **Storage:** AWS S3
- **Architecture:** Domain-Driven Design (DDD)

---

## New member checklist

If you just joined the team, do these in order:

```
□ 1. Clone the repo
□ 2. Run `make env` — fill in your .env (ask the team for real values)
□ 3. Run `make setup-hooks` — installs the pre-push hook that blocks direct master pushes
□ 4. Run `make up` — starts Postgres + app in Docker, applies migrations automatically
□ 5. Hit `curl http://localhost:8080/health` — expect {"status":"ok"}
□ 6. Read CLAUDE.md — architecture rules, critical error conventions, gotchas
□ 7. Install Claude Code and open this project — the /commands below become slash commands
□ 8. Install the Jira CLI (`jira init`) and link your account
□ 9. Run `make lint` — fix any lint issues on your first branch before opening a PR
□ 10. Read the "Architecture rules" section below before writing any code
```

---

## Quick start

### Prerequisites

| Tool              | Version  | Install |
|-------------------|----------|---------|
| Go                | ≥ 1.26   | [go.dev/dl](https://go.dev/dl) |
| PostgreSQL        | ≥ 14     | via Docker (see below) |
| Docker + Compose  | latest   | [docs.docker.com](https://docs.docker.com/get-docker) |
| golangci-lint     | v2       | `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest` |
| swag              | latest   | `go install github.com/swaggo/swag/cmd/swag@latest` |

---

## Running locally (without Docker)

**1. Clone and enter the repo**
```bash
git clone <repo-url> hof_backend
cd hof_backend
```

**2. Create your `.env` file**
```bash
make env
```
This copies `.env.example` → `.env`. Open `.env` and fill in your database credentials and secrets.

**3. Start a local Postgres** (or point `DATABASE_URL` at an existing one)
```bash
docker run -d --name hofdb \
  -e POSTGRES_DB=hofdb \
  -e POSTGRES_USER=hofuser \
  -e POSTGRES_PASSWORD=hofpassword \
  -p 5432:5432 \
  postgres:16-alpine
```

**4. Run the server**
```bash
make run
```

The server starts on `http://localhost:8080`.
Migrations are applied automatically on startup.

---

## Running with Docker Compose

The compose file starts **PostgreSQL + the app** together.

**1. Create your `.env` file**
```bash
make env
# Edit .env — set JWT_SIGNING_KEY, PAYSTACK_SECRET, MAILER_PASSWORD, AWS_* etc.
```

**2. Start all services**
```bash
make up
```

**3. Follow logs**
```bash
make logs
```

**4. Stop everything**
```bash
make down
```

> The compose file automatically overrides `DATABASE_URL` to point the app at the
> internal `db` service, so you don't need to change it manually for local Docker use.

---

## All `make` targets

```
make help          Show all targets with descriptions
make setup-hooks   Install git hooks (run once after cloning)
make env           Copy .env.example → .env (skips if .env already exists)
make seed-admin    Bootstrap the first admin user (see below)
make run           Run the app locally (loads .env automatically)
make build         Compile binary → bin/server
make clean         Remove compiled binaries
make swagger       Regenerate Swagger/OpenAPI docs from source annotations
make test          Run all tests
make lint          Run golangci-lint (installs if missing)
make docker-build  Build the Docker image
make up            docker compose up --build -d
make down          docker compose down
make logs          docker compose logs -f
make ps            docker compose ps
make db-shell      Open a psql shell inside the compose postgres container
```

---

## Creating the first admin

Admin accounts are **not publicly self-registerable**. The `POST /session/sign_up/admin` endpoint no longer exists. Admin creation works in two phases:

### Phase 1 — Bootstrap (no admins exist yet)

Use the `seed-admin` command. It connects directly to the database and will **refuse to run** if any `church_admin` already exists, preventing accidental overwrites.

```bash
# Make sure your .env is populated (DATABASE_URL is required)
make seed-admin \
  EMAIL=admin@hofng.org \
  FIRST=John \
  LAST=Doe \
  PASS='SecretPass123!'
```

Or run the Go command directly:

```bash
go run ./cmd/seed \
  -email admin@hofng.org \
  -first-name John \
  -last-name Doe \
  -password 'SecretPass123!'
```

On success you'll see:
```
✓ Admin created: John Doe <admin@hofng.org>
  Sign in at POST /session/sign_in/admin
```

### Phase 2 — Adding more admins (after the first admin exists)

Once an admin exists and can sign in, new admin accounts are created via the admin-protected API:

```
POST /admin/user/create
Authorization: Bearer <admin_jwt>

{
  "first_name": "Jane",
  "last_name":  "Smith",
  "email":      "jane@hofng.org",
  "password":   "SecretPass456!"
}
```

This endpoint requires a valid admin JWT. Regular users cannot reach it.

---

## Email delivery

All outbound emails (password reset, email verification) go through an async queue — the HTTP request returns immediately and delivery happens in the background.

### How it works

```
Request → EmailQueue.Enqueue()
              │
              ├── 1. Write job to email_jobs table (status = pending)
              └── 2. Push to in-memory channel (buffer: 200)
                            │
                     Background worker
                            │
                     SMTP send via Brevo
                            │
                  ┌─────────┴──────────┐
                Success              Failure
                  │                    │
            status = sent        attempt < 3?
                                 ├── yes → schedule retry
                                 │         (1 min / 5 min / 15 min)
                                 └── no  → status = failed (logged)
```

### Retry behaviour

| Attempt | Delay before retry |
|---------|--------------------|
| 1 → 2   | 1 minute           |
| 2 → 3   | 5 minutes          |
| 3       | Permanently failed |

### Resilience on restart

A 5-minute DB poll picks up any `pending` jobs whose `scheduled_at` has arrived. This means:
- Jobs survive server restarts (they're in the DB)
- Jobs missed due to a full channel are retried automatically
- Failed sends are visible in the `email_jobs` table for debugging

### Monitoring failed emails

```sql
-- See all permanently failed emails
SELECT id, "to", subject, attempts, last_error, created_at
FROM email_jobs
WHERE status = 'failed'
ORDER BY created_at DESC;

-- Manually requeue a failed job
UPDATE email_jobs
SET status = 'pending', attempts = 0, scheduled_at = NOW()
WHERE id = '<job-id>';
```

---

## Environment variables

| Variable            | Default                         | Description                                |
|---------------------|---------------------------------|--------------------------------------------|
| `APP_ENV`           | `dev`                           | `dev` or `production`                      |
| `PORT`              | `8080`                          | HTTP listen port                           |
| `SERVER_URL`        | `http://localhost:8080`         | Public base URL (used in email links)      |
| `DATABASE_URL`      | —                               | Full Postgres DSN (overrides fields below) |
| `DB_HOST`           | `localhost`                     | Postgres host                              |
| `DB_PORT`           | `5432`                          | Postgres port                              |
| `DB_NAME`           | `hofdb`                         | Database name                              |
| `DB_USERNAME`       | —                               | Postgres user                              |
| `DB_PASSWORD`       | —                               | Postgres password                          |
| `DB_SSL_MODE`       | `disable`                       | `disable` / `require` / `verify-full`      |
| `JWT_SIGNING_KEY`   | —                               | **Required.** HS256 secret (≥ 32 chars)    |
| `AWS_REGION`        | `us-east-1`                     | S3 region                                  |
| `AWS_ENDPOINT`      | —                               | S3 access key ID                           |
| `AWS_SECRET`        | —                               | S3 secret access key                       |
| `AWS_BUCKET`        | `hof-s3`                        | S3 bucket name                             |
| `PAYSTACK_SECRET`   | —                               | Paystack secret key (`sk_...`)             |
| `MAILER_HOST`       | `smtp-relay.sendinblue.com`     | SMTP host                                  |
| `MAILER_USERNAME`   | —                               | SMTP username                              |
| `MAILER_PASSWORD`   | —                               | SMTP password                              |
| `MAILER_PORT`       | `2525`                          | SMTP port                                  |

See [`.env.example`](.env.example) for the full list with inline comments.

---

## Architecture rules

This project follows **Clean Architecture / DDD**. Every new feature must follow the same layered path — no exceptions.

```
domain → application → infrastructure → interfaces/http
```

### The layers

| Layer | Folder | What belongs here |
|---|---|---|
| **Domain** | `internal/domain/` | Entities, repository interfaces, domain errors. No framework imports. |
| **Application** | `internal/application/` | Use cases (service methods), DTOs. Imports domain only. |
| **Infrastructure** | `internal/infrastructure/` | GORM queries, Paystack client, S3, mailer, JWT. Implements domain interfaces. |
| **HTTP** | `internal/interfaces/http/` | Handlers, middleware, router. Imports application only. |

### The single most important rule — typed errors

All errors that should produce a specific HTTP status code **must** use the typed errors from `internal/domain/shared/errors.go`. A plain `errors.New()` or bare `fmt.Errorf()` will always produce **HTTP 500**.

```go
// ✅ Returns 400
return shared.ErrInvalidInput{Message: "passwords do not match"}

// ✅ Returns 404
return shared.ErrNotFound{Resource: "plan", ID: planID.String()}

// ✅ Returns 409
return shared.ErrAlreadyExists{Resource: "user", Field: "email", Value: email}

// ❌ Returns 500 — wrong
return errors.New("passwords do not match")
```

| Typed error | HTTP |
|---|---|
| `shared.ErrNotFound` | 404 |
| `shared.ErrAlreadyExists` | 409 |
| `shared.ErrInvalidInput` | 400 |
| `shared.ErrUnauthorized` | 401 |
| `shared.ErrForbidden` | 403 |
| anything else | 500 |

### Adding a new feature — checklist

```
□ 1. Add struct/field to entity.go (domain layer)
□ 2. Add method signature to repository.go interface (domain layer)
□ 3. Create migration SQL in migrations/NNN_description.sql
□ 4. Add request/response types to dto.go (application layer)
□ 5. Add method to Service interface and implement it (application layer)
□ 6. Implement the repository method in persistence/ (infrastructure layer)
□ 7. Add HTTP handler (interfaces/http/handler/)
□ 8. Register the route in router.go
□ 9. go build ./... — must be zero errors
□ 10. Write tests, run /deskcheck, run /commit, run /pr
```

### Common gotchas

- **GORM column names** — GORM snake_cases field names. Use `gorm:"column:..."` when the DB column doesn't match. `CreatedAt` must be `gorm:"column:date_added"`, `UpdatedAt` must be `gorm:"column:last_updated"`.
- **Stale query plan error (`SQLSTATE 0A000`)** — happens after `ALTER TABLE` while the server is running. Fix: restart the server.
- **`UserIDFromContext` vs JWT claims** — use `middleware.UserIDFromContext(r.Context())` in handlers to get the authenticated user's ID. Never call `jwtSvc.Parse()` directly in handlers.
- **Never modify an applied migration** — always create a new numbered file. The runner tracks applied versions in `schema_migrations`.
- **`go vet`, `make lint`, and `govulncheck`** must all pass before a PR is mergeable.

---

## Developer workflow

### Day-to-day

```bash
# 1. Pick up a Jira ticket
/jira HOF-123 start

# 2. Create a feature branch
git checkout -b hof-123-my-feature master

# 3. Implement the feature
/implement HOF-123 add phone number to user profile

# 4. Verify end-to-end
/deskcheck user

# 5. Commit
/commit

# 6. Open a PR
/pr

# 7. Monitor CI, address review comments, merge
```

### Branch naming

```
hof-<ticket-id>-<short-description-in-kebab-case>

Examples:
  hof-123-add-phone-to-profile
  hof-124-fix-signup-devices-array
  hof-125-restore-webhook-events
```

### Commit message format

Follows [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short description> [HOF-123]

Types:  feat | fix | refactor | test | chore | docs | migration
Scopes: auth | user | content | subscription | infra | http | ci
```

Examples:
```
feat(subscription): restore Paystack webhook event handling [HOF-98]
fix(user): restore devices array in signup request [HOF-101]
migration: add sub_code column to subscriptions [HOF-99]
test(shared): 100% coverage on domain error types [HOF-105]
```

---

## Claude AI slash commands

When Claude Code is open on this project, the following slash commands are available.
Each one is a complete, project-aware workflow — not a generic helper.

| Command | What it does |
|---|---|
| `/implement` | Full DDD feature workflow — reads the ticket, identifies which layers need changing, writes the code layer by layer (domain → app → infra → HTTP → router), then hands off to `/deskcheck` |
| `/deskcheck` | Runs every API endpoint with curl, checks expected status codes and response shapes, reports any unexpected 500s. Use after every implementation. |
| `/commit` | Build gate → `go vet` → `make lint` → `go test` → stage specific files → write a Conventional Commits message → push |
| `/pr` | Pre-flight checks → `gh pr create` with the full checklist template → links Jira ticket → monitors CI jobs |
| `/deploy` | Verifies Docker build locally → publishes the release target → runs post-deploy smoke test → shows rollback guidance if needed |
| `/logs` | Tails structured logs locally, in Docker, or in production. Includes a diagnosis table for every common error pattern (500, 401, webhook sig failure, stale plan, etc.) |
| `/migrate` | Finds the next migration number, generates the SQL file from a template, cross-references GORM struct tags, and confirms the migration applied on restart |
| `/test` | Runs the test suite with coverage, shows the coverage report, and lists the priority packages to test next |
| `/jira` | Fetches ticket details, transitions state (In Progress → In Review → Done), links PRs to tickets, adds work log comments |
| `/debug` | Systematic diagnosis for HTTP 500 / 401 / 404 / 400 errors, DB errors, Paystack webhook issues, and auth/password problems. Always starts by confirming the right binary is running. |

> These commands live in `.claude/commands/`. Each is a markdown file — read them directly
> if you want to understand or adapt a workflow.

---

## CI/CD pipeline

Every PR and push to `master` or `develop` runs the full pipeline on GitHub Actions
(`.github/workflows/ci.yml`).

### Jobs

| Job | Runs on | What it does |
|---|---|---|
| **Build & Vet** | PR + push | `go build`, `go vet`, `go mod verify` |
| **Lint** | PR + push | `golangci-lint` — config in `.golangci.yml` |
| **Security** | PR + push | `govulncheck` — checks for known vulnerabilities in dependencies |
| **Unit Tests + Coverage** | PR + push | `go test -race`, coverage report, fails if coverage drops below threshold |
| **Integration Tests** | PR + push | Spins up a real Postgres 16 container, runs all migrations, tests the full sign-up/sign-in/webhook flow |
| **Swagger Docs** | push to master only | Regenerates OpenAPI spec and deploys to Netlify |

### Coverage threshold

Current threshold: **2%** (baseline — only two packages have tests so far).
Raise the threshold in `.github/workflows/ci.yml` as coverage grows:

```
2% (now) → 30% → 50% → 70%
```

Priority packages to cover next (highest value):
1. `internal/application/auth` — login, token refresh
2. `internal/application/subscription` — webhook handlers
3. `internal/application/user` — signup, password flows
4. `internal/interfaces/http/handler` — HTTP layer

### Required GitHub secrets

Before the docs workflow can deploy, set these in **GitHub → Settings → Secrets and variables → Actions**:

| Secret | Where to find it |
|---|---|
| `NETLIFY_AUTH_TOKEN` | Netlify → User settings → Applications |
| `NETLIFY_SITE_ID` | Netlify → Site settings → Site details |

The integration test job uses ephemeral values for `JWT_SIGNING_KEY` and `DATABASE_URL` — these are set directly in the workflow and do not need secrets.

### PR rules

- All 4 non-deploy jobs must be green before merging
- Direct pushes to `master` are blocked by the pre-push git hook
- All changes go through a PR — no exceptions

---

## Testing

```bash
# Run all unit tests
make test

# With coverage report
go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out

# View coverage in browser
go tool cover -html=coverage.out

# Integration tests (require a running Postgres)
TEST_DATABASE_URL="postgres://hofuser:hofpassword@localhost:5432/hofdb_test?sslmode=disable" \
  go test ./... -tags integration -v
```

### Test file locations

| Package | Test file |
|---|---|
| `internal/domain/shared` | `errors_test.go` — all error types, 100% coverage |
| `internal/interfaces/http/response` | `response_test.go` — HTTP envelope, status mapping, 100% coverage |
| `internal/interfaces/http` | `integration_test.go` — full API tests (build tag: `integration`) |

### Writing new tests

- Unit tests: same package as the code, `_test.go` suffix, no build tags
- Integration tests: `//go:build integration` at the top, require `TEST_DATABASE_URL`
- Mock the repository interface (not GORM) in unit tests
- Use `t.Skip("requires live <service>")` for anything that needs real credentials

---

## API documentation

| Interface        | URL                                      | Notes                        |
|------------------|------------------------------------------|------------------------------|
| **Scalar UI**    | `http://localhost:8080/docs`             | Modern interactive docs      |
| **Swagger UI**   | `http://localhost:8080/swagger/`         | Classic Swagger explorer     |
| **Raw JSON**     | `http://localhost:8080/swagger/doc.json` | OpenAPI 2.0 spec             |
| **Netlify** | `https://<your-site>.netlify.app`        | Auto-synced on every `master` push |

### Enabling Netlify (auto-sync)

Every push to `master` runs a GitHub Actions workflow that:
1. Regenerates the OpenAPI spec via `swag init`
2. Copies `docs/swagger.json` → `api-docs/swagger.json`
3. Deploys the [Scalar](https://scalar.com) interactive UI to Netlify

**To activate:**
1. Go to your Netlify site → **Site settings**
2. **Build command:** use the generated `api-docs/` folder as the publish directory
3. **Publish directory:** `api-docs`
4. Save — your docs will be live at your Netlify site URL

To regenerate docs locally at any time:
```bash
make swagger
```

---

## Project structure

```
cmd/main.go                          ← Entry point, DI wiring
internal/
  domain/
    shared/errors.go                 ← Typed domain errors (NotFound, Forbidden, etc.)
    user/                            ← User aggregate + Roles many2many + Repository interface
    content/                         ← AudioMessage (allow_steward) + series + Repository
    subscription/                    ← Plans, Subscription, PaymentProvider interface
  application/
    auth/                            ← Login (bcrypt/MD5 upgrade), token refresh
    user/                            ← SignUp, ForgotPassword, AssignRoles
    content/                         ← CRUD messages / series / meditations
    subscription/                    ← VerifySubscription, InitializeTransaction, webhook dispatch
  infrastructure/
    config/config.go                 ← Env-driven config (caarlos0/env + godotenv)
    database/gorm.go                 ← GORM connect + SQL migration runner
    persistence/                     ← GORM implementations of all repo interfaces
    security/jwt.go password.go      ← JWT + bcrypt (transparent MD5 upgrade)
    payment/paystack/                ← Paystack REST adapter
    mailer/ storage/ logger/
  interfaces/http/
    handler/                         ← auth, user, content, subscription, upload, admin
    middleware/auth.go               ← JWT enforcement, UUID extraction
    response/response.go             ← Standard JSON envelope + error→status mapping
    router.go                        ← Chi routing (Scalar at /docs, Swagger at /swagger/*)
    server.go                        ← Graceful shutdown (30 s)
migrations/                          ← Sequential SQL files (NNN_description.sql)
.claude/commands/                    ← Claude Code slash commands (see Developer workflow)
.github/workflows/                   ← GitHub Actions CI/CD pipeline
.golangci.yml                        ← golangci-lint configuration
CLAUDE.md                            ← Full project context for AI-assisted development
api-docs/                            ← Static Scalar page deployed to Netlify
docs/                                ← Generated Swagger spec (do not edit manually)
```

---

## Authentication

All protected endpoints require a `Bearer` token:

```
Authorization: Bearer <access_token>
```

| Token         | TTL  | Obtained via                         |
|---------------|------|--------------------------------------|
| Access token  | 48 h | `POST /session/sign_in`              |
| Refresh token | 30 d | Returned alongside the access token  |

Use `POST /session/authenticate` with the refresh token to get a new pair without re-logging in.

Every sign-in response includes `global_parameters` (feature flags) and `subscription` status
alongside the tokens — clients do not need a separate call for these.

---

## User roles

A user can hold **multiple roles** simultaneously.

| Role            | Description                          |
|-----------------|--------------------------------------|
| `member`        | Default — assigned on sign-up        |
| `steward`       | Access to steward-gated content      |
| `church_friend` | External supporter                   |
| `team_lead`     | Internal team lead                   |
| `church_admin`  | Full administrative access           |

Manage via `POST /user/roles` (assign) and `GET /user/roles` (list).

---

## Content access control (`allow_steward`)

The `allow_steward` flag on audio messages allows steward-role users to access non-free
content without an active subscription:

| Has active sub? | `is_free` | `allow_steward` | Is steward? | Access? |
|-----------------|-----------|-----------------|-------------|---------|
| Yes             | any       | any             | any         | ✓       |
| No              | true      | any             | any         | ✓       |
| No              | false     | true            | Yes         | ✓       |
| No              | false     | false           | any         | ✗       |

---

## Key API routes

### Session (public)
| Method | Path                          | Description                |
|--------|-------------------------------|----------------------------|
| POST   | `/session/sign_in`            | Login — returns JWT pair + subscription + global_parameters |
| POST   | `/session/sign_up`            | Register (accepts `devices: []`) |
| POST   | `/session/authenticate`       | Refresh tokens             |
| POST   | `/session/forgot_password`    | Send OTP reset email       |
| PUT    | `/session/verify_token`       | Verify OTP                 |
| POST   | `/subscription/webhook`       | Paystack webhook (HMAC-verified, always 200) |
| GET    | `/verify_email/{token}`       | Email verification link    |

### User (JWT required)
| Method | Path                                    | Description                |
|--------|-----------------------------------------|----------------------------|
| POST   | `/user/update`                          | Update profile             |
| POST   | `/user/reset_password`                  | Reset with OTP token       |
| POST   | `/user/change_password`                 | Change password (needs old password) |
| POST   | `/user/verify_email`                    | Send email verification link |
| GET    | `/user/roles`                           | List user's roles          |
| POST   | `/user/roles`                           | Assign roles               |
| POST   | `/user/favourite/`                      | Add favourite              |
| GET    | `/user/favourite/favs`                  | List favourites            |
| DELETE | `/user/favourite/delete/{message_id}`   | Remove favourite           |
| GET    | `/user/devices/all`                     | List registered devices    |
| POST   | `/user/devices/add`                     | Register a device          |
| DELETE | `/user/devices/delete/{identifier}`     | Remove a device            |
| PUT    | `/user/devices/update/{identifier}/{status}` | Update device status  |

### Content (JWT required)
| Method | Path                                       | Description                 |
|--------|--------------------------------------------|-----------------------------|
| GET    | `/audio_message/`                          | List messages (paginated)   |
| POST   | `/audio_message/`                          | Create message              |
| GET    | `/audio_message/id/message/{id}`           | Get message                 |
| PUT    | `/audio_message/update/{id}`               | Update message              |
| DELETE | `/audio_message/delete/{id}`               | Soft-delete message         |
| POST   | `/audio_message/meditation`                | Create meditation           |
| GET    | `/audio_message/meditations`               | List meditations            |
| GET    | `/audio_message/meditation/{id}`           | Get meditation              |
| PUT    | `/audio_message/meditation/{id}`           | Update meditation           |
| DELETE | `/audio_message/meditation/delete/{id}`    | Delete meditation           |
| GET    | `/audio_series/`                           | List series                 |
| POST   | `/audio_series/`                           | Create series               |
| GET    | `/audio_series/id/series/{id}`             | Get series                  |
| PUT    | `/audio_series/update/{id}`                | Update series               |
| DELETE | `/audio_series/delete/{id}`                | Delete series               |
| GET    | `/audio_series/home`                       | Homepage (series + meditations) |

### Subscriptions (JWT required)
| Method | Path                                | Description                  |
|--------|-------------------------------------|------------------------------|
| GET    | `/subscription`                     | List all subscriptions       |
| POST   | `/subscription/transaction`         | Initialize Paystack payment  |
| POST   | `/subscription/verify`              | Verify payment + activate sub |
| DELETE | `/subscription/disable/{code}`      | Disable a subscription       |
| GET    | `/subscription/plan`                | List plans                   |
| POST   | `/subscription/plan`                | Create plan                  |
| GET    | `/subscription/plan/{id}`           | Get plan                     |
| DELETE | `/subscription/plan/{id}`           | Delete plan                  |
| GET    | `/subscription/plan/offering`       | List plan offerings          |
| POST   | `/subscription/plan/offering`       | Create plan offering         |
| GET    | `/subscription/offering`            | List offerings               |
| POST   | `/subscription/offering`            | Create offering              |
| DELETE | `/subscription/offering/delete/{id}`| Delete offering              |

### Admin (JWT required)
| Method | Path                   | Description                      |
|--------|------------------------|----------------------------------|
| GET    | `/admin/global`        | Get global parameters            |
| PUT    | `/admin/global`        | Update global parameters         |

### Upload (JWT required)
| Method | Path       | Description        |
|--------|------------|--------------------|
| POST   | `/upload`  | Upload file to S3  |

---

## Migrations

SQL migrations live in `migrations/` and are named `NNN_description.sql`.
The app applies them automatically on startup and tracks applied versions
in a `schema_migrations` table — no external tool needed.

**Rules:**
- Always use `IF NOT EXISTS` / `IF EXISTS` — migrations must be safe to re-run
- Never edit an already-applied migration — create a new numbered file
- Every table needs `date_added`, `last_updated`, `deleted_at` columns
- Always update the GORM entity struct after adding a column

To add a new migration:
```bash
# Find the next number
ls migrations/ | sort | tail -1

# Create the file
touch migrations/026_my_change.sql
# Write SQL, restart the server — it applies automatically
```

See `/migrate` slash command for the full guided workflow.

---

## Paystack webhook

`POST /subscription/webhook` is public but protected by HMAC-SHA512 signature verification.
Set `PAYSTACK_SECRET` in your environment to match your Paystack dashboard webhook secret.

Events handled automatically:
| Event | Action |
|---|---|
| `charge.success` | Sets subscription to Active, updates next payment date |
| `subscription.create` | Creates/updates subscription record |
| `invoice.update` | Updates next payment date |
| `subscription.not_renew` | Sets subscription to Past Due |
| `invoice.payment_failed` | Sets subscription to Canceled |

The webhook always returns HTTP 200 to prevent Paystack retries.

---

## Health

```
GET /         →  200  HOF Backend — running
GET /health   →  200  {"status":"ok"}
```
