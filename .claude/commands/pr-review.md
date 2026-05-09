# /pr-review — Pull Request Code Review

Exhaustive code review for the HOF backend. Covers architecture, security, correctness, and style.
Arguments: $ARGUMENTS — branch name, PR number (e.g. `gh pr view 42`), or blank to review the current branch vs master.

---

## Step 0 — Scope the diff

```bash
# If a PR number was given
gh pr diff $ARGUMENTS

# Otherwise review current branch vs master
git diff master...HEAD --stat
git diff master...HEAD
git log master..HEAD --oneline
```

Read every changed file in full before commenting. Do not review from the diff alone.

---

## Step 1 — Architecture / Layer violations

Check import paths in every changed `.go` file.

```bash
# Domain must not import application or infrastructure
grep -rn "application\|infrastructure\|interfaces" internal/domain/

# Application must not import infrastructure or interfaces
grep -rn "infrastructure\|interfaces" internal/application/

# HTTP handlers must not import infrastructure directly
grep -rn "infrastructure" internal/interfaces/http/handler/
```

Flag any violation — this breaks the dependency rule and makes the code untestable.

Check the DDD path is complete for every new feature:
- [ ] New entity field → added to `domain/<domain>/entity.go`
- [ ] New DB operation → added to **both** `domain/<domain>/repository.go` (interface) **and** `infrastructure/persistence/<domain>_repository.go` (implementation)
- [ ] New use case → `application/<domain>/service.go` (interface + impl) and `dto.go`
- [ ] New endpoint → handler + route in `router.go`
- [ ] Schema change → new migration file

---

## Step 2 — Security

### Authentication on routes

```bash
# Find all r.Get / r.Post / r.Put / r.Delete calls in router.go
grep -n "r\.\(Get\|Post\|Put\|Delete\)" internal/interfaces/http/router.go
```

Verify every write route (`Post`, `Put`, `Delete`) that touches user data is inside the `middleware.Authenticate` group.
Verify every admin-only route is inside the `middleware.RequireAdmin` group under `/admin`.

Public routes that are intentionally unauthenticated:
- `/session/*` — sign-in, sign-up, forgot-password, verify-token, send-verify-email
- `/verify_email/{token}` — email link
- `/subscription/webhook` — Paystack (HMAC-verified internally)
- `/health`, `/`

### Response DTO leakage

```bash
# Check handlers that return raw domain entities (exposes password hash, internal IDs, etc.)
grep -n "response\.JSON.*\bu\b\|response\.JSON.*user\b" internal/interfaces/http/handler/user.go
grep -n "response\.JSON" internal/interfaces/http/handler/*.go
```

Flag any handler that passes a raw `*domainUser.User` directly to `response.JSON` — must use `ToUserResponse(u)` or a DTO.
Flag any handler returning full GORM entities with unexported fields or sensitive columns.

### Input validation

```bash
grep -n "validate.Struct\|validate.Var" internal/application/
```

Every `service.go` method that accepts a request struct must call `validate.Struct(req)` or equivalent before processing.
Check that email fields use `validate:"required,email"`, UUIDs use `validate:"required,uuid"`, passwords use `validate:"required,min=6"`.

### OTP / token flow

Check that `ResetPassword` calls `GetPasswordToken` and verifies `token.Validated == true` before updating the password. Without this, the OTP step can be bypassed.

### SQL injection

```bash
grep -n "\.Raw\|\.Exec" internal/infrastructure/persistence/
```

Every `Raw` / `Exec` call must use `?` placeholders, never string interpolation. Flag `fmt.Sprintf` inside any SQL query.

### Paystack webhook

```bash
grep -n "PaystackWebhook" internal/interfaces/http/handler/subscription.go
```

Verify the handler validates the `X-Paystack-Signature` HMAC before processing the payload. The handler must always return 200 regardless of outcome to prevent Paystack retries.

---

## Step 3 — Error handling

```bash
# Find plain errors.New / fmt.Errorf that are not wrapping a typed error
# These map to HTTP 500 — check each one is intentional
grep -n "errors\.New\|fmt\.Errorf" internal/application/ internal/interfaces/
```

Every error that maps to a specific HTTP code **must** use a typed error:

| Situation | Correct typed error |
|---|---|
| Resource not found | `shared.ErrNotFound{Resource: "x", ID: id}` |
| Duplicate / conflict | `shared.ErrAlreadyExists{Resource, Field, Value}` |
| Bad input | `shared.ErrInvalidInput{Field, Message}` |
| No credentials | `shared.ErrUnauthorized{Message}` |
| Insufficient permission | `shared.ErrForbidden{Message}` |
| State conflict | `shared.ErrConflict{Message}` |

Plain `errors.New` or bare `fmt.Errorf` → HTTP 500. Flag every occurrence in application or handler code.

```bash
# Check all error type assertions use the generic helper (not manual errors.As)
grep -n "errors\.As\b" internal/
# Should use errors.AsType[T](err) pattern instead
```

Check service methods handle `shared.IsNotFound(err)` correctly when a missing record is expected (e.g. creating a record that may not exist yet).

---

## Step 4 — Database / GORM patterns

```bash
# Every GORM call must use WithContext
grep -n "r\.db\." internal/infrastructure/persistence/
# Flag any call missing .WithContext(ctx)
```

```bash
# Every query on soft-deleted tables must filter deleted_at
grep -n "\.First\|\.Find\|\.Where" internal/infrastructure/persistence/user_repository.go
```

Tables with `deleted_at`: `users`, `audio_messages`, `audio_series`. Every query on these must include `WHERE deleted_at IS NULL`.

```bash
# Check for Save() vs Updates()
grep -n "\.Save(" internal/infrastructure/persistence/
```

`Save()` overwrites ALL fields including zero values — flag any `Save()` call where only specific fields should change; use `Updates(map[string]any{...})` instead.

Check migrations:
- [ ] File is `NNN_description.sql` with correct sequential number
- [ ] Uses `ADD COLUMN IF NOT EXISTS` / `CREATE TABLE IF NOT EXISTS`
- [ ] Contains `---- create above / drop below ----` separator with matching `DROP` statements
- [ ] Never modifies an already-applied migration — always a new file
- [ ] TIMESTAMPTZ not TIMESTAMP for all time columns
- [ ] Index named `idx_<table>_<col>`
- [ ] Reserved SQL keywords quoted (e.g. `"to"`) or renamed

```bash
# Check GORM struct tags for non-obvious column names
grep -n "gorm:\"" internal/domain/
# Every field whose snake_case name differs from the column name must have gorm:"column:..."
# Reserved words (to, from, order, type) must always have explicit column tags
```

---

## Step 5 — HTTP / API

### Status codes

```bash
grep -n "http\.Status" internal/interfaces/http/handler/
```

Expected codes:
- `201` for successful creation (`POST` that creates a resource)
- `200` for successful reads, updates, deletes
- `400` for invalid input
- `401` for missing/invalid token
- `403` for insufficient permission
- `404` for not found
- `409` for conflict/duplicate
- `500` for unexpected errors (should be rare)

Flag `200` returned from a `POST` handler that creates a new entity (should be `201`).

### Swagger annotation accuracy

```bash
# Extract all @Router annotations and compare to actual router.go registrations
grep -n "@Router" internal/interfaces/http/handler/*.go
grep -n "r\.\(Get\|Post\|Put\|Delete\)" internal/interfaces/http/router.go
```

Every `@Router` path and method must exactly match a registered route.
Every protected route must have `@Security BearerAuth`.
Every admin route must have `@Security BearerAuth`.

```bash
# Check for missing security annotations on protected handlers
grep -B5 "@Router" internal/interfaces/http/handler/*.go | grep -v "Security\|Summary\|Router\|Tags\|Accept\|Produce\|Param\|Success\|Failure\|Description\|^--$"
```

### URL param validation

```bash
grep -n "chi\.URLParam" internal/interfaces/http/handler/
```

Every UUID param must be parsed with `uuid.Parse` and the error handled:
```go
id, err := uuid.Parse(chi.URLParam(r, "some_id"))
if err != nil {
    response.BadRequest(w, "invalid some_id")
    return
}
```

Bare `chi.URLParam` used directly as a string without validation is acceptable only for non-UUID string identifiers.

### Pagination on list endpoints

```bash
grep -n "List\|list" internal/interfaces/http/handler/*.go
```

Every list endpoint must accept and apply `page` / `page_size` query params. Response must go through `response.JSONList(w, status, data, total)`.

---

## Step 6 — Business logic

### Access control for content

```bash
grep -n "access_level\|AccessLevel\|isAccessAllowed\|AccessIn" internal/
```

`GetMessage` must enforce access level — viewer role must be checked against the message's `access_level`. `ListMessages` must filter by `AccessIn` based on the viewer's role.

Valid access levels: `members` → `stewards` → `leaders` (broadest to narrowest).

### Device registration

```bash
grep -n "RegisterDevice\|UpsertDeviceRecord" internal/
```

Every call path to `UpsertDeviceRecord` must deduplicate by `Identifier` before appending. Flag any path that appends without checking for an existing entry with the same identifier.

Device upsert must happen on login (`Login` in `authService`) if a device is provided in the request body.

### Date format

```bash
grep -n "date_released\|DateReleased\|parseDate" internal/
```

`date_released` must be parsed with the `DD/MM/YYYY` format (`"02/01/2006"` Go layout). Any RFC3339 or other format is a bug. Invalid formats must return `shared.ErrInvalidInput`.

### Password flows

```bash
grep -n "PasswordVersion\|MD5Hash\|bcrypt" internal/
```

New users → always `bcrypt`. Legacy MD5 users → upgrade hash on successful login. No new code should introduce MD5 hashing.

`ChangePassword` → must verify the old password before updating.
`ResetPassword` → must verify `token.Validated == true` before updating.

### Admin operations

```bash
grep -n "DeleteAdmin\|AdminSignup" internal/
```

`DeleteAdmin`: caller cannot delete themselves, target must be `church_admin` (otherwise 404, not 403).
`AdminSignup`: only callable with a valid admin JWT (`/admin` group).

---

## Step 7 — Email / Mailer

```bash
grep -rn "\*mailer\.Mailer\|mailer\.New(" internal/application/
```

Application code must **never** depend on `*mailer.Mailer` directly — only on the `mailer.EmailSender` interface. Flag any service that takes or calls `*Mailer`.

```bash
grep -n "go func.*mailer\|go func.*Send" internal/application/
```

`EmailQueue.Send*` methods are already non-blocking — wrapping them in a goroutine is redundant and swallows errors. Flag any `go func() { s.mailer.Send...() }()` pattern.

Adding a new email type requires **all three**:
1. Method on `EmailQueue` calling `q.enqueue(...)`
2. Same method signature on `EmailSender` interface in `mailer.go`
3. Template file under `templates/`

---

## Step 8 — Storage

```bash
grep -rn "GeneratePresignedURL" internal/
```

`GeneratePresignedURL` is **not supported** on Cloudinary — calling it returns `ErrForbidden`. Any code path that calls this must handle the Forbidden error gracefully.

```bash
grep -n "GetMaxFileSize\|fileStorage" internal/interfaces/http/handler/
```

File size must be validated against `fileStorage.GetMaxFileSize()` before uploading. Flag any upload handler that skips this check.

---

## Step 9 — Concurrency

```bash
grep -n "go func\|goroutine\|sync\." internal/application/ internal/infrastructure/
```

Every goroutine must:
- Receive `ctx context.Context` and respect `ctx.Done()`
- Not capture loop variables by reference (use parameter copy: `go func(v val) {...}(variable)`)
- Not write to shared state without a mutex or channel

```bash
grep -n "var.*=.*\[\]" internal/
# Check for package-level slices/maps modified after init (race condition)
```

In tests: package-level vars overridden for test isolation must use `defer` to restore:
```go
original := retryDelays
retryDelays = []time.Duration{...}
defer func() { retryDelays = original }()
```

---

## Step 10 — Code quality

### Logging

```bash
grep -n "s\.log\.\|r\.log\.\|h\.log\." internal/application/ internal/infrastructure/ internal/interfaces/
```

Every log line must include enough context to debug without guessing:
- `zap.String("user_id", ...)` or `zap.String("email", ...)`
- `zap.Error(err)` always on error/warn lines
- Errors logged at `Error` level; expected/recoverable situations at `Warn`; audit events at `Info`

### Nil pointer risks

```bash
grep -n "\*string\|\*time\.Time\|\*uuid\.UUID" internal/domain/
```

Every pointer field access must be guarded:
```go
// Bad
u.Mobile = &req.Mobile  // even when req.Mobile == ""

// Good
if req.Mobile != "" {
    u.Mobile = &req.Mobile
}
```

### Magic values

```bash
grep -n "\"members\"\|\"stewards\"\|\"leaders\"\|\"pending\"\|\"active\"\|\"ACTIVE\"\|\"INACTIVE\"" internal/application/
```

Hard-coded status/role strings in application code should use the domain constants (`domainUser.RoleMember`, `domainUser.DeviceStatusActive`, etc.). Flag raw string literals where a typed constant exists.

### Unused code

```bash
go vet ./...
# Also check for unexported functions with no callers
grep -rn "^func [a-z]" internal/ | grep -v "_test\.go"
```

---

## Step 11 — Config / Environment

```bash
grep -n "os\.Getenv" internal/
```

New environment variables must not be read with `os.Getenv` directly — they must be added to `internal/infrastructure/config/config.go` and `.env.example`. Flag any direct `os.Getenv` outside of `config.go` or `main.go`.

---

## Step 12 — Test coverage

```bash
# Check what is new/changed but has no test file
git diff master...HEAD --name-only | grep -v "_test\.go"

# Find test files for new packages
git diff master...HEAD --name-only | grep "_test\.go"
```

New service methods → unit test required.
New HTTP handlers → integration test under `internal/interfaces/http/` with `//go:build integration` tag.
New queue/retry logic → test with shortened `retryDelays` and atomic counters (see `queue_test.go`).

Integration tests must not run without `TEST_DATABASE_URL`. Guard with build tag:
```go
//go:build integration
```

---

## Step 13 — Final checks

```bash
# Build must be clean
go build ./...

# Vet must be clean
go vet ./...

# All unit tests must pass with race detector
go test ./... -race -count=1

# No new secrets committed
git diff master...HEAD -- '*.env' '*.env.*' | head -20

# No binary files committed
git diff master...HEAD --name-only | xargs file 2>/dev/null | grep -v text
```

```bash
# Lint
make lint

# Vulnerability scan
govulncheck ./...
```

---

## Review verdict

After completing every section, give:

### Summary
One paragraph describing the overall quality and intent of the PR.

### Must fix (blocks merge)
Numbered list — architecture violations, security holes, data-loss bugs, test failures.

### Should fix (high priority)
Numbered list — wrong HTTP codes, missing validation, error-type mismatches, missing soft-delete filters.

### Nice to have (non-blocking)
Numbered list — logging improvements, missing tests for edge cases, minor clarity issues.

### Approved / Changes requested
One line verdict with reason.
