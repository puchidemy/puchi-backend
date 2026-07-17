# Puchi Auth — Limen Design Specification

**Status:** Current  
**Date:** 2026-07-17  
**Replaces:** `2026-07-16-auth-service-design.md` (custom JWT stack)

---

## 1. Executive Summary

Auth dùng **[Limen](https://limenauth.dev/)** — thin Go service mount handler tại `/auth/`. Session opaque (cookie + Bearer), không JWT/JWKS.

### Why Limen?

| Trước (custom JWT) | Limen |
|--------------------|-------|
| Tự build login/OAuth/session/MFA | Plugins chuẩn + thin wrapper |
| JWT RS256 + JWKS + refresh rotation | Opaque session + Bearer |
| Nhiều endpoint tự maintain | Limen HTTP API + `/internal/session` |

### Scope (hiện tại)

- Email/password (`credential-password`)
- OAuth: Google, Facebook, TikTok (custom Login Kit provider)
- Email verify + password reset (qua Limen + NATS `email.send`)
- Session cookie cross-subdomain + Bearer cho Go APIs
- NATS `auth.user.created` → Core profile sync

### Out of scope (ngày 1)

- Magic link, TOTP MFA, RBAC trong auth-service

---

## 2. Architecture

```
Browser (Next.js + limen-auth/react + bearerPlugin)
  │
  ├─► https://api.puchi.io.vn/auth/*  →  auth-service:8000 (Limen)
  │      ├── credential-password
  │      ├── oauth-google / oauth-facebook / oauth-tiktok
  │      └── session cookie + opaque Bearer
  │
  └─► https://api.puchi.io.vn/{core,user,...}  →  Go services
         Authorization: Bearer <opaque>
         pkg/auth → GET auth-service/internal/session (+ cache ~60s)
```

| Layer | Path / Package | Role |
|-------|----------------|------|
| Frontend | `src/lib/limen-auth.ts`, `AuthProvider` | Limen client |
| Auth service | `puchi-backend/app/auth/` | Limen + TikTok provider + NATS |
| Shared middleware | `puchi-backend/pkg/auth/` | Bearer introspect |
| Infra | `puchi-infra/infra/backend/auth/` | Deploy + cluster secrets |

---

## 3. Session model

| Mechanism | Use |
|-----------|-----|
| HttpOnly session cookie | Browser ↔ Limen (`/auth/*`), domain `puchi.io.vn` |
| Opaque Bearer | Frontend → Go APIs; Go → `GET /internal/session` |

Không còn: JWT access 15m, opaque refresh 30d rotation, JWKS, Next `/api/auth/set-session`.

---

## 4. OAuth

| Provider | Plugin / code | Env |
|----------|---------------|-----|
| Google | `oauth-google` | `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` |
| Facebook | `oauth-facebook` | `FACEBOOK_CLIENT_ID`, `FACEBOOK_CLIENT_SECRET` |
| TikTok | `internal/oauth/tiktok` | `TIKTOK_CLIENT_KEY`, `TIKTOK_CLIENT_SECRET` |

**Callback URIs** (đăng ký trên provider console):

```
https://api.puchi.io.vn/auth/oauth/google/callback
https://api.puchi.io.vn/auth/oauth/facebook/callback
https://api.puchi.io.vn/auth/oauth/tiktok/callback
```

Facebook: phải set **Valid OAuth Redirect URIs** trong Facebook Login settings (không chỉ App Domains).

TikTok: synthetic email `{open_id}@tiktok.oauth.puchi.local` khi provider không trả email.

---

## 5. Secrets (cluster)

| Secret | Keys | Git? |
|--------|------|------|
| `auth-limen-secret` | `LIMEN_SECRET` (đúng 32 bytes) | Không — cluster-managed |
| `auth-oauth-credentials` | `GOOGLE_*`, `FACEBOOK_*`, `TIKTOK_*` | Không — cluster-managed |

Deploy `envFrom` / `secretKeyRef` trỏ các secret trên. Không commit giá trị rỗng vào kustomize (Argo sẽ ghi đè).

---

## 6. Database

Migration: `app/auth/migrations/001_limen_schema.up.sql`

- Drop schema auth cũ (JWT stack)
- Tạo bảng Limen (`users`, `sessions`, `accounts`, verifications, …)
- UUID PKs

Schema: `auth.*` (`search_path=auth`).

---

## 7. Events

| Subject | When |
|---------|------|
| `auth.user.created` | User mới (credential hoặc OAuth) |
| `email.send` | Verify / reset email |

---

## 8. Frontend integration

```ts
// limen-auth client
baseURL = NEXT_PUBLIC_API_URL   // https://api.puchi.io.vn
basePath = "/auth"
plugins = [credentialPassword, oauthClient, bearer]
```

- Forms gọi `authClient.signIn/signUp.credential`
- Social: `authClient.oauth.signIn({ provider })`
- Go API: `fetchWithAuth` + Bearer từ bearerPlugin / token-manager

---

## 9. Dev local

```bash
# Postgres: apply 001_limen_schema.up.sql (search_path=auth)
export LIMEN_SECRET="$(openssl rand -base64 24 | head -c 32)"  # đúng 32 chars
cd puchi-backend/app/auth && go run ./cmd/auth/ -conf ../../configs

# Frontend
cd puchi-frontend && bun dev
# NEXT_PUBLIC_API_URL=http://localhost:8000 (hoặc URL auth local)
```

---

## 10. Related docs

| Doc | Role |
|-----|------|
| `.cursor/rules/ecosystem.mdc`, `puchi.mdc` | Workspace overview |
| `puchi-backend/.cursor/rules/auth-service.mdc` | Auth service conventions |
| `puchi-frontend/.cursor/rules/project.mdc` | FE auth flow |
| `docs/superpowers/specs/2026-07-16-auth-service-design.md` | **SUPERSEDED** |
