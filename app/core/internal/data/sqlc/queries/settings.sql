-- name: GetUserSettings :one
SELECT * FROM core.user_settings WHERE user_id = $1;

-- name: UpsertUserSettings :one
INSERT INTO core.user_settings (
  user_id, sound_effects, animations, motivational_messages, listening_exercises,
  theme, locale, privacy_json, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8, now())
ON CONFLICT (user_id) DO UPDATE SET
  sound_effects = EXCLUDED.sound_effects,
  animations = EXCLUDED.animations,
  motivational_messages = EXCLUDED.motivational_messages,
  listening_exercises = EXCLUDED.listening_exercises,
  theme = EXCLUDED.theme,
  locale = EXCLUDED.locale,
  privacy_json = EXCLUDED.privacy_json,
  updated_at = now()
RETURNING *;

-- name: CreateUserSettingsDefaults :one
INSERT INTO core.user_settings (user_id)
VALUES ($1)
ON CONFLICT (user_id) DO UPDATE SET updated_at = core.user_settings.updated_at
RETURNING *;
