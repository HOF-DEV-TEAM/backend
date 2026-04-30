# /commit — Git Commit

Stage, validate, and commit changes following HOF backend conventions.
Arguments: $ARGUMENTS — optional commit message override.

---

## Step 1 — Status check

```bash
git status
git diff --stat
```

Review every modified file. Never commit:
- `.env` (contains secrets)
- `bin/` (compiled binaries)
- `*.log` files
- `server` (local dev binary)

---

## Step 2 — Build gate

```bash
go build ./...
```
Must be zero errors before committing. Fix any compile errors first.

---

## Step 3 — Vet + lint

```bash
go vet ./...
make lint
```

Fix all vet and lint issues. If a lint rule is genuinely incorrect for this codebase,
add a `//nolint:<linter> // reason` comment — never disable globally.

---

## Step 4 — Tests

```bash
go test ./... -race
```

All tests must pass. If a test is failing due to a flaky external dependency, skip it
with `t.Skip("requires live <service>")` and note it in the commit message.

---

## Step 5 — Stage files

Stage specific files — never `git add -A` blindly:

```bash
# Stage by file
git add internal/application/subscription/service.go
git add internal/domain/subscription/repository.go
git add migrations/025_add_plan_offering_fields.sql
# etc.
```

Always verify what's staged:
```bash
git diff --cached --stat
```

---

## Step 6 — Commit message

Follow Conventional Commits format:
```
<type>(<scope>): <short description>

<body — what and why, not how>
```

Types:
- `feat` — new feature
- `fix` — bug fix
- `refactor` — code change with no behaviour change
- `test` — adding or fixing tests
- `chore` — tooling, deps, CI
- `docs` — documentation only
- `migration` — database schema change

Scopes: `auth`, `user`, `content`, `subscription`, `infra`, `http`, `ci`

Examples:
```
feat(subscription): restore Paystack webhook event handling

Ported charge.success, invoice.update, subscription.create,
subscription.not_renew, invoice.payment_failed handlers from
the pre-DDD codebase. Adds HMAC-SHA512 signature verification.

feat(user): change signup devices field from singular to array

Old API expected devices:[...]; new singular device field broke
existing mobile clients. Restored array format with multi-device
support at signup.

migration: add name/fee/code/currency fields to plan_offerings

These fields were added to the PlanOffering entity in the DDD
rewrite but the migration was never applied. Fixes CreatePlanOffering.
```

---

## Step 7 — Commit

```bash
git commit -m "$(cat <<'EOF'
<type>(<scope>): <description>

<body>

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Step 8 — Post-commit check

```bash
git log --oneline -5
git status  # should be clean
```

If on a feature branch, push:
```bash
git push origin HEAD
```
