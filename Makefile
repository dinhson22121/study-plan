.PHONY: help up down logs build run test test-integration cover migrate-up migrate-down migrate-version lint tidy \
        deploy deploy-down deploy-logs docker-build admin-install admin-dev admin-build worker-install

# Local infra connection strings for running the Go backend from the host.
export EDU_POSTGRES_URL ?= postgres://eduapp:secret@localhost:5432/eduapp?sslmode=disable
export EDU_REDIS_URL    ?= redis://localhost:6379/0
export EDU_KAFKA_BROKERS ?= localhost:9092
export EDU_JWT_SECRET   ?= dev-only-change-me

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

## ---- Infra (Postgres, Redis, Kafka, MinIO) for local dev ----

up: ## Start infra only for local `make run`
	docker compose up -d postgres redis kafka minio minio-init

down: ## Stop and remove all containers
	docker compose down

logs: ## Tail infra logs
	docker compose logs -f postgres redis kafka minio

## ---- Backend (server/) ----

build: ## Compile the Go backend
	cd server && go build ./...

run: ## Run the API server
	cd server && go run ./cmd/server

test: ## Run Go unit tests
	cd server && go test ./...

test-integration: migrate-up ## Run Go integration tests (requires infra)
	cd server && EDU_TEST_POSTGRES_URL=$(EDU_POSTGRES_URL) EDU_TEST_REDIS_URL=redis://localhost:6379/1 \
		go test -tags=integration ./...

cover: ## Coverage summary (business logic)
	cd server && go test -coverprofile=coverage.out ./config/... ./internal/... && go tool cover -func=coverage.out | tail -1

migrate-up: ## Apply all migrations
	cd server && go run ./cmd/migrate up

migrate-down: ## Roll back one migration
	cd server && go run ./cmd/migrate down 1

migrate-version: ## Print current schema version
	cd server && go run ./cmd/migrate version

lint: ## Run golangci-lint (if installed)
	cd server && golangci-lint run ./...

tidy: ## Tidy go.mod/go.sum
	cd server && go mod tidy

## ---- Self-hosted deploy (server + worker + infra in containers) ----

deploy: ## Build and run the full stack
	docker compose up -d --build

deploy-down: ## Stop the full stack
	docker compose down

deploy-logs: ## Tail the app container logs
	docker compose logs -f app

docker-build: ## Build the backend image only
	docker build -t edu-app:latest ./server

## ---- Admin web (admin/) ----

admin-install: ## Install admin deps
	cd admin && npm install

admin-dev: ## Run the admin dev server
	cd admin && npm run dev

admin-build: ## Build the admin app
	cd admin && npm run build

## ---- PDF parse worker (worker/) ----

worker-install: ## Install worker deps
	cd worker && pip install -r requirements.txt
