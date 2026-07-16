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
	ErrSessionExpired     = errors.New("session expired")
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

// Logout revokes the current session.
func (uc *AuthUsecase) Logout(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) error {
	if err := uc.sessionRepo.Revoke(ctx, sessionID); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

// LogoutAllDevices revokes all sessions except the current one.
func (uc *AuthUsecase) LogoutAllDevices(ctx context.Context, userID uuid.UUID, currentSessionID uuid.UUID) error {
	sessions, err := uc.sessionRepo.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("list sessions: %w", err)
	}
	for _, s := range sessions {
		if s.ID != currentSessionID {
			if err := uc.sessionRepo.Revoke(ctx, s.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

// ChangePassword changes the user's password, verifying the old password first.
// Revokes all sessions except the current one as a security measure.
func (uc *AuthUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, oldPassword, newPassword string) error {
	if len(newPassword) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return ErrInvalidCredentials
	}

	if user.PasswordHash == nil {
		return ErrInvalidCredentials
	}
	ok, err := VerifyPassword(oldPassword, *user.PasswordHash)
	if err != nil || !ok {
		return ErrInvalidCredentials
	}

	hash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	if err := uc.userRepo.UpdatePassword(ctx, userID, hash); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	// Revoke all sessions except current
	sessions, err := uc.sessionRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil // don't leak session errors
	}
	for _, s := range sessions {
		if s.ID != sessionID {
			_ = uc.sessionRepo.Revoke(ctx, s.ID)
		}
	}

	return nil
}

// RequestPasswordReset generates a reset token and (in future) sends via NATS.
func (uc *AuthUsecase) RequestPasswordReset(ctx context.Context, email string) error {
	emailNorm := strings.ToLower(strings.TrimSpace(email))

	// Always return success (prevent email enumeration)
	user, err := uc.userRepo.GetByEmail(ctx, emailNorm)
	if err != nil {
		return nil
	}

	// Generate reset token
	token, _, err := uc.tokenUC.GenerateRefreshToken() // reuse random generator
	if err != nil {
		return fmt.Errorf("generate reset token: %w", err)
	}

	// Store token in DB (in Phase 3 this goes through outbox + NATS)
	// For now just return — email sending will be implemented in Phase 3
	_ = token
	_ = user
	_ = emailNorm

	return nil
}

// CompletePasswordReset completes the password reset flow.
// In Phase 3: look up token, verify not expired, update password, revoke all sessions.
func (uc *AuthUsecase) CompletePasswordReset(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	return fmt.Errorf("password reset not fully implemented")
}

func (uc *AuthUsecase) RefreshTokens(ctx context.Context, rawRefreshToken string) (*TokenPair, error) {
	hash := uc.tokenUC.HashRefreshToken(rawRefreshToken)

	session, err := uc.sessionRepo.GetByRefreshTokenHash(ctx, hash)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !session.IsActive {
		return nil, ErrInvalidCredentials
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	// Revoke old session first
	if err := uc.sessionRepo.Revoke(ctx, session.ID); err != nil {
		return nil, fmt.Errorf("revoke old session: %w", err)
	}

	// REUSE DETECTION: check if any active session still exists in this family
	// (besides the one we just revoked). If yes, this is a replay attack
	// using an older refresh token — revoke the entire family.
	hasActive, err := uc.sessionRepo.HasActiveInFamily(ctx, session.TokenFamily, session.ID)
	if err != nil {
		return nil, fmt.Errorf("check active family: %w", err)
	}
	if hasActive {
		// Replay attack detected — revoke entire family
		if err := uc.sessionRepo.RevokeFamily(ctx, session.TokenFamily); err != nil {
			return nil, fmt.Errorf("revoke family on reuse: %w", err)
		}
		return nil, ErrInvalidCredentials
	}

	newSession := &Session{
		UserID:      session.UserID,
		TokenFamily: session.TokenFamily,
		ChildNumber: session.ChildNumber + 1,
		ExpiresAt:   time.Now().Add(uc.tokenUC.RefreshTokenTTL()),
	}

	raw, hashNew, err := uc.tokenUC.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate new refresh token: %w", err)
	}
	newSession.RefreshTokenHash = hashNew

	user, err := uc.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	if err := uc.sessionRepo.Create(ctx, newSession); err != nil {
		return nil, fmt.Errorf("create new session: %w", err)
	}

	accessToken, err := uc.tokenUC.IssueAccessToken(AccessTokenClaims{
		UserID:        user.ID,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Roles:         []string{"user"},
		PermVersion:   1,
		SessionID:     newSession.ID,
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
