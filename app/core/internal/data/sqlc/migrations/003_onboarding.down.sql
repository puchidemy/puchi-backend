-- +goose Down
DROP TABLE IF EXISTS core.user_onboarding;
ALTER TABLE core.users DROP COLUMN IF EXISTS onboarding_completed;
ALTER TABLE core.users DROP COLUMN IF EXISTS age_range;
