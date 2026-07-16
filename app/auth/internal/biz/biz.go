package biz

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewAuthUsecase, NewTokenUsecase, NewSocialUsecase, NewMagicLinkUsecase, NewMFAUsecase)

const MinPasswordLength = 8

// User represents an auth user entity.
type User struct {
	ID              uuid.UUID
	Email           string
	EmailNormalized string
	EmailVerified   bool
	PasswordHash    *string
	DisplayName     string
	Locale          string
	IsActive        bool
	IsSuperAdmin    bool
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// UserRepo defines the data access interface for users.
type UserRepo interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, emailNormalized string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	SetEmailVerified(ctx context.Context, id uuid.UUID) error
	SetActive(ctx context.Context, id uuid.UUID, active bool) error
}

// RegisterInput holds the registration request data.
type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

// LoginInput holds the login request data.
type LoginInput struct {
	Email     string
	Password  string
	IP        string
	UserAgent string
}

// TokenPair contains the access and refresh tokens returned on login.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	User         User
}

// Session represents an authenticated user session / refresh token record.
type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	TokenFamily      uuid.UUID
	ChildNumber      int
	IPAddress        string
	UserAgent        string
	DeviceName       string
	DeviceType       string
	OS               string
	Location         string
	IsActive         bool
	ExpiresAt        time.Time
	LastUsedAt       time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
}

// SocialConnection represents an OAuth2 social login connection.
type SocialConnection struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Provider       string
	ProviderUserID string
	ProviderEmail  string
	AvatarURL      string
	LinkedAt       time.Time
}

// SocialConnectionRepo defines the data access interface for social connections.
type SocialConnectionRepo interface {
	Create(ctx context.Context, conn *SocialConnection) error
	GetByProvider(ctx context.Context, provider, providerUserID string) (*SocialConnection, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*SocialConnection, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

// MagicLink represents a passwordless login link.
type MagicLink struct {
	ID         uuid.UUID
	Email      string
	UserID     uuid.UUID // nil if user doesn't exist yet
	Token      string
	RedirectTo string
	ExpiresAt  time.Time
	UsedAt     *time.Time
	CreatedAt  time.Time
}

// MagicLinkRepo defines the data access interface for magic links.
type MagicLinkRepo interface {
	Create(ctx context.Context, ml *MagicLink) error
	GetByToken(ctx context.Context, token string) (*MagicLink, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
}

// TOTPSecret represents an encrypted TOTP secret with recovery codes.
type TOTPSecret struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	EncryptedSecret string
	EncryptedCodes  string
	IsEnabled       bool
	VerifiedAt      *time.Time
	LastUsedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TOTPRepo defines the data access interface for TOTP secrets.
type TOTPRepo interface {
	Upsert(ctx context.Context, totp *TOTPSecret) error
	GetByUser(ctx context.Context, userID uuid.UUID) (*TOTPSecret, error)
	Enable(ctx context.Context, userID uuid.UUID) error
	Disable(ctx context.Context, userID uuid.UUID) error
}

// SessionRepo defines the data access interface for sessions.
type SessionRepo interface {
	Create(ctx context.Context, session *Session) error
	GetByRefreshTokenHash(ctx context.Context, hash string) (*Session, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeFamily(ctx context.Context, family uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Session, error)
	HasActiveInFamily(ctx context.Context, family uuid.UUID, excludeSessionID uuid.UUID) (bool, error)
}
