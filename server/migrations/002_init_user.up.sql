-- User bounded context: profile data. id mirrors the auth identity (user_id).
-- Created during registration via the UserRegisteredEvent, so later modules
-- (e.g. notification.device_token) can safely reference users(id).
CREATE TABLE IF NOT EXISTS users (
    id           UUID PRIMARY KEY,
    email        VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
