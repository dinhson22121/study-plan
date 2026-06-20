-- Parse jobs: claimed and processed by the Python PDF parse worker.
CREATE TABLE IF NOT EXISTS parse_job (
    id             UUID PRIMARY KEY,
    asset_id       UUID NOT NULL REFERENCES uploaded_asset (id) ON DELETE CASCADE,
    status         VARCHAR(20) NOT NULL
                   CHECK (status IN ('QUEUED', 'PROCESSING', 'PARSED', 'REVIEW_REQUIRED', 'FAILED')),
    parser_version VARCHAR(50),
    attempt_count  INT NOT NULL DEFAULT 0,
    error_message  TEXT,
    claimed_by     VARCHAR(100),
    claimed_at     TIMESTAMPTZ,
    started_at     TIMESTAMPTZ,
    finished_at    TIMESTAMPTZ,
    raw_text       TEXT, -- extracted PDF text (decision: kept in DB for MVP)
    created_by     UUID,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_parse_job_status ON parse_job (status, created_at);
CREATE INDEX IF NOT EXISTS idx_parse_job_asset ON parse_job (asset_id, created_at DESC);
