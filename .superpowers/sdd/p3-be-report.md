# P3: Social Service — Implementation Report

**Date:** 2026-07-16
**Service:** `app/user/` — User Social Service (replaces Todo template)

---

## Summary

Successfully implemented the social service (`app/user/`) for puchi-backend with follows, leaderboard, and user search features. The service builds and the database migration has been applied.

---

## What was created

### 1. Database Migration
- `internal/data/sqlc/migrations/001_social.up.sql` — `user_social` schema with `follows` and `weekly_leaderboard` tables
- `internal/data/sqlc/migrations/001_social.down.sql` — Rollback

### 2. sqlc Config + Queries
- `internal/data/sqlc/sqlc.yaml` — v2 config (references core migrations for JOIN queries)
- `internal/data/sqlc/queries/follows.sql` — Follow, Unfollow, IsFollowing, ListFollowers, ListFollowing, SearchUsers, GetFollowCounts
- `internal/data/sqlc/queries/leaderboard.sql` — GetWeeklyLeaderboard, UpsertWeeklyXP, GetUserWeeklyXP
- Generated code in `internal/data/sqlc/gen/`

### 3. Proto Definition
- `api/social/v1/social.proto` — SocialService with Follow, Unfollow, ListFollowing, ListFollowers, SearchUsers, GetWeeklyLeaderboard
- Generated code in `api/social/v1/`

### 4. Full Service Structure (Kratos v3 layered architecture)

| Layer | Files | Description |
|-------|-------|-------------|
| **conf** | `internal/conf/conf.proto`, `conf.pb.go`, `configs/config.yaml` | Minimal config (server + database), no auth section |
| **data** | `internal/data/data.go` | pgxpool connection |
| | `internal/data/social_repo.go` | SocialRepo wrapping sqlc generated queries |
| **biz** | `internal/biz/biz.go` | Wire ProviderSet |
| | `internal/biz/social.go` | SocialUsecase + SocialRepoInterface |
| **service** | `internal/service/service.go` | Wire ProviderSet |
| | `internal/service/social.go` | SocialService handler (reads X-User-ID header) |
| **server** | `internal/server/server.go` | Wire ProviderSet |
| | `internal/server/http.go` | HTTP server with CORS (no auth middleware) |
| | `internal/server/grpc.go` | gRPC server |
| **cmd** | `cmd/user/main.go` | Entry point |
| | `cmd/user/wire.go` + `wire_gen.go` | Wire DI |

### 5. Build Files
- `Dockerfile` — Single-stage Go 1.26 build
- `go.mod` — Updated with `github.com/jackc/pgx/v5`

---

## Auth Approach

Since the user service has NO auth middleware, all endpoints get the authenticated user ID from the `X-User-ID` HTTP header, which **Envoy Gateway injects after verifying the Supertokens session**. The `userIDFromContext()` helper in the service layer extracts this via `transport.FromServerContext(ctx)`.

---

## Database Changes

Migration `001_social.up.sql` was applied to `puchi@192.168.100.201:30433/puchi`:

- Created schema `user_social`
- Created `user_social.follows` (follower_id, following_id, created_at)
- Created `user_social.weekly_leaderboard` (uuid PK, user_id, week_start, weekly_xp, rank)

---

## Build Verification

- `buf generate` — proto stubs generated ✓
- `sqlc generate` — Go code from queries generated ✓
- `wire gen` — DI wiring generated ✓
- `go build ./cmd/user/` — compiles successfully ✓

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/social/follow` | Follow a user |
| DELETE | `/v1/social/follow/{following_id}` | Unfollow a user |
| GET | `/v1/social/following` | List who you follow |
| GET | `/v1/social/followers` | List your followers |
| GET | `/v1/social/search` | Search users by name |
| GET | `/v1/social/leaderboard` | Get weekly XP leaderboard |

All endpoints require `X-User-ID` header (injected by Envoy Gateway after auth).

---

## Design Decisions

1. **No auth middleware** — User service relies on Envoy Gateway for auth verification. X-User-ID header is injected after session verification.
2. **Direct DB JOINs** — Social service JOINs against `core.users` and `core.user_stats` tables directly (same PostgreSQL, different schemas). This avoids needing a CoreInternalService gRPC client.
3. **sqlc with core schema** — The sqlc config references both `user_social` and `core` migration directories so generated queries can JOIN across schemas.
4. **Follows use ON CONFLICT DO NOTHING** — Idempotent follow operations.
5. **Weekly leaderboard uses standalone `week_start`** — Not tied to `core.xp_history`. XP sync is expected via `UpsertWeeklyXP` call from the lesson/completion service.

---

## Next Steps

- [ ] Add `UpsertWeeklyXP` call from the lesson service when a user completes a lesson
- [ ] Set up Envoy Gateway `SecurityPolicy` with `ext_auth` to inject `X-User-ID` header
- [ ] Add monitoring/observability (metrics, tracing) via Kratos middleware
- [ ] Write integration tests for social endpoints
