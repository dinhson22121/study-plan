-- Admin action audit trail: one row per successful mutating request made by an
-- ADMIN actor. Written best-effort by the AuditAdmin middleware (closes
-- SECURITY_CLOSURE item #9).
CREATE TABLE IF NOT EXISTS admin_audit_log (
    id             UUID PRIMARY KEY,
    actor_user_id  TEXT NOT NULL,
    action         TEXT NOT NULL, -- "METHOD path", e.g. "POST /api/v1/questions"
    method         VARCHAR(10),
    path           TEXT,
    status_code    INT,
    correlation_id TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_actor
    ON admin_audit_log (actor_user_id, created_at);
