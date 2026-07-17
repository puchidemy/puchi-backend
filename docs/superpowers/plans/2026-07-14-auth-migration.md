# Auth Migration Implementation Plan (Updated)

> **?? SUPERSEDED (2026-07-17)** � Auth �? chuy?n sang **Limen**. Spec hi?n t?i: `docs/superpowers/specs/2026-07-17-limen-auth-design.md`. Plan d�?i ��y l� l?ch s? JWT custom � **kh�ng** follow �? implement.


> **⚠️ SUPERSEDED** — Auth đã migrate từ Supertokens → Auth-service tự xây dựng (Go/Kratos), xoá hoàn toàn Zitadel.
> Xem main spec: `docs/superpowers/specs/2026-07-16-auth-service-design.md`

> **Status:** 28/31 tasks completed ✅

## Phase 1: Backend Auth Middleware ✅ All Done

| # | Task | Status |
|---|------|--------|
| 1 | Add Supertokens SDK dependency | ✅ `f11817a` |
| 2 | Update conf.proto — add Auth config | ✅ `f11817a` |
| 3 | Update config.yaml | ✅ `f11817a` |
| 4 | Create auth struct directory | ✅ `f11817a` |
| 5 | Create auth context helpers | ✅ `f11817a` |
| 6 | Create auth HTTP middleware | ✅ `f11817a` |
| 7 | Create auth Wire ProviderSet | ✅ `f11817a` |
| 8 | Update HTTP server | ✅ `f11817a` |
| 9 | Update gRPC server | ✅ `f11817a` |
| 10 | Update Wire injection + main.go | ✅ `f11817a` |
| 11 | Write middleware unit test | ✅ `f11817a` |

## Phase 2: Frontend ✅ All Done

| # | Task | Status |
|---|------|--------|
| 12 | Remove Clerk, add Supertokens deps | ✅ `b74f939` |
| 13 | Delete Clerk files + env | ✅ `b74f939` |
| 14 | Create Supertokens client config | ✅ `b74f939` |
| 15 | Create Supertokens server config | ✅ `b74f939` |
| 16 | Create SupertokensProvider | ✅ `b74f939` |
| 17 | Create API route for Supertokens | ✅ `b74f939` |
| 18 | Create auth page routes | ✅ `b74f939` |
| 19 | Create AuthCard | ✅ `b74f939` |
| 20 | Create SignInForm | ✅ `b74f939` |
| 21 | Create SignUpForm | ✅ `b74f939` |
| 22 | Create Forgot/Reset forms | ✅ `b74f939` |
| 23 | Create SocialLoginButtons | ✅ `b74f939` |
| 24 | Rewrite proxy.ts | ✅ `b74f939` |
| 25 | Update protected layout | ✅ `b74f939` |
| 26 | Update user.service.ts | ✅ `b74f939` |
| 27 | Update Clerk-dependent components | ✅ `b74f939` |

## Phase 3: Infrastructure ✅ Mostly Done

| # | Task | Status |
|---|------|--------|
| 28 | Update Supertokens Helm values (NodePort) | ✅ `3109ca1` |
| 29 | Create Envoy Gateway API resources | ✅ `3109ca1` |
| 30 | Create ArgoCD app for Envoy routes | ✅ `3109ca1` |
| 31 | Create Supertokens auth secret | ⏳ Need kubectl apply |

## Remaining Work

### Task 31: Supertokens Auth Secret

Apply to cluster (one-time):
```bash
kubectl create secret generic supertokens-auth-secret \
  -n puchi-infra \
  --from-literal=supertokens-api-key=<api-key> \
  --from-literal=google-client-id=<google-client-id> \
  --from-literal=google-client-secret=<google-client-secret> \
  --from-literal=facebook-client-id=<facebook-client-id> \
  --from-literal=facebook-client-secret=<facebook-client-secret> \
  --from-literal=tiktok-client-key=<tiktok-client-key> \
  --from-literal=tiktok-client-secret=<tiktok-client-secret>
```

### Verify Build

```bash
cd puchi-frontend && bun run build
```

### Clean up old directory

```powershell
Remove-Item -Recurse -Force D:\Github\puchidemy-old
```
