# Task 6: Auth Sync — Lazy Creation Middleware + UserSyncer

## Status: ✅ Completed

## Changes Made

### 1. Created `internal/auth/sync.go`
- **`SyncUserRepo` interface** — minimal repo contract for `GetUser` and `CreateUserFromAuth`
- **`UserSyncer` struct** — orchestrates lazy user creation:
  1. Calls `repo.GetUser(ctx, userID)` to check existence
  2. If not found, fetches email from Supertokens core via `emailpassword.GetUserByID` or `thirdparty.GetUserByID`
  3. Calls `repo.CreateUserFromAuth(ctx, userID, email)` to create user in DB
- **`profileUsecaseAdapter`** — adapts `*biz.ProfileUsecase` to `SyncUserRepo` (delegates `GetProfile` → `GetUser` and `CreateUserFromAuth` directly)
- **`NewUserSyncerFromUsecase`** — convenience constructor taking `*biz.ProfileUsecase`
- Logs via `slog` at debug/info/warn levels for observability

### 2. Modified `internal/auth/middleware.go`
- `Middleware(cfg, syncer)` now accepts an **optional** `*UserSyncer` parameter (nil-safe)
- After session verification succeeds, calls `syncer.EnsureUserExists(ctx, userID)`
- On sync failure: logs a warning but **does not block the request** (user has valid session; transient DB errors should not reject)

### 3. Updated `internal/server/http.go`
- `NewHTTPServer` now accepts `*biz.ProfileUsecase` as 4th parameter
- Creates syncer via `auth.NewUserSyncerFromUsecase(profileUC)` and passes to `auth.Middleware(authCfg, syncer)`

### 4. Updated `cmd/core/wire_gen.go`
- Updated injector to pass `profileUsecase` to `NewHTTPServer`

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| Email from Supertokens API | Uses `emailpassword.GetUserByID` / `thirdparty.GetUserByID` instead of requiring caller to provide email — works for both email/password and social login users |
| Non-blocking on sync failure | The user holds a valid Supertokens session; a transient DB issue shouldn't reject the request |
| Interface-based repo | `SyncUserRepo` keeps auth package decoupled from biz layer; `profileUsecaseAdapter` bridges them |
| Lazy creation on first request | Instead of webhook-based user creation during signup, we create the DB record on first API call after auth |

## Files Changed

| File | Action |
|------|--------|
| `internal/auth/sync.go` | **Created** — UserSyncer + SyncUserRepo interface |
| `internal/auth/middleware.go` | **Modified** — accept syncer param, call EnsureUserExists |
| `internal/server/http.go` | **Modified** — accept ProfileUsecase, create syncer |
| `cmd/core/wire_gen.go` | **Modified** — pass profileUsecase to NewHTTPServer |

## Build
- `go build ./cmd/core/` — **passes** ✅
