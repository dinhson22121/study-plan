.PHONY: help up down logs build run test test-integration cover migrate-up migrate-down migrate-version lint tidy

# Local infra connection strings (override as needed).
export EDU_POSTGRES_URL ?= postgres://eduapp:secret@localhost:5432/eduapp?sslmode=disable
export EDU_REDIS_URL    ?= redis://localhost:6379/0
export EDU_KAFKA_BROKERS ?= localhost:9092
export EDU_JWT_SECRET   ?= dev-only-change-me

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

up: ## Start Postgres, Redis, Kafka
	docker compose up -d

down: ## Stop and remove infra containers
	docker compose down

logs: ## Tail infra logs
	docker compose logs -f

build: ## Compile all packages
	go build ./...

run: ## Run the API server
	go run ./cmd/server

test: ## Run unit tests
	go test ./...

test-integration: migrate-up ## Run integration tests (requires running infra)
	EDU_TEST_POSTGRES_URL=$(EDU_POSTGRES_URL) EDU_TEST_REDIS_URL=redis://localhost:6379/1 \
		go test -tags=integration ./...

cover: ## Run tests with coverage summary (test packages only)
	go test -coverprofile=coverage.out ./config/... ./internal/... && go tool cover -func=coverage.out | tail -1

migrate-up: ## Apply all migrations
	go run ./cmd/migrate up

migrate-down: ## Roll back one migration
	go run ./cmd/migrate down 1

migrate-version: ## Print current schema version
	go run ./cmd/migrate version

lint: ## Run golangci-lint (if installed)
	golangci-lint run ./...

tidy: ## Tidy go.mod/go.sum
	go mod tidy
