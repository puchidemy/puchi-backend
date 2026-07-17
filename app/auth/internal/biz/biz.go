package biz

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/google/wire"
)

// OAuthStateData holds the temporary state stored in Valkey during the OAuth2 flow.
type OAuthStateData struct {
	Provider      string `json:"provider"`
	CodeVerifier  string `json:"code_verifier"`
	RedirectAfter string `json:"redirect_after,omitempty"`
}

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewAuthUsecase, NewTokenUsecase, NewSocialUsecase, NewMagicLinkUsecase, NewMFAUsecase, NewRBACUsecase, NewAuditUsecase, NewEmailVerificationUsecase)

// EventPublisher defines the interface for publishing domain events.
type EventPublisher interface {
	Publish(ctx context.Context, topic string, payload any) error
}

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
	IP          string
	UserAgent   string
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

// AuditLog represents an audit log entry.
type AuditLog struct {
	ID         uuid.UUID
	UserID     *uuid.UUID
	Action     string
	Resource   string
	ResourceID string
	IPAddress  string
	UserAgent  string
	Metadata   json.RawMessage // map[string]any
	CreatedAt  time.Time
}

// AuditRepo defines the data access interface for audit logs.
type AuditRepo interface {
	Create(ctx context.Context, log *AuditLog) error
}

// SessionCache defines the interface for caching session metadata.
type SessionCache interface {
	Set(ctx context.Context, sessionID string, data []byte, ttl time.Duration) error
	Get(ctx context.Context, sessionID string) ([]byte, error)
	Del(ctx context.Context, sessionID string) error
}

// TokenBlacklist defines the interface for blacklisting revoked JTIs.
type TokenBlacklist interface {
	BlacklistJTI(ctx context.Context, jti string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
}

// RateLimiter defines the interface for rate limiting actions.
type RateLimiter interface {
	Allow(ctx context.Context, key string, maxAttempts int, window time.Duration) (bool, error)
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

// PasswordResetToken represents a password reset token entity.
type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// PasswordResetTokenRepo defines the data access interface for password reset tokens.
type PasswordResetTokenRepo interface {
	Create(ctx context.Context, token *PasswordResetToken) error
	GetByToken(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
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

// Role represents an auth role entity.
type Role struct {
	ID          uuid.UUID
	Name        string
	Description string
	IsSystem    bool
	CreatedAt   time.Time
}

// Permission represents an auth permission entity.
type Permission struct {
	ID          uuid.UUID
	Name        string
	Resource    string
	Action      string
	Description string
	CreatedAt   time.Time
}

// RoleRepo defines the data access interface for roles.
type RoleRepo interface {
	GetByName(ctx context.Context, name string) (*Role, error)
	List(ctx context.Context) ([]*Role, error)
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*Permission, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*Role, error)
	AssignToUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID, grantedBy uuid.UUID) error
	RemoveFromUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error
}

// PermissionRepo defines the data access interface for permissions.
type PermissionRepo interface {
	List(ctx context.Context) ([]*Permission, error)
	IncrementVersion(ctx context.Context) error
	GetVersion(ctx context.Context) (int64, error)
}
