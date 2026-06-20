# Edu App — Implementation Plan

> Derived from `plan/edu-app-prd-tdd-v3.html`. Backend: Go 1.23 / Gin, DDD modular monolith.

## Decisions (locked)

| Axis | Decision |
|------|----------|
| **Scope** | All 12 domain modules, fully implemented |
| **Messaging** | Full Kafka from day one (segmentio/kafka-go) — 4 topics + DLQ |
| **Go module path** | `github.com/son-ngo/edu-app` |
| **Testing** | TDD (write tests first, target 80%+ coverage) |
| **FCM** | Real Firebase adapter wired (mock used in tests) |

## Prerequisites you must provide

These are runtime/secret inputs I cannot generate. They do **not** block scaffolding or TDD (the FCM port is mocked in tests), but they are required to run the real notification pipeline end-to-end:

1. **Firebase service-account JSON** — for the real FCM adapter. Provide the file path or drop it at `config/firebase-service-account.json` (gitignored). The plan assumes Firebase Admin SDK (`firebase.google.com/go/v4`) reading `GOOGLE_APPLICATION_CREDENTIALS`.
2. **Local toolchain** — Go 1.23+, Docker + Docker Compose (for Postgres/Redis/Kafka), `make`. Confirm these are installed.
3. **`golang-migrate` CLI** (or we run migrations via the library in `main`).

## Assumptions for under-specified modules

The PRD details only auth/user/notification. Below is the domain scope I will build for the other 9. **Correct anything wrong before Phase 4.**

| Module | Assumed responsibility | Key aggregate(s) |
|--------|------------------------|------------------|
| `curriculum` | Subjects → grades → chapters → topics hierarchy (catalog) | `Curriculum`, `Chapter`, `Topic` |
| `question` | Question bank (MCQ + free-text), difficulty, topic linkage | `Question`, `AnswerOption` |
| `placement` | Placement test that assesses level on signup; produces a level per subject | `PlacementTest`, `PlacementResult` |
| `goal` | Student learning goals (target level/date per subject) | `Goal` |
| `studyplan` | Generated daily/weekly plan from goal + placement; milestones; emits reminder-due events | `StudyPlan`, `Milestone` |
| `quiz` | Quiz sessions (assemble questions, grade, score); weekly quiz | `QuizSession`, `QuizAttempt` |
| `progress` | Tracks per-topic mastery, streaks, completion; emits achievement events | `Progress`, `Streak` |
| `content` | Learning content/lessons attached to topics (text/video refs) | `Lesson`, `ContentItem` |
| `analytics` | Read-model aggregations: activity, weak topics, last-active (feeds re-engagement) | `ActivitySnapshot` (read model) |

Inter-module contracts are **domain events** (the eventbus / Kafka), never direct cross-module struct imports.

---

> **Progress:** Phase 0 ✅ · Phase 1 ✅ · Phase 2 ✅ · Phase 3 next.

## Phase 0 — Foundation & tooling ✅

**Goal:** compiling skeleton, infra up, CI-ready, one trivial test passing.

- `go mod init github.com/son-ngo/edu-app`; pin deps from the PRD `init.sh`.
- Directory tree per PRD §2 (`cmd/`, `config/`, `internal/shared/`, `internal/<module>/{domain,application,infrastructure,interfaces}`, `pkg/`, `migrations/`).
- `pkg/`: `postgres` (pgx pool), `redis` (go-redis), `kafka` (segmentio client wrapper: producer/consumer factory), `fcm` (Firebase Admin init).
- `internal/shared/`: `domain/errors.go` (DomainError + sentinels from PRD §9), `domain/event.go` (DomainEvent interface), `eventbus/eventbus.go` (in-proc bus), `middleware/` (auth JWT, logger via zap, recovery), HTTP error→status mapper (PRD §9 table).
- `config/`: Viper loader (`config.go` + `config.yaml`) — DB/Redis/Kafka URLs, JWT secret, FCM key path, port, timezone.
- `cmd/server/main.go`: Gin bootstrap with `Register(...)` wiring stubs for all modules (PRD §3).
- `docker-compose.yml` (Postgres 16, Redis 7, Kafka — upgrade to KRaft, dropping Zookeeper, since the PRD's Zookeeper image is legacy), `Makefile` (run/test/migrate/lint/compose), `.golangci.yml`.
- Migration tooling wired; `001_init` placeholder.
- **Tests:** eventbus pub/sub unit test, config loader test, error→HTTP mapper test.

**Exit:** `make up && make test` green; server boots and serves `/health`.

## Phase 1 — Auth + User (identity foundation) ✅

Everything depends on identity, so this comes before notification.

- **auth** (PRD §8): `domain` (UserCredential, Token), `application` (Register, Login, Refresh, Logout commands; ValidateToken query), `infrastructure` (bcrypt, jwt v5 access+refresh, refresh-token store in Redis, credential repo), `interfaces/http`. Endpoints: `POST /auth/{register,login,refresh,logout}`.
- **user**: `User` aggregate, `UserRegisteredEvent` (consumed by notification to seed default preferences, and by placement to trigger onboarding). Repo + minimal profile endpoints.
- Migrations `001_init_auth`, `002_init_user`.
- **Tests (TDD):** password hashing, JWT issue/validate/expiry, refresh rotation & revoke, register/login handler integration (httptest), repo integration against test Postgres.

**Exit:** register→login→authenticated request works; `UserRegisteredEvent` observable on the bus.

## Phase 2 — Notification subsystem (the PRD focus) ✅

Full pipeline per PRD §4–§7, §9.

- **domain:** `NotificationLog` (aggregate), `NotificationType`/`Status` VOs, `NotificationTemplate`, `NotificationPreference`, `events.go`; ports `NotificationRepository`, `FCMPort`.
- **migrations:** `device_token`, `notification_template`, `notification_log` (with `retry_count`, `correlation_id`, `notification_type`, `SKIPPED` status), `notification_preference` (PRD §5). Seed default templates per type.
- **infrastructure:** `pg_repository`, `fcm_adapter` (real Firebase + `SendWithRetry` exponential backoff 1/2/4s, token-invalid → deactivate, PRD §7), `kafka_producer`, `kafka_consumer`.
- **Kafka topics** (PRD §6): `notification.schedule`, `notification.send`, `notification.result`, `notification.dlq`. Every message carries `correlationId` + `idempotencyKey`.
- **application:** `SendNotification`, `UpdatePreference` commands; `GetNotificationLog` query; `scheduler.go` (robfig/cron v3, timezone-aware) — daily reminder, weekly quiz (Sun 19:00), re-engagement batch.
- **Flow** (PRD §7): trigger → **preference gate** (disabled ⇒ log `SKIPPED`) → Redis idempotency (TTL 24h) → enqueue → consumer resolves active device token + renders template → `PENDING` log → FCM send → `notification.result` → log updater → 3 fails ⇒ DLQ.
- **interfaces:** http (`POST/DELETE /devices/token`, `GET/PUT /notifications/preferences[/:type]`, `GET /notifications/history`, `POST /admin/notifications/broadcast`) + kafka consumers.
- **Tests (TDD):** template rendering, preference gate (skip path), idempotency dedup, retry/backoff, token-invalid deactivation, DLQ routing after 3 fails, each Kafka consumer with an embedded/dockerized broker, all handlers.

**Exit:** end-to-end — scheduled trigger lands as a push via real FCM (creds present) or mock (tests); DLQ + log statuses correct.

## Phase 3 — Learning core (curriculum → question → content)

Static/catalog domains first; they have no upstream deps beyond identity.

- **curriculum:** hierarchy CRUD + read APIs; `CurriculumPublished` event.
- **question:** question bank CRUD, link to topics, difficulty; query API for quiz assembly.
- **content:** lessons/content items attached to topics; read APIs.
- Migrations + repos + handlers + TDD per module (domain rules, repo integration, handlers).

## Phase 4 — Assessment & planning (placement → goal → studyplan)

- **placement:** placement test assembled from `question`; on submit produce `PlacementResult` + `PlacementCompletedEvent`.
- **goal:** create/track goals per subject; validation (target date future, level bounds).
- **studyplan:** generate plan from goal+placement; milestones; **emit `StudyPlanReminderDue`** consumed by notification scheduler (PRD §7 step 1). Register with Kafka producer.
- TDD: plan generation logic, milestone scheduling, event emission.

## Phase 5 — Activity loop (quiz → progress → analytics)

- **quiz:** session assembly, grading, scoring; weekly quiz; emit `QuizCompletedEvent`.
- **progress:** consume quiz/lesson events → update mastery & streaks; emit `AchievementUnlockedEvent` / `MilestoneReached` (→ achievement notifications).
- **analytics:** read-model aggregations (activity, weak topics, last-active) feeding re-engagement batch + dashboards.
- TDD: grading correctness, streak math, achievement triggers, analytics aggregation.

## Phase 6 — Hardening & integration

- Full end-to-end integration test across modules (register → placement → goal → studyplan → quiz → progress → achievement notification).
- Scaling toggles from PRD §10 (partition count, consumer count config-driven).
- Observability: structured zap logs with `correlationId` propagation, basic metrics.
- Coverage report ≥80%; `golangci-lint` clean; code review pass (per global rules); security review (auth, input validation, secrets).
- README run instructions; finalize `docker-compose` + `Makefile`.

---

## Cross-cutting standards (applied every phase)

- **DDD layering:** domain → application → infrastructure → interfaces. No cross-module struct imports; communicate via events.
- **Immutability** in domain operations; small focused files (<800 lines).
- **TDD:** RED → GREEN → REFACTOR; 80%+ coverage gate.
- **Error handling:** DomainError sentinels → consistent HTTP status (PRD §9).
- **Secrets:** env / file only, never committed; `.gitignore` covers creds.
- **Per-module checklist:** migration → domain (+tests) → application (+tests) → infra (+tests) → interfaces (+tests) → wire `Register()` in `main.go` → code review.

## Suggested commit cadence

One feature branch per phase; conventional commits (`feat:`, `test:`, `chore:`); PR per phase with the test plan from that phase's "Tests" bullet.
