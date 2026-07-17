# Puchi Backend

Monorepo Go microservices cho **Puchi** — nền tảng học tiếng Việt.

Kiến trúc: **Kratos v3** monorepo, 7 service modules, Go workspace.

## Tech Stack

| Thành phần | Công nghệ |
|------------|-----------|
| Ngôn ngữ | Go 1.26 |
| Framework | Kratos v3 (HTTP + gRPC hybrid, Protobuf-first) |
| DI | Wire (compile-time) |
| Auth | **Limen** (`app/auth/`) — opaque session + Bearer introspect |
| Database | PostgreSQL 18 (CloudNativePG) |
| Cache | Valkey / Redis |
| Message Queue | NATS (event-driven) |
| Object Storage | Garage (S3 API) |
| Migration | goose |

## Cấu trúc monorepo

```
puchi-backend/
├── api/                      # Protobuf definitions (shared)
├── app/                      # Go service modules
│   ├── auth/                 # Limen auth-service (identity, OAuth, session)
│   ├── core/                 # Profile + gamification stats
│   ├── learn/                # Curriculum, attempts, guest trial, grading
│   ├── user/                 # Social features
│   ├── media/                # Upload (R2)
│   └── notification/         # Push, email
├── pkg/                      # Shared Kit library
│   └── auth/                 # Bearer session introspect middleware
├── go.work                   # Go workspace (dev only, không build CI)
└── Makefile
```

## Auth (Limen)

Auth-service mount [Limen](https://limenauth.dev/) tại `/auth/`. Các Go service khác verify qua `pkg/auth/`:

1. Parse `Authorization: Bearer <opaque>`
2. `GET {auth_service_url}/internal/session` (cache ~60s)
3. Inject user id vào context

Methods: Email/Password + Google + Facebook + TikTok.

OAuth callbacks: `https://api.puchi.io.vn/auth/oauth/{google|facebook|tiktok}/callback`

Chi tiết: xem `.cursor/rules/auth-service.mdc` và workspace spec `docs/superpowers/specs/2026-07-17-limen-auth-design.md`.

## Services

| Service | Module | Ports | Docker image | Envoy path |
|---------|--------|-------|-------------|------------|
| Auth | `app/auth` | **8000** | `puchi-auth` | `/auth` |
| Core | `app/core` | 8000/9000 | `puchi-core` | `/core` |
| Learn | `app/learn` | 8000/9000 | `puchi-learn` | `/v1/learn` |
| User | `app/user` | 8000/9000 | `puchi-user` | `/user` |
| Media | `app/media` | 8000/9000 | `puchi-media` | `/media` |
| Notification | `app/notification` | 8000/9000 | `puchi-notification` | `/notification` |

## Dev Local

```bash
# Auth (cần LIMEN_SECRET đúng 32 bytes + migration 001_limen_schema)
export LIMEN_SECRET="$(openssl rand -base64 24 | head -c 32)"
cd app/auth && go run ./cmd/auth/ -conf ../../configs

# Core
cd app/core && go run ./cmd/core/

# auth.auth_service_url = http://localhost:8000
```

## Deployment

CI/CD: GitHub Actions (`backend.yml`) — matrix build → `ghcr.io/puchidemy/puchi-{service}`.
