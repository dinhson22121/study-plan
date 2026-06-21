# PDF Parse Worker

Single-file Python worker (`main.py`) that turns admin-uploaded PDFs into
reviewable question drafts. Part of the edu-app admin upload pipeline.

## What it does
1. Recovers any jobs stuck in `PROCESSING` (see [Stuck-job recovery](#stuck-job-recovery)).
2. Polls `parse_job` for `QUEUED` jobs and claims one with `FOR UPDATE SKIP LOCKED`.
3. Downloads the asset PDF from S3/MinIO (bounded to the 20MB upload limit).
4. Validates the file against the [support matrix](#supported-format-matrix).
5. Extracts text (PyMuPDF) and parses MCQ questions (`Câu N` / `Question N` + A–D).
6. Writes `question_draft` + `question_draft_option` rows.
7. Marks the job `PARSED`, `REVIEW_REQUIRED` when a question is incomplete, or
   `FAILED` (with a clear `error_message`) on a permanent error. Re-running a
   job replaces its drafts (idempotent).

MVP scope: text-based PDFs and MCQ only. No OCR, no scanned images, no free-text.

## Run locally
```bash
pip install -r requirements.txt
export EDU_POSTGRES_URL=postgres://eduapp:secret@localhost:5432/eduapp?sslmode=disable
export EDU_S3_ENDPOINT=http://localhost:9000
export EDU_S3_ACCESS_KEY=minioadmin EDU_S3_SECRET_KEY=minioadmin
export EDU_S3_BUCKET=edu-assets EDU_S3_USE_PATH_STYLE=true
python main.py
```

In Docker it runs as the `worker` service in the repo's `docker-compose.yml`
(`restart: unless-stopped`). Stop with `docker stop`; the worker finishes its
in-flight job, then exits cleanly (SIGTERM-aware).

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `EDU_POSTGRES_URL` | _(required)_ | Postgres DSN. |
| `EDU_S3_BUCKET` | _(required)_ | Default bucket (per-asset `bucket_name` wins if set). |
| `EDU_S3_ENDPOINT` | `""` | S3/MinIO endpoint; empty = real AWS. |
| `EDU_S3_REGION` | `us-east-1` | S3 region. |
| `EDU_S3_ACCESS_KEY` / `EDU_S3_SECRET_KEY` | `""` | S3 credentials. |
| `EDU_S3_USE_PATH_STYLE` | `true` | Path-style addressing (required for MinIO). |
| `EDU_PARSER_POLL_INTERVAL` | `5` | Seconds to sleep when the queue is empty. |
| `EDU_PARSER_VERSION` | `pdf-parser-mvp-1` | Stamped onto claimed jobs. |
| `EDU_PARSER_STUCK_TIMEOUT` | `600` | Seconds a job may sit in `PROCESSING` before recovery. |
| `EDU_PARSER_MAX_ATTEMPTS` | `3` | Claim attempts before a transient/stuck job is failed for good. |
| `EDU_PARSER_MAX_RETRIES` | `3` | boto3 S3 client-level retry budget per request. |
| `EDU_PARSER_RETRY_BACKOFF` | `1.5` | Base for exponential DB-reconnect backoff (capped at 30s). |
| `EDU_PARSER_LOG_LEVEL` | `INFO` | Log level (`DEBUG`/`INFO`/`WARNING`/`ERROR`). |
| `EDU_SENTRY_DSN` | `""` | Optional. If set (and `sentry-sdk` installed), errors are reported. No-op when unset. |

## Supported-format matrix

| Input | Outcome |
|---|---|
| Text-based PDF, MCQ layout (`Câu N` / `Question N` + A–D) | `PARSED` (or `REVIEW_REQUIRED`) |
| Text-based PDF, no recognizable questions | `REVIEW_REQUIRED` (0 drafts) |
| Empty file (0 bytes) | `FAILED` — "uploaded file is empty" |
| Non-PDF (wrong magic bytes / content-type, e.g. DOCX/PNG/JPG) | `FAILED` — "file is not a PDF … only text-based PDFs are supported" |
| Corrupt / unopenable PDF | `FAILED` — "could not open PDF (corrupt or unsupported)" |
| Scanned / image-only PDF (no extractable text) | `FAILED` — "little or no extractable text … OCR is not supported" |
| PDF over 20MB (header or body) | `FAILED` — exceeds the 20MB limit |
| Missing asset row / object missing in storage | `FAILED` — "asset row not found" / "object not found in storage" |

All `FAILED` outcomes write an admin-readable message to `parse_job.error_message`.

## Reliability behavior

### Error handling
Errors are classified:
- **Permanent** (corrupt/unsupported/empty PDF, missing object) → job is marked
  `FAILED` immediately with a clear message. One bad file never crash-loops the worker.
- **Transient** (S3 connection error, throttling) → the job is returned to
  `QUEUED` and retried on a later poll, up to `EDU_PARSER_MAX_ATTEMPTS`
  (`attempt_count` is incremented at each claim). After that it is `FAILED`.
- **DB connection loss** mid-job → the exception bubbles to the main loop, which
  reconnects with exponential backoff; the job is left `PROCESSING` and picked up
  by stuck-job recovery.

### Stuck-job recovery
On every loop iteration, jobs left in `PROCESSING` longer than
`EDU_PARSER_STUCK_TIMEOUT` (uses `updated_at` as a heartbeat) are:
- returned to `QUEUED` if `attempt_count < EDU_PARSER_MAX_ATTEMPTS`, or
- marked `FAILED` once attempts are exhausted.

The schema has no `PENDING` status, so recovered jobs go back to `QUEUED`.
No schema changes / migrations are added.

### Graceful shutdown
`SIGTERM`/`SIGINT` set a flag checked between jobs. The in-flight job always
finishes and is committed; then the process exits cleanly. Safe for `docker stop`
and systemd restarts.

### Resilience
- DB connection is re-established automatically (bounded exponential backoff).
- Memory is bounded: `head_object` size check + a capped read enforce the 20MB
  upload limit before extraction.

### Observability
Structured `key=value` logs (level, timestamp, logger) with per-job context:
`job_id`, `asset_id`, `status`, `drafts`, `attempt`, `duration_s`, `error`.
Optional Sentry via `EDU_SENTRY_DSN`.

## Runbook

- **A file keeps failing** — read `parse_job.error_message`; the support matrix
  above maps each message to its cause. Non-PDF / scanned / corrupt inputs are
  expected `FAILED`s, not bugs.
- **Jobs are stuck in `PROCESSING`** — they self-recover after
  `EDU_PARSER_STUCK_TIMEOUT`. To force it, lower the env value and restart, or
  manually `UPDATE parse_job SET status='QUEUED', claimed_by=NULL WHERE id=…`.
- **Worker is crash-looping on startup** — check `EDU_POSTGRES_URL` /
  `EDU_S3_BUCKET` are set (they are required) and the DB/S3 are reachable.
- **Worker can't reach the DB** — it backs off and retries; logs show
  `db connect failed; backing off`. It does not exit, so `restart: unless-stopped`
  is not triggered into a tight loop.
- **Reprocess a job** — set it back to `QUEUED`; drafts are replaced idempotently.
- **Inspect throughput** — grep logs for `msg="job complete"` and `status=` /
  `duration_s=` fields.
