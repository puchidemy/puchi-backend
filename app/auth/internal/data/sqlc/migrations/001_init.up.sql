-- +goose Up

-- Auth schema: identity, sessions, TOTP, magic links, password reset, audit logs
-- PostgreSQL 18: uuidv7() natively available for time-ordered PK tables

CREATE SCHEMA IF NOT EXISTS auth;

-- Core identity
CREATE TABLE auth.users (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email            TEXT NOT NULL,
    email_normalized TEXT NOT NULL UNIQUE,
    email_verified   BOOLEAN NOT NULL DEFAULT false,
    password_hash    TEXT,
    display_name     TEXT NOT NULL DEFAULT '',
    locale           TEXT NOT NULL DEFAULT 'vi',
    is_active        BOOLEAN NOT NULL DEFAULT true,
    is_super_admin   BOOLEAN NOT NULL DEFAULT false,
    last_login_at    TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_users_email_normalized ON auth.users(email_normalized);
CREATE INDEX idx_users_created_at ON auth.users(created_at);

-- Email verification
CREATE TABLE auth.email_verifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    token       TEXT NOT NULL UNIQUE,
    email       TEXT NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_email_verifications_token ON auth.email_verifications(token);
CREATE INDEX idx_email_verifications_user ON auth.email_verifications(user_id);

-- Social connections
CREATE TABLE auth.social_connections (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    provider          TEXT NOT NULL,
    provider_user_id  TEXT NOT NULL,
    provider_email    TEXT,
    avatar_url        TEXT,
    linked_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(provider, provider_user_id)
);
CREATE INDEX idx_social_connections_user ON auth.social_connections(user_id);
CREATE INDEX idx_social_connections_provider ON auth.social_connections(provider, provider_user_id);

-- Sessions & refresh tokens (uuidv7 PK — time-ordered, B-tree friendly)
CREATE TABLE auth.sessions (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id             UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    refresh_token_hash  TEXT NOT NULL UNIQUE,
    token_family        UUID NOT NULL,
    child_number        INTEGER NOT NULL DEFAULT 1,
    ip_address          INET,
    user_agent          TEXT,
    device_name         TEXT,
    device_type         TEXT,
    os                  TEXT,
    location            TEXT,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    expires_at          TIMESTAMPTZ NOT NULL,
    last_used_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_sessions_user_active ON auth.sessions(user_id, is_active) WHERE is_active = true;
CREATE INDEX idx_sessions_refresh_hash ON auth.sessions(refresh_token_hash);
CREATE INDEX idx_sessions_token_family ON auth.sessions(token_family);

-- TOTP secrets (encrypted at rest, Phase 2)
CREATE TABLE auth.totp_secrets (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    encrypted_secret TEXT NOT NULL,
    encrypted_codes  TEXT NOT NULL,
    is_enabled       BOOLEAN NOT NULL DEFAULT false,
    verified_at      TIMESTAMPTZ,
    last_used_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id)
);

-- Magic links (uuidv7 PK — time-ordered for efficient lookup)
CREATE TABLE auth.magic_links (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    email       TEXT NOT NULL,
    user_id     UUID REFERENCES auth.users(id) ON DELETE CASCADE,
    token       TEXT NOT NULL UNIQUE,
    redirect_to TEXT,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_magic_links_token ON auth.magic_links(token);
CREATE INDEX idx_magic_links_email_expires ON auth.magic_links(email, expires_at DESC);

-- Password reset tokens (uuidv7 PK — time-ordered for efficient cleanup)
CREATE TABLE auth.password_reset_tokens (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id     UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    token       TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_password_reset_token ON auth.password_reset_tokens(token);
CREATE INDEX idx_password_reset_user ON auth.password_reset_tokens(user_id);

-- Audit logs (uuidv7 PK — time-ordered for efficient time-range queries)
CREATE TABLE auth.audit_logs (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id     UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    action      TEXT NOT NULL,
    resource    TEXT,
    resource_id TEXT,
    ip_address  INET,
    user_agent  TEXT,
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_logs_user ON auth.audit_logs(user_id, created_at DESC);
CREATE INDEX idx_audit_logs_action ON auth.audit_logs(action, created_at DESC);
CREATE INDEX idx_audit_logs_created ON auth.audit_logs(created_at DESC);
