-- +goose Up
-- Seed a single trial course/unit so guests can try a lesson before signing up.
-- The unit id below MUST match `learn.trial_unit_id` in configs/config.yaml.
INSERT INTO learn.courses (id, slug, title) VALUES
  ('00000000-0000-0000-0000-000000000001', 'trial-course', 'Trial Course');

INSERT INTO learn.units (id, course_id, position, title) VALUES
  ('11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000001', 1, 'Trial Unit');

INSERT INTO learn.skills (id, unit_id, position, title) VALUES
  ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 1, 'Greetings');

INSERT INTO learn.lessons (id, skill_id, position, title, xp_reward) VALUES
  ('33333333-3333-3333-3333-333333333331', '22222222-2222-2222-2222-222222222222', 1, 'Say Hello', 10),
  ('33333333-3333-3333-3333-333333333332', '22222222-2222-2222-2222-222222222222', 2, 'Introduce Yourself', 10),
  ('33333333-3333-3333-3333-333333333333', '22222222-2222-2222-2222-222222222222', 3, 'Polite Farewells', 10);

INSERT INTO learn.exercises (id, lesson_id, position, type, prompt, answer) VALUES
  ('44444444-4444-4444-4444-444444444441', '33333333-3333-3333-3333-333333333331', 1, 'select',
    '{"question": "How do you say \"Hello\" in English?", "options": ["Hello", "Goodbye", "Thanks", "Sorry"]}',
    '{"correct": "Hello"}'),
  ('44444444-4444-4444-4444-444444444442', '33333333-3333-3333-3333-333333333332', 1, 'match',
    '{"pairs": [["I", "T\u00f4i"], ["You", "B\u1ea1n"], ["We", "Ch\u00fang ta"]]}',
    '{"pairs": [["I", "T\u00f4i"], ["You", "B\u1ea1n"], ["We", "Ch\u00fang ta"]]}'),
  ('44444444-4444-4444-4444-444444444443', '33333333-3333-3333-3333-333333333333', 1, 'select',
    '{"question": "How do you say \"Goodbye\"?", "options": ["Hello", "Goodbye", "Please", "Thanks"]}',
    '{"correct": "Goodbye"}');

-- +goose Down
DELETE FROM learn.exercises WHERE id IN (
  '44444444-4444-4444-4444-444444444441',
  '44444444-4444-4444-4444-444444444442',
  '44444444-4444-4444-4444-444444444443'
);
DELETE FROM learn.lessons WHERE id IN (
  '33333333-3333-3333-3333-333333333331',
  '33333333-3333-3333-3333-333333333332',
  '33333333-3333-3333-3333-333333333333'
);
DELETE FROM learn.skills WHERE id = '22222222-2222-2222-2222-222222222222';
DELETE FROM learn.units WHERE id = '11111111-1111-1111-1111-111111111111';
DELETE FROM learn.courses WHERE id = '00000000-0000-0000-0000-000000000001';
