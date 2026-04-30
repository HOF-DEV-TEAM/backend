## ──────────────────────────────────────────────────────────────────────────
##  HOF Backend — Makefile
## ──────────────────────────────────────────────────────────────────────────
##  Usage: make <target>
##  Run `make help` to list all targets with descriptions.
## ──────────────────────────────────────────────────────────────────────────

SHELL := bash

# Load .env into the environment for every sub-command (silent if missing).
-include .env
export

BINARY     := server
BUILD_DIR  := bin
CMD        := ./cmd/main.go
IMAGE_NAME := hof-backend
GO         := go

.PHONY: help env run build clean swagger test lint \
        docker-build up down logs ps db-shell setup-hooks

## ── Help ─────────────────────────────────────────────────────────────────────

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

## ── Setup ────────────────────────────────────────────────────────────────────

setup-hooks: ## Install git hooks (run once after cloning)
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "Copy-Item -Path 'scripts/hooks/pre-push' -Destination '.git/hooks/pre-push' -Force; Write-Host 'Git hooks installed.'"
else
	@cp scripts/hooks/pre-push .git/hooks/pre-push
	@chmod +x .git/hooks/pre-push
	@echo "Git hooks installed."
endif

## ── Environment ──────────────────────────────────────────────────────────────

env: ## Create .env from .env.example, or sync missing/empty keys from .env.example into .env
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env from .env.example — fill in your secrets."; \
	else \
		added=0; filled=0; \
		while IFS= read -r line; do \
			key=$$(echo "$$line" | sed -n 's/^\([A-Za-z_][A-Za-z0-9_]*\)=.*/\1/p'); \
			val=$$(echo "$$line" | sed -n 's/^[A-Za-z_][A-Za-z0-9_]*=\(.*\)/\1/p'); \
			if [ -z "$$key" ]; then continue; fi; \
			if ! grep -q "^$$key=" .env; then \
				echo "$$line" >> .env; added=$$((added+1)); \
			elif [ -n "$$val" ] && grep -q "^$$key=$$" .env; then \
				sed -i "s|^$$key=$$|$$line|" .env; filled=$$((filled+1)); \
			fi; \
		done < .env.example; \
		echo "env: $$added key(s) added, $$filled empty value(s) filled from .env.example."; \
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
	$(GO) run github.com/swaggo/swag/cmd/swag@latest init -g $(CMD) -o docs --parseDependency
	@cp docs/swagger.json api-docs/swagger.json
	@echo "Docs regenerated → docs/ and api-docs/"

test: ## Run all tests
	$(GO) test ./... -v

lint: ## Run golangci-lint (installs v2 if missing)
	@command -v golangci-lint > /dev/null 2>&1 || \
		$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	golangci-lint run --timeout=5m

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
