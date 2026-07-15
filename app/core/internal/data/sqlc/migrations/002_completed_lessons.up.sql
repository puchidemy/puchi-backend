-- +goose Up
-- Add completed_lessons column — currently both total_lessons and completed_lessons
-- mapped to same column, making completion rate show 100% always.
ALTER TABLE core.user_stats ADD COLUMN completed_lessons INT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE core.user_stats DROP COLUMN IF EXISTS completed_lessons;
