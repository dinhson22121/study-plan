-- Admin upload pipeline: uploaded asset metadata (bytes live in S3/MinIO).
CREATE TABLE IF NOT EXISTS uploaded_asset (
    id                UUID PRIMARY KEY,
    object_key        VARCHAR(1024) NOT NULL,
    bucket_name       VARCHAR(255) NOT NULL,
    original_filename VARCHAR(512) NOT NULL,
    content_type      VARCHAR(255) NOT NULL,
    file_size         BIGINT NOT NULL DEFAULT 0,
    checksum_sha256   VARCHAR(64) NOT NULL DEFAULT '',
    status            VARCHAR(20) NOT NULL
                      CHECK (status IN ('PENDING', 'UPLOADED', 'VERIFIED', 'DELETED', 'FAILED')),
    uploaded_by       UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    entity_type       VARCHAR(20), -- optional 1-1 attach (deferred): QUESTION/EXAM/CONTENT
    entity_id         UUID,
    storage_provider  VARCHAR(20) NOT NULL DEFAULT 'S3',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at       TIMESTAMPTZ,
    deleted_at        TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_uploaded_asset_object_key ON uploaded_asset (object_key);
CREATE INDEX IF NOT EXISTS idx_uploaded_asset_status ON uploaded_asset (status, created_at DESC);
