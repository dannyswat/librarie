.PHONY: dev dev-backend dev-frontend build build-backend build-frontend \
        migrate migrate-create sqlc-gen test test-backend test-frontend lint clean

# ── Local development ──────────────────────────────────────────────────────────

dev: ## Start backend (hot-reload) and frontend dev servers concurrently
	@make -j2 dev-backend dev-frontend

dev-backend: ## Start backend with air hot-reload
	@which air > /dev/null 2>&1 || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air -c .air.toml

dev-frontend: ## Start Vite dev server
	cd web && npm run dev

# ── Build ──────────────────────────────────────────────────────────────────────

build: build-backend build-frontend ## Build both backend and frontend

build-backend: ## Compile the Go backend binary
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/librarie ./cmd/server

build-frontend: ## Bundle the React frontend
	cd web && npm run build

# ── Database ───────────────────────────────────────────────────────────────────

migrate: ## Apply all pending migrations (requires DATABASE_URL or uses default)
	@which migrate > /dev/null 2>&1 || (echo "Installing golang-migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	migrate -path db/migrations -database "$${LIBRARIE_DATABASE_URL:-postgres://postgres:postgres@localhost:5432/librarie?sslmode=disable}" up

migrate-create: ## Create a new migration pair  (usage: make migrate-create NAME=add_foo)
	@[ -n "$(NAME)" ] || (echo "Usage: make migrate-create NAME=<name>" && exit 1)
	@which migrate > /dev/null 2>&1 || (echo "Installing golang-migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	migrate create -ext sql -dir db/migrations -seq $(NAME)

sqlc-gen: ## Generate Go code from SQL queries
	@which sqlc > /dev/null 2>&1 || go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	$$(go env GOPATH)/bin/sqlc generate

# ── Testing ────────────────────────────────────────────────────────────────────

test: test-backend test-frontend ## Run all tests

test-backend: ## Run Go tests
	go test ./... -race -count=1

test-frontend: ## Run frontend tests
	cd web && npm test

# ── Lint ───────────────────────────────────────────────────────────────────────

lint: ## Lint Go and TypeScript code
	@which golangci-lint > /dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...
	cd web && npm run lint

# ── Docker ─────────────────────────────────────────────────────────────────────

docker-build: ## Build Docker images
	docker compose build

docker-up: ## Start containers (backend + frontend; Postgres runs externally)
	docker compose up -d

docker-down: ## Stop and remove containers
	docker compose down

# ── Misc ───────────────────────────────────────────────────────────────────────

clean: ## Remove build artifacts
	rm -rf bin/ web/dist

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
