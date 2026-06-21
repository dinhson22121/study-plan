-- Curriculum bounded context: Subject -> Chapter -> Topic catalog hierarchy.
CREATE TABLE IF NOT EXISTS subject (
    id          UUID PRIMARY KEY,
    code        VARCHAR(50) NOT NULL UNIQUE,
    name        VARCHAR(255) NOT NULL,
    grade_level INT NOT NULL CHECK (grade_level BETWEEN 1 AND 12)
);

CREATE TABLE IF NOT EXISTS chapter (
    id          UUID PRIMARY KEY,
    subject_id  UUID NOT NULL REFERENCES subject (id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    order_index INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_chapter_subject ON chapter (subject_id, order_index);

CREATE TABLE IF NOT EXISTS topic (
    id          UUID PRIMARY KEY,
    chapter_id  UUID NOT NULL REFERENCES chapter (id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    order_index INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_topic_chapter ON topic (chapter_id, order_index);
