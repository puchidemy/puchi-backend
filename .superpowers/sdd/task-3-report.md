# Task 3 Report: Biz Layer — ProfileUsecase

## Completed

### Created
- **`internal/biz/profile.go`** — ProfileUsecase with:
  - `UserRepoInterface` (dependency inversion contract with flat params)
  - `UpdateProfileInput` validated input struct
  - `GetProfile(ctx, userID)` — fetch user by ID
  - `UpdateProfile(ctx, userID, input)` — validate username uniqueness, update profile
  - `CreateUserFromAuth(ctx, userID, email)` — create user with unique username from email
  - `generateUsername(email)` — sanitize email into username
  - Domain errors: `ErrUserNotFound`, `ErrUsernameTaken`
- **`internal/biz/biz.go`** — Wire `ProviderSet` with `NewProfileUsecase`

### Updated
- **`internal/data/user_repo.go`** — Refactored to satisfy `biz.UserRepoInterface`:
  - `CreateUser` now takes flat params `(id, username, email, firstName, lastName)` instead of `gen.CreateUserParams`
  - `UpdateUser` now takes flat params `(id, firstName, lastName, username, bio, avatarKey)` instead of `gen.UpdateUserParams`

### Deleted
- **`internal/biz/todo.go`** — stub Todo model + TodoUsecase (replaced by ProfileUsecase)
- **`internal/service/todo.go`** — TodoService (depended on deleted biz types)
- **`internal/service/todo_test.go`** — TodoService tests

### Updated (cleanup)
- **`internal/service/service.go`** — ProviderSet emptied (todo service removed)

## Build Verification
```bash
cd app/core && go build ./...
```
Exit code: 0 — all packages compile successfully.

## Notes
- The `UserRepoInterface` uses `*gen.CoreUser` (actual sqlc model name) instead of `*gen.User` from the original spec
- Flat-param interface in biz ensures the service layer doesn't need to construct sqlc param structs
- Data layer `UserRepo` now satisfies the biz interface by wrapping sqlc params internally
- Todo service was removed because it depended on deleted biz types and is not imported by the main app
