# P4: Achievements Endpoint — Implementation Report

## Summary

Added `ListAchievements` RPC to the Core service's ProfileService. The endpoint returns all achievement definitions with the authenticated user's progress, following the existing layered architecture (proto → service → biz → data → sqlc).

## Files Changed

### Proto (generated)
- `api/profile/v1/profile.proto` — Added `rpc ListAchievements` to `ProfileService` and `Achievement`/`AchievementList` messages
- `api/profile/v1/profile.pb.go`, `profile_grpc.pb.go`, `profile_http.pb.go` — Generated via `buf generate`

### New Go files
- `internal/biz/achievement.go` — `AchievementUsecase` with `ListAchievements` method, `AchievementRepoInterface` contract
- `internal/data/achievement_repo.go` — `AchievementRepo` wrapping sqlc-generated queries

### Modified Go files
- `internal/service/profile.go` — Added `achievementUC` field to `ProfileService`, `ListAchievements` handler, and `achievementItemsToProto` converter
- `internal/biz/biz.go` — Added `NewAchievementUsecase` to `ProviderSet`
- `internal/data/data.go` — Added `NewAchievementRepo` + `wire.Bind` to `ProviderSet`
- `cmd/core/wire_gen.go` — Regenerated via `wire` to wire the new dependencies

## Architecture

```
HTTP GET /v1/profile/achievements
  → ProfileService.ListAchievements (service)
    → auth.UserIDFromContext (extract userID)
    → AchievementUsecase.ListAchievements (biz)
      → AchievementRepo.ListAchievementDefs (data/sqlc) — all defs
      → AchievementRepo.ListUserAchievements (data/sqlc) — user progress
      → Merge defs + progress into AchievementItem slice
    → achievementItemsToProto (AchievementList proto)
  ← JSON/gRPC response
```

## Logic Details

1. Fetch all `achievements_def` rows (unordered, sorted by id)
2. Fetch all `user_achievements` rows for the authenticated user
3. Build a map keyed by `achievement_id` for O(1) lookup
4. For each def: merge progress/unlocked/unlocked_at from user records if present
5. Compute `progress_label` as `"progress/requirement_value"` (e.g. `"3/10"`)
6. Return `AchievementList` proto — unlocked_at omitted when null

## Build Verification

- `go build ./cmd/core/` — passed
- `buf generate` — passed
- `wire` injection — regenerated successfully
