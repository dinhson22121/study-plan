# Edu App

A modular-monolith education platform: a Go/Gin DDD backend, a React admin
console, a Python PDF-parsing worker, and a mobile client prototype.

## Layout

| Folder      | What it is                                                                 |
|-------------|---------------------------------------------------------------------------|
| `server/`   | Go backend (Gin, DDD, 12 domain modules, Postgres, Redis, Kafka, S3/FCM). |
| `admin/`    | React 19 + Vite + TypeScript admin web console.                           |
| `worker/`   | Python PDF → draft-question parse worker.                                  |
| `client/`   | Mobile client design + single-file prototype.                             |
| `plan/`     | Product/architecture docs and enhancement specs.                          |

## Quick start

```bash
# 1. Bring up infra (Postgres, Redis, Kafka, MinIO)
make up

# 2. Apply migrations and run the API on :8080
make migrate-up
make run

# Or run the entire stack (API + worker + infra) in containers:
make deploy
```

Run `make help` to list every target. Component-specific docs live in each
folder's own `README.md`.

## Tech stack

- **Backend** — Go 1.23, Gin, pgx/v5, golang-migrate, segmentio/kafka-go, aws-sdk-go-v2, Firebase FCM.
- **Admin** — React 19, Vite 6, TanStack Query, React Hook Form + Zod, Tailwind 4.
- **Worker** — Python, PyMuPDF, psycopg, boto3.
- **Infra** — Postgres 16, Redis 7, Kafka (KRaft), MinIO, Docker Compose.
