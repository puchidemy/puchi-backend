-- +goose Down
DROP TABLE IF EXISTS auth.audit_logs CASCADE;
DROP TABLE IF EXISTS auth.password_reset_tokens CASCADE;
DROP TABLE IF EXISTS auth.magic_links CASCADE;
DROP TABLE IF EXISTS auth.totp_secrets CASCADE;
DROP TABLE IF EXISTS auth.sessions CASCADE;
DROP TABLE IF EXISTS auth.social_connections CASCADE;
DROP TABLE IF EXISTS auth.email_verifications CASCADE;
DROP TABLE IF EXISTS auth.users CASCADE;
DROP SCHEMA IF EXISTS auth;
