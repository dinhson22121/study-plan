-- Placement bounded context: per-subject placement tests and their results.
CREATE TABLE IF NOT EXISTS placement_test (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subject (id) ON DELETE CASCADE,
    status     VARCHAR(20) NOT NULL CHECK (status IN ('IN_PROGRESS', 'COMPLETED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_placement_test_user ON placement_test (user_id);

CREATE TABLE IF NOT EXISTS placement_test_question (
    test_id     UUID NOT NULL REFERENCES placement_test (id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES question (id) ON DELETE CASCADE,
    order_index INT NOT NULL DEFAULT 0,
    PRIMARY KEY (test_id, question_id)
);

CREATE TABLE IF NOT EXISTS placement_result (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    subject_id   UUID NOT NULL REFERENCES subject (id) ON DELETE CASCADE,
    score        NUMERIC(5, 2) NOT NULL,
    level        VARCHAR(20) NOT NULL CHECK (level IN ('BEGINNER', 'INTERMEDIATE', 'ADVANCED')),
    completed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_placement_result_user_subject ON placement_result (user_id, subject_id, completed_at DESC);
