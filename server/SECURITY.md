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

## Deferred / accepted for a learning project

These are documented trade-offs, acceptable for local/learning use but **required before any real deployment**:

1. **Rate limiting** — login/register are unthrottled. Add IP-based throttling (e.g. `golang.org/x/time/rate`) before exposing publicly.
2. **Dev secrets in committed config** — `config/config.yaml` and `docker-compose.yml` carry dev-only credentials (JWT secret, Postgres password). Production must supply these via `EDU_*` env vars; remove the committed defaults and ship a `.example`.
3. **Transport security** — local Postgres uses `sslmode=disable` and Kafka is PLAINTEXT/unauthenticated. Production needs `sslmode=require` and SASL/mTLS on Kafka.
4. **Password policy** — minimum length 8 only; consider complexity / breached-password checks.
5. **Access-token revocation** — access tokens are valid until their 15-minute TTL (refresh tokens are revocable). Add a Redis blocklist if immediate revocation is required.

## Known design trade-offs (intentional)

- **Cross-module reads** (analytics→progress/quiz, studyplan→curriculum/placement/goal, placement→curriculum/question) compose sibling services through ports+adapters. The adapters construct a second stateless, DB-backed service instance — acceptable because these services hold no mutable in-memory state.
- **Event handlers are best-effort** on the synchronous in-process eventbus: progress/analytics failures on `quiz.completed` are logged and swallowed so they never fail the user's quiz submission. This favors availability over strong consistency; an outbox/retry would close the gap.
