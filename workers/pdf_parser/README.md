# PDF Parse Worker

Single-file Python worker that turns admin-uploaded PDFs into reviewable question
drafts. Part of the edu-app admin upload pipeline.

## What it does
1. Polls `parse_job` for `QUEUED` jobs and claims one with `FOR UPDATE SKIP LOCKED`.
2. Downloads the asset PDF from S3/MinIO.
3. Extracts text (PyMuPDF) and parses MCQ questions (`Câu N` / `Question N` + A–D).
4. Writes `question_draft` + `question_draft_option` rows.
5. Marks the job `PARSED`, or `REVIEW_REQUIRED` when a question is incomplete, or
   `FAILED` on error. Re-running a job replaces its drafts (idempotent).

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

In Docker it runs as the `worker` service in the repo's `docker-compose.yml`.
