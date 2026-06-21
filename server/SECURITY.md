# Security Notes

Summary of the Phase 6 security & code review (run via the `security-reviewer` and
`go-reviewer` agents) — what was fixed, and what is consciously deferred for this
learning project. No CRITICAL issues were found.

## Fixed in Phase 6

| Area | Fix |
|------|-----|
| **Logout auth** | `POST /auth/logout` now requires a valid access token, so a leaked refresh token alone cannot revoke a session. |
| **IDOR / data-access** | `GET /quizzes/:id` enforces ownership in the SQL query (`GetResultForUser` with `WHERE user_id`), not just a post-fetch check. |
| **Resource exhaustion** | `limit` query param clamped on `/questions` (≤100) and `/analytics/me/weak-topics` (≤50). |
| **Atomicity (data integrity)** | Quiz and placement submit now persist the result **and** mark the session complete in a single transaction (`SaveResultAndComplete` / `CompleteWithResult`) — no more "result saved but session still IN_PROGRESS" on crash. |
| **Streak correctness** | Streak day comparison normalized to UTC, avoiding spurious resets near midnight on non-UTC servers. |
| **Plan window** | `WeeksUntilTarget` rounds up so a generated plan covers the full window to the goal date. |
| **Kafka poison messages** | Undecodable messages are dead-lettered to `notification.dlq` instead of being silently dropped. |
| **Graceful shutdown** | FCM retry backoff is context-aware; Kafka consumers use a `WaitGroup` so shutdown waits for in-flight handlers. |
| **Log hygiene** | Inbound `X-Correlation-ID` is length-bounded (≤64) before being logged/echoed. |
| **Broadcast idempotency** | Admin broadcast uses a deterministic per-day key so a repeated same-day broadcast does not double-send. |

## Verified safe (no change needed)

- **SQL injection** — every repository uses pgx parameterized queries; the one dynamic clause (question `LIMIT`) interpolates only the placeholder position, value stays parameterized.
- **Answer-key leakage** — `/questions` hides `is_correct`/`explanation` from students (admin-only); quiz answers are revealed only in the post-submit review.
- **Auth** — bcrypt hashing, refresh-token rotation+revocation (Redis), HS256 pinned via `jwt.WithValidMethods` (no `alg:none`), login returns a uniform error (no account enumeration via message).
- **AuthZ** — every write/admin route is guarded by `RequireRole(ADMIN)`; `/me`-scoped and ownership-checked reads prevent cross-user access (studyplan, placement, quiz, analytics).
- **Secrets** — no hardcoded production secrets; FCM creds and `.env` are gitignored; config validates required secrets at startup.

## Closed in Phase 1 go-live hardening (Workstream B)

| Deferred item | How it was closed |
|---------------|-------------------|
| **Rate limiting** | Redis fixed-window limiter (`internal/shared/ratelimit`) applied to `/auth/login`, `/auth/register`, `/auth/refresh` via `middleware.RateLimit`. Per-IP, configurable (`ratelimit.auth_requests` / `auth_window`, default 10/min). INCR+EXPIRE run as one atomic Lua call (no orphan-TTL race). Returns `429 TOO_MANY_REQUESTS`. The middleware fails **open** on a Redis error — acceptable because the auth operations it guards (`issueTokenPair` → refresh-token `Save`) already hard-depend on Redis, so a Redis outage breaks login/register itself rather than opening a brute-force window. |
| **Dev secrets in committed config** | Production startup validation (`config.validateProduction`) rejects known default and short (<32-char) JWT secrets. Shipped `config/config.example.yaml` and `server/.env.example`; production supplies all secrets via `EDU_*`. |
| **Transport security (Postgres)** | Production validation requires `postgres.url` to use `sslmode=require` (or `verify-ca` / `verify-full`); `sslmode=disable` is rejected when `env=production`. |
| **Password policy** | Minimum length raised to 10, capped at 72 bytes (bcrypt's effective limit — prevents silent truncation), and must contain at least one letter and one digit (`authdomain.ValidatePassword`). |
| **Access-token revocation** | Access tokens now carry a `jti`; `POST /auth/logout` revokes both the refresh token and the presented access token via a Redis blocklist (`auth:revoked:<jti>`, TTL = remaining access lifetime), checked in `ValidateAccessToken`. |

Operational add-ons:
- `GET /health/ready` performs a Postgres + Redis ping for readiness probes; `GET /health` remains the liveness check.
- **Metrics:** `GET /metrics` exposes Prometheus metrics — per-route request count + latency histogram (`edu_http_*`), Go/process collectors, and an `edu_app_failures_total{domain,kind}` business-failure counter. See `internal/shared/metrics`.
- **Error reporting:** Sentry is wired (`internal/shared/observability`) and captures recovered panics with correlation-id/path tags. Disabled (no-op) unless `EDU_SENTRY_DSN` is set.
- **API contract:** frozen for Phase 1 clients in `docs/API.md`.

## Deferred / still required before production

These remain open and are tracked in the go-live plan (Workstream B/C/G):

1. **Kafka transport security** — brokers are PLAINTEXT/unauthenticated locally. Production needs SASL/mTLS (e.g. Amazon MSK with IAM/TLS); not enforceable from the broker list at config-load time, so this is an infra deployment requirement.
2. **Metrics wiring depth** — HTTP, Go and process metrics are live and `RecordFailure` is available, but the per-domain failure counters (parse/upload/notification) and Kafka consumer lag are not yet instrumented at every call site. Kafka lag is best scraped from the broker/exporter rather than the app.
3. **Breached-password checks** — password policy enforces length + composition but does not yet check against a breached-password corpus (e.g. HIBP k-anonymity).

## Known design trade-offs (intentional)

- **Cross-module reads** (analytics→progress/quiz, studyplan→curriculum/placement/goal, placement→curriculum/question) compose sibling services through ports+adapters. The adapters construct a second stateless, DB-backed service instance — acceptable because these services hold no mutable in-memory state.
- **Event handlers are best-effort** on the synchronous in-process eventbus: progress/analytics failures on `quiz.completed` are logged and swallowed so they never fail the user's quiz submission. This favors availability over strong consistency; an outbox/retry would close the gap.
