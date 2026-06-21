# Edu App — API Contract (Phase 1 freeze)

This is the frozen Phase 1 contract for the mobile and admin clients. Breaking
changes after freeze require a new version prefix.

- **Base URL:** `https://<host>/api/v1`
- **Operational endpoints (no prefix, no auth):** `GET /health` (liveness),
  `GET /health/ready` (Postgres+Redis readiness), `GET /metrics` (Prometheus).
- **Content type:** `application/json` for all request and response bodies.

## Response envelope

Every `/api/v1` response uses a uniform envelope:

```jsonc
// success
{ "success": true, "data": { /* ... */ }, "meta": { "total": 0, "page": 1, "limit": 20 } } // meta only on lists
// error
{ "success": false, "error": { "code": "VALIDATION_ERROR", "message": "..." } }
```

## Authentication

- Bearer JWT: `Authorization: Bearer <access_token>`.
- Access token TTL 15m; refresh token TTL 30d (rotated on every refresh).
- `POST /auth/logout` revokes the refresh token **and** the presented access
  token (Redis blocklist), so the access token stops working immediately.
- Roles: `STUDENT` (default on register) and `ADMIN`. Admin-only routes return
  `403 FORBIDDEN` for students.

## Error codes → HTTP status

| Code | HTTP | Meaning |
|------|------|---------|
| `VALIDATION_ERROR` | 422 | Request body/params failed validation |
| `UNAUTHORIZED` | 401 | Missing/invalid/expired/revoked token |
| `FORBIDDEN` | 403 | Authenticated but not permitted (role/preference) |
| `NOT_FOUND` | 404 | Resource does not exist |
| `CONFLICT` | 409 | Duplicate or state conflict |
| `TOO_MANY_REQUESTS` | 429 | Rate limit exceeded |
| `FCM_TOKEN_INVALID` | 410 | Push token revoked/expired |
| `INTERNAL` | 500 | Unexpected server error |

## Conventions

- **Password policy:** 10–72 chars, must contain a letter and a digit.
- **Rate limiting:** `/auth/login`, `/auth/register`, `/auth/refresh` are
  throttled per client IP (default 10 requests/minute) → `429`.
- **Pagination:** list endpoints accept `?limit=` (and where applicable `?page=`);
  `limit` is clamped server-side (`/questions` ≤100, `/analytics/me/weak-topics` ≤50).

---

## Endpoints

Legend: 🔓 public · 🔒 authenticated · 🛡 admin-only

### Auth
| Method | Path | Access | Notes |
|--------|------|--------|-------|
| POST | `/auth/register` | 🔓 | Returns token pair; creates a STUDENT |
| POST | `/auth/login` | 🔓 | Returns token pair |
| POST | `/auth/refresh` | 🔓 | Body `{refresh_token}`; rotates the pair |
| POST | `/auth/logout` | 🔒 | Body `{refresh_token}`; revokes both tokens |

### Users
| Method | Path | Access |
|--------|------|--------|
| GET | `/users/me` | 🔒 |
| PUT | `/users/me` | 🔒 |

### Curriculum
| Method | Path | Access |
|--------|------|--------|
| GET | `/curriculum/subjects` | 🔒 |
| POST | `/curriculum/subjects` | 🛡 |
| GET | `/curriculum/subjects/:id/chapters` | 🔒 |
| POST | `/curriculum/subjects/:id/chapters` | 🛡 |
| GET | `/curriculum/chapters/:id/topics` | 🔒 |
| POST | `/curriculum/chapters/:id/topics` | 🛡 |
| GET | `/curriculum/topics/:id` | 🔒 |

### Content (lessons)
| Method | Path | Access |
|--------|------|--------|
| GET | `/topics/:id/lessons` | 🔒 |
| POST | `/topics/:id/lessons` | 🛡 |
| GET | `/lessons/:id` | 🔒 |

### Questions
| Method | Path | Access | Notes |
|--------|------|--------|-------|
| GET | `/questions` | 🔒 | `is_correct`/`explanation` hidden from students |
| GET | `/questions/:id` | 🔒 | |
| POST | `/questions` | 🛡 | |

### Goals
| Method | Path | Access |
|--------|------|--------|
| GET | `/goals` | 🔒 |
| PUT | `/goals` | 🔒 |

### Placement
| Method | Path | Access |
|--------|------|--------|
| POST | `/placement/tests` | 🔒 |
| POST | `/placement/tests/:id/submit` | 🔒 |
| GET | `/placement/results` | 🔒 |

### Study plans
| Method | Path | Access |
|--------|------|--------|
| POST | `/studyplans/generate` | 🔒 |
| GET | `/studyplans` | 🔒 |
| GET | `/studyplans/:id` | 🔒 |

### Quizzes
| Method | Path | Access | Notes |
|--------|------|--------|-------|
| POST | `/quizzes` | 🔒 | Start a quiz |
| POST | `/quizzes/:id/submit` | 🔒 | Submit answers; reveals review |
| GET | `/quizzes/:id` | 🔒 | Owner-scoped |
| GET | `/quizzes` | 🔒 | List own quizzes |

### Progress & Analytics
| Method | Path | Access |
|--------|------|--------|
| GET | `/progress` | 🔒 |
| GET | `/progress/topics` | 🔒 |
| GET | `/analytics/me` | 🔒 |
| GET | `/analytics/me/weak-topics` | 🔒 |

### Notifications & Devices
| Method | Path | Access |
|--------|------|--------|
| POST | `/devices/token` | 🔒 |
| DELETE | `/devices/token` | 🔒 |
| GET | `/notifications/preferences` | 🔒 |
| PUT | `/notifications/preferences/:type` | 🔒 |
| GET | `/notifications/history` | 🔒 |
| POST | `/admin/notifications/broadcast` | 🛡 |

### Admin — Uploads & question drafts
| Method | Path | Access |
|--------|------|--------|
| POST | `/admin/uploads/init` | 🛡 |
| POST | `/admin/uploads/complete` | 🛡 |
| GET | `/admin/uploads` | 🛡 |
| GET | `/admin/uploads/:id` | 🛡 |
| POST | `/admin/uploads/:id/parse` | 🛡 |
| POST | `/admin/uploads/:id/link` | 🛡 |
| GET | `/admin/uploads/:id/parse-jobs` | 🛡 |
| DELETE | `/admin/uploads/:id` | 🛡 |
| GET | `/admin/uploads/:id/draft-questions` | 🛡 |
| POST | `/admin/uploads/:id/publish` | 🛡 |
| PUT | `/admin/question-drafts/:id` | 🛡 |
| PUT | `/admin/question-drafts/:id/options/:optionId` | 🛡 |
| POST | `/admin/question-drafts/:id/publish` | 🛡 |
