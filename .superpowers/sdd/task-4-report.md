# Task 4: Proto definition for ProfileService

## Summary
Created the protobuf definition for `ProfileService` with two RPCs (`GetProfile`, `UpdateProfile`) and generated Go code using `buf generate`.

## Files created

| File | Description |
|------|-------------|
| `app/core/api/profile/v1/profile.proto` | Proto source defining `ProfileService`, `User`, and `UpdateProfileRequest` |
| `app/core/api/profile/v1/profile.pb.go` | Generated message types (`User`, `UpdateProfileRequest`) |
| `app/core/api/profile/v1/profile_grpc.pb.go` | Generated gRPC client/server interfaces |
| `app/core/api/profile/v1/profile_http.pb.go` | Generated Kratos HTTP server interface |

## Proto details

- **Package**: `puchi.core.profile.v1`
- **Service**: `ProfileService`
  - `GetProfile(google.protobuf.Empty) → User` — GET /v1/profile
  - `UpdateProfile(UpdateProfileRequest) → User` — PUT /v1/profile (body: "*")
- **Messages**: `User` (9 fields), `UpdateProfileRequest` (4 fields)

## Generation

- **Working dir**: `app/core`
- **Command**: `buf generate` (buf.yaml + buf.gen.yaml already configured)
- **buf version**: v2 config with `buf.build/googleapis/googleapis` dependency
- **Plugins**: protoc-gen-go v1.36.11, protoc-gen-go-grpc v1.6.2, protoc-gen-go-http (kratos) v2.9.2
- **Go package**: `github.com/puchidemy/puchi-backend/app/core/api/profile/v1;v1`
- **Output**: All files generated with `paths=source_relative`

## Notable

- The `buf.gen.yaml` already existed with correct configuration (4 plugins: go, go-grpc, go-http/kratos, openapi)
- `buf.yaml` declares dependency on `buf.build/googleapis/googleapis` which provides `google/api/annotations.proto` and `google/protobuf/timestamp.proto`
- Followed existing convention from `api/todo/v1/todo.proto` for `go_package` and options
