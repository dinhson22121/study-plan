-- Content bounded context: lessons and their content items, attached to a topic.
CREATE TABLE IF NOT EXISTS lesson (
    id          UUID PRIMARY KEY,
    topic_id    UUID NOT NULL REFERENCES topic (id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    order_index INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_lesson_topic ON lesson (topic_id, order_index);

CREATE TABLE IF NOT EXISTS content_item (
    id          UUID PRIMARY KEY,
    lesson_id   UUID NOT NULL REFERENCES lesson (id) ON DELETE CASCADE,
    kind        VARCHAR(10) NOT NULL CHECK (kind IN ('PDF', 'SLIDE', 'NOTE', 'VIDEO')),
    url         TEXT NOT NULL DEFAULT '',
    body        TEXT NOT NULL DEFAULT '',
    order_index INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_content_item_lesson ON content_item (lesson_id, order_index);
