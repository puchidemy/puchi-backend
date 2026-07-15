# Task 1 Report: Core Schema Migration + sqlc Setup

## Files Created/Modified

### Pre-existing (verified)
- `app/core/internal/data/sqlc/migrations/001_init.up.sql` — Up migration with core schema (users, user_stats, daily_activities, xp_history, achievements_def, user_achievements, level_thresholds + seed data)
- `app/core/internal/data/sqlc/migrations/001_init.down.sql` — Down migration (drop all)

### Created/Updated
| File | Status |
|------|--------|
| `app/core/internal/data/sqlc/sqlc.yaml` | Updated — added `overrides` section (uuid→string, timestamptz→time.Time) |
| `app/core/internal/data/sqlc/queries/users.sql` | Pre-existed, unchanged |
| `app/core/internal/data/sqlc/queries/stats.sql` | Pre-existed, unchanged |
| `app/core/internal/data/sqlc/queries/achievements.sql` | Created — ListAchievementDefs, GetUserAchievement, ListUserAchievements, UpsertUserAchievement |
| `app/core/internal/data/sqlc/queries/stats.sql` | Pre-existed, unchanged |

## sqlc Generate Status

**Result: SUCCESS** (sqlc v1.31.1)

Generated files in `app/core/internal/data/sqlc/gen/`:
| File | Description |
|------|-------------|
| `db.go` | DBTX interface, Queries struct, New(), WithTx() |
| `models.go` | All model types (CoreUser, CoreUserStat, CoreDailyActivity, CoreXpHistory, CoreAchievementsDef, CoreUserAchievement, CoreLevelThreshold) |
| `querier.go` | Querier interface with all 12 methods |
| `users.sql.go` | CreateUser, GetUser, GetUserByEmail, UpdateUser, UsernameExists, DeleteUser |
| `stats.sql.go` | GetUserStats, CreateUserStats, UpdateUserStats |
| `achievements.sql.go` | ListAchievementDefs, GetUserAchievement, ListUserAchievements, UpsertUserAchievement |

## Database Migration Status

**Result: SUCCESS**

Using Docker postgres:18 image to run psql against `postgresql://puchi:puchi-db-prod@192.168.100.201:30433/puchi`.

All 7 tables created in `core` schema:
| Table | Columns |
|-------|---------|
| `core.users` | 13 columns (id, username, email, first_name, last_name, avatar_key, bio, created_at, updated_at, st_sign_up_at, st_third_party_provider, st_third_party_user_id) |
| `core.user_stats` | 14 columns (1:1 with users, xp/streak/lesson stats) |
| `core.daily_activities` | 6 columns (uuidv7 PK, user_id, activity_date, lessons_completed, xp_earned, minutes_spent) |
| `core.xp_history` | 4 columns (uuidv7 PK, user_id, week_start, xp_earned) |
| `core.achievements_def` | 7 columns (id, title, description, icon, color, category, requirement_type, requirement_value) |
| `core.user_achievements` | 5 columns (composite PK: user_id + achievement_id, progress, unlocked, unlocked_at) |
| `core.level_thresholds` | 2 columns (level, xp_required) + 10 seed rows |

## Issues / Notes

### Migration naming convention
File `001_init.up.sql` uses a 3-digit prefix (`001`). Goose convention typically expects exactly 5 digits (`00001`). The current naming may work with goose since it accepts variable-length numeric prefixes, but should be renamed to `00001_init.up.sql` if strict goose compatibility is required.

### psql not available locally
psql was not installed on the Windows development machine. Used Docker (`postgres:18` image) to run the migration. For CI/CD pipelines, either install psql or use a migration tool like goose.

### Migration ran both Up and Down sections
The `001_init.up.sql` file contains both `-- +goose Up` and `-- +goose Down` sections in the same file. When run through raw psql, both sections execute. A temporary up-only file was used (now cleaned up). When using proper goose tooling, this is handled automatically.

### timestamptz nullable columns
For nullable `TIMESTAMPTZ` columns (e.g., `st_sign_up_at`, `unlocked_at`), sqlc generates `pgtype.Timestamptz` instead of `time.Time`. This is expected behavior since `time.Time` cannot represent NULL in Go. The NOT NULL columns (`created_at`, `updated_at`) correctly use `time.Time`.

## Dependencies Added
- `github.com/jackc/pgx/v5 v5.10.0` — PostgreSQL driver
- `github.com/sqlc-dev/sqlc v1.31.1` — sqlc library (added as go.mod dependency for version tracking)
- sqlc CLI installed via `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
