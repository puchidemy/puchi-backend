package biz

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MagicLinkUsecase handles magic link (passwordless email login) business logic.
type MagicLinkUsecase struct {
	mlRepo      MagicLinkRepo
	userRepo    UserRepo
	session     SessionRepo
	tokenUC     *TokenUsecase
	publisher   EventPublisher
	frontendURL string
}

// NewMagicLinkUsecase creates a new MagicLinkUsecase.
func NewMagicLinkUsecase(mlRepo MagicLinkRepo, userRepo UserRepo, session SessionRepo, tokenUC *TokenUsecase, publisher EventPublisher, frontendURL string) *MagicLinkUsecase {
	return &MagicLinkUsecase{mlRepo: mlRepo, userRepo: userRepo, session: session, tokenUC: tokenUC, publisher: publisher, frontendURL: frontendURL}
}

// Send creates a magic link token and returns the raw token (email sending deferred).
func (uc *MagicLinkUsecase) Send(ctx context.Context, email string, redirectTo string) (string, error) {
	emailNorm := strings.ToLower(strings.TrimSpace(email))

	// Check if user exists
	var userID uuid.UUID
	user, err := uc.userRepo.GetByEmail(ctx, emailNorm)
	if err == nil {
		userID = user.ID
	}

	// Generate token (store hash in DB, return raw to caller)
	raw, hash, err := uc.tokenUC.GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate magic link token: %w", err)
	}

	ml := &MagicLink{
		Email:      emailNorm,
		UserID:     userID,
		Token:      hash,
		RedirectTo: redirectTo,
	}

	if err := uc.mlRepo.Create(ctx, ml); err != nil {
		return "", fmt.Errorf("create magic link: %w", err)
	}

	// Publish email.send event
	if err := uc.publisher.Publish(ctx, "email.send", map[string]any{
		"to":       emailNorm,
		"template": "magic-link",
		"data": map[string]any{
			"link": fmt.Sprintf("%s/auth/magic-link/verify?token=%s", uc.frontendURL, raw),
		},
	}); err != nil {
		// Log but don't fail — the outbox will retry
		slog.Warn("publish magic link email", "error", err)
	}

	return raw, nil
}

// Verify validates a magic link token and logs the user in.
func (uc *MagicLinkUsecase) Verify(ctx context.Context, token string) (*TokenPair, error) {
	tokenHash := uc.tokenUC.HashRefreshToken(token)
	ml, err := uc.mlRepo.GetByToken(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	var user *User

	if ml.UserID == uuid.Nil {
		// New user — auto-create account
		displayName := strings.Split(ml.Email, "@")[0]
		newUser := &User{
			Email:           ml.Email,
			EmailNormalized: ml.Email,
			EmailVerified:   true, // verified by email delivery
			PasswordHash:    nil,
			DisplayName:     displayName,
			Locale:          "vi",
		}
		if err := uc.userRepo.Create(ctx, newUser); err != nil {
			return nil, fmt.Errorf("create user via magic link: %w", err)
		}
		user = newUser
	} else {
		// Existing user
		user, err = uc.userRepo.GetByID(ctx, ml.UserID)
		if err != nil {
			return nil, ErrInvalidCredentials
		}
		if !user.IsActive {
			return nil, ErrUserNotActive
		}
	}

	// Create session + tokens
	family := uuid.New()
	session := &Session{
		UserID:      user.ID,
		TokenFamily: family,
		ChildNumber: 1,
		ExpiresAt:   time.Now().Add(uc.tokenUC.RefreshTokenTTL()),
	}

	raw, hash, err := uc.tokenUC.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	session.RefreshTokenHash = hash

	if err := uc.session.Create(ctx, session); err != nil {
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

	// Mark used (prevent replay) — AFTER user + session creation succeed
	if err := uc.mlRepo.MarkUsed(ctx, ml.ID); err != nil {
		return nil, fmt.Errorf("mark magic link used: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: raw,
		ExpiresIn:    int64(uc.tokenUC.AccessTokenTTL().Seconds()),
		User:         *user,
	}, nil
}
