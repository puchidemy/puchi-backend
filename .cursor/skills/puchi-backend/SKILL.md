---
name: puchi-backend
description: Backend conventions for the Puchi Go microservices monorepo — sqlc patterns, database design, FE↔BE communication, gRPC inter-service patterns, testing, auth sync, and error handling. Use when writing Go backend code, creating services, defining database schemas, or setting up Server Actions.
---

# Puchi Backend — sqlc, Database, FE↔BE, gRPC

Existing rules cover Kratos v3 basics, auth middleware, proto conventions, CI/CD, and infra. This skill adds patterns NOT in those rules.

---

## 1. sqlc — Database Layer

### 1.1 Project Setup

```bash
# Each service has its own sqlc project
mkdir -p app/core/internal/data/sqlc/queries
mkdir -p app/core/internal/data/sqlc/migrations
```

```
app/{service}/internal/data/sqlc/
├── sqlc.yaml           # sqlc config
├── queries/            # .sql query files
│   ├── users.sql
│   └── stats.sql
├── migrations/         # goose up/down SQL files
│   ├── 001_init.up.sql
│   └── 001_init.down.sql
└── gen/                # GENERATED — do not edit
    ├── db.go
    ├── models.go
    └── queries.sql.go
```

### 1.2 sqlc Config

```yaml
# sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries/"
    schema: "migrations/"
    gen:
      go:
        package: "gen"
        out: "gen"
        sql_package: "pgx/v5"
        emit_db_tags: true
        emit_interface: true        # Querier interface for mocking
        emit_empty_slices: true
        emit_pointers_for_null_types: true
        overrides:
          - db_type: "timestamptz"
            go_type: "time.Time"
          - db_type: "uuid"
            go_type: "github.com/google/uuid.UUID"
          - column: "core.users.id"
            go_type: "github.com/google/uuid.UUID"
```

### 1.3 Query File Format

```sql
-- queries/users.sql
-- name: GetUser :one
SELECT * FROM core.users WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM core.users WHERE username = $1;

-- name: CreateUser :one
INSERT INTO core.users (id, username, email, first_name, last_name)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateUser :one
UPDATE core.users
SET first_name = $2, last_name = $3, username = $4, bio = $5, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListUsers :many
SELECT * FROM core.users ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: BatchGetUsers :many
SELECT * FROM core.users WHERE id = ANY($1::text[]);
```

### 1.4 Repository Wrapper (in data layer)

```go
// internal/data/user_repo.go
package data

import "github.com/jackc/pgx/v5/pgxpool"
import "github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"

type UserRepo struct {
    pool *pgxpool.Pool
    q    *gen.Queries  // sqlc-generated
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
    return &UserRepo{pool: pool, q: gen.New(pool)}
}

func (r *UserRepo) GetUser(ctx context.Context, id uuid.UUID) (*gen.User, error) {
    return r.q.GetUser(ctx, id)
}
```

### 1.5 Database Connection Setup

```go
// internal/data/data.go
import "github.com/jackc/pgx/v5/pgxpool"

type Data struct {
    DB *pgxpool.Pool
}

func NewData(cfg *conf.Data) (*Data, func(), error) {
    pool, err := pgxpool.New(context.Background(), cfg.Database.Source)
    return &Data{DB: pool}, func() { pool.Close() }, nil
}
```

### 1.6 Migrations (goose)

```bash
# Create migration
goose -dir app/core/internal/data/sqlc/migrations create init_schema sql
```

Result:

```
migrations/
├── 001_init_schema.up.sql
└── 001_init_schema.down.sql
```

Run in main.go or wire:

```go
import "github.com/pressly/goose/v3"

func runMigrations(pool *pgxpool.Pool) error {
    conn, _ := pool.Acquire(ctx)
    defer conn.Release()
    return goose.Up(conn.Conn().RawConn(), "migrations")
}
```

---

## 2. Database Design Conventions

### 2.1 Naming

| Concept | Convention | Example |
|---|---|---|
| Schema | `{service}_name` | `core`, `user_social` |
| Table | `snake_case`, plural only if collection | `users`, `daily_activities` |
| Column | `snake_case` | `first_name`, `created_at` |
| Primary key | `id TEXT` (UUID from Zitadel for users), `BIGSERIAL` for auto-increment | |
| Foreign key | `{referenced_table}_id` | `user_id` |
| Timestamp | `TIMESTAMPTZ`, suffix `_at` | `created_at`, `unlocked_at` |
| Boolean | `IS_` prefix or `_enabled` suffix | `unlocked`, `push_enabled` |

### 2.2 Schema Ownership

Each service OWNS one schema. Cross-schema access MUST go through gRPC — never through direct SQL JOINs.

| Service | Schema | Cross-service access via |
|---|---|---|
| Core | `core.*` | `CoreInternalService` gRPC |
| User | `user_social.*` | `SocialService` gRPC |
| Notification | `notification.*` | gRPC |
| Media | `media.*` | gRPC + HTTP presigned |

### 2.3 Migration Rules

- Always pair `up.sql` + `down.sql`
- Never modify existing migrations — always add new ones
- Seed data goes in separate `seed.sql` run after migrations
- Timestamps use `now()` in SQL, never Go-side default

---

## 3. FE ↔ BE Communication

### 3.1 Architecture

```
FE (Next.js)                    Go Backend
─────────                       ──────────
Server Component ──fetch()──→   HTTP REST (gRPC-gateway)
  → backendFetch("/v1/profile")

Client Component                Server Action
  <form action={serverAction}>  → backendFetch("/v1/profile")
                                  → 1. getSession()
                                  → 2. Zod validate
                                  → 3. gRPC call → Go service
                                  → 4. revalidatePath()
                                  → 5. Return ActionResult<T>
```

### 3.2 backendFetch — FE utility

```typescript
// src/lib/backend.ts — "server-only"
export async function backendFetch<T>(path: string, opts?: FetchOptions): Promise<T>
// Auto-attaches Authorization: Bearer {accessToken}
// 15s timeout
// Error response → BackendError(code, status, path)
// → FE maps to i18n key via getErrorI18nKey()
```

### 3.3 ActionResult — Standard Response

```typescript
// Both FE and BE use this pattern
type ActionResult<T> =
  | { success: true; data: T }
  | { success: false; error: string }  // i18n key
```

### 3.4 Validation

- FE: Zod schemas in `src/lib/schemas/` — validate before sending
- BE: Kratos validate.Validator middleware + biz layer validation
- Error messages use i18n keys, not raw strings

### 3.5 Upload Flow (2-step)

```
1. FE → POST /v1/media/upload-url {category, contentType, size}
   ← {uploadUrl, objectKey, mediaId}

2. FE → PUT uploadUrl + file (direct to S3, bypasses backend)

3. FE → POST /v1/media/finalize {mediaId}
   ← {id, url, width, height, durationMs}
```

---

## 4. gRPC Inter-Service Communication

### 4.1 Pattern: Service A calls Service B

```go
// User Service cần lấy user data từ Core Service
import corePb "github.com/puchidemy/puchi-backend/app/core/api/core/v1"

type SocialUsecase struct {
    coreClient corePb.CoreInternalServiceClient
}

func (uc *SocialUsecase) GetFollowers(ctx context.Context, userID string) {
    // Gọi gRPC sang Core để lấy thông tin users
    resp, err := uc.coreClient.BatchGetUsers(ctx, &corePb.BatchGetUsersRequest{
        Ids: followerIDs,
    })
    // Merge với data local
}
```

### 4.2 Wire Integration (DI for gRPC clients)

```go
// ProviderSet for gRPC clients
var ClientSet = wire.NewSet(
    NewCoreInternalClient,
)

func NewCoreInternalClient() corePb.CoreInternalServiceClient {
    conn, _ := grpc.Dial(
        "core-service:9000",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    return corePb.NewCoreInternalServiceClient(conn)
}
```

### 4.3 NATS Events (async)

| Event | Publisher | Subscribers |
|---|---|---|
| `lesson.completed` | Grading | Core (update stats/xp), Notification (congrats) |
| `user.followed` | User | Notification (new follower alert) |
| `user.created` | Core (webhook/lazy) | Notification (welcome), User (init social) |
| `achievement.unlocked` | Core | Notification (badge notification) |

Subject format: `puchi.{domain}.{event}`

---

## 5. Auth Sync — Zitadel JWT → Core DB

### Hybrid Strategy

**Path A — JWT lazy creation (primary):**
```
FE request → Core middleware verify JWT → userId từ sub claim
  → SELECT users WHERE id = userId → NOT FOUND
  → Extract email từ JWT claims
  → INSERT users (auto-generate username), INSERT user_stats
  → NATS publish "user.created"
```

**Path B — Fallback via Zitadel API:**
```
Khi JWT không có email claim, gọi Zitadel Management API để lấy user info
  → GET /v2beta/users/{userId}
  → INSERT users
```

### Username Auto-generation

```go
func generateUsername(ctx context.Context, email string, repo UserRepo) string {
    localPart := strings.Split(email, "@")[0]
    // Clean: keep only [a-z0-9]
    localPart = regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(strings.ToLower(localPart), "")
    if len(localPart) < 3 { localPart = "puchi_user" }
    // Uniqueness check with counter
    username := localPart
    for i := 1; ; i++ {
        exists, _ := repo.UsernameExists(ctx, username)
        if !exists { return username }
        username = fmt.Sprintf("%s%d", localPart, i)
    }
}
```

---

## 6. Testing

### 6.1 Layer Testing Strategy

| Layer | Test with | Example |
|---|---|---|
| **biz** (usecase) | Mock repo interface | `mockUserRepo := &MockUserRepo{}` |
| **data** (repo) | Testcontainers PostgreSQL | Real DB in Docker |
| **service** (gRPC handler) | Mock biz usecase | `mockUsecase := &MockProfileUsecase{}` |
| **integration** | Testcontainers + real services | End-to-end gRPC calls |

### 6.2 Biz Layer Test (Mock)

```go
func TestProfileUsecase_UpdateProfile(t *testing.T) {
    repo := &MockUserRepo{
        GetUserFn: func(ctx context.Context, id uuid.UUID) (*gen.User, error) {
            return &gen.User{Username: "old"}, nil
        },
    }
    uc := NewProfileUsecase(repo)
    user, err := uc.UpdateProfile(ctx, userID, UpdateInput{FirstName: "New"})
    assert.NoError(t, err)
    assert.Equal(t, "New", user.FirstName)
}
```

### 6.3 Data Layer Test (Testcontainers)

```go
func TestUserRepo_GetUser(t *testing.T) {
    ctx := context.Background()
    container, _ := testcontainers.GenericContainer(ctx, ...)
    pool := connectToContainer(container)
    goose.Up(pool, "migrations")

    repo := NewUserRepo(pool)
    user, err := repo.GetUser(ctx, testUserID)
    assert.NoError(t, err)
    assert.Equal(t, "puchiuser", user.Username)
}
```

### 6.4 sqlc Querier Mocking

```go
// sqlc generates Querier interface — use for mocking
type MockQuerier struct {
    GetUserFunc func(ctx context.Context, id uuid.UUID) (*gen.User, error)
}

func (m *MockQuerier) GetUser(ctx context.Context, id uuid.UUID) (*gen.User, error) {
    return m.GetUserFunc(ctx, id)
}
```

---

## 7. Error Handling

### 7.1 Go Backend Errors

```go
import "github.com/go-kratos/kratos/v3/errors"

// Defined errors
var (
    ErrUserNotFound    = errors.NotFound("USER_NOT_FOUND", "user not found")
    ErrUsernameTaken   = errors.AlreadyExists("USERNAME_TAKEN", "username already taken")
    ErrInvalidInput    = errors.BadRequest("INVALID_INPUT", "invalid input")
    ErrUnauthorized    = errors.Unauthorized("UNAUTHORIZED", "unauthorized")
    ErrRateLimited     = errors.ResourceExhausted("RATE_LIMITED", "rate limited")
)
```

### 7.2 Error Code → i18n Key Mapping (FE)

```go
// Backend returns error code → FE maps to i18n key
"USER_NOT_FOUND"   → "errors.profile.notFound"
"USERNAME_TAKEN"   → "errors.validation.usernameTaken"
"INVALID_INPUT"    → "errors.validation.invalidInput"
"UNAUTHORIZED"     → "errors.auth.unauthorized"
"RATE_LIMITED"     → "errors.server.rateLimited"
```

### 7.3 Service Layer Error Handling

```go
func (s *ProfileService) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.User, error) {
    user, err := s.uc.UpdateProfile(ctx, auth.UserIDFromContext(ctx), req)
    if errors.Is(err, biz.ErrUsernameTaken) {
        return nil, kerrors.AlreadyExists("USERNAME_TAKEN", err.Error())
    }
    if err != nil {
        return nil, kerrors.InternalServer("INTERNAL_ERROR", err.Error())
    }
    return toProto(user), nil
}
```

---

## 8. Config Pattern

```yaml
# configs/config.yaml
server:
  http:
    addr: 0.0.0.0:8000
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9000
    timeout: 1s
data:
  database:
    source: "postgresql://user:pass@host:5432/puchi?sslmode=disable"
  redis:
    addr: valkey:6379
auth:
  zitadel:
    issuer_url: "https://auth.puchi.io.vn"
    jwks_url: "https://auth.puchi.io.vn/oauth/v2/keys"
  public_paths:
    - /v1/health
    - /v1/healthz
```

Config proto for data section (update existing template):

```protobuf
message Data {
  message Database {
    string source = 1;
  }
  message Redis {
    string addr = 1;
  }
  Database database = 1;
  Redis redis = 2;
}
```

---

## 9. Service Scaffold Checklist

When creating a NEW service:

- [ ] Create `app/{service}/` with go.mod referencing `github.com/puchidemy/puchi-backend/app/{service}`
- [ ] Setup sqlc: `sqlc.yaml` + `queries/` + `migrations/` + `gen/`
- [ ] Define proto: `api/{service}/v1/{service}.proto` with HTTP + gRPC annotations
- [ ] Implement layers: `service → biz → data` with Wire ProviderSet per layer
- [ ] Add auth middleware if service needs session verification
- [ ] Configure gRPC client connections in Wire for cross-service calls
- [ ] Add NATS publisher/subscriber if service produces/consumes events
- [ ] Add config section in `configs/config.yaml`
- [ ] Add Dockerfile + CI matrix entry
