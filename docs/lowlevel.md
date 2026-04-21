# HOF Backend — Low-Level Design

## Module map

```
bitbucket.org/hofng/hofApp/
├── cmd/main.go
└── internal/
    ├── domain/
    │   ├── shared/errors.go
    │   ├── user/         entity.go  repository.go  errors.go
    │   ├── content/      entity.go  repository.go  errors.go
    │   └── subscription/ entity.go  repository.go  errors.go
    ├── application/
    │   ├── auth/         dto.go  service.go
    │   ├── user/         dto.go  service.go
    │   ├── content/      dto.go  service.go
    │   └── subscription/ dto.go  service.go
    ├── infrastructure/
    │   ├── config/       config.go
    │   ├── database/     gorm.go
    │   ├── persistence/  user_repository.go  content_repository.go  subscription_repository.go
    │   ├── security/     jwt.go  password.go
    │   ├── mailer/       mailer.go
    │   ├── storage/      s3.go
    │   └── payment/paystack/  client.go  service.go
    └── interfaces/http/
        ├── router.go
        ├── server.go
        ├── middleware/auth.go
        ├── response/response.go
        └── handler/  auth.go  user.go  content.go  subscription.go  upload.go  admin.go
```

---

## Domain layer

### Typed errors (`internal/domain/shared/errors.go`)

All error types are structs with an `Error() string` method. Predicate helpers
(`IsNotFound`, `IsInvalidInput`, etc.) allow the HTTP handler layer to map errors
to status codes without importing domain packages directly.

```go
type ErrNotFound     struct{ Resource, ID string }
type ErrAlreadyExists struct{ Resource, Field, Value string }
type ErrInvalidInput  struct{ Field, Message string }
type ErrUnauthorized  struct{ Message string }
type ErrForbidden     struct{ Message string }
type ErrConflict      struct{ Message string }
```

### User aggregate (`internal/domain/user/entity.go`)

```go
type User struct {
    ID              uuid.UUID
    Email           string           // unique index
    Password        string           // bcrypt or legacy MD5 hash
    PasswordVersion PasswordVersion  // "bcrypt" | "md5"
    IsVerified      VerificationStatus
    Roles           []Role           // many2many:user_roles
    ...
}

type Role struct {
    ID   uuid.UUID
    Name RoleName  // "steward" | "member" | "church_friend" | "team_lead" | "church_admin"
}
```

**Many-to-many relationship** is managed by GORM through the `user_roles` join table:
```sql
user_roles(user_id UUID, role_id UUID, PRIMARY KEY(user_id, role_id))
```

### AudioMessage (`internal/domain/content/entity.go`)

```go
type AudioMessage struct {
    ID           uuid.UUID
    Title        string
    AudioURL     string
    IsFree       bool
    AllowSteward bool       // ← new: stewards can access without subscription
    SeriesID     *uuid.UUID
    DeletedAt    *time.Time // soft delete
    ...
}
```

Access control logic:

| User has active sub? | `is_free` | `allow_steward` | User is steward? | Access? |
|----------------------|-----------|-----------------|------------------|---------|
| Yes                  | any       | any             | any              | ✓       |
| No                   | true      | any             | any              | ✓       |
| No                   | false     | true            | Yes              | ✓       |
| No                   | false     | false           | any              | ✗       |

---

## Infrastructure layer

### GORM ORM (`internal/infrastructure/database/gorm.go`)

GORM v2 with the `gorm.io/driver/postgres` adapter (backed by pgx/v5).
All repository implementations use GORM's fluent API rather than raw SQL, except for
complex JSONB array operations (devices, favourites) which use `db.Exec` / `db.Raw`.

```go
// Example: paginated message query
db.WithContext(ctx).
    Where("deleted_at IS NULL").
    Where("title ILIKE ?", "%"+search+"%").
    Order("date_added DESC").
    Offset((page-1)*pageSize).
    Limit(pageSize).
    Find(&messages)
```

### Password hashing (`internal/infrastructure/security/password.go`)

| Function           | Algorithm | Use                              |
|--------------------|-----------|----------------------------------|
| `HashPassword`     | bcrypt 12 | New password storage             |
| `CheckPasswordBcrypt` | bcrypt | Verify bcrypt-stored passwords  |
| `MD5Hash`          | MD5 (hex) | Legacy password comparison only  |

**Upgrade flow** (in `internal/application/auth/service.go`):
1. Login: password_version == "md5" → compare MD5 hashes
2. Match → bcrypt the plaintext → `UpdatePassword(..., PasswordVersionBcrypt)`
3. Next login uses bcrypt path

### JWT service (`internal/infrastructure/security/jwt.go`)

```
Claims { UserID string; jwt.RegisteredClaims }

IssueAccessToken(userID)  → HS256, 48h TTL
IssueRefreshToken(userID) → HS256, 30d TTL
Parse(tokenStr)           → *Claims or error
```

Middleware attaches parsed claims to `context.Context` via `security.WithClaims`.
The `middleware.Authenticate` middleware then extracts the `UserID` uuid and stores it
under a private key for handler use.

### JSONB fields

Devices and favourites are stored as JSONB arrays in PostgreSQL. The domain types
(`DeviceList`, `FavouriteList`) implement `driver.Valuer` and `sql.Scanner` for
transparent JSON serialisation. GORM serializes these using its built-in JSON serializer tag.

Complex mutations (delete a single element from a JSONB array by index) use raw SQL
with PostgreSQL's `jsonb_array_elements` + ordinality pattern.

---

## Application layer

Each service follows the same pattern:

```
Service interface (exported)
    ↓
serviceImpl struct (unexported)
    ↓  injects:
    ├── domain.Repository (interface)
    ├── other application services (interface)
    └── infrastructure helpers (concrete)
```

### Auth service (`internal/application/auth/service.go`)

```
Login(ctx, LoginRequest) → SessionResponse
  1. GetByEmail
  2. checkPassword (bcrypt / MD5 upgrade)
  3. buildSession (issue JWT pair + resolve subscription status)

Authenticate(ctx, AuthenticateRequest) → SessionResponse
  1. Parse refresh token (must be valid)
  2. Parse access token (ignore expiry)
  3. Confirm user IDs match
  4. buildSession
```

### User service (`internal/application/user/service.go`)

Key flows:

| Use case       | Steps                                                        |
|----------------|--------------------------------------------------------------|
| SignUp         | Validate → bcrypt password → Create user → assign `member` role → register device |
| ForgotPassword | Lookup user → generate 6-digit OTP → upsert token → send email (goroutine) |
| AssignRoles    | Parse role names → find Role records → GORM many2many append |

---

## HTTP interface layer

### Standard response envelope

All responses use a consistent JSON structure:

```json
{
  "success": true,
  "data": { ... },
  "total": 42        // only for list endpoints
}

{
  "success": false,
  "error": "audio message not found"
}
```

### Error → HTTP status mapping (`internal/interfaces/http/response/response.go`)

| Domain error        | HTTP status |
|---------------------|-------------|
| `ErrNotFound`       | 404         |
| `ErrAlreadyExists`  | 409         |
| `ErrInvalidInput`   | 400         |
| `ErrUnauthorized`   | 401         |
| `ErrForbidden`      | 403         |
| everything else     | 500         |

### Middleware chain

```
CORS → RequestID → Logger → Recoverer → JWT.Middleware (claims → ctx)
   └─ protected group: middleware.Authenticate (uuid → ctx)
```

`JWT.Middleware` is non-blocking — it runs on every request and attaches claims
if a valid token is present, but does **not** reject requests without a token.
`middleware.Authenticate` enforces presence and returns 401 if no claims are found.

---

## Database schema (key tables)

```sql
users (id, first_name, last_name, email UNIQUE, password, password_version, is_verified, ...)
roles (id, name UNIQUE, description)
user_roles (user_id FK users, role_id FK roles, PK(user_id, role_id))

audio_messages (id, title, audio_url, is_free, allow_steward, series_id FK audio_series, deleted_at, ...)
audio_series   (id, title, of_the_month, deleted_at, ...)
meditations    (id, name, status, deleted_at, ...)

subscription_plans     (id, name, type, freq, fee, currency, code, ...)
subscription_offerings (id, name, status, ...)
subscription_plan_offerings (id, plan_id FK, offering_id FK, ...)
subscriptions (id, user_id FK users, subscription_plan_id FK, status, next_payment_date, ...)

devices   (id, user_id FK UNIQUE, devices JSONB)
favourites (id, user_id FK UNIQUE, fav JSONB)
app_version (id, version, force)
global_parameters (id, activate_subscription)
user_password_token (id, email, password_reset_token, password_reset_at, validated)
```

---

## Startup sequence (`cmd/main.go`)

```
1. Init logger (Zap)
2. Load config (env vars via caarlos0/env)
3. Connect GORM (postgres driver)
4. Run Tern migrations (embedded SQL files)
5. Create repositories (userRepo, contentRepo, subRepo)
6. Create infrastructure services (JWT, mailer, S3, Paystack)
7. Create application services (auth, user, content, subscription)
8. Build HTTP server (Chi router + all handlers)
9. srv.Run() — blocks until SIGINT/SIGTERM, then graceful shutdown (30s)
```
