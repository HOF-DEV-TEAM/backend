# /debug — Debug & Diagnose

Systematic approach to debugging issues in the HOF backend.
Arguments: $ARGUMENTS — error message, endpoint, or symptom to investigate.

---

## First steps — always

```bash
# 1. Check the server is actually the latest build
ps aux | grep server
# Compare start time — if old binary, kill and rebuild:
make build && ./bin/server &

# 2. Check server logs for the actual error
# (response.Error masks everything as "an unexpected error occurred" in HTTP 500)
tail -50 server.log | grep -E "error|FATAL|warn" | tail -20

# 3. Hit health check
curl -s http://localhost:8080/health
```

---

## Diagnosing HTTP 500 "an unexpected error occurred"

This means an unclassified error reached `response.Error`. The real error is in the logs.

```bash
# Find the real error in logs
tail -100 server.log | grep -E '"level":"error"|"level":"fatal"' | tail -10

# Common causes:
# 1. Plain errors.New() or fmt.Errorf() without typed error → fix: use shared.ErrXxx
# 2. DB connection failure → check DATABASE_URL
# 3. Missing table or column → check migrations applied
# 4. Paystack/S3 provider unreachable → expected in dev, check env vars
```

---

## Diagnosing HTTP 401 Unauthorized

```bash
# 1. Is the token included?
curl -v http://localhost:8080/user/roles -H "Authorization: Bearer $TOKEN" 2>&1 | grep "< HTTP"

# 2. Is the token expired? (JWT payload is base64)
echo "$TOKEN" | cut -d. -f2 | base64 -d 2>/dev/null | grep exp

# 3. Is the route in the protected group?
grep -n "verify_email\|devices\|roles" internal/interfaces/http/router.go

# 4. Does the handler call UserIDFromContext?
# Check handler uses middleware.UserIDFromContext(r.Context())
# NOT jwtSvc.Parse() directly
```

---

## Diagnosing HTTP 404

```bash
# 1. Is the route registered?
grep -n "route-path" internal/interfaces/http/router.go

# 2. Is the handler correct?
grep -n "FunctionName" internal/interfaces/http/router.go

# 3. Check URL path exactly — Chi is strict about trailing slashes
curl -v http://localhost:8080/subscription/plan/     # vs
curl -v http://localhost:8080/subscription/plan      # different!
```

---

## Diagnosing HTTP 400 validation errors

```bash
# The error message tells you which field failed:
# "invalid input: Key: 'CreatePlanRequest.Name' Error:Field validation for 'Name' failed on 'required'"
#
# Check:
# 1. Field names in request body match JSON tags in DTO
# 2. DTO struct is the correct one for this endpoint
# 3. validate:"required" fields are present

# Example: check DTO tags
grep -A10 "CreatePlanRequest" internal/application/subscription/dto.go
```

---

## Diagnosing database errors

```bash
# SQLSTATE 0A000 — cached plan mismatch (after ALTER TABLE)
# Fix: restart server to clear connection pool
taskkill //F //IM server.exe && ./bin/server &

# "relation does not exist"
# Fix: check if migration was applied
make db-shell
SELECT * FROM schema_migrations ORDER BY applied_at DESC LIMIT 10;

# "null value in column violates not-null constraint"
# Fix: check migration added DEFAULT or entity sends value

# "duplicate key value violates unique constraint"
# Fix: use upsert pattern or check IsAlreadyExists before inserting
```

---

## Diagnosing Paystack webhook issues

```bash
# 1. Test webhook receipt (no signature — should 200)
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/subscription/webhook \
  -H "Content-Type: application/json" \
  -d '{"event":"charge.success","data":{}}'
# Expected: 200

# 2. Signature mismatch (logs "invalid signature")
# Check: PAYSTACK_SECRET in .env matches Paystack dashboard webhook secret
echo $PAYSTACK_SECRET

# 3. User not found by customer code
# Check: UpdatePaystackInfo was called after VerifySubscription
# Check: user.paystack_customer_code in DB matches event.data.customer.customer_code
make db-shell
SELECT id, email, paystack_customer_code FROM users WHERE paystack_customer_code IS NOT NULL;

# 4. Plan not found by plan code
# Check: subscription_plans.code matches event.data.plan.plan_code
SELECT id, name, code FROM subscription_plans WHERE deleted_at IS NULL;
```

---

## Diagnosing auth / login issues

```bash
# "invalid credentials" on correct password
# Check password_version field
make db-shell
SELECT email, password_version, LEFT(password, 20) FROM users WHERE email = 'test@example.com';

# If password_version = 'md5':
# Legacy hash — uses MD5Hash(plaintext), not bcrypt
# The login flow auto-upgrades on success

# If password was corrupted (e.g. empty string hashed):
# Reset it:
go run tools/reset_password.go  # or direct DB update with a known bcrypt hash

# bcrypt hash of "admin123" for direct DB update:
# $2a$12$... (generate with: go run -e 'import "golang.org/x/crypto/bcrypt"; ...')
```

---

## Diagnosing signup / device issues

```bash
# "could not register devices during sign up" (WARN, non-fatal)
# Check devices table exists
make db-shell
\d user_devices

# "device record with id X not found" on GET /user/devices/all
# Normal — user never added a device. No device record is created until first device added.

# Signup with devices not persisting
# Check request body uses "devices": [...] (array), not "device": {...}
```

---

## Enable verbose GORM SQL logging temporarily

```go
// internal/infrastructure/database/gorm.go
// Change: logger.Silent → logger.Info
// Rebuild and run — all SQL will be logged
// ALWAYS revert before committing
```

---

## Useful one-liners

```bash
# What port is the server on?
netstat -ano | grep ":8080.*LISTEN"

# Kill server on port 8080 (Windows)
netstat -ano | grep ":8080.*LISTEN" | awk '{print $5}' | xargs taskkill //F //PID

# Find which handler handles a route
grep -rn "route-pattern" internal/interfaces/http/

# Find where a service method is called from
grep -rn "MethodName" internal/interfaces/http/

# Check all routes registered
grep -n "r\.Get\|r\.Post\|r\.Put\|r\.Delete\|r\.Patch" internal/interfaces/http/router.go
```
