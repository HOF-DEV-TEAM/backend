# HOF Backend API v2

REST API for the **Heritage of Faith Church** audio content and subscription platform.

---

## Quick start

### Prerequisites
| Tool       | Version  |
|------------|----------|
| Go         | ≥ 1.21   |
| PostgreSQL | ≥ 14     |

### Run locally

```bash
# Install dependencies
go mod download

# Set required env vars (see table below)
export DATABASE_URL="postgres://..."
export JWT_SIGNING_KEY="your-secret"

# Start the server
go run ./cmd/main.go
```

Default port: **8080** — override with `PORT`.  
Swagger UI: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

---

## Environment variables

| Variable          | Default | Required | Description                        |
|-------------------|---------|----------|------------------------------------|
| `DATABASE_URL`    | —       | **Yes**  | Full PostgreSQL connection string  |
| `JWT_SIGNING_KEY` | —       | **Yes**  | HMAC key for signing JWTs          |
| `PORT`            | `8080`  | No       | HTTP listen port                   |
| `APP_ENV`         | `dev`   | No       | `dev` / `staging` / `production`   |
| `AWS_ENDPOINT`    | —       | No       | AWS access key ID                  |
| `AWS_SECRET`      | —       | No       | AWS secret access key              |
| `AWS_BUCKET`      | `hof-s3`| No       | S3 bucket name                     |
| `PAYSTACK_SECRET` | —       | No       | Paystack secret key                |
| `MAILER_HOST`     | —       | No       | SMTP host                          |
| `MAILER_USERNAME` | —       | No       | SMTP username                      |
| `MAILER_PASSWORD` | —       | No       | SMTP password                      |

---

## API routes

| Group         | Base path               | Auth |
|---------------|-------------------------|------|
| Auth          | `/session`              | No   |
| Users         | `/user`                 | Yes  |
| Audio content | `/audio_message`        | Yes  |
| Audio series  | `/audio_series`         | Yes  |
| Subscriptions | `/subscription`         | Yes  |
| Admin         | `/admin`                | Yes  |
| File upload   | `/upload`               | Yes  |
| Webhook       | `/subscription/webhook` | No   |

Full interactive docs at `/swagger/index.html`.

---

## User roles (1:M)

A user may hold **multiple roles simultaneously**.

| Role            | Description                                     |
|-----------------|-------------------------------------------------|
| `member`        | Default; assigned on sign-up                    |
| `steward`       | Elevated content access via `allow_steward`     |
| `church_friend` | Friend or visitor of the church                 |
| `team_lead`     | Internal team lead                              |
| `church_admin`  | Full administrator                              |

Assign roles: `POST /user/roles`

### `allow_steward` on AudioMessage

When `allow_steward = true`, users with the `steward` role can access that message
without an active subscription, even when `is_free = false`.

---

## Authentication

```
Authorization: Bearer <access_token>
```

Access tokens expire after **48 h**. Refresh at `POST /session/authenticate`.

Passwords are hashed with **bcrypt** (cost 12). Legacy MD5 accounts upgrade automatically on next login.

---

## Database migrations

Migrations run automatically on startup via [Tern](https://github.com/jackc/tern).

| File                                              | Purpose                          |
|---------------------------------------------------|----------------------------------|
| `018_create_roles.sql`                            | Roles table + seed data          |
| `019_create_user_roles.sql`                       | User↔Role many-to-many join      |
| `020_add_allow_steward_and_password_version.sql`  | New message + user columns       |
| `021_create_global_parameters.sql`                | App-wide feature flags table     |

---

## Project layout

```
cmd/main.go                        — Entry point & DI wiring
internal/
  domain/                          — Pure business logic (no external deps)
    shared/errors.go               — Typed domain errors
    user/                          — User aggregate + roles + repository interface
    content/                       — AudioMessage (allow_steward), AudioSeries, Meditation
    subscription/                  — Plans, subscriptions, payment provider port
  application/                     — Use-case orchestration layer
    auth/                          — Login, token refresh
    user/                          — Sign-up, profile, devices, favourites, roles
    content/                       — CRUD for messages, series, meditations
    subscription/                  — Plans, offerings, Paystack flows, global params
  infrastructure/                  — Concrete implementations
    config/                        — Env-var driven configuration
    database/                      — GORM connection + Tern migration runner
    persistence/                   — GORM repository implementations
    security/                      — JWTService + bcrypt password helpers
    mailer/                        — SMTP email delivery
    storage/                       — AWS S3 upload
    payment/paystack/              — Paystack REST client + domain adapter
  interfaces/http/                 — HTTP delivery
    router.go                      — Chi route wiring
    server.go                      — Graceful HTTP server
    middleware/auth.go             — JWT extraction + context injection
    response/response.go           — Standard JSON envelope helpers
    handler/                       — auth, user, content, subscription, upload, admin
migrations/                        — SQL migration files
templates/                         — Email HTML templates
docs/                              — Swagger + architecture docs
```

---

## Docker

```bash
docker build -t hof-backend .
docker run -p 8080:8080 --env-file .env hof-backend
```
