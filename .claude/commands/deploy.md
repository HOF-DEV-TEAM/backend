# /deploy — Deployment

Deploy HOF backend to the target environment.
Arguments: $ARGUMENTS — target environment: `staging` (default) or `production`.

---

## Environments

| Environment | Platform | Branch | URL |
|---|---|---|---|
| Staging | Heroku | `develop` | Configured in Heroku dashboard |
| Production | Heroku | `master` | `https://my-heritage-app-1e457dfa2e9c.herokuapp.com` |

---

## Pre-deploy checklist

```bash
# 1. Confirm all tests pass
go test ./... -race -count=1
make lint
govulncheck ./...

# 2. Confirm build is clean
go build ./...

# 3. Confirm migrations are ready
ls migrations/ | sort | tail -5
# Verify latest migration has been tested locally

# 4. Check for any pending env var changes
diff .env.example .env
# Any new keys in .env.example must be set in Heroku config vars before deploy
```

---

## Docker build (local verification)

```bash
# Build the production image
make docker-build

# Run it locally with env vars
docker run --rm \
  --env-file .env \
  -e DATABASE_URL="$DATABASE_URL" \
  -p 8080:8080 \
  hof-backend

curl http://localhost:8080/health
# Expected: {"status":"ok"}
```

---

## Deploy to Heroku (manual)

### Staging

```bash
# Login to Heroku container registry
heroku container:login

# Build and push
heroku container:push web --app <staging-app-name>

# Release
heroku container:release web --app <staging-app-name>

# Confirm running
heroku logs --tail --app <staging-app-name>
```

### Production

```bash
heroku container:login
heroku container:push web --app <prod-app-name>
heroku container:release web --app <prod-app-name>
heroku logs --tail --app <prod-app-name>
```

---

## Deploy via GitHub Actions (automated)

On merge to `master`, the CI pipeline:
1. Builds the Docker image
2. Pushes to Heroku container registry
3. Releases the new image
4. Runs health check

Monitor:
```bash
gh run list --workflow=ci.yml --limit 5
gh run view <run-id>
```

---

## Post-deploy verification

```bash
BASE=https://my-heritage-app-1e457dfa2e9c.herokuapp.com  # or staging URL

# Health check
curl -s $BASE/health
# Expected: {"status":"ok"}

# Sign in smoke test
curl -s -X POST $BASE/session/sign_in \
  -H "Content-Type: application/json" \
  -d '{"email":"<admin-email>","password":"<admin-pass>"}' | \
  grep -o '"success":[^,]*'
# Expected: "success":true

# Check migrations applied
heroku run ./server --app <app-name> -- migrate-status 2>/dev/null || \
heroku logs --app <app-name> | grep "migration applied"
```

---

## Rollback

```bash
# List recent releases
heroku releases --app <app-name>

# Roll back one version
heroku rollback --app <app-name>

# Roll back to specific version
heroku rollback v42 --app <app-name>
```

---

## Environment variables (Heroku config vars)

```bash
# List current
heroku config --app <app-name>

# Set a new var
heroku config:set PAYSTACK_SECRET=sk_live_xxx --app <app-name>

# Required vars (must all be set before first deploy):
# DATABASE_URL, JWT_SIGNING_KEY, JWT_SECRET
# PAYSTACK_ADDR, PAYSTACK_SECRET
# MAILER_EMAIL, MAILER_HOST, MAILER_USERNAME, MAILER_PASSWORD, MAILER_PORT
# AWS_REGION, AWS_ENDPOINT, AWS_SECRET, AWS_BUCKET, AWS_BASE_URL
# SERVER_URL, APP_ENV=production
```

---

## Migrations on deploy

Migrations run automatically at server startup via `database.RunMigrations`.
The `schema_migrations` table tracks applied versions — re-running is safe (idempotent).
If a migration fails, the server will fatal-exit — check `heroku logs` immediately.
