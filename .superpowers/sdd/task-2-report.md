# Task 2 — Data Layer with pgxpool and UserRepo

## Summary

Implemented PostgreSQL data layer using `pgxpool` connection pool and `UserRepo` wrapping sqlc-generated queries.

## Files Changed

| File | Action |
|------|--------|
| `internal/conf/conf.proto` | Updated — removed `driver` field from `Database` message, kept only `source` |
| `internal/conf/conf.pb.go` | Regenerated via `protoc` to reflect proto change |
| `internal/data/data.go` | Replaced stub — now uses `pgxpool.Pool`, provides `NewUserRepo` in `ProviderSet` |
| `internal/data/user_repo.go` | **Created** — `UserRepo` wrapping `gen.Queries` with `CreateUser`, `GetUser`, `GetUserByEmail`, `UpdateUser`, `UsernameExists` |
| `internal/data/todo.go` | **Deleted** — old in-memory `TodoRepo` stub |
| `internal/server/http.go` | Removed `todo *service.TodoService` parameter and v1 registration |
| `internal/server/grpc.go` | Removed `todo *service.TodoService` parameter and v1 registration |
| `cmd/core/wire.go` | Updated — removed `biz`/`service` imports, only `data` + `server` ProviderSets |
| `cmd/core/wire_gen.go` | Updated — removed all `todo`/`biz`/`service` wiring |

## Data Layer Architecture

```
NewData(cfg.Data) → *Data { pool: *pgxpool.Pool }
NewUserRepo(pool) → *UserRepo { q: *gen.Queries }
                     ├── CreateUser(ctx, CreateUserParams) → *CoreUser
                     ├── GetUser(ctx, id) → *CoreUser
                     ├── GetUserByEmail(ctx, email) → *CoreUser
                     ├── UpdateUser(ctx, UpdateUserParams) → *CoreUser
                     └── UsernameExists(ctx, username) → bool
```

- `Data` struct holds a `*pgxpool.Pool` — cleanup closes the pool
- `UserRepo` initializes `gen.Queries` with the pool via `gen.New(pool)`
- `pgxpool.Pool` satisfies the `gen.DBTX` interface (has `Exec`, `Query`, `QueryRow`)
- `ProviderSet` exposes `NewData` and `NewUserRepo` for Wire DI

## Verification

- `cd app/core && go build ./...` — **passes** without errors
- All old Todo references removed from non-test production code
