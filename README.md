# Puchi Backend

Monorepo Go microservices cho **Puchi** — nền tảng học tiếng Việt.

Kiến trúc: **Kratos v3** monorepo, 8 service modules, Go workspace.

## Tech Stack

| Thành phần | Công nghệ |
|------------|-----------|
| Ngôn ngữ | Go 1.26 |
| Framework | Kratos v3 (HTTP + gRPC hybrid, Protobuf-first) |
| DI | Wire (compile-time) |
| Auth | Auth-service tự xây dựng — JWT RS256, JWKS local cache |
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
├── app/                      # 8 Go service modules
│   ├── auth/                 # Identity: login, register, OAuth2, session, MFA, RBAC
│   ├── core/                 # Auth verify + User + Game (Phase 1)
│   ├── content/              # Courses, units, lessons
│   ├── grading/              # Dictation, listening grading
│   ├── user/                 # Profile, settings (Phase 3)
│   ├── game/                 # XP, leaderboard (Phase 3)
│   ├── media/                # Upload, resize (Phase 1)
│   └── notification/         # Push, email (Phase 3)
├── pkg/                      # Shared Kit library
│   └── auth/                 # JWT verification middleware (JWKS local, 15-min cache)
├── go.work                   # Go workspace (dev only, không build CI)
└── Makefile                  # Build helper
```

## Auth

Backend xác thực qua **JWT verification** từ auth-service tự xây dựng. Middleware trong `pkg/auth/`:

1. Parse `Authorization: Bearer <JWT>` từ request
2. Fetch JWKS từ auth-service (cached 15 phút)
3. Verify RS256 signature + issuer claim
4. Lấy userID từ `sub` claim → inject vào context

Double token pattern: JWT access (RS256, 15min) + opaque refresh (SHA-256, 30d rotation).

Auth methods: Email/Password (Argon2id) + Google + Facebook + TikTok + Magic Link + TOTP MFA.

## Services

| Service | Module | Ports | Docker image |
|---------|--------|-------|-------------|
| Auth | `app/auth` | 8080 | `puchi-auth` |
| Core | `app/core` | 8000/9000 | `puchi-core` |
| Content | `app/content` | 8000/9000 | `puchi-content` |
| Grading | `app/grading` | 8000/9000 | `puchi-grading` |
| User | `app/user` | 8000/9000 | `puchi-user` |
| Game | `app/game` | 8000/9000 | `puchi-game` |
| Media | `app/media` | 8000/9000 | `puchi-media` |
| Notification | `app/notification` | 8000/9000 | `puchi-notification` |

## Dev Local

```bash
# Auth service (cần private.pem)
cd app/auth && go run ./cmd/auth/ -conf ../../configs

# Core service
cd app/core && go run ./cmd/core/

# Auth config trong configs/config.yaml
# auth.issuer = https://api.puchi.io.vn
# auth.private_key_path = configs/private.pem
```

## Deployment

CI/CD: GitHub Actions (`backend.yml`) — matrix build 8 services → `ghcr.io/puchidemy/puchi-{service}`.
