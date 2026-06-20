-- Progress bounded context: topic mastery, streaks, achievements.
CREATE TABLE IF NOT EXISTS topic_progress (
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    topic_id   UUID NOT NULL REFERENCES topic (id) ON DELETE CASCADE,
    status     VARCHAR(20) NOT NULL CHECK (status IN ('NOT_STARTED', 'IN_PROGRESS', 'COMPLETED')),
    best_score NUMERIC(5, 2) NOT NULL DEFAULT 0,
    attempts   INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, topic_id)
);

CREATE TABLE IF NOT EXISTS streak (
    user_id          UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    current_streak   INT NOT NULL DEFAULT 0,
    longest_streak   INT NOT NULL DEFAULT 0,
    last_active_date TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS achievement (
    user_id     UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type        VARCHAR(40) NOT NULL,
    ref         VARCHAR(100) NOT NULL,
    unlocked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, type, ref)
);
