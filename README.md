# Puchi Backend

Monorepo Go microservices cho **Puchi** — nền tảng học tiếng Việt.

Kiến trúc: **Kratos v3** monorepo, 7 service modules, Go workspace.

## Tech Stack

| Thành phần | Công nghệ |
|------------|-----------|
| Ngôn ngữ | Go 1.26 |
| Framework | Kratos v3 (HTTP + gRPC hybrid, Protobuf-first) |
| DI | Wire (compile-time) |
| Auth | Zitadel (self-host) — JWT verification qua JWKS |
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
│   ├── core/                 # Auth verify + User + Game (Phase 1)
│   ├── content/              # Courses, units, lessons
│   ├── grading/              # Dictation, listening grading
│   ├── user/                 # Profile, settings (Phase 3)
│   ├── game/                 # XP, leaderboard (Phase 3)
│   ├── media/                # Upload, resize (Phase 1)
│   └── notification/         # Push, email (Phase 3)
├── pkg/                      # Shared Kit library
├── go.work                   # Go workspace (dev only, không build CI)
└── Makefile                  # Build helper
```

## Auth

Backend xác thực qua **JWT verification** từ Zitadel. Middleware trong `app/core/internal/auth/`:
1. Parse `Authorization: Bearer <JWT>` từ request
2. Fetch JWKS từ `https://auth.puchi.io.vn/oauth/v2/keys` (cached 15 phút)
3. Verify signature + issuer claim
4. Lấy userID từ `sub` claim → inject vào context

Signin/signup do frontend (Next.js) xử lý qua Zitadel Session API + Auth.js OIDC.

## Services

| Service | Module | Ports | Docker image |
|---------|--------|-------|-------------|
| Core | `app/core` | 8000/9000 | `puchi-core` |
| Content | `app/content` | 8000/9000 | `puchi-content` |
| Grading | `app/grading` | 8000/9000 | `puchi-grading` |
| User | `app/user` | 8000/9000 | `puchi-user` |
| Game | `app/game` | 8000/9000 | `puchi-game` |
| Media | `app/media` | 8000/9000 | `puchi-media` |
| Notification | `app/notification` | 8000/9000 | `puchi-notification` |

## Dev Local

```bash
# Core service
cd app/core && go run ./cmd/core/

# Auth config trong configs/config.yaml
# auth.zitadel.issuer_url = https://auth.puchi.io.vn
# auth.zitadel.jwks_url = https://auth.puchi.io.vn/oauth/v2/keys
```

## Deployment

CI/CD: GitHub Actions (`backend.yml`) — matrix build 7 services → `ghcr.io/puchidemy/puchi-{service}`.
