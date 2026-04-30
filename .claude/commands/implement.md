# /implement — Feature Implementation

Full end-to-end workflow for implementing a new feature in the HOF backend.
Arguments: $ARGUMENTS — the feature description, Jira ticket ID, or both.

---

## Step 0 — Understand the ticket

If a Jira ticket ID was provided (e.g. `HOF-123`), fetch it:
```bash
# via Jira CLI if configured
jira issue view $TICKET_ID
```
Otherwise read the feature description from $ARGUMENTS carefully.
Extract:
- What the feature does (user story)
- What API contract is expected (endpoints, request/response shape)
- Any business rules (validation, edge cases)

---

## Step 1 — Locate where the change lives

Use the layer map from CLAUDE.md to decide which files need touching:

| Layer | When needed |
|---|---|
| `internal/domain/<domain>/entity.go` | New struct field, new type, new constant |
| `internal/domain/<domain>/repository.go` | New DB operation signature |
| `internal/infrastructure/persistence/<domain>_repository.go` | DB query implementation |
| `internal/application/<domain>/service.go` + `dto.go` | New use case or request/response type |
| `internal/interfaces/http/handler/<domain>.go` | New HTTP handler |
| `internal/interfaces/http/router.go` | New route |
| `migrations/` | New table or column |

Read the relevant existing files before writing anything.

---

## Step 2 — Create the migration (if schema changes)

```bash
# Find the next migration number
ls migrations/ | sort | tail -1

# Create the file
# Format: NNN_description_of_change.sql
```

Template:
```sql
-- Write your migrate up statements here
ALTER TABLE <table>
    ADD COLUMN IF NOT EXISTS <col> <type> [DEFAULT <val>];

---- create above / drop below ----

ALTER TABLE <table>
    DROP COLUMN IF EXISTS <col>;
```

Rules:
- Always use `IF NOT EXISTS` / `IF EXISTS`
- Never modify an existing applied migration — always new file
- Index naming: `idx_<table>_<column>`

---

## Step 3 — Domain layer

Add the entity field or new struct to `entity.go`.
Add the repository method signature to `repository.go` (interface only — no implementation).

---

## Step 4 — Application layer

In `dto.go`, add the request/response types:
```go
type CreateXxxRequest struct {
    Name string `json:"name" validate:"required"`
}
```

In `service.go`, add the method to the `Service` interface and implement it on `xyzService`.

**Error rules (critical):**
- Validation errors → `shared.ErrInvalidInput{Message: err.Error()}`
- Not found → `shared.ErrNotFound{Resource: "thing", ID: id.String()}`
- Already exists → `shared.ErrAlreadyExists{Resource: "thing", Field: "email", Value: email}`
- Provider/network errors → `fmt.Errorf("context: %w", err)` (becomes 500)

---

## Step 5 — Infrastructure layer

Implement the new repository method in `internal/infrastructure/persistence/<domain>_repository.go`.

Pattern:
```go
func (r *xyzRepository) CreateThing(ctx context.Context, t *domain.Thing) error {
    if result := r.db.WithContext(ctx).Create(t); result.Error != nil {
        return fmt.Errorf("creating thing: %w", result.Error)
    }
    return nil
}
```

Always:
- Use `WithContext(ctx)` on every GORM call
- Wrap GORM errors with `fmt.Errorf`
- Map `gorm.ErrRecordNotFound` → `shared.ErrNotFound`

---

## Step 6 — HTTP layer

Add the handler to `internal/interfaces/http/handler/<domain>.go`:
```go
func (h *XyzHandler) CreateThing(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.UserIDFromContext(r.Context())
    if !ok {
        response.Unauthorized(w)
        return
    }

    var req appXyz.CreateXxxRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.BadRequest(w, "invalid request body")
        return
    }

    result, err := h.svc.CreateThing(r.Context(), userID, req)
    if err != nil {
        response.Error(w, err)
        return
    }

    response.JSON(w, http.StatusCreated, result)
}
```

Add the route to `internal/interfaces/http/router.go` in the correct group (public vs protected).

---

## Step 7 — Build and verify

```bash
go build ./...          # must be zero errors
go vet ./...            # must be zero warnings
make lint               # fix any lint issues
```

---

## Step 8 — Write tests

Add unit tests for the service logic, and integration tests for the HTTP endpoint.
Run `/test` for the full test workflow.

---

## Step 9 — Deskcheck

Run `/deskcheck` to verify the endpoint works end-to-end against the live server.

---

## Step 10 — Commit and PR

Run `/commit` then `/pr` to ship.
