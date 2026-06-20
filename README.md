# Edu App — Backend

Go/Gin **modular monolith** built with **Domain-Driven Design**, implementing the
spec in [`plan/edu-app-prd-tdd-v3.html`](plan/edu-app-prd-tdd-v3.html). See
[`plan/IMPLEMENTATION_PLAN.md`](plan/IMPLEMENTATION_PLAN.md) for the phased roadmap.

## Stack

Go 1.23 · Gin · PostgreSQL (pgx) · Redis · Kafka (segmentio, KRaft mode) ·
Firebase Cloud Messaging · JWT auth · golang-migrate · zap.

## Status

| Phase | Scope | State |
|-------|-------|-------|
| 0 | Foundation (config, shared errors/eventbus/middleware, pkg adapters, server bootstrap) | ✅ Done |
| 1 | Auth + User (JWT access/refresh, registration event, profile) | ✅ Done |
| 2 | Notification (Kafka pipeline, FCM w/ retry, scheduler, DLQ, preferences) | ✅ Done |
| 3 | Learning core (curriculum, question, content) | ⏳ Next |
| 4–6 | Assessment/planning, activity loop, hardening | ⬜ Planned |

## Prerequisites

- Go 1.23+
- Docker + Docker Compose (for Postgres, Redis, Kafka)
- A Firebase service-account JSON at `config/firebase-service-account.json`
  (only required once Phase 2's notification pipeline runs; gitignored).

## Quick start

```bash
# 1. Start infrastructure (Postgres, Redis, Kafka)
make up

# 2. Apply database migrations
make migrate-up

# 3. Run the API server (listens on :8080)
make run

# Health check
curl localhost:8080/health
```

Configuration is read from `config/config.yaml`, overridable by `EDU_*` env vars
(e.g. `EDU_POSTGRES_URL`, `EDU_JWT_SECRET`, `EDU_KAFKA_BROKERS`). The Makefile
exports sensible local defaults.

## Testing

```bash
make test              # unit tests (no infra required)
make cover             # unit tests + coverage summary
make test-integration  # adapter integration tests (requires `make up`)
```

Unit tests use in-memory fakes and run without Docker. Integration tests are
gated behind `//go:build integration` and skip unless `EDU_TEST_POSTGRES_URL` /
`EDU_TEST_REDIS_URL` are set (the Makefile target sets them).

## API (current)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET`  | `/health` | — | Liveness |
| `POST` | `/api/v1/auth/register` | Public | Create account, returns token pair |
| `POST` | `/api/v1/auth/login` | Public | Login, returns token pair |
| `POST` | `/api/v1/auth/refresh` | Refresh token | Rotate tokens |
| `POST` | `/api/v1/auth/logout` | Refresh token | Revoke refresh token |
| `GET`  | `/api/v1/users/me` | JWT | Get own profile |
| `PUT`  | `/api/v1/users/me` | JWT | Update display name |
| `POST` | `/api/v1/devices/token` | JWT | Register/update FCM device token |
| `DELETE` | `/api/v1/devices/token` | JWT | Remove device token on logout |
| `GET`  | `/api/v1/notifications/preferences` | JWT | List notification preferences |
| `PUT`  | `/api/v1/notifications/preferences/:type` | JWT | Enable/disable a notification type |
| `GET`  | `/api/v1/notifications/history` | JWT | Paginated delivery history |
| `POST` | `/api/v1/admin/notifications/broadcast` | ADMIN | Broadcast to all active users |

### Notification pipeline

`scheduler/trigger → preference gate → Redis idempotency → Kafka notification.schedule → resolve token + render → notification.send → FCM (retry/backoff) → notification.result → log update → notification.dlq (terminal failures)`. Topics are auto-provisioned at startup. Without Firebase credentials the pipeline runs end-to-end using a logging fallback sender.

## Project layout

```
cmd/server      # API entrypoint (graceful shutdown)
cmd/migrate     # golang-migrate runner
config          # Viper config loader
internal/app    # shared dependency container
internal/shared # cross-cutting: domain errors, events, eventbus, middleware, httpx
internal/<module>/{domain,application,infrastructure,interfaces}  # DDD layers
pkg/{postgres,redis,kafka,fcm}  # infra client wrappers
migrations      # SQL migrations (golang-migrate)
```

Modules communicate only via domain **events** (in-process `eventbus`, and Kafka
for the notification pipeline) — never by importing each other's structs.

License: MIT — see [LICENSE](LICENSE).
