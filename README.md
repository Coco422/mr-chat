# MrChat

MrChat is a Go + Vue based multi-model chat platform currently entering the first implementation phase.

## Current Focus

- `M0` / `M1` engineering bootstrap
- backend skeleton with Gin + GORM + PostgreSQL
- frontend skeleton with Vue 3 + Vite
- local development workflow with PostgreSQL + Redis

## Repository Layout

```text
mr-chat/
├── cmd/api/                    # API entrypoint
├── internal/
│   ├── app/                    # app bootstrap and config
│   ├── http/                   # router and middleware
│   ├── modules/                # business modules
│   ├── platform/               # db, cache, logger
│   └── shared/                 # shared helpers
├── db/migrations/              # goose SQL migrations
├── web/                        # Vue 3 + Vite frontend
├── docs/                       # product and engineering docs
├── docker-compose.yml          # local PostgreSQL / Redis
└── Makefile                    # local developer shortcuts
```

## Quick Start

1. Copy `.env.example` to `.env`
2. Start local services:
   - `docker compose up -d`
3. Start backend:
   - `go run ./cmd/api`
4. Start frontend:
   - `cd web && pnpm install && pnpm dev`

## Notes

- Database schema changes must go through `goose` migrations
- In the current early development stage, the API server auto-runs pending migrations on startup by default
- `go run ./cmd/migrate <status|up|down>` provides a repo-local migration CLI without requiring a separate manual goose install
- GORM is used as the default ORM, but critical paths may fall back to raw SQL
- Redis is an acceleration layer only and must be safe to lose at runtime
