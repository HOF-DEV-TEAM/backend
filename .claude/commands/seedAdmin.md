# /seedAdmin — Fold seed into main binary as subcommand

Refactor the standalone `cmd/seed/main.go` into the main server binary so that
the same `bin/server` artifact handles both the HTTP server and the one-off seed task.

---

## What to implement

### Goal

```
./bin/server              → starts the HTTP server (existing behaviour, unchanged)
./bin/server seed ...     → runs seed logic and exits
```

### 1. Restructure `cmd/main.go`

Extract the existing `main()` body into a `runServer()` function.
Add a mode check at the very top of `main()`:

```go
func main() {
    if len(os.Args) > 1 && os.Args[1] == "seed" {
        runSeed(os.Args[2:])
        return
    }
    runServer()
}
```

### 2. Add `runSeed(args []string)` to `cmd/main.go`

Move the logic from `cmd/seed/main.go` into this function.
- Parse flags from `args` (not `os.Args`) using a dedicated `flag.FlagSet`
- Reuse the same config load, DB connect, migration run, and persistence layer already wired in `runServer()`
- Keep the Option C guard: refuse if any `church_admin` already exists
- On success print confirmation and exit 0; on error exit 1

```go
func runSeed(args []string) {
    fs := flag.NewFlagSet("seed", flag.ExitOnError)
    email     := fs.String("email", "", "Admin email (required)")
    firstName := fs.String("first-name", "", "First name (required)")
    lastName  := fs.String("last-name", "", "Last name (required)")
    password  := fs.String("password", "", "Password min 8 chars (required)")
    _ = fs.Parse(args)
    // ... validation, DB setup, create user, assign role
}
```

### 3. Delete `cmd/seed/main.go` and `cmd/seed/` directory

The standalone package is replaced entirely.

### 4. Update `Makefile`

Change the `seed-admin` target to use the server binary instead of `./cmd/seed`:

```makefile
seed-admin: ## Seed the first admin user (fails if any admin already exists)
	@test -n "$(EMAIL)" || (echo "usage: make seed-admin EMAIL=… FIRST=… LAST=… PASS=…" && exit 1)
	@test -n "$(FIRST)" || (echo "usage: make seed-admin EMAIL=… FIRST=… LAST=… PASS=…" && exit 1)
	@test -n "$(LAST)"  || (echo "usage: make seed-admin EMAIL=… FIRST=… LAST=… PASS=…" && exit 1)
	@test -n "$(PASS)"  || (echo "usage: make seed-admin EMAIL=… FIRST=… LAST=… PASS=…" && exit 1)
	$(GO) run ./cmd/main.go seed \
		-email "$(EMAIL)" \
		-first-name "$(FIRST)" \
		-last-name "$(LAST)" \
		-password "$(PASS)"
```

Or, if building first: `$(BUILD_DIR)/$(BINARY) seed ...`

### 5. Update `README.md`

In the "Creating the first admin" section, update the direct Go command example:

```bash
# Before (standalone)
go run ./cmd/seed -email=… -first-name=… -last-name=… -password=…

# After (subcommand)
go run ./cmd/main.go seed -email=… -first-name=… -last-name=… -password=…

# Or with compiled binary
./bin/server seed -email=… -first-name=… -last-name=… -password=…
```

Also update the Docker and platform one-off examples:
```bash
# Heroku
heroku run ./bin/server seed -email=… -first-name=… -last-name=… -password=…

# Docker
docker run --env-file .env hof-backend seed -email=… …

# Railway / Render / SSH
./bin/server seed -email=… …
```

### 6. Update `CLAUDE.md`

In the "Admin bootstrapping" section, update the seed script reference:
- Remove mention of `cmd/seed/main.go`
- Note that the seed subcommand is built into the main binary: `./bin/server seed …`

---

## Verification

```bash
go build ./...                          # must compile cleanly
go run ./cmd/main.go seed --help        # must print flag usage
make seed-admin EMAIL=a@b.com FIRST=A LAST=B PASS=pass1234   # must run (or fail with "admin exists")
make run                                # server must still start normally
```
