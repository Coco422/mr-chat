# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MrChat is an aggregated LLM chat platform for teams/operations (聚合镜像的 LLM 对话站). It routes user conversations through multiple upstream LLM providers with priority-based failover, manages token quotas/billing, and provides an admin panel for operations.

**Current stage:** Design & documentation phase transitioning into implementation. No source code exists yet — only design docs.

**v0.1 scope:** User auth + self-built Chat UI + internal LLM router + token quota billing + minimal admin panel.

## Tech Stack (Decided via ADRs)

| Layer | Choice | ADR |
|---|---|---|
| Frontend | Vue 3 + Vite (SPA) | ADR-0001 |
| Chat UI | Self-built (not iframe/portal) | ADR-0002 |
| Backend framework | Go + Gin | ADR-0003 |
| Database | PostgreSQL | ADR-0004 |
| Migrations | goose (SQL-first, not GORM AutoMigrate) | ADR-0005 |
| ORM | GORM (default); raw SQL for critical paths like quota settlement, `FOR UPDATE` locks, redemption idempotency | ADR-0006 |
| Redis | Cache/rate-limit/ephemeral state only; must degrade gracefully on failure | ADR-0007 |
| Streaming | SSE (not WebSocket for v0.1) |
| Chat protocol | OpenAI-compatible format |

## Architecture

- **Go backend (Gin)** is the single business entry point and gateway. Frontend connects directly to Go's SSE endpoint.
- **Layering:** handler → service → repository. Keep business logic out of Gin handlers.
- **Routing algorithm:** Priority-based failover with per-provider cooldown blacklist. For each request, iterate providers by priority; skip blacklisted ones; on failure, increment count and blacklist after threshold. See `router-design.md`.
- **Billing flow:** Pre-deduct quota → stream → settle by actual usage → refund unused on cancel/failure. `quota_logs` is the source of truth (not `users.quota`).
- **Auth:** JWT access token in response body + httpOnly refresh token cookie. Three roles: `Root > Admin > User`.
- **Data:** All business PKs are UUID strings. Soft-delete for conversations and messages. Timestamps in UTC.

## Key Documentation (Read Order)

1. `docs/Requirements-Baseline-v0.1.md` — single source of truth for requirements (overrides `Plan.md` and `Architecture-Design-OUI-Integration.md` on conflicts)
2. `docs/Page-and-Route-Spec-v0.1.md` — frontend routes and page specs
3. `docs/API-Contract-v0.1.md` — API endpoints, request/response shapes, SSE event format, error codes
4. `docs/Data-Model-and-State-v0.1.md` — table schemas, relationships, state machines, indexes
5. `docs/Development-Task-Breakdown-v0.1.md` — milestones M0–M7, task IDs, dependencies, sprint plan
6. `docs/adr/` — all architectural decision records

## Build & Development Commands (TBD)

Commands will be established during M0 (INF-01 through INF-08). Expected structure:

- **Backend:** `go run`, `go test ./...`, `go build`
- **Frontend:** `npm run dev`, `npm run build`, `npm run lint`
- **Database:** `goose -dir db/migrations postgres "$DSN" up` / `status` / `down`
- **Local deps:** Docker Compose for PostgreSQL + Redis (INF-04)
- **CI:** lint + test + build (INF-05)

## API Conventions

- Base path: `/api/v1`
- Response envelope: `{ "success": bool, "data": {}, "meta": {}, "error": {}, "request_id": "" }`
- Pagination: `page`, `page_size`, `sort_by`, `sort_order`
- Error codes follow `DOMAIN_SPECIFIC_ERROR` pattern (e.g., `AUTH_INVALID_CREDENTIALS`, `CHAT_UPSTREAM_UNAVAILABLE`, `BILLING_INSUFFICIENT_QUOTA`)

## SSE Event Types

Chat streaming uses these event types in `data` field:
- `response.start` — includes `conversation_id` and `assistant_message_id`
- `response.delta` — content chunks
- `reasoning.delta` — reasoning content chunks
- `response.completed` — usage and billing summary
- `error` — error with code and message
- `[DONE]` — stream end marker

Stop generation: client disconnects SSE; server must abort upstream request and enter settlement/refund.

## Critical Design Rules

- **Redis is never the source of truth** for quota, redemptions, messages, or audit. PostgreSQL transactions + explicit locks protect these paths.
- **goose manages all schema changes.** Never use GORM `AutoMigrate` in production.
- **Streaming writes:** Aggregate in memory during SSE streaming; persist final content to DB only after completion.
- **Redeem codes:** Store only hash in DB, never plaintext. Redemption must be atomic + idempotent via PostgreSQL transaction.
- **Upstream credentials** (`auth_config_encrypted`) must be encrypted at rest.

## Project Language

Documentation is written in Chinese (Simplified). Code, API responses, error codes, and commit messages should be in English.
