# P5: Media Service — Implementation Report

## Status: Complete ✓

## Tổng quan

Media service (`app/media/`) đã được implement đầy đủ theo kiến trúc Kratos v3 với avatar upload support. Service hỗ trợ các category: `avatar`, `lesson_image`, `lesson_audio`, `recording`.

## Cấu trúc files

```
app/media/
├── api/media/v1/
│   ├── media.proto              # Proto định nghĩa MediaService
│   ├── media.pb.go              # Generated protobuf
│   ├── media_grpc.pb.go         # Generated gRPC stubs
│   └── media_http.pb.go         # Generated Kratos HTTP stubs
├── configs/
│   └── config.yaml              # Config với Server + Data + Media sections
├── internal/
│   ├── conf/
│   │   ├── conf.proto           # Bootstrap: Server, Data, Media (Storage + Upload)
│   │   └── conf.pb.go           # Generated protobuf config
│   ├── data/
│   │   ├── data.go              # NewData với pgxpool + ProviderSet
│   │   ├── storage.go           # StorageProvider interface + MockStorage
│   │   ├── media_repo.go        # MediaRepo wrapping sqlc queries
│   │   └── sqlc/
│   │       ├── sqlc.yaml        # sqlc config (pgx/v5, pointers for nulls)
│   │       ├── queries/media.sql # CRUD queries: Create, Get, List, UpdateStatus, Delete
│   │       ├── migrations/
│   │       │   ├── 001_media.up.sql    # CREATE SCHEMA media + TABLE objects
│   │       │   └── 001_media.down.sql  # DROP TABLE + SCHEMA
│   │       └── gen/             # sqlc generated code
│   ├── biz/
│   │   ├── biz.go               # ProviderSet with NewMediaUsecase
│   │   └── media.go             # MediaUsecase: object key generation, validation, mock storage
│   ├── service/
│   │   ├── service.go           # ProviderSet with NewMediaService
│   │   └── media.go             # MediaService handlers + error mapping
│   └── server/
│       ├── server.go            # ProviderSet
│       ├── http.go              # HTTP server with MediaService registration
│       └── grpc.go              # gRPC server with MediaService registration
└── cmd/media/
    ├── main.go                  # Entry point, config loading
    ├── wire.go                  # Wire injection
    └── wire_gen.go              # Generated Wire code
```

## API Endpoints

| Method | Path | Handler |
|--------|------|---------|
| POST | `/v1/media/upload-url` | RequestUploadURL |
| POST | `/v1/media/finalize` | FinalizeUpload |
| gRPC | GetMedia | GetMedia |
| gRPC | DeleteMedia | DeleteMedia |

## Key Design Decisions

### 1. Object Key Pattern
`{category}/{user_id}/{uuid}.{ext}` — ví dụ: `avatar/user_abc/550e8400-e29b-41d4-a716-446655440000.jpg`

Extension được lấy từ `mime.ExtensionsByType()` (ví dụ `image/webp` → `.webp`, `audio/mpeg` → `.mpeg`).

### 2. Storage Provider (Mock)
Hiện tại dùng `MockStorage` — tạo fake URL `http://localhost:3900/puchi-media/{objectKey}`. Khi tích hợp MinIO/Garage thật, chỉ cần:
1. Implement `StorageProvider` interface với MinIO client
2. Inject config từ `conf.Media.Storage`
3. Replace `NewStorageProvider()` trong data.go

### 3. User Authentication
Không có auth middleware. User ID được lấy từ header `X-User-ID` trong gRPC metadata (HTTP gateway map). Frontend cần gửi header này hoặc tích hợp Supertokens session.

### 4. Content Validation
- Category phải thuộc: `avatar`, `lesson_image`, `lesson_audio`, `recording`
- Content type phải bắt đầu bằng: `image/`, `audio/`, `video/`

### 5. Config (conf.proto)
```protobuf
message Bootstrap {
  Server server = 1;
  Data data = 2;
  Media media = 3;
}
message Media {
  message Storage {
    string endpoint = 1;
    string access_key_id = 2;
    string secret_access_key = 3;
    string bucket = 4;
    bool use_ssl = 5;
    string region = 6;
  }
  message Upload {
    int64 max_size_bytes = 1;
    int64 presigned_url_ttl = 2;
  }
}
```

## Build & Migration

- **Code generation**: `buf generate` (cả API và config) + `sqlc generate` + `wire` — đều thành công
- **Build**: `go build ./cmd/media/` — thành công
- **Go vet**: `go vet ./...` — không lỗi
- **Migration**: đã chạy qua Docker psql tới PostgreSQL cluster (192.168.100.201:30433)

## Dependencies added

- `github.com/jackc/pgx/v5 v5.10.0` (PostgreSQL driver)
- `github.com/google/uuid v1.6.0` (was indirect, now direct)

## Cleanup

Đã xoá skeleton files từ Kratos template (todo service):
- `internal/data/todo.go`, `internal/biz/todo.go`, `internal/service/todo.go`, `internal/service/todo_test.go`
- `api/todo/v1/` directory
- Various `README.md` trong internal layers

## Future work

1. **MinIO/Garage integration**: Replace `MockStorage` với real S3 client, dùng config từ `conf.Media.Storage`
2. **Auth middleware**: Thêm middleware để verify Supertokens session thay vì dùng header
3. **Image processing**: Thêm resize/optimize cho avatar uploads
4. **CDN integration**: Thêm CDN URL generation thay vì presigned GET URLs
