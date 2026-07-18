# Task 2 Report — Core biz + proto settings APIs

**Status:** DONE  
**Branch:** `feat/guest-settings-sync`  
**Commit:** `feat(core): profile settings GET/PATCH/merge APIs`

## What landed

### Proto (`app/core/api/profile/v1/profile.proto`)
- RPCs: `GetSettings`, `UpdateSettings`, `MergeSettings`
- HTTP: `GET/PATCH /v1/profile/settings`, `POST /v1/profile/settings/merge`
- Messages: `UserSettings`, `UpdateSettingsRequest` (optional fields), `MergeSettingsRequest`, `MergeSettingsResponse`
- Regenerated via `buf generate --template buf.gen.yaml`

### Biz (`app/core/internal/biz/settings.go`)
- Kept `biz.SettingsRepo` interface from Task 1
- `SettingsValues`, `UpdateSettingsInput`, `MergeSettingsResult`
- Product defaults: sound/animations/motivational/listening=`true`, theme=`system`, locale=`en`, privacy=`{}`
- `MergeSettingsValues`: guest wins only when guest ≠ default AND server == default
- `ProfileUsecase` methods: `GetSettings` (EnsureDefaults), `UpdateSettings` (partial), `MergeSettings`
- `NewProfileUsecase` now takes `SettingsRepo`

### Service + wire
- Handlers on `ProfileService` with auth + proto mapping
- `settings` reserved in public-profile path matcher
- `wire_gen.go` injects `data.NewSettingsRepo` into `NewProfileUsecase`

### Tests
- `TestMergeSettings_GuestWinsOnlyVsDefaults` (TDD): server sound=false custom, animations=true default; guest sound=true, animations=false → keep sound=false, merge animations only

## Test evidence

```text
=== RUN   TestMergeSettings_GuestWinsOnlyVsDefaults
--- PASS: TestMergeSettings_GuestWinsOnlyVsDefaults (0.00s)
PASS
ok  	github.com/puchidemy/puchi-backend/app/core/internal/biz
```

Also: `go build ./...` (app/core) OK; `go test ./internal/biz/ ./internal/server/` OK.

## Manual curl

Not run in this session (no live Bearer against local/core). Routes registered in generated HTTP:

- `GET /v1/profile/settings`
- `PATCH /v1/profile/settings`
- `POST /v1/profile/settings/merge`

## Concerns

- Manual Bearer curl deferred (unit coverage only).
- `privacy_json` compare is trimmed string equality against `"{}"` (not deep JSON equality).

---

## Review fixes (Important findings)

**Status:** DONE  
**Commit:** `fix(core): validate settings theme/locale and merge guest payload`

### Changes
- Biz: validate `theme ∈ {system, light, dark}` and `locale` (non-empty, 2–16 chars, supported codes or BCP47-like pattern) on `UpdateSettings` and `MergeSettings`
- Biz: `MergeSettings` rejects nil guest with `ErrGuestSettingsRequired`
- Proto: comment on `MergeSettingsRequest` — clients must send full guest snapshot (bool `false` is meaningful)
- Service: map validation errors to `InvalidArgument`
- Test: `TestUpdateSettings_RejectsInvalidTheme`

### Test evidence (post-fix)

```text
=== RUN   TestMergeSettings_GuestWinsOnlyVsDefaults
--- PASS: TestMergeSettings_GuestWinsOnlyVsDefaults (0.00s)
=== RUN   TestUpdateSettings_RejectsInvalidTheme
--- PASS: TestUpdateSettings_RejectsInvalidTheme (0.00s)
PASS
ok  	github.com/puchidemy/puchi-backend/app/core/internal/biz	0.561s
```
