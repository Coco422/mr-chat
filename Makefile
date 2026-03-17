APP_NAME ?= mrchat-api
GOOSE ?= go run ./cmd/migrate
POSTGRES_DSN ?= postgres://mrchat:mrchat@127.0.0.1:5432/mrchat?sslmode=disable

.PHONY: compose-up compose-down api-run web-dev test fmt migrate-status migrate-up migrate-down

compose-up:
	docker compose up -d

compose-down:
	docker compose down

api-run:
	go run ./cmd/api

web-dev:
	cd web && pnpm dev

fmt:
	go fmt ./...

test:
	go test ./...

migrate-status:
	POSTGRES_DSN="$(POSTGRES_DSN)" $(GOOSE) status

migrate-up:
	POSTGRES_DSN="$(POSTGRES_DSN)" $(GOOSE) up

migrate-down:
	POSTGRES_DSN="$(POSTGRES_DSN)" $(GOOSE) down
