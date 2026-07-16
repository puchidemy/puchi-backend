### Task 1: Backend — DB Migration + SQLC Queries

**Files:**
- Create: `app/core/internal/data/sqlc/migrations/003_onboarding.up.sql`
- Create: `app/core/internal/data/sqlc/migrations/003_onboarding.down.sql`
- Modify: `app/core/internal/data/sqlc/queries/users.sql`
- Regenerate: `app/core/internal/data/sqlc/gen/` (sqlc generate)

**Interfaces:**
- Consumes: existing `core.users` table schema
- Produces: `UpdateOnboardingInfo :one` query, `UpsertUserOnboarding :one` query, `GetUserByUsername :one` query

- [ ] **Step 1: Create up migration**

```sql
-- app/core/internal/data/sqlc/migrations/003_onboarding.up.sql

-- +goose Up
ALTER TABLE core.users ADD COLUMN age_range TEXT NOT NULL DEFAULT '';
ALTER TABLE core.users ADD COLUMN onboarding_completed BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE core.user_onboarding (
    user_id    TEXT PRIMARY KEY REFERENCES core.users(id) ON DELETE CASCADE,
    how_heard  TEXT NOT NULL DEFAULT '',
    why_learn  TEXT NOT NULL DEFAULT '',
    level      TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose Down
```

- [ ] **Step 2: Create down migration**

```sql
-- app/core/internal/data/sqlc/migrations/003_onboarding.down.sql

-- +goose Down
DROP TABLE IF EXISTS core.user_onboarding;
ALTER TABLE core.users DROP COLUMN IF EXISTS onboarding_completed;
ALTER TABLE core.users DROP COLUMN IF EXISTS age_range;
```

- [ ] **Step 3: Add queries to users.sql**

```sql
-- app/core/internal/data/sqlc/queries/users.sql — thêm 3 queries mới

-- name: GetUserByUsername :one
SELECT * FROM core.users WHERE username = $1;

-- name: UpdateOnboardingInfo :one
UPDATE core.users 
SET first_name = $2, last_name = $3, age_range = $4, 
    onboarding_completed = true, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpsertUserOnboarding :one
INSERT INTO core.user_onboarding (user_id, how_heard, why_learn, level)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id) 
DO UPDATE SET how_heard = EXCLUDED.how_heard, 
              why_learn = EXCLUDED.why_learn, 
              level = EXCLUDED.level,
              updated_at = now()
RETURNING *;
```

- [ ] **Step 4: Regenerate sqlc models**

Run: `cd app/core && sqlc generate`
Expected: updated `internal/data/sqlc/gen/models.go` with `AgeRange string` and `OnboardingCompleted bool` fields in `CoreUser`, new `CoreUserOnboarding` struct.

- [ ] **Step 5: Commit**

```bash
git add app/core/internal/data/sqlc/migrations/003_onboarding.up.sql
git add app/core/internal/data/sqlc/migrations/003_onboarding.down.sql
git add app/core/internal/data/sqlc/queries/users.sql
git add app/core/internal/data/sqlc/gen/
git commit -m "feat(core): add onboarding fields and user_onboarding table"
```
