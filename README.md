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
| 3 | Learning core (curriculum, question bank, content/lessons) | ✅ Done |
| 4 | Assessment & planning (placement test, goal, studyplan generation) | ✅ Done |
| 5 | Activity loop (quiz, progress/streaks/achievements, analytics, re-engagement) | ✅ Done |
| 6 | Hardening (e2e test, security/code review + fixes, observability) | ✅ Done |

All 12 PRD modules implemented. See [SECURITY.md](SECURITY.md) for the review findings and fixes.

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

## Self-hosted deploy (Docker)

The whole stack — Postgres, Redis, Kafka, **the app**, and a one-shot migration
job — runs from `docker-compose.yml` with the multi-stage [`Dockerfile`](Dockerfile):

```bash
# Build the image, start infra, auto-apply migrations, then start the API.
EDU_JWT_SECRET=$(openssl rand -hex 32) make deploy

curl localhost:8080/health
make deploy-logs   # tail app logs
make deploy-down   # stop (add `docker compose down -v` to drop data)
```

`make deploy` runs `docker compose up -d --build`. Startup order is enforced:
`postgres/redis/kafka` become healthy → the `migrate` job applies all migrations →
the `app` starts and provisions Kafka topics. Override secrets/URLs via the
environment or a `.env` file (the compose `x-app-env` block has the defaults).

**Notes**
- Set a real `EDU_JWT_SECRET` — the compose default is a placeholder.
- For real FCM push, drop `config/firebase-service-account.json` beside the
  compose file and uncomment the volume mount under the `app` service; otherwise
  a logging fallback sender is used.
- Kafka advertises two listeners: `kafka:29092` (in-network, used by the app) and
  `localhost:9092` (host tools).

## Testing

```bash
make test              # unit tests (no infra required)
make cover             # unit tests + coverage summary
make test-integration  # adapter integration tests (requires `make up`)
```

Unit tests use in-memory fakes and run without Docker. The full-flow **e2e test**
lives in `internal/bootstrap` (register → goal → studyplan → quiz → progress →
analytics over the real router). Coverage of domain + application (business
logic) runs **68–97%** per package (~83% avg); repository adapters and HTTP
handlers are covered by the integration/e2e tests rather than unit tests, so the
unit-only headline is lower by design. Integration tests are
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
| `GET/POST` | `/api/v1/curriculum/subjects` | JWT / ADMIN | List / create subjects |
| `GET/POST` | `/api/v1/curriculum/subjects/:id/chapters` | JWT / ADMIN | List / create chapters |
| `GET/POST` | `/api/v1/curriculum/chapters/:id/topics` | JWT / ADMIN | List / create topics |
| `GET` | `/api/v1/curriculum/topics/:id` | JWT | Get a topic |
| `POST` | `/api/v1/questions` | ADMIN | Author a question (MCQ/free-text) |
| `GET` | `/api/v1/questions/:id` | JWT | Get a question (answers hidden from students) |
| `GET` | `/api/v1/questions?topic_id=&difficulty=&limit=` | JWT | Query the question bank |
| `GET/POST` | `/api/v1/topics/:id/lessons` | JWT / ADMIN | List / create lessons |
| `GET` | `/api/v1/lessons/:id` | JWT | Get a lesson with content items |
| `PUT/GET` | `/api/v1/goals` | JWT | Set / get learning goal (onboarding) |
| `POST` | `/api/v1/placement/tests` | JWT | Start a placement test for a subject |
| `POST` | `/api/v1/placement/tests/:id/submit` | JWT | Submit & grade → assessed level |
| `GET` | `/api/v1/placement/results` | JWT | List placement results |
| `POST` | `/api/v1/studyplans/generate` | JWT | Generate a study plan for a subject |
| `GET` | `/api/v1/studyplans` | JWT | List own study plans |
| `GET` | `/api/v1/studyplans/:id` | JWT | Get a study plan with milestones |
| `POST` | `/api/v1/quizzes` | JWT | Start a topic quiz |
| `POST` | `/api/v1/quizzes/:id/submit` | JWT | Submit & grade → result with review |
| `GET` | `/api/v1/quizzes/:id` | JWT | Get a quiz result |
| `GET` | `/api/v1/quizzes` | JWT | Quiz history |
| `GET` | `/api/v1/progress` | JWT | Streak + per-topic mastery overview |
| `GET` | `/api/v1/progress/topics` | JWT | Per-topic progress |
| `GET` | `/api/v1/analytics/me` | JWT | Dashboard (completion, quiz avg, streak) |
| `GET` | `/api/v1/analytics/me/weak-topics` | JWT | Lowest-scoring topics |

### Activity loop

`quiz submit → graded result + review → quiz.completed (eventbus)`. Two subscribers react: **progress** updates per-topic mastery (≥80% = mastered) and the daily streak, awarding `TOPIC_COMPLETED / STREAK_7 / STREAK_30 / PERFECT_SCORE` achievements (recorded once) that push an `ACHIEVEMENT` notification; **analytics** appends an activity event. The notification re-engagement cron reads analytics' inactive-user feed (≥3 days idle) and enqueues `REENGAGEMENT` pushes.

### Plan generation flow

`goal (timing) + placement (level) + curriculum (topics) → studyplan.GeneratePlan → weekly milestones → enqueues a STUDY_PLAN reminder via the notification pipeline`. Levels: `BEGINNER` (<40%), `INTERMEDIATE` (40–75%), `ADVANCED` (>75%). Milestones distribute topics sequentially (curriculum order) across the weeks until the goal's target date.

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
