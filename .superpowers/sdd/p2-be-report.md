# P2: Gamification Stats Endpoints — Implementation Report

## Summary

Implemented `GET /v1/profile/stats` endpoint returning gamification stats for the authenticated user. The endpoint is part of the existing `ProfileService` proto, gRPC + HTTP dual-registered.

## Changes

### 1. Proto (`api/profile/v1/profile.proto`)
- Added `Stats` message with 14 fields: `total_lessons`, `completed_lessons`, `total_minutes`, `accuracy`, `words_learned`, `current_xp`, `total_xp`, `level`, `xp_to_next_level`, `streak`, `longest_streak`, `streak_freezes`, `crowns`, `gems`
- Added `rpc GetStats(google.protobuf.Empty) returns (Stats)` with HTTP mapping `GET /v1/profile/stats`

### 2. Data layer (`internal/data/stats_repo.go`)
- `StatsRepo` wraps `gen.Queries` for `core.user_stats` table
- Methods: `GetUserStats`, `UpsertStats` (create), `UpdateStats`
- Follows the same pattern as `UserRepo`
- Registered in `data.ProviderSet` with `wire.Bind(new(biz.StatsRepoInterface), new(*StatsRepo))`

### 3. Business logic (`internal/biz/stats.go`)
- `StatsRepoInterface` — repository contract (dependency inversion)
- `StatsUsecase` with:
  - `GetStats(ctx, userID)` — delegates to repo
  - `GetXPToNextLevel(level)` — formula `ceil(level * 60 * 1.5)`
- Registered in `biz.ProviderSet`

### 4. Service layer (`internal/service/stats.go` + `profile.go`)
- `StatsService` — handles `GetStats` (auth check → usecase → proto conversion)
- `ProfileService` extended with `statsSvc *StatsService` field
- `ProfileService.GetStats` delegates to `StatsService.GetStats`
- `statsToProto` converts `gen.CoreUserStat` → `pb.Stats`
- No separate server registration needed — `ProfileService` already registered in both HTTP and gRPC

### 5. Wire DI
- Generated `wire_gen.go` creates the full chain: `pool → StatsRepo → StatsUsecase → StatsService → ProfileService`

## Files Changed

| File | Action |
|------|--------|
| `api/profile/v1/profile.proto` | Modified — added Stats message + GetStats RPC |
| `api/profile/v1/profile.pb.go` | Regenerated (buf) |
| `api/profile/v1/profile_grpc.pb.go` | Regenerated (buf) |
| `api/profile/v1/profile_http.pb.go` | Regenerated (buf) |
| `internal/data/stats_repo.go` | New |
| `internal/data/data.go` | Modified — ProviderSet |
| `internal/biz/stats.go` | New |
| `internal/biz/biz.go` | Modified — ProviderSet |
| `internal/service/stats.go` | New |
| `internal/service/profile.go` | Modified — stats delegation |
| `internal/service/service.go` | Modified — ProviderSet |
| `cmd/core/wire_gen.go` | Regenerated (wire) |

## Verification

- `buf generate` — exited 0
- `wire` — exited 0, wrote `wire_gen.go`
- `go build ./cmd/core/` — exited 0

## NOT included (future P2 tasks per SDD)

- `PATCH /v1/profile/stats` endpoint (update stats)
- `POST /v1/profile/stats` endpoint (create stats on signup)
- Achievement integration
- Leaderboard endpoints
