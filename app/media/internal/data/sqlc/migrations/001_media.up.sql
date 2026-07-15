CREATE SCHEMA IF NOT EXISTS media;

CREATE TABLE media.objects (
    id           BIGSERIAL PRIMARY KEY,
    user_id      TEXT NOT NULL,
    object_key   TEXT NOT NULL UNIQUE,
    bucket       TEXT NOT NULL DEFAULT 'puchi-media',
    content_type TEXT NOT NULL,
    category     TEXT NOT NULL,
    size_bytes   BIGINT NOT NULL,
    width        INT,
    height       INT,
    duration_ms  INT,
    status       TEXT NOT NULL DEFAULT 'uploading',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_media_user ON media.objects(user_id);
