-- +goose Up
ALTER TABLE core.users DROP COLUMN IF EXISTS st_sign_up_at;
ALTER TABLE core.users DROP COLUMN IF EXISTS st_third_party_provider;
ALTER TABLE core.users DROP COLUMN IF EXISTS st_third_party_user_id;
