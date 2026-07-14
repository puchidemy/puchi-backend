# Puchi Backend

Monorepo Go microservices cho **Puchi** — nền tảng học tiếng Việt.

Kiến trúc: **Kratos v3** monorepo, 7 service modules, Go workspace.

## Tech Stack

| Thành phần | Công nghệ |
|------------|-----------|
| Ngôn ngữ | Go 1.26 |
| Framework | Kratos v3 (HTTP + gRPC hybrid, Protobuf-first) |
| DI | Wire (compile-time) |
| Auth | Supertoken (self-host) |
| Database | PostgreSQL 18 (CloudNativePG) |
| Cache | Valkey / Redis |
| Message Queue | NATS (event-driven) |
| Object Storage | Garage (S3 API) |
| Migration | goose |

## Cấu trúc monorepo

```
puchi-backend/
├── api/                      # Protobuf definitions (shared)
│   ├── auth/v1/
│   ├── content/v1/
│   ├── grading/v1/
│   ├── user/v1/
│   ├── game/v1/
│   ├── media/v1/
│   └── notification/v1/
├── app/                      # 7 Go service modules
│   ├── core/                 # Auth + User + Game (Phase 1)
│   ├── content/              # Courses, units, lessons
│   ├── grading/              # Dictation, listening grading
│   ├── user/                 # Profile, settings (Phase 3)
│   ├── game/                 # XP, leaderboard (Phase 3)
│   ├── media/                # Upload, resize, transcode
│   └── notification/         # Push, email (Phase 3)
├── pkg/                      # Shared Kit library
├── .cursor/rules/
│   ├── project.mdc           # Project overview + infra connections
│   ├── kratos.mdc            # Kratos v3 layered architecture
│   └── proto.mdc             # Protobuf conventions
├── go.work                   # Go workspace (dev only)
└── Makefile
```

## Development

### Requirements

- Go 1.26+
- Kratos CLI v3 (`go install github.com/go-kratos/kratos/cmd/kratos/v3@latest`)
- Docker + Docker Compose (cho infra local)

### Local dev

```bash
# Clone
git clone https://github.com/puchidemy/puchi-backend.git
cd puchi-backend

# Với Go workspace (dùng local)
# go.work đã config sẵn, chỉ cần:
go work sync

# Build 1 service
cd app/core && go build -o ../../bin/core ./cmd/core/

# Hoặc dùng Makefile
make build-all
```

### Thêm service mới

```bash
cd app
kratos new <service-name>
# Sau đó fix module name + go.work
```

### Layered architecture (Kratos)

Mỗi service follow:

```
internal/
├── service/     # Transport — validate input, gọi biz
├── biz/         # Domain logic — usecases, entities, repo interfaces
└── data/        # Repository implementations — DB, cache, S3
```

**Rule:** `service → biz → data` (biz không bao giờ import data)

## CI/CD

GitHub Actions — Matrix build 7 services:

| Service | Image | Ports |
|---------|-------|-------|
| core | `ghcr.io/puchidemy/puchi-core` | 8080 HTTP, 9090 gRPC |
| content | `ghcr.io/puchidemy/puchi-content` | 8080, 9090 |
| grading | `ghcr.io/puchidemy/puchi-grading` | 8080, 9090 |
| user | `ghcr.io/puchidemy/puchi-user` | 8080, 9090 |
| game | `ghcr.io/puchidemy/puchi-game` | 8080, 9090 |
| media | `ghcr.io/puchidemy/puchi-media` | 8080, 9090 |
| notification | `ghcr.io/puchidemy/puchi-notification` | 8080, 9090 |

## Kết nối Infra (K3s)

| Service | Internal DNS |
|---------|--------------|
| PostgreSQL | `pg-puchi-rw.puchi-db.svc.cluster.local:5432` |
| Supertoken DB | `pg-supertokens-rw.puchi-db.svc.cluster.local:5432` |
| NATS | `nats.platform.svc.cluster.local:4222` |
| Garage (S3) | `http://garage.platform.svc.cluster.local:3900` |
| Valkey | `valkey-node.platform.svc.cluster.local:6379` |
| Supertoken | `supertokens.puchi-infra.svc.cluster.local:3567` |

## Lộ trình

| Phase | Thời gian | Mô tả |
|-------|-----------|-------|
| 1 | 1-2 tháng | Monolith modular: core (auth+user+game) + content + media |
| 2 | tháng 3-4 | Tách auth + grading service, thêm NATS |
| 3 | tháng 5-6 | Tách game + user + notification, bật Istio |
