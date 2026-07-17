-- name: InsertProcessedLearnEvent :execrows
INSERT INTO core.processed_learn_events (event_type, user_id, source_id, xp_applied)
VALUES ($1, $2, $3, $4)
ON CONFLICT (event_type, user_id, source_id) DO NOTHING;
