-- +goose Up
-- MVP: 8 city shells + 1 playable Hanoi A1 story (3 scenes, mixed activities).

INSERT INTO learn.cities (id, slug, name, position, map_x, map_y, cover_url, blurb) VALUES
  ('a1000000-0000-4000-8000-000000000001', 'hanoi', 'Hà Nội', 1, 0.42, 0.28, NULL,
   'Capital mornings, street food, and lakeside walks.'),
  ('a1000000-0000-4000-8000-000000000002', 'ha-long', 'Hạ Long', 2, 0.58, 0.22, NULL,
   'Limestone cliffs rising from emerald water.'),
  ('a1000000-0000-4000-8000-000000000003', 'hue', 'Huế', 3, 0.48, 0.48, NULL,
   'Imperial city on the Perfume River.'),
  ('a1000000-0000-4000-8000-000000000004', 'hoi-an', 'Hội An', 4, 0.52, 0.55, NULL,
   'Lanterns, tailor shops, and riverside evenings.'),
  ('a1000000-0000-4000-8000-000000000005', 'da-nang', 'Đà Nẵng', 5, 0.50, 0.52, NULL,
   'Beaches, bridges, and a modern coastal hub.'),
  ('a1000000-0000-4000-8000-000000000006', 'hcmc', 'TP. Hồ Chí Minh', 6, 0.45, 0.78, NULL,
   'Energy, coffee, and endless street corners.'),
  ('a1000000-0000-4000-8000-000000000007', 'can-tho', 'Cần Thơ', 7, 0.40, 0.82, NULL,
   'Floating markets and Mekong delta life.'),
  ('a1000000-0000-4000-8000-000000000008', 'phu-quoc', 'Phú Quốc', 8, 0.22, 0.85, NULL,
   'Island sunsets and pepper farms by the sea.');

INSERT INTO learn.stories (
  id, city_id, slug, title, summary, cover_url, cefr, tags, audio_url,
  vocab_focus, grammar_focus, cultural_fact, est_minutes, position, status
) VALUES (
  'a2000000-0000-4000-8000-000000000001',
  'a1000000-0000-4000-8000-000000000001',
  'pho-morning',
  'Morning Phở',
  'You wake up hungry in Hà Nội and find a steaming bowl of phở on a quiet street corner.',
  NULL,
  'A1',
  ARRAY['food', 'daily', 'travel'],
  NULL,
  ARRAY['phở', 'nóng', 'bát', 'cảm ơn'],
  ARRAY['là', 'muốn'],
  'In Hà Nội, locals often eat phở standing or on low plastic stools for a quick morning meal.',
  8,
  1,
  'published'
);

INSERT INTO learn.scenes (id, story_id, position, title, narration, dialogue_json, illustration_url, audio_url) VALUES
  (
    'a3000000-0000-4000-8000-000000000001',
    'a2000000-0000-4000-8000-000000000001',
    1,
    'At the stall',
    'Steam rises from a big pot. You smell beef and herbs. A small table waits for you with chopsticks and a spoon.',
    NULL,
    NULL,
    NULL
  ),
  (
    'a3000000-0000-4000-8000-000000000002',
    'a2000000-0000-4000-8000-000000000001',
    2,
    'Ordering',
    'The cook smiles. You want one hot bowl. You say you want phở. She nods and fills a white bowl.',
    '{"turns":[{"speaker":"cook","text":"Phở bò?"},{"speaker":"you","text":"Vâng, một bát."}]}'::jsonb,
    NULL,
    NULL
  ),
  (
    'a3000000-0000-4000-8000-000000000003',
    'a2000000-0000-4000-8000-000000000001',
    3,
    'First sip',
    'The broth is hot and clear. You add lime and chili. You say thank you. The morning feels perfect.',
    NULL,
    NULL,
    NULL
  );

INSERT INTO learn.activities (id, scene_id, position, type, prompt, answer) VALUES
  (
    'a4000000-0000-4000-8000-000000000001',
    'a3000000-0000-4000-8000-000000000001',
    1,
    'select',
    '{"question":"What food is steaming in the pot?","options":["Phở","Pizza","Sushi","Bread"]}',
    '{"correct":"Phở"}'
  ),
  (
    'a4000000-0000-4000-8000-000000000002',
    'a3000000-0000-4000-8000-000000000001',
    2,
    'listen',
    '{"prompt":"Type the word you hear for the noodle soup.","hint":"phở"}',
    '{"text":"phở"}'
  ),
  (
    'a4000000-0000-4000-8000-000000000003',
    'a3000000-0000-4000-8000-000000000002',
    1,
    'match',
    '{"pairs":[["phở","noodle soup"],["nóng","hot"],["bát","bowl"]]}',
    '{"pairs":[["phở","noodle soup"],["nóng","hot"],["bát","bowl"]]}'
  ),
  (
    'a4000000-0000-4000-8000-000000000004',
    'a3000000-0000-4000-8000-000000000002',
    2,
    'select',
    '{"question":"How do you ask for one bowl?","options":["Một bát","Hai ly","Ba ổ","Không"]}',
    '{"correct":"Một bát"}'
  ),
  (
    'a4000000-0000-4000-8000-000000000005',
    'a3000000-0000-4000-8000-000000000003',
    1,
    'dictate',
    '{"prompt":"Type \"thank you\" in Vietnamese as used in the scene."}',
    '{"text":"cảm ơn"}'
  ),
  (
    'a4000000-0000-4000-8000-000000000006',
    'a3000000-0000-4000-8000-000000000003',
    2,
    'select',
    '{"question":"How does the broth taste in the story?","options":["Hot and clear","Cold and sweet","Dry and spicy","Salty only"]}',
    '{"correct":"Hot and clear"}'
  );

-- +goose Down
DELETE FROM learn.activity_attempt_answers;
DELETE FROM learn.activity_attempts;
DELETE FROM learn.user_scene_progress;
DELETE FROM learn.user_story_progress;
DELETE FROM learn.activities WHERE id IN (
  'a4000000-0000-4000-8000-000000000001',
  'a4000000-0000-4000-8000-000000000002',
  'a4000000-0000-4000-8000-000000000003',
  'a4000000-0000-4000-8000-000000000004',
  'a4000000-0000-4000-8000-000000000005',
  'a4000000-0000-4000-8000-000000000006'
);
DELETE FROM learn.scenes WHERE id IN (
  'a3000000-0000-4000-8000-000000000001',
  'a3000000-0000-4000-8000-000000000002',
  'a3000000-0000-4000-8000-000000000003'
);
DELETE FROM learn.stories WHERE id = 'a2000000-0000-4000-8000-000000000001';
DELETE FROM learn.cities WHERE id IN (
  'a1000000-0000-4000-8000-000000000001',
  'a1000000-0000-4000-8000-000000000002',
  'a1000000-0000-4000-8000-000000000003',
  'a1000000-0000-4000-8000-000000000004',
  'a1000000-0000-4000-8000-000000000005',
  'a1000000-0000-4000-8000-000000000006',
  'a1000000-0000-4000-8000-000000000007',
  'a1000000-0000-4000-8000-000000000008'
);
