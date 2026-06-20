-- Notification bounded context (PRD section 5).

CREATE TABLE IF NOT EXISTS device_token (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    device_token VARCHAR(512) NOT NULL UNIQUE,
    platform     VARCHAR(10) NOT NULL CHECK (platform IN ('android', 'ios')),
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_device_token_user ON device_token (user_id);
CREATE INDEX IF NOT EXISTS idx_device_token_active ON device_token (user_id, is_active);

CREATE TABLE IF NOT EXISTS notification_template (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code              VARCHAR(100) NOT NULL UNIQUE,
    title             VARCHAR(255) NOT NULL,
    body              TEXT NOT NULL,
    variables         JSONB NOT NULL DEFAULT '[]'::jsonb,
    notification_type VARCHAR(50) NOT NULL,
    is_active         BOOLEAN NOT NULL DEFAULT true,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notification_log (
    id                UUID PRIMARY KEY,
    user_id           UUID NOT NULL,
    template_code     VARCHAR(100) NOT NULL,
    notification_type VARCHAR(50) NOT NULL,
    correlation_id    UUID NOT NULL,
    status            VARCHAR(20) NOT NULL
                      CHECK (status IN ('PENDING', 'SENT', 'FAILED', 'RETRYING', 'SKIPPED')),
    retry_count       INT NOT NULL DEFAULT 0,
    sent_at           TIMESTAMPTZ,
    error_message     TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_notification_log_user ON notification_log (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notification_log_correlation ON notification_log (correlation_id);

CREATE TABLE IF NOT EXISTS notification_preference (
    user_id           UUID NOT NULL,
    notification_type VARCHAR(50) NOT NULL,
    enabled           BOOLEAN NOT NULL DEFAULT true,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, notification_type)
);

-- Seed default templates (Vietnamese copy, aligned with the app design doc).
INSERT INTO notification_template (code, title, body, variables, notification_type) VALUES
    ('DAILY_REMINDER_V1', 'Đã đến giờ học rồi! 🎯',
     'Hôm nay bạn vẫn còn nhiệm vụ đang chờ. Vào học ngay nhé!', '[]'::jsonb, 'DAILY_REMINDER'),
    ('WEEKLY_QUIZ_V1', 'Bài kiểm tra tuần đang chờ 📝',
     'Bạn chưa hoàn thành Quiz tuần này. Dành ít phút hoàn thành nhé!', '[]'::jsonb, 'WEEKLY_QUIZ'),
    ('STUDY_PLAN_V1', 'Sắp đến hạn lộ trình ⏰',
     'Bạn sắp trễ mốc "{milestone}". Hoàn thành sớm để giữ tiến độ nhé!', '["milestone"]'::jsonb, 'STUDY_PLAN'),
    ('ACHIEVEMENT_V1', 'Chúc mừng! 🎉',
     'Bạn đã hoàn thành {topic}. Tiếp tục phát huy nhé!', '["topic"]'::jsonb, 'ACHIEVEMENT'),
    ('REENGAGEMENT_V1', 'Chúng tôi nhớ bạn! 👋',
     'Đã {days} ngày bạn chưa học. Quay lại chinh phục mục tiêu nào!', '["days"]'::jsonb, 'REENGAGEMENT'),
    ('ADMIN_BROADCAST_V1', '{title}', '{message}', '["title","message"]'::jsonb, 'ADMIN_BROADCAST')
ON CONFLICT (code) DO NOTHING;
