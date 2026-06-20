-- Question bounded context: question bank with answer options.
CREATE TABLE IF NOT EXISTS question (
    id          UUID PRIMARY KEY,
    topic_id    UUID NOT NULL REFERENCES topic (id) ON DELETE CASCADE,
    type        VARCHAR(20) NOT NULL CHECK (type IN ('MCQ', 'FREE_TEXT')),
    stem        TEXT NOT NULL,
    difficulty  VARCHAR(10) NOT NULL CHECK (difficulty IN ('EASY', 'MEDIUM', 'HARD')),
    explanation TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_question_topic ON question (topic_id, difficulty);

CREATE TABLE IF NOT EXISTS answer_option (
    id          UUID PRIMARY KEY,
    question_id UUID NOT NULL REFERENCES question (id) ON DELETE CASCADE,
    text        TEXT NOT NULL,
    is_correct  BOOLEAN NOT NULL DEFAULT false,
    order_index INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_answer_option_question ON answer_option (question_id, order_index);
