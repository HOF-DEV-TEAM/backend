# /pr — Pull Request

Create or update a GitHub Pull Request for the current branch.
Arguments: $ARGUMENTS — optional PR title override or Jira ticket ID (e.g. HOF-123).

---

## Step 1 — Pre-flight

```bash
# Confirm you're on a feature branch (not master)
git branch --show-current

# Confirm everything is committed
git status   # should be clean

# Confirm CI will pass locally
go build ./...
go vet ./...
go test ./... -race -count=1
make lint
govulncheck ./...
```

---

## Step 2 — Push branch

```bash
git push origin HEAD -u
```

---

## Step 3 — Check what's in this PR

```bash
# All commits vs master
git log master..HEAD --oneline

# All files changed vs master
git diff master...HEAD --stat
```

---

## Step 4 — Create the PR

```bash
gh pr create \
  --base master \
  --title "<type>(<scope>): <description>" \
  --body "$(cat <<'EOF'
## Summary
- <bullet: what changed>
- <bullet: why it changed>
- <bullet: any migration or breaking change>

## Jira
<!-- HOF-123 or N/A -->

## Type of change
- [ ] Bug fix
- [ ] New feature
- [ ] Refactor
- [ ] Migration
- [ ] CI/tooling

## Test plan
- [ ] `go build ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `make lint` passes
- [ ] `govulncheck ./...` clean
- [ ] `/deskcheck` run against local server
- [ ] Migrations applied cleanly on fresh DB

## Breaking changes
<!-- List any API contract changes, removed fields, renamed endpoints -->

## Checklist
- [ ] No `.env` or secrets committed
- [ ] All new errors use `shared.ErrXxx` typed errors (not plain `errors.New`)
- [ ] New repository methods added to both interface and implementation
- [ ] Migration uses `IF NOT EXISTS` / `IF EXISTS`
- [ ] Swagger annotations updated (if endpoint changed)
EOF
)"
```

---

## Step 5 — Link Jira (if applicable)

```bash
# If using Jira CLI
jira issue transition HOF-123 "In Review"
jira issue comment HOF-123 "PR opened: $(gh pr view --json url -q .url)"
```

---

## Step 6 — Request review

```bash
gh pr edit --add-reviewer <github-username>
```

---

## Step 7 — Monitor CI

```bash
gh pr checks --watch
```

Expected CI jobs (all must pass):
- `build` — go build + go vet
- `lint` — golangci-lint
- `security` — govulncheck
- `test` — go test with coverage ≥ threshold

If a job fails, check:
```bash
gh run view --log-failed
```

---

## PR title conventions

```
feat(subscription): add plan offering price fields
fix(user): restore devices array in signup request
migration: add sub_code column to subscriptions
chore(ci): add govulncheck to PR pipeline
```

Never merge directly to master — all changes go through PRs.
The pre-push hook blocks direct master pushes.
