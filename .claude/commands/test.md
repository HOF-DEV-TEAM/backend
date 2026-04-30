# /test — Testing

Write, run, and check test coverage for the HOF backend.
Arguments: $ARGUMENTS — optional package path or test name filter.

---

## Running tests

```bash
# All tests
go test ./...

# All tests with race detector (use in CI and before committing)
go test ./... -race -count=1

# Specific package
go test ./internal/domain/shared/... -v
go test ./internal/interfaces/http/... -v

# Specific test by name
go test ./... -run TestSignUp -v

# With coverage
go test ./... -race -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out        # summary
go tool cover -html=coverage.out        # open in browser

# Integration tests only (require TEST_DATABASE_URL)
TEST_DATABASE_URL="postgres://hofuser:hofpassword@localhost:5432/hofdb_test?sslmode=disable" \
  go test ./... -tags integration -v -race
```

---

## Test file locations

| What to test | File location |
|---|---|
| Domain errors, entities | `internal/domain/<domain>/<file>_test.go` |
| Application service logic | `internal/application/<domain>/service_test.go` |
| HTTP handlers + routes | `internal/interfaces/http/handler/<domain>_test.go` |
| Response helpers | `internal/interfaces/http/response/response_test.go` |
| Repository (integration) | `internal/infrastructure/persistence/<domain>_repository_test.go` |
| Full API integration | `internal/interfaces/http/integration_test.go` |

---

## Unit test patterns

### Testing domain error types

```go
func TestErrNotFound_Error(t *testing.T) {
    err := shared.ErrNotFound{Resource: "user", ID: "abc-123"}
    assert.Equal(t, "user with id 'abc-123' not found", err.Error())
    assert.True(t, shared.IsNotFound(err))
    assert.False(t, shared.IsInvalidInput(err))
}
```

### Testing service with mocked repository

Use a hand-written mock or `github.com/stretchr/testify/mock`:

```go
type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domainUser.User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil { return nil, args.Error(1) }
    return args.Get(0).(*domainUser.User), args.Error(1)
}

func TestLogin_InvalidCredentials(t *testing.T) {
    repo := &mockUserRepo{}
    repo.On("GetByEmail", mock.Anything, "bad@example.com").
        Return(nil, shared.ErrNotFound{Resource: "user", ID: "bad@example.com"})

    svc := auth.NewService(repo, nil, jwtSvc, zaptest.NewLogger(t))
    _, err := svc.Login(context.Background(), auth.LoginRequest{
        Email: "bad@example.com", Password: "wrong",
    })
    assert.ErrorIs(t, err, domainUser.ErrInvalidCredentials)
}
```

### Testing HTTP handlers

```go
func TestVerifySubscription_Unauthorized(t *testing.T) {
    svc := &mockSubService{}
    h := handler.NewSubscriptionHandler(svc, "", zaptest.NewLogger(t))

    req := httptest.NewRequest(http.MethodPost, "/subscription/verify", nil)
    w := httptest.NewRecorder()
    h.VerifySubscription(w, req)

    assert.Equal(t, http.StatusUnauthorized, w.Code)
}
```

---

## Integration test pattern

Integration tests use the `integration` build tag and require a real database.

```go
//go:build integration

package http_test

import (
    "testing"
    "net/http"
    "net/http/httptest"
    // ...
)

func TestIntegration_HealthCheck(t *testing.T) {
    srv := setupTestServer(t)  // starts server with test DB
    defer srv.Close()

    resp, err := http.Get(srv.URL + "/health")
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func setupTestServer(t *testing.T) *httptest.Server {
    t.Helper()
    dsn := os.Getenv("TEST_DATABASE_URL")
    if dsn == "" {
        t.Skip("TEST_DATABASE_URL not set")
    }
    // wire up DB, run migrations, build router
    // ...
}
```

---

## Coverage requirements

Current coverage threshold: **30%** (enforced in CI, raise as test suite grows).

To check current coverage locally:
```bash
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out | grep total
```

Priority packages to cover first (highest value):
1. `internal/domain/shared` — error types (easy, high confidence)
2. `internal/interfaces/http/response` — HTTP response mapping (easy)
3. `internal/application/auth` — login, token refresh (high risk)
4. `internal/application/subscription` — webhook handlers (business critical)
5. `internal/application/user` — signup, password flows (user-facing)

---

## Test dependencies to add (if not present)

```bash
go get github.com/stretchr/testify@latest
go get github.com/stretchr/testify/mock@latest
```

Both are already likely in the dependency tree via other packages.

---

## CI coverage gate

The CI pipeline enforces `COVERAGE_THRESHOLD=2` (current baseline — raise as tests are added).
Progression target: 2 → 30 → 50 → 70.

To update the threshold:
```yaml
# .github/workflows/ci.yml
env:
  COVERAGE_THRESHOLD: 30  # raise this as test suite matures
```
