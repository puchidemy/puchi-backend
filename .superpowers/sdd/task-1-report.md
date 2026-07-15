# Task 1 Report: Database Migration + sqlc Setup

**Date:** 2026-07-16  
**Service:** app/core (Kratos v3)

---

## Files Created

| File | Path | Status |
|------|------|--------|
| Migration | `internal/data/sqlc/migrations/001_init.up.sql` | Created |
| sqlc config | `internal/data/sqlc/sqlc.yaml` | Created |
| User queries | `internal/data/sqlc/queries/users.sql` | Created |
| Stats queries | `internal/data/sqlc/queries/stats.sql` | Created |

Additionally, an existing `internal/data/sqlc/queries/achievements.sql` was already present and was picked up by sqlc (references `achievements_def` and `user_achievements` tables).

---

## sqlc Generate

**Result:** Success

sqlc v1.31.1 installed via `go install`. Generated 6 Go files in `internal/data/sqlc/gen/`:

| File | Size | Description |
|------|------|-------------|
| `db.go` | 565 B | Database connection interface |
| `models.go` | 2.8 KB | Go structs for `core.users` and `core.user_stats` |
| `querier.go` | 1.1 KB | Querier interface |
| `users.sql.go` | 3.9 KB | User queries implementation |
| `stats.sql.go` | 2.5 KB | Stats queries implementation |
| `achievements.sql.go` | 3.5 KB | Achievement queries (from pre-existing file) |

All generated code compiles cleanly (`go vet` passes).

---

## Dependencies Added

- `github.com/jackc/pgx/v5 v5.10.0` — PostgreSQL driver
- `github.com/jackc/pgx/v5/pgxpool` — Connection pool

---

## Migration Test

**Result:** Already applied (tables exist with correct schema)

Connectivity to PostgreSQL at `192.168.100.201:30433` confirmed. The database already contains the `core` schema with `users` and `user_stats` tables matching the migration exactly.

### Verified Structure

**core.users** (12 columns):
- `id` (TEXT PK), `username` (UNIQUE NOT NULL), `first_name`, `last_name`, `email` (UNIQUE NOT NULL)
- `avatar_key` (nullable), `bio` (default '')
- `created_at`, `updated_at` (TIMESTAMPTZ, default now())
- `st_sign_up_at`, `st_third_party_provider`, `st_third_party_user_id` (nullable)

**core.user_stats** (14 columns):
- `user_id` (TEXT PK, FK to users), `current_xp`, `total_xp`, `level` (default 1)
- `current_streak`, `longest_streak`, `streak_freezes`, `crowns`, `gems`
- `total_lessons`, `total_minutes`, `accuracy` (REAL), `words_learned`, `updated_at`

**Indexes:**
- `idx_users_username` on core.users(username)
- `idx_users_email` on core.users(email)

**Additional tables found** (from subsequent migrations): `achievements_def`, `daily_activities`, `level_thresholds`, `user_achievements`, `xp_history`

---

## Notes

- The migration uses `-- +goose Up` / `-- +goose Down` annotations compatible with the [goose](https://github.com/pressly/goose) migration tool
- sqlc config emits `emit_db_tags`, `emit_interface`, `emit_empty_slots`, and `emit_pointers_for_null_types` for ergonomic Go codegen
- `psql` not available on Windows; Docker Desktop was also unavailable. Used a Go test program with `pgx/v5/pgxpool` to verify database state
- PostgreSQL pgxpool subpackage required explicit `go get` (`pgx/v5` alone doesn't include `pgxpool`)

## Next Steps

The `gen` package is ready to be wired into the `data` layer. The existing `data.go` file needs updating to initialize a `pgxpool.Pool` and expose sqlc queries to the repository layer.
