package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Sentinel errors for social auth operations.
var (
	ErrSocialAlreadyLinked = errors.New("this social account is already linked to another user")
	ErrCannotUnlinkLast    = errors.New("cannot unlink the only authentication method")
	ErrInvalidOAuthState   = errors.New("invalid OAuth state")
)

// ProviderUserInfo is the normalized user info returned by an OAuth2 provider
// after a successful code exchange.
type ProviderUserInfo struct {
	ProviderID    string
	Email         string
	EmailVerified bool
	Name          string
	AvatarURL     string
}

// SocialUsecase handles social login: OAuth2 callback processing, linking,
// and unlinking social connections to user accounts.
type SocialUsecase struct {
	userRepo       UserRepo
	socialConnRepo SocialConnectionRepo
	sessionRepo    SessionRepo
	tokenUC        *TokenUsecase
	providers      map[string]interface{} // OAuth2Provider instances (wired in Phase 3)
	eventPublisher EventPublisher
}

// NewSocialUsecase creates a new SocialUsecase.
func NewSocialUsecase(
	userRepo UserRepo,
	socialConnRepo SocialConnectionRepo,
	sessionRepo SessionRepo,
	tokenUC *TokenUsecase,
	eventPublisher EventPublisher,
) *SocialUsecase {
	return &SocialUsecase{
		userRepo:       userRepo,
		socialConnRepo: socialConnRepo,
		sessionRepo:    sessionRepo,
		tokenUC:        tokenUC,
		providers:      make(map[string]interface{}),
		eventPublisher: eventPublisher,
	}
}

// LoginOrLink handles the OAuth2 callback for a social login provider.
// If the social connection already exists, it logs the user in and returns tokens.
// If a user with the same email exists, it links the social connection and logs in.
// Otherwise, it creates a new user and links the social connection.
func (uc *SocialUsecase) LoginOrLink(ctx context.Context, providerName string, providerUser ProviderUserInfo) (*TokenPair, error) {
	// Check if this provider account is already connected to a user
	conn, err := uc.socialConnRepo.GetByProvider(ctx, providerName, providerUser.ProviderID)
	if err == nil && conn != nil {
		// Existing connection — login
		user, err := uc.userRepo.GetByID(ctx, conn.UserID)
		if err != nil {
			return nil, ErrInvalidCredentials
		}
		if !user.IsActive {
			return nil, ErrUserNotActive
		}
		return uc.createTokensForUser(ctx, user)
	}

	// Check if a user already exists with this email
	emailNorm := normalizeEmail(providerUser.Email)
	user, err := uc.userRepo.GetByEmail(ctx, emailNorm)

	if err == nil {
		// Existing user with same email — link social connection and login
		conn := &SocialConnection{
			UserID:         user.ID,
			Provider:       providerName,
			ProviderUserID: providerUser.ProviderID,
			ProviderEmail:  providerUser.Email,
			AvatarURL:      providerUser.AvatarURL,
		}
		if err := uc.socialConnRepo.Create(ctx, conn); err != nil {
			return nil, fmt.Errorf("link social connection: %w", err)
		}

		_ = uc.eventPublisher.Publish(ctx, "auth.social.linked", map[string]any{
			"user_id":        user.ID.String(),
			"provider":       providerName,
			"provider_user":  providerUser.ProviderID,
		})

		return uc.createTokensForUser(ctx, user)
	}

	// New user — create account with social connection
	displayName := providerUser.Name
	if displayName == "" {
		// Fallback: providerName + first 8 chars of provider ID
		shortID := providerUser.ProviderID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		displayName = providerName + "_" + shortID
	}

	newUser := &User{
		Email:           providerUser.Email,
		EmailNormalized: emailNorm,
		EmailVerified:   providerUser.EmailVerified,
		PasswordHash:    nil, // social-only account, no password
		DisplayName:     displayName,
		Locale:          "vi",
	}
	if err := uc.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Create the social connection
	conn = &SocialConnection{
		UserID:         newUser.ID,
		Provider:       providerName,
		ProviderUserID: providerUser.ProviderID,
		ProviderEmail:  providerUser.Email,
		AvatarURL:      providerUser.AvatarURL,
	}
	if err := uc.socialConnRepo.Create(ctx, conn); err != nil {
		return nil, fmt.Errorf("create social connection: %w", err)
	}

	_ = uc.eventPublisher.Publish(ctx, "auth.social.linked", map[string]any{
		"user_id":        newUser.ID.String(),
		"provider":       providerName,
		"provider_user":  providerUser.ProviderID,
	})

	return uc.createTokensForUser(ctx, newUser)
}

// LinkToExistingUser adds a social connection to an already-authenticated user.
// Returns ErrSocialAlreadyLinked if the social account is already linked to another user.
func (uc *SocialUsecase) LinkToExistingUser(ctx context.Context, userID uuid.UUID, providerName string, providerUser ProviderUserInfo) error {
	existing, err := uc.socialConnRepo.GetByProvider(ctx, providerName, providerUser.ProviderID)
	if err == nil && existing != nil {
		if existing.UserID != userID {
			return ErrSocialAlreadyLinked
		}
		return nil // already linked to this user
	}

	conn := &SocialConnection{
		UserID:         userID,
		Provider:       providerName,
		ProviderUserID: providerUser.ProviderID,
		ProviderEmail:  providerUser.Email,
		AvatarURL:      providerUser.AvatarURL,
	}
	return uc.socialConnRepo.Create(ctx, conn)
}

// Unlink removes a social connection from a user.
// Returns ErrCannotUnlinkLast if this would leave the user with no authentication method.
func (uc *SocialUsecase) Unlink(ctx context.Context, userID uuid.UUID, connectionID uuid.UUID) error {
	conns, err := uc.socialConnRepo.ListByUser(ctx, userID)
	if err != nil {
		return err
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	hasPassword := user.PasswordHash != nil
	socialCount := len(conns)

	if !hasPassword && socialCount <= 1 {
		return ErrCannotUnlinkLast
	}

	return uc.socialConnRepo.Delete(ctx, connectionID, userID)
}

// ListConnections returns all social connections for a user.
func (uc *SocialUsecase) ListConnections(ctx context.Context, userID uuid.UUID) ([]*SocialConnection, error) {
	return uc.socialConnRepo.ListByUser(ctx, userID)
}

// createTokensForUser creates a session and issues access+refresh tokens for the given user.
func (uc *SocialUsecase) createTokensForUser(ctx context.Context, user *User) (*TokenPair, error) {
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
