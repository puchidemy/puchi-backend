package biz

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Sentinel errors for auth operations.
var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotActive      = errors.New("user account is not active")
	ErrPasswordTooShort   = errors.New("password too short")
)

// AuthUsecase handles authentication business logic: register, login.
type AuthUsecase struct {
	userRepo    UserRepo
	sessionRepo SessionRepo
	tokenUC     *TokenUsecase
}

// NewAuthUsecase creates a new AuthUsecase.
func NewAuthUsecase(userRepo UserRepo, sessionRepo SessionRepo, tokenUC *TokenUsecase) *AuthUsecase {
	return &AuthUsecase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		tokenUC:     tokenUC,
	}
}

// Register creates a new user account with email and password.
func (uc *AuthUsecase) Register(ctx context.Context, input RegisterInput) (*User, error) {
	if len(input.Password) < MinPasswordLength {
		return nil, ErrPasswordTooShort
	}

	emailNorm := strings.ToLower(strings.TrimSpace(input.Email))

	hash, err := HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	displayName := input.DisplayName
	if displayName == "" {
		displayName = strings.Split(input.Email, "@")[0]
	}

	user := &User{
		Email:           input.Email,
		EmailNormalized: emailNorm,
		PasswordHash:    &hash,
		DisplayName:     displayName,
		Locale:          "vi",
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user with email and password, returning a token pair.
func (uc *AuthUsecase) Login(ctx context.Context, input LoginInput) (*TokenPair, error) {
	emailNorm := strings.ToLower(strings.TrimSpace(input.Email))

	user, err := uc.userRepo.GetByEmail(ctx, emailNorm)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrUserNotActive
	}
	if user.PasswordHash == nil {
		return nil, ErrInvalidCredentials
	}

	ok, err := VerifyPassword(input.Password, *user.PasswordHash)
	if err != nil || !ok {
		return nil, ErrInvalidCredentials
	}

	family := uuid.New()
	session := &Session{
		UserID:      user.ID,
		TokenFamily: family,
		ChildNumber: 1,
		IPAddress:   input.IP,
		UserAgent:   input.UserAgent,
		ExpiresAt:   time.Now().Add(uc.tokenUC.RefreshTokenTTL()),
	}

	raw, hash, err := uc.tokenUC.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	session.RefreshTokenHash = hash

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	_ = uc.userRepo.UpdateLastLogin(ctx, user.ID)

	accessToken, err := uc.tokenUC.IssueAccessToken(AccessTokenClaims{
		UserID:        user.ID,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Roles:         []string{"user"},
		PermVersion:   1,
		SessionID:     session.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: raw,
		ExpiresIn:    int64(uc.tokenUC.AccessTokenTTL().Seconds()),
		User:         *user,
	}, nil
}
