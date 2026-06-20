-- Goal bounded context: one learning goal per user, plus per-subject targets.
CREATE TABLE IF NOT EXISTS goal (
    user_id           UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    target_university VARCHAR(255) NOT NULL,
    target_major      VARCHAR(255) NOT NULL DEFAULT '',
    target_date       TIMESTAMPTZ NOT NULL,
    hours_per_day     INT NOT NULL CHECK (hours_per_day > 0),
    days_per_week     INT NOT NULL CHECK (days_per_week BETWEEN 1 AND 7),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS subject_target (
    user_id       UUID NOT NULL REFERENCES goal (user_id) ON DELETE CASCADE,
    subject_id    UUID NOT NULL REFERENCES subject (id) ON DELETE CASCADE,
    current_score NUMERIC(4, 2) NOT NULL CHECK (current_score BETWEEN 0 AND 10),
    target_score  NUMERIC(4, 2) NOT NULL CHECK (target_score BETWEEN 0 AND 10),
    PRIMARY KEY (user_id, subject_id)
);
