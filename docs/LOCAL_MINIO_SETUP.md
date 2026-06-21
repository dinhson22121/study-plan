# Local MinIO Setup

This project already supports a **local S3-compatible object store** through
**MinIO**. Use this document when you want local object storage now, and switch
to real AWS S3 later without changing the app code.

---

## What MinIO is used for

- Admin uploads exam PDFs through **presigned PUT URLs**
- The backend verifies uploaded objects
- The Python worker downloads PDFs from object storage to parse them into draft questions

The integration is already wired through:

- `server/pkg/s3`
- `server/internal/content` upload module
- `worker/main.py`
- `docker-compose.yml`

---

## Default local MinIO settings

| Variable | Default |
|---|---|
| `EDU_S3_ENDPOINT` | `http://localhost:9000` |
| `EDU_S3_REGION` | `us-east-1` |
| `EDU_S3_ACCESS_KEY` | `minioadmin` |
| `EDU_S3_SECRET_KEY` | `minioadmin` |
| `EDU_S3_BUCKET` | `edu-assets` |
| `EDU_S3_USE_PATH_STYLE` | `true` |

These values are exported by the root `Makefile` for local runs and match:

- `server/config/config.yaml`
- `docker-compose.yml`

---

## Quick start

### 1. Start only MinIO

```bash
make s3-up
```

If you are on Windows and do not have `make`:

```powershell
powershell -File deploy/scripts/minio-local.ps1 start
```

This starts:

- `minio`
- `minio-init`

`minio-init` creates the bucket automatically if it does not exist.

### 2. Inspect console / credentials

```bash
make s3-console
```

Windows:

```powershell
powershell -File deploy/scripts/minio-local.ps1 console
```

Console:

- API: `http://localhost:9000`
- Web console: `http://localhost:9001`

Default credentials:

- access key: `minioadmin`
- secret key: `minioadmin`

### 3. Smoke test MinIO

```bash
make s3-smoke
```

Windows:

```powershell
powershell -File deploy/scripts/minio-local.ps1 smoke
```

This verifies:

1. MinIO health endpoint is alive
2. The bucket exists and can be listed

---

## Run the full local stack

```bash
make up
make migrate-up
make run
```

This gives you:

- Postgres
- Redis
- Kafka
- MinIO
- bucket bootstrap
- Go API on `:8080`

If you want the Python worker too, use:

```bash
make deploy
```

---

## Local integration points

### Backend (`server/`)

The backend uses:

- `EDU_S3_ENDPOINT=http://localhost:9000`
- `EDU_S3_USE_PATH_STYLE=true`

So presigned uploads and object verification work against MinIO.

### Worker (`worker/`)

The worker reads the same `EDU_S3_*` env vars and pulls PDFs from MinIO.

### Admin web (`admin/`)

No direct MinIO config is needed in the admin UI. The admin app only calls the
backend upload endpoints and uses the presigned URL returned by the backend.

---

## Data persistence

Local MinIO stores data in the Docker volume:

- `minio_data`

To keep local files:

```bash
docker volume ls
```

To completely reset MinIO data:

```bash
docker compose down -v
```

> Warning: this removes uploaded assets and bucket data.

---

## How to switch to real S3 later

When you have budget for AWS S3, you do **not** need to rewrite the upload flow.
Just change configuration:

- `EDU_S3_ENDPOINT=` (empty)
- `EDU_S3_REGION=<aws-region>`
- `EDU_S3_ACCESS_KEY=<real-key>` or instance role
- `EDU_S3_SECRET_KEY=<real-secret>`
- `EDU_S3_BUCKET=<real-bucket>`
- `EDU_S3_USE_PATH_STYLE=false`

The same backend/worker integration continues to work.

---

## Recommended next step

If you want, the next implementation step should be:

1. run `make s3-up`
2. verify `make s3-smoke`
3. run `make up && make migrate-up && make run`
4. test admin upload → complete → parse flow against local MinIO

For backend integration coverage against MinIO:

```bash
make test-integration
```

This now forwards `EDU_TEST_S3_*` values into Go integration tests, including
the S3-compatible client check in `server/pkg/s3`.
