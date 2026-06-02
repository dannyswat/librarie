# Librarie

A school-facing learning management platform. Single-school deployment, Stage 1 MVP.

## Prerequisites

- Go 1.26+
- Node.js 22+
- PostgreSQL (running externally — not in Docker Compose)

## Quick start (local dev)

```bash
# 1. Copy and fill in required secrets
cp .env.example .env
# Edit .env — set LIBRARIE_SESSION_SECRET and LIBRARIE_ENCRYPTION_KEY

# 2. Create the database
createdb librarie   # or use psql

# 3. Run migrations
make migrate

# 4. Install frontend deps
cd web && npm install && cd ..

# 5. Start backend + frontend (hot-reload)
make dev
```

Backend: http://localhost:8080  
Frontend: http://localhost:5173 (proxies `/api` to backend)

## Environment variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `LIBRARIE_SESSION_SECRET` | ✅ | — | Secret for signing session tokens (min 32 chars) |
| `LIBRARIE_ENCRYPTION_KEY` | ✅ | — | Key for encrypting AI provider credentials (32 hex chars) |
| `LIBRARIE_PORT` | | `8080` | Backend listen port |
| `LIBRARIE_DATABASE_URL` | | `postgres://postgres:postgres@localhost:5432/librarie?sslmode=disable` | Postgres connection string |
| `LIBRARIE_ADMIN_USERS` | | — | Comma-separated usernames auto-elevated to admin on first login |
| `LIBRARIE_STORAGE_PATH` | | `./data/uploads` | Base path for local file storage |

## Makefile targets

```
make dev              Start backend (air) + frontend (vite) concurrently
make build            Build backend binary + frontend bundle
make migrate          Apply pending DB migrations
make migrate-create   Scaffold a new migration  (NAME=<name>)
make sqlc-gen         Regenerate sqlc Go code from SQL queries
make test             Run all tests
make lint             Run golangci-lint + eslint
make docker-build     Build Docker images
make docker-up        Start containers
make docker-down      Stop containers
```

## Project structure

```
cmd/server/         Go backend entry point
internal/
  config/           Environment config loader
  auth/             Session, passkey, password, rate limiting  (Phase 2)
  content/          Subjects, topics, contents, pages, blocks  (Phase 4)
  assessment/       Questions, question bank, assessments       (Phase 5)
  ai/               Provider abstraction + adapters             (Phase 6)
  storage/          Local disk driver + S3-compatible interface (Phase 4)
  user/             User management, invitations                (Phase 3)
db/
  migrations/       SQL migration files (golang-migrate)
  queries/          sqlc SQL query files
web/                React + TypeScript frontend (Vite)
Dockerfile          Backend multi-stage image
web/Dockerfile      Frontend multi-stage image (nginx)
docker-compose.yml  Backend + frontend services
```
