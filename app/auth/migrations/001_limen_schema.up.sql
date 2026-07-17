-- Limen auth schema (fresh install). Run against DB with search_path=auth.
CREATE SCHEMA IF NOT EXISTS auth;
SET search_path TO auth;

DROP TABLE IF EXISTS rate_limits CASCADE;
DROP TABLE IF EXISTS accounts CASCADE;
DROP TABLE IF EXISTS verifications CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
-- Legacy custom auth tables
DROP TABLE IF EXISTS outbox CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS user_permissions CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS totp_secrets CASCADE;
DROP TABLE IF EXISTS magic_links CASCADE;
DROP TABLE IF EXISTS password_resets CASCADE;
DROP TABLE IF EXISTS email_verifications CASCADE;
DROP TABLE IF EXISTS social_connections CASCADE;
DROP TABLE IF EXISTS oauth_states CASCADE;

CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255),
    email_verified_at TIMESTAMPTZ,
    username VARCHAR(30) UNIQUE,
    first_name VARCHAR(255) NOT NULL DEFAULT '',
    last_name VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    token VARCHAR(255) NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    last_access TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);

CREATE TABLE verifications (
    id UUID PRIMARY KEY,
    subject VARCHAR(255) NOT NULL,
    value VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_verifications_subject ON verifications(subject);

CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(255) NOT NULL,
    provider_account_id VARCHAR(255),
    access_token TEXT NOT NULL DEFAULT '',
    refresh_token TEXT,
    access_token_expires_at TIMESTAMPTZ,
    scope VARCHAR(255) NOT NULL DEFAULT '',
    id_token TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_account_id)
);

CREATE INDEX idx_accounts_user_id ON accounts(user_id);

CREATE TABLE rate_limits (
    id UUID PRIMARY KEY,
    key VARCHAR(255) NOT NULL UNIQUE,
    count INT NOT NULL DEFAULT 0,
    last_request_at BIGINT NOT NULL DEFAULT 0
);
