# Task 5: Service Layer — ProfileService implementation

## Summary
Implemented the service layer for `ProfileService` with `GetProfile` and `UpdateProfile` methods, wired the full dependency chain (service → biz → data → pgxpool) using Google Wire, and registered both HTTP and gRPC transport.

## Files created / modified

| File | Change |
|------|--------|
| `app/core/internal/service/profile.go` | **Created** — `ProfileService` implementing both gRPC and HTTP server interfaces |
| `app/core/internal/service/service.go` | **Modified** — added `NewProfileService` to `ProviderSet` |
| `app/core/internal/server/http.go` | **Modified** — accepts `*service.ProfileService`, registers HTTP routes |
| `app/core/internal/server/grpc.go` | **Modified** — accepts `*service.ProfileService`, registers gRPC service |
| `app/core/internal/data/data.go` | **Modified** — exported `Pool` field, added `wire.FieldsOf` + `wire.Bind` for DI |
| `app/core/cmd/core/wire.go` | **Modified** — added `biz.ProviderSet` and `service.ProviderSet` |
| `app/core/cmd/core/wire_gen.go` | **Regenerated** by Wire with full dependency injection chain |

## Key details

### Service layer (`internal/service/profile.go`)

- `ProfileService` embeds `pb.UnimplementedProfileServiceServer`, satisfying both gRPC and HTTP interfaces from the generated proto
- `GetProfile` takes `*emptypb.Empty` (no `GetProfileRequest` message) — verified from `profile_grpc.pb.go`
- Both methods extract `userID` from context via `auth.UserIDFromContext(ctx)`
- `UpdateProfile` maps `biz.ErrUsernameTaken` → gRPC `codes.AlreadyExists`
- `userToProto` converts `gen.CoreUser` (domain model) → `pb.User` (proto), handling `*string` → `string` via `safePtr`

### Wire dependency chain

```
wireApp params
  ├─ data.NewData(cfg) → *data.Data
  │   └─ data.Data.Pool → *pgxpool.Pool
  │       └─ data.NewUserRepo(pool) → *data.UserRepo
  │           └─ [wire.Bind: *data.UserRepo → biz.UserRepoInterface]
  │               └─ biz.NewProfileUsecase(repo) → *biz.ProfileUsecase
  │                   └─ service.NewProfileService(uc) → *service.ProfileService
  │                       ├─ server.NewGRPCServer(cfg, auth, svc) → *grpc.Server
  │                       └─ server.NewHTTPServer(cfg, auth, svc) → *http.Server
  └─ newApp(logger, grpc, http) → *kratos.App
```

### Data package changes

- Exported `Data.Pool` field (was `pool`) — required because `wire_gen.go` lives in `main` package and needs to access the field
- Added `wire.FieldsOf(new(*Data), "Pool")` to extract `*pgxpool.Pool` from `*Data`
- Added `wire.Bind(new(biz.UserRepoInterface), new(*data.UserRepo))` to bind implementation to interface

### Registration

- **HTTP**: `GET /v1/profile`, `PUT /v1/profile` via `pb.RegisterProfileServiceHTTPServer`
- **gRPC**: via `pb.RegisterProfileServiceServer`

## Build

- `go build ./cmd/core/` — **success** (exit code 0)
- `wire ./cmd/core/` — generated `wire_gen.go` with full DI chain
