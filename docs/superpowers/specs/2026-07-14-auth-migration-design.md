# Auth Migration: Clerk → Supertokens (Self-hosted)

> **⚠️ SUPERSEDED** — Auth hiện dùng **Limen**. Spec: `docs/superpowers/specs/2026-07-17-limen-auth-design.md` (workspace root).  
> Lịch sử: Clerk → Supertokens → Zitadel → custom JWT auth-service → **Limen**.

**Date:** 2026-07-14
**Status:** Historical
**Author:** Hoan

## Overview

Migrated authentication from Clerk (SaaS) to Supertokens (self-hosted on K3s) across puchi-frontend, puchi-backend, puchi-infra.

**Approach:** Hybrid D — Envoy ext_authz (layer 1) + Backend Go SDK verify (layer 2, zero trust)

**Auth methods:** Email/Password + Google + Facebook + TikTok

## Repository Changes

### puchi-backend — Phase 1 ✅

| Change | Files |
|--------|-------|
| Supertokens Go SDK (`github.com/supertokens/supertokens-golang v0.25.2`) | `app/core/go.mod` |
| Auth config proto (`Auth`, `Supertokens` messages) | `internal/conf/conf.proto` |
| Config YAML (connection_uri, api_key, public_paths) | `configs/config.yaml` |
| Auth middleware (verify session, inject user_id into context) | `internal/auth/middleware.go`, `context.go` |
| Supertokens client wrapper | `internal/auth/supertokens/client.go` |
| Wire ProviderSet + Init | `internal/auth/auth.go` |
| HTTP server integration (`http.Filter`) | `internal/server/http.go` |
| gRPC server placeholder | `internal/server/grpc.go` |
| Wire DI + main.go integration | `cmd/core/wire.go`, `main.go` |

### puchi-frontend — Phase 2 ✅

| Action | Details |
|--------|---------|
| Removed Clerk | `@clerk/nextjs`, `@clerk/localizations`, `@clerk/themes`, ClerkProvider, hooks, env vars |
| Added Supertokens | `supertokens-web-js@0.16.0`, `supertokens-node@24.0.2` |
| Client config | `src/config/supertokens.ts` |
| Server config (with TikTok custom OAuth) | `src/config/supertokens-server.ts` |
| API proxy route | `src/app/api/auth/[...path]/route.ts` |
| Auth pages | sign-in, sign-up, forgot-password, reset-password |
| Auth components | AuthCard, SignInForm, SignUpForm, ForgotPasswordForm, ResetPasswordForm, SocialLoginButtons |
| Middleware rewrite | `src/proxy.ts` — remove Clerk, use Supertokens `getSession()` |
| Provider | `src/providers/SupertokensProvider.tsx` |
| Service cleanup | `src/services/user.service.ts` — remove `getAuthHeaders` |
| Component cleanup | SidebarLeft, ProfileForm, ProfileActions, profile page, RightBarSection |

### puchi-infra — Phase 3 ✅

| Change | File |
|--------|------|
| NodePort 30567 for Supertokens | `argocd/apps/supertokens.yaml` |
| Envoy Gateway resource | `infra/envoy-gateway/gateway.yaml` |
| HTTP routes to backend services | `infra/envoy-gateway/httproute.yaml` |
| SecurityPolicy ext_authz | `infra/envoy-gateway/security-policy.yaml` |
| ArgoCD app for routes | `argocd/apps/envoy-gateway-routes.yaml` |

## Architecture

```
User → puchi.io.vn (FE)
   │
   ├─ auth → auth.puchi.io.vn (Supertokens Core, K3s)
   │            │ Email/Password, Google, Facebook, TikTok
   │            │ Sets session cookie
   │
   └─ API → Envoy Gateway (api.puchi.io.vn)
                │ SecurityPolicy → /session/verify
                │ → Go Backend
                       │ middleware → Supertokens SDK verify
                       │ inject user_id into context
                       │ biz layer reads auth.UserIDFromContext(ctx)
```

## Dev Local

```
FE (localhost:3000) → auth.puchi.io.vn (Supertokens on K3s, same LAN)
                    → localhost:8000 (Go binary)
                        → middleware verify → K3s NodePort :30567
```

## Remaining

- Replace `change-in-production-please` API key with real secret
- Apply `supertokens-auth-secret` to cluster (via kubectl or Sealed Secrets)
- Verify build FE: `bun run build`
- Wire up TikTok OAuth credentials in Supertokens env vars
