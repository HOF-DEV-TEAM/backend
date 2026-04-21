## ──────────────────────────────────────────────────────────────────────────
##  HOF Backend — Makefile
## ──────────────────────────────────────────────────────────────────────────
##  Usage: make <target>
##  Run `make help` to list all targets with descriptions.
## ──────────────────────────────────────────────────────────────────────────

# Load .env into the environment for every sub-command (silent if missing).
-include .env
export

BINARY     := server
BUILD_DIR  := bin
CMD        := ./cmd/main.go
IMAGE_NAME := hof-backend
GO         := go

.PHONY: help env run build clean swagger test lint \
        docker-build up down logs ps db-shell

## ── Help ─────────────────────────────────────────────────────────────────────

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

## ── Environment ──────────────────────────────────────────────────────────────

env: ## Copy .env.example → .env (skips if .env already exists)
	@if [ -f .env ]; then \
		echo ".env already exists — skipping. Delete it first to reset."; \
	else \
		cp .env.example .env; \
		echo "Created .env from .env.example — fill in your secrets."; \
	fi

## ── Development ──────────────────────────────────────────────────────────────

run: ## Run the application locally (loads .env automatically)
	$(GO) run $(CMD)

build: ## Compile a production binary → bin/server
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY) $(CMD)
	@echo "Binary: $(BUILD_DIR)/$(BINARY)"

clean: ## Remove compiled binaries
	@rm -rf $(BUILD_DIR)
	@echo "Cleaned."

swagger: ## Regenerate Swagger docs from source annotations
	@which swag > /dev/null 2>&1 || $(GO) install github.com/swaggo/swag/cmd/swag@latest
	swag init -g $(CMD) -o docs --parseDependency
	@echo "Docs regenerated → docs/"

test: ## Run all tests
	$(GO) test ./... -v

lint: ## Run golangci-lint (installs if missing)
	@which golangci-lint > /dev/null 2>&1 || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$($(GO) env GOPATH)/bin
	golangci-lint run ./...

## ── Docker ───────────────────────────────────────────────────────────────────

docker-build: ## Build the Docker image
	docker build -t $(IMAGE_NAME) .

up: ## Start all services (app + postgres) with docker-compose
	docker compose up --build -d
	@echo "App → http://localhost:$(or $(PORT),8080)"
	@echo "API Docs → http://localhost:$(or $(PORT),8080)/docs"

down: ## Stop and remove all docker-compose services
	docker compose down

logs: ## Follow docker-compose logs
	docker compose logs -f

ps: ## Show running docker-compose containers
	docker compose ps

db-shell: ## Open a psql shell inside the postgres container
	docker compose exec db psql -U $${DB_USERNAME:-hofuser} -d $${DB_NAME:-hofdb}
