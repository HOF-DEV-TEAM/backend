# Contributing Guidelines

## Setup

After cloning, run this once to install the git hooks:

```bash
make setup-hooks
```

This installs a `pre-push` hook that **blocks direct pushes to `master`** on your machine.

## Branch Rules

- **Never push directly to `master`.** It is the production branch.
- All work must be done on a feature or fix branch and merged via a Pull Request.
- PRs require at least **one approval** before merging.
- Delete your branch after it is merged.

## Branch Naming

| Type | Pattern | Example |
|------|---------|---------|
| Feature | `feat/<short-description>` | `feat/add-subscription-plan` |
| Bug fix | `fix/<short-description>` | `fix/auth-token-expiry` |
| Hotfix | `hotfix/<short-description>` | `hotfix/missing-migration` |
| Chore | `chore/<short-description>` | `chore/update-dependencies` |

## Workflow

```bash
# 1. Always branch from master
git checkout master
git pull github master

# 2. Create your branch
git checkout -b feat/your-feature

# 3. Work, commit, push
git push github feat/your-feature

# 4. Open a PR on GitHub → master
# 5. Get approval → merge → delete branch
```

## Commit Messages

Follow the format: `type: short description`

```text
feat: [ticket-id] add plan code to subscriptions
fix: [ticket-id] resolve token refresh race condition
chore: [ticket-id] bump Go version to 1.26
refactor: [ticket-id] replace interface{} with any
```

Types: `feat`, `fix`, `chore`, `refactor`, `docs`, `test`
