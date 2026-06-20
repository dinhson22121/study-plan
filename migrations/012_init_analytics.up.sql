-- Analytics bounded context: append-only activity log for last-active /
-- inactivity detection (re-engagement). Dashboard metrics are computed
-- on-demand from progress + quiz.
CREATE TABLE IF NOT EXISTS activity_event (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_activity_event_user_time ON activity_event (user_id, occurred_at DESC);
