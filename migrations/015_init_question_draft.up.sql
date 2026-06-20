-- Parsed question drafts (written by the worker, reviewed/published by admin).
CREATE TABLE IF NOT EXISTS question_draft (
    id                    UUID PRIMARY KEY,
    asset_id              UUID NOT NULL REFERENCES uploaded_asset (id) ON DELETE CASCADE,
    parse_job_id          UUID NOT NULL REFERENCES parse_job (id) ON DELETE CASCADE,
    question_number       INT NOT NULL DEFAULT 0,
    question_type         VARCHAR(20) NOT NULL DEFAULT 'MCQ',
    stem                  TEXT NOT NULL,
    explanation_raw       TEXT,
    answer_key_raw        TEXT,
    parse_confidence      NUMERIC(4, 3) NOT NULL DEFAULT 0,
    status                VARCHAR(20) NOT NULL DEFAULT 'DRAFT'
                          CHECK (status IN ('DRAFT', 'PUBLISHED')),
    reviewed_by           UUID,
    reviewed_at           TIMESTAMPTZ,
    published_question_id UUID,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_question_draft_asset ON question_draft (asset_id, question_number);

CREATE TABLE IF NOT EXISTS question_draft_option (
    id                  UUID PRIMARY KEY,
    question_draft_id   UUID NOT NULL REFERENCES question_draft (id) ON DELETE CASCADE,
    option_label        VARCHAR(8) NOT NULL DEFAULT '',
    option_text         TEXT NOT NULL,
    is_correct_inferred BOOLEAN NOT NULL DEFAULT false,
    order_index         INT NOT NULL DEFAULT 0,
    UNIQUE (question_draft_id, order_index)
);
