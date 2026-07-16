### Task 2: Backend — Proto + Biz + Service

**Files:**
- Modify: `app/core/api/profile/v1/profile.proto`
- Modify: `app/core/internal/biz/profile.go`
- Modify: `app/core/internal/data/user_repo.go`
- Modify: `app/core/internal/service/profile.go`

**Context:** This builds on Task 1's queries (`GetUserByUsername`, `UpdateOnboardingInfo`, `UpsertUserOnboarding`) by wiring them through data → biz → service layers. It also adds proto definitions for the new RPCs.

**Import paths:**
- Auth: `"github.com/puchidemy/puchi-backend/app/core/internal/auth"` (exports `UserIDFromContext(ctx) (string, bool)`)
- Biz: `"github.com/puchidemy/puchi-backend/app/core/internal/biz"`
- Data sqlc gen: `"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"`
- Supertokens: `"github.com/supertokens/supertokens-golang/supertokens"` (exports `GetUser(userID string)`)
- Proto libs: `"google.golang.org/protobuf/types/known/timestamppb"`, `"google.golang.org/protobuf/types/known/emptypb"`

- [ ] **Step 1: Add proto messages and RPCs to profile.proto**

Add these new RPCs to the `ProfileService`:
```protobuf
service ProfileService {
  // ... existing RPCs unchanged ...

  rpc GetProfileByUsername(GetProfileByUsernameRequest) returns (User) {
    option (google.api.http) = {
      get: "/v1/profile/{username}"
    };
  }

  rpc CompleteOnboarding(CompleteOnboardingRequest) returns (User) {
    option (google.api.http) = {
      post: "/v1/onboarding/complete"
      body: "*"
    };
  }

  rpc GetLinkedAccounts(google.protobuf.Empty) returns (LinkedAccountsResponse) {
    option (google.api.http) = {
      get: "/v1/profile/linked-accounts"
    };
  }
}
```

Add these new messages:
```protobuf
message UpdateProfileRequest {
  string first_name = 1;
  string last_name = 2;
  string username = 3;
  string bio = 4;
  string age_range = 5;  // NEW field
}

message GetProfileByUsernameRequest {
  string username = 1;
}

message CompleteOnboardingRequest {
  string first_name = 1;
  string last_name = 2;
  string age_range = 3;
  string how_heard = 4;
  string why_learn = 5;
  string level = 6;
}

message LinkedAccount {
  string provider = 1;
  string email = 2;
  string linked_at = 3;
}

message LinkedAccountsResponse {
  repeated LinkedAccount accounts = 1;
}
```

- [ ] **Step 2: Extend biz layer (biz/profile.go)**

```go
// Thêm vào UserRepoInterface
type UserRepoInterface interface {
    CreateUser(ctx context.Context, id, username, email, firstName, lastName string) (*gen.CoreUser, error)
    GetUser(ctx context.Context, id string) (*gen.CoreUser, error)
    GetUserByEmail(ctx context.Context, email string) (*gen.CoreUser, error)
    GetUserByUsername(ctx context.Context, username string) (*gen.CoreUser, error)            // NEW
    UpdateUser(ctx context.Context, id, firstName, lastName, username string, bio, avatarKey *string) (*gen.CoreUser, error)
    UpdateOnboardingInfo(ctx context.Context, id, firstName, lastName, ageRange string) (*gen.CoreUser, error)  // NEW
    UpsertUserOnboarding(ctx context.Context, userID, howHeard, whyLearn, level string) error                   // NEW
    UsernameExists(ctx context.Context, username string) (bool, error)
}

// Thêm UpdateProfileInput.AgeRange
type UpdateProfileInput struct {
    FirstName string
    LastName  string
    Username  string
    Bio       string
    AgeRange  string  // NEW
}

// Thêm OnboardingInput
type OnboardingInput struct {
    FirstName string
    LastName  string
    AgeRange  string
    HowHeard  string
    WhyLearn  string
    Level     string
}

// GetProfileByUsername
func (uc *ProfileUsecase) GetProfileByUsername(ctx context.Context, username string) (*gen.CoreUser, error) {
    user, err := uc.repo.GetUserByUsername(ctx, username)
    if err != nil {
        return nil, fmt.Errorf("%w: %w", ErrUserNotFound, err)
    }
    return user, nil
}

// CompleteOnboarding
func (uc *ProfileUsecase) CompleteOnboarding(ctx context.Context, userID string, input OnboardingInput) (*gen.CoreUser, error) {
    user, err := uc.repo.UpdateOnboardingInfo(ctx, userID, input.FirstName, input.LastName, input.AgeRange)
    if err != nil {
        return nil, fmt.Errorf("complete onboarding: %w", err)
    }

    if input.HowHeard != "" || input.WhyLearn != "" || input.Level != "" {
        if err := uc.repo.UpsertUserOnboarding(ctx, userID, input.HowHeard, input.WhyLearn, input.Level); err != nil {
            return nil, fmt.Errorf("save onboarding answers: %w", err)
        }
    }

    return user, nil
}
```

- [ ] **Step 3: Add data layer methods (user_repo.go)**

```go
// GetUserByUsername retrieves a user by username.
func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*gen.CoreUser, error) {
    row, err := r.q.GetUserByUsername(ctx, username)
    if err != nil {
        return nil, err
    }
    return &row, nil
}

// UpdateOnboardingInfo updates user's first_name, last_name, age_range and sets onboarding_completed=true.
func (r *UserRepo) UpdateOnboardingInfo(ctx context.Context, id, firstName, lastName, ageRange string) (*gen.CoreUser, error) {
    row, err := r.q.UpdateOnboardingInfo(ctx, gen.UpdateOnboardingInfoParams{
        ID:        id,
        FirstName: firstName,
        LastName:  lastName,
        AgeRange:  ageRange,
    })
    if err != nil {
        return nil, err
    }
    return &row, nil
}

// UpsertUserOnboarding inserts or updates onboarding answers.
func (r *UserRepo) UpsertUserOnboarding(ctx context.Context, userID, howHeard, whyLearn, level string) error {
    _, err := r.q.UpsertUserOnboarding(ctx, gen.UpsertUserOnboardingParams{
        UserID:   userID,
        HowHeard: howHeard,
        WhyLearn: whyLearn,
        Level:    level,
    })
    return err
}
```

- [ ] **Step 4: Add service layer handlers (service/profile.go)**

```go
// GetProfileByUsername returns a user's public profile by username.
func (s *ProfileService) GetProfileByUsername(ctx context.Context, req *pb.GetProfileByUsernameRequest) (*pb.User, error) {
    user, err := s.uc.GetProfileByUsername(ctx, req.Username)
    if err != nil {
        return nil, status.Error(codes.NotFound, "user not found")
    }

    // If user is logged in and it's their own profile, show email
    currentUserID, isLoggedIn := auth.UserIDFromContext(ctx)
    userProto := userToProto(user)
    if !isLoggedIn || currentUserID != user.ID {
        userProto.Email = "" // hide email for others
    }
    return userProto, nil
}

// CompleteOnboarding completes onboarding and saves profile + answers.
func (s *ProfileService) CompleteOnboarding(ctx context.Context, req *pb.CompleteOnboardingRequest) (*pb.User, error) {
    userID, ok := auth.UserIDFromContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "not authenticated")
    }

    user, err := s.uc.CompleteOnboarding(ctx, userID, biz.OnboardingInput{
        FirstName: req.FirstName,
        LastName:  req.LastName,
        AgeRange:  req.AgeRange,
        HowHeard:  req.HowHeard,
        WhyLearn:  req.WhyLearn,
        Level:     req.Level,
    })
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    return userToProto(user), nil
}

// GetLinkedAccounts returns linked third-party accounts.
func (s *ProfileService) GetLinkedAccounts(ctx context.Context, _ *emptypb.Empty) (*pb.LinkedAccountsResponse, error) {
    userID, ok := auth.UserIDFromContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "not authenticated")
    }

    accounts := fetchLinkedAccountsFromSupertokens(userID)
    return &pb.LinkedAccountsResponse{Accounts: accounts}, nil
}

// fetchLinkedAccountsFromSupertokens calls SuperTokens to get linked providers.
func fetchLinkedAccountsFromSupertokens(userID string) []*pb.LinkedAccount {
    user, err := supertokens.GetUser(userID)
    if err != nil || user == nil {
        return nil
    }
    var accounts []*pb.LinkedAccount
    for _, ru := range user.RecipeUsers {
        if ru.ThirdParty != nil {
            accounts = append(accounts, &pb.LinkedAccount{
                Provider:  ru.ThirdParty.ThirdPartyId,
                Email:     ru.Email,
                LinkedAt:  timestamppb.New(ru.TimeJoined),
            })
        }
    }
    return accounts
}
```

- [ ] **Step 5: Regenerate proto Go code and verify compilation**

```bash
# From app/core directory, run buf generate or protoc
cd D:\Github\puchidemy\puchi-backend\app\core
# Check if buf is available, or run protoc manually
# Then verify `go build ./...` compiles
```

- [ ] **Step 6: Commit**

```bash
git add app/core/api/profile/v1/profile.proto
git add app/core/internal/biz/profile.go
git add app/core/internal/data/user_repo.go
git add app/core/internal/service/profile.go
git commit -m "feat(core): add GetProfileByUsername, CompleteOnboarding, GetLinkedAccounts"
```
