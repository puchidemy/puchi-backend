package biz

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/puchidemy/puchi-backend/app/core/internal/data/sqlc/gen"
)

// UpdateProfileInput — validated input from service layer
type UpdateProfileInput struct {
	FirstName string
	LastName  string
	Username  string
	Bio       string
	AgeRange  string
}

// OnboardingInput — input for completing onboarding
type OnboardingInput struct {
	FirstName string
	LastName  string
	AgeRange  string
	HowHeard  string
	WhyLearn  string
	Level     string
}

// UserRepoInterface — biz layer defines the repo contract (dependency inversion)
type UserRepoInterface interface {
	CreateUser(ctx context.Context, id, username, email, firstName, lastName string) (*gen.CoreUser, error)
	GetUser(ctx context.Context, id string) (*gen.CoreUser, error)
	GetUserByEmail(ctx context.Context, email string) (*gen.CoreUser, error)
	GetUserByUsername(ctx context.Context, username string) (*gen.CoreUser, error)
	UpdateUser(ctx context.Context, id, firstName, lastName, username string, bio, avatarKey *string) (*gen.CoreUser, error)
	UpdateOnboardingInfo(ctx context.Context, id, firstName, lastName, ageRange string) (*gen.CoreUser, error)
	UpsertUserOnboarding(ctx context.Context, userID, howHeard, whyLearn, level string) error
	UsernameExists(ctx context.Context, username string) (bool, error)
}

// Domain errors — business-level sentinel errors
var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUsernameTaken = errors.New("username already taken")
)

// ProfileUsecase handles user profile operations
type ProfileUsecase struct {
	repo UserRepoInterface
}

// NewProfileUsecase creates a new ProfileUsecase
func NewProfileUsecase(repo UserRepoInterface) *ProfileUsecase {
	return &ProfileUsecase{repo: repo}
}

// GetProfile returns user by ID
func (uc *ProfileUsecase) GetProfile(ctx context.Context, userID string) (*gen.CoreUser, error) {
	user, err := uc.repo.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUserNotFound, err)
	}
	return user, nil
}

// UpdateProfile validates and updates user profile
func (uc *ProfileUsecase) UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*gen.CoreUser, error) {
	exists, err := uc.repo.UsernameExists(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("check username: %w", err)
	}

	// Allow if username hasn't changed (same user's own username)
	current, err := uc.repo.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUserNotFound, err)
	}
	if exists && current.Username != input.Username {
		return nil, ErrUsernameTaken
	}

	var bioPtr *string
	if input.Bio != "" {
		bioPtr = &input.Bio
	}

	user, err := uc.repo.UpdateUser(ctx, userID, input.FirstName, input.LastName, input.Username, bioPtr, nil)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

// CreateUserFromAuth creates a new user during auth sync
func (uc *ProfileUsecase) CreateUserFromAuth(ctx context.Context, userID, email string) (*gen.CoreUser, error) {
	username := generateUsername(email)

	// Ensure uniqueness
	base := username
	for i := 1; ; i++ {
		exists, _ := uc.repo.UsernameExists(ctx, username)
		if !exists {
			break
		}
		username = fmt.Sprintf("%s%d", base, i)
	}

	user, err := uc.repo.CreateUser(ctx, userID, username, email, "", "")
	if err != nil {
		return nil, fmt.Errorf("create user from auth: %w", err)
	}
	return user, nil
}

// GetProfileByUsername returns a user's public profile by username.
func (uc *ProfileUsecase) GetProfileByUsername(ctx context.Context, username string) (*gen.CoreUser, error) {
	user, err := uc.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUserNotFound, err)
	}
	return user, nil
}

// CompleteOnboarding updates user's profile and saves onboarding answers.
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

func generateUsername(email string) string {
	localPart, _, _ := strings.Cut(email, "@")
	localPart = regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(strings.ToLower(localPart), "")
	if len(localPart) < 3 {
		localPart = "puchi_user"
	}
	return localPart
}
