-- Studyplan bounded context: generated plans with weekly milestones.
CREATE TABLE IF NOT EXISTS study_plan (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    subject_id  UUID NOT NULL REFERENCES subject (id) ON DELETE CASCADE,
    level       VARCHAR(20) NOT NULL,
    start_date  TIMESTAMPTZ NOT NULL,
    target_date TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_study_plan_user ON study_plan (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS milestone (
    id          UUID PRIMARY KEY,
    plan_id     UUID NOT NULL REFERENCES study_plan (id) ON DELETE CASCADE,
    week_number INT NOT NULL,
    due_date    TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_milestone_plan ON milestone (plan_id, week_number);

CREATE TABLE IF NOT EXISTS milestone_topic (
    milestone_id UUID NOT NULL REFERENCES milestone (id) ON DELETE CASCADE,
    topic_id     UUID NOT NULL REFERENCES topic (id) ON DELETE CASCADE,
    order_index  INT NOT NULL DEFAULT 0,
    PRIMARY KEY (milestone_id, topic_id)
);
