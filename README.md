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

## Quick start

### Prerequisites

| Tool              | Version  |
|-------------------|----------|
| Go                | ≥ 1.26   |
| PostgreSQL        | ≥ 14     |
| Docker + Compose  | latest   |

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

## API documentation

| Interface        | URL                                      | Notes                        |
|------------------|------------------------------------------|------------------------------|
| **Scalar UI**    | `http://localhost:8080/docs`             | Modern interactive docs      |
| **Swagger UI**   | `http://localhost:8080/swagger/`         | Classic Swagger explorer     |
| **Raw JSON**     | `http://localhost:8080/swagger/doc.json` | OpenAPI 2.0 spec             |
| **GitHub Pages** | `https://hof-dev-team.github.io/backend`        | Auto-synced on every `master` push |

### Enabling GitHub Pages (auto-sync)

Every push to `master` runs a GitHub Actions workflow that:
1. Regenerates the OpenAPI spec via `swag init`
2. Copies `docs/swagger.json` → `api-docs/swagger.json`
3. Deploys the [Scalar](https://scalar.com) interactive UI to the `gh-pages` branch

**To activate:**
1. Go to your repo → **Settings → Pages**
2. **Source:** `Deploy from a branch`
3. **Branch:** `gh-pages` / `/ (root)`
4. Save — your docs will be live at `https://<org>.github.io/<repo>/`

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
    subscription/                    ← VerifySubscription, InitializeTransaction
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
migrations/                          ← Sequential SQL files (001–021_*.sql)
api-docs/                            ← Static Scalar page deployed to GitHub Pages
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
| POST   | `/session/sign_in`            | Login — returns JWT pair   |
| POST   | `/session/sign_up`            | Register new user          |
| POST   | `/session/authenticate`       | Refresh tokens             |
| POST   | `/session/forgot_password`    | Send OTP reset email       |
| PUT    | `/session/verify_token`       | Verify OTP                 |
| POST   | `/session/verify_email`       | Send email verification    |

### User (JWT required)
| Method | Path                                    | Description                |
|--------|-----------------------------------------|----------------------------|
| POST   | `/user/update`                          | Update profile             |
| POST   | `/user/reset_password`                  | Reset with OTP             |
| POST   | `/user/change_password`                 | Change password            |
| GET    | `/user/roles`                           | List user's roles          |
| POST   | `/user/roles`                           | Assign roles               |
| POST   | `/user/favourite/`                      | Add favourite              |
| GET    | `/user/favourite/favs`                  | List favourites            |
| DELETE | `/user/favourite/delete/{message_id}`   | Remove favourite           |

### Content (JWT required)
| Method | Path                                       | Description                 |
|--------|--------------------------------------------|-----------------------------|
| GET    | `/audio_message/`                          | List messages (paginated)   |
| POST   | `/audio_message/`                          | Create message              |
| GET    | `/audio_message/id/message/{id}`           | Get message                 |
| PUT    | `/audio_message/update/{id}`               | Update message              |
| DELETE | `/audio_message/delete/{id}`               | Soft-delete message         |
| GET    | `/audio_series/`                           | List series                 |
| POST   | `/audio_series/`                           | Create series               |
| GET    | `/audio_series/home`                       | Homepage (series + meditations) |
| GET    | `/audio_message/meditations`               | List meditations            |

### Subscriptions (JWT required)
| Method | Path                             | Description                  |
|--------|----------------------------------|------------------------------|
| POST   | `/subscription/transaction`      | Initialize Paystack payment  |
| POST   | `/subscription/verify`           | Paystack webhook (public)    |
| DELETE | `/subscription/disable/{code}`   | Disable subscription         |
| GET    | `/subscription/plan/`            | List plans                   |
| POST   | `/subscription/plan/`            | Create plan                  |
| GET    | `/subscription/offering/`        | List offerings               |

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

To add a new migration:
```bash
touch migrations/026_my_change.sql
# write your SQL, then restart the app
```

---

## Health

```
GET /         →  200  HOF Backend — running
GET /health   →  200  {"status":"ok"}
```
