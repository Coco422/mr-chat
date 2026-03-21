APP_NAME ?= mrchat-api
GOOSE ?= go run ./cmd/migrate
POSTGRES_DSN ?= postgres://mrchat:mrchat@172.16.99.32:5432/mrchat?sslmode=disable

.PHONY: compose-up compose-down api-run web-dev test fmt swagger migrate-status migrate-up migrate-down

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

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -d cmd/api,internal -o internal/http/swagger --parseInternal

test:
	go test ./...

migrate-status:
	POSTGRES_DSN="$(POSTGRES_DSN)" $(GOOSE) status

migrate-up:
	POSTGRES_DSN="$(POSTGRES_DSN)" $(GOOSE) up

migrate-down:
	POSTGRES_DSN="$(POSTGRES_DSN)" $(GOOSE) down
