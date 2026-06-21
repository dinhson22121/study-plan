.PHONY: help up down logs s3-up s3-down s3-logs s3-smoke s3-console build run test test-integration cover migrate-up migrate-down migrate-version lint tidy \
        deploy deploy-down deploy-logs docker-build admin-install admin-dev admin-build worker-install

# Local infra connection strings for running the Go backend from the host.
export EDU_POSTGRES_URL ?= postgres://eduapp:secret@localhost:5432/eduapp?sslmode=disable
export EDU_REDIS_URL    ?= redis://localhost:6379/0
export EDU_KAFKA_BROKERS ?= localhost:9092
export EDU_JWT_SECRET   ?= dev-only-change-me
export EDU_S3_ENDPOINT ?= http://localhost:9000
export EDU_S3_REGION ?= us-east-1
export EDU_S3_ACCESS_KEY ?= minioadmin
export EDU_S3_SECRET_KEY ?= minioadmin
export EDU_S3_BUCKET ?= edu-assets
export EDU_S3_USE_PATH_STYLE ?= true

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

## ---- Infra (Postgres, Redis, Kafka, MinIO) for local dev ----

up: ## Start infra only for local `make run`
	docker compose up -d postgres redis kafka minio minio-init

down: ## Stop and remove all containers
	docker compose down

logs: ## Tail infra logs
	docker compose logs -f postgres redis kafka minio

s3-up: ## Start only local MinIO + bucket bootstrap
	docker compose up -d minio minio-init

s3-down: ## Stop only local MinIO services
	docker compose stop minio minio-init

s3-logs: ## Tail MinIO logs
	docker compose logs -f minio minio-init

s3-smoke: ## Verify local MinIO health and bucket availability
	curl -fsS $(EDU_S3_ENDPOINT)/minio/health/live
	docker compose run --rm minio-init /bin/sh -c "mc alias set local http://minio:9000 $(EDU_S3_ACCESS_KEY) $(EDU_S3_SECRET_KEY) >/dev/null && mc ls local/$(EDU_S3_BUCKET)"

s3-console: ## Print local MinIO console URL and default credentials
	@echo "MinIO API:     $(EDU_S3_ENDPOINT)"
	@echo "MinIO console: http://localhost:9001"
	@echo "Access key:    $(EDU_S3_ACCESS_KEY)"
	@echo "Secret key:    $(EDU_S3_SECRET_KEY)"

## ---- Backend (server/) ----

build: ## Compile the Go backend
	cd server && go build ./...

run: ## Run the API server
	cd server && go run ./cmd/server

test: ## Run Go unit tests
	cd server && go test ./...

test-integration: migrate-up ## Run Go integration tests (requires infra)
	cd server && EDU_TEST_POSTGRES_URL=$(EDU_POSTGRES_URL) EDU_TEST_REDIS_URL=redis://localhost:6379/1 \
		EDU_TEST_S3_ENDPOINT=$(EDU_S3_ENDPOINT) EDU_TEST_S3_REGION=$(EDU_S3_REGION) EDU_TEST_S3_ACCESS_KEY=$(EDU_S3_ACCESS_KEY) \
		EDU_TEST_S3_SECRET_KEY=$(EDU_S3_SECRET_KEY) EDU_TEST_S3_BUCKET=$(EDU_S3_BUCKET) EDU_TEST_S3_USE_PATH_STYLE=$(EDU_S3_USE_PATH_STYLE) \
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
