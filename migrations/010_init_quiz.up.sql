-- Quiz bounded context: practice quiz sessions, results, and per-question review.
CREATE TABLE IF NOT EXISTS quiz_session (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    topic_id   UUID NOT NULL REFERENCES topic (id) ON DELETE CASCADE,
    status     VARCHAR(20) NOT NULL CHECK (status IN ('IN_PROGRESS', 'COMPLETED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_quiz_session_user ON quiz_session (user_id);

CREATE TABLE IF NOT EXISTS quiz_session_question (
    session_id  UUID NOT NULL REFERENCES quiz_session (id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES question (id) ON DELETE CASCADE,
    order_index INT NOT NULL DEFAULT 0,
    PRIMARY KEY (session_id, question_id)
);

CREATE TABLE IF NOT EXISTS quiz_result (
    session_id    UUID PRIMARY KEY REFERENCES quiz_session (id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    topic_id      UUID NOT NULL REFERENCES topic (id) ON DELETE CASCADE,
    score         NUMERIC(5, 2) NOT NULL,
    correct_count INT NOT NULL,
    total         INT NOT NULL,
    passed        BOOLEAN NOT NULL,
    completed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_quiz_result_user ON quiz_result (user_id, completed_at DESC);
CREATE INDEX IF NOT EXISTS idx_quiz_result_topic ON quiz_result (user_id, topic_id);

CREATE TABLE IF NOT EXISTS quiz_answer (
    session_id         UUID NOT NULL REFERENCES quiz_session (id) ON DELETE CASCADE,
    question_id        UUID NOT NULL,
    selected_option_id UUID,
    is_correct         BOOLEAN NOT NULL,
    PRIMARY KEY (session_id, question_id)
);
