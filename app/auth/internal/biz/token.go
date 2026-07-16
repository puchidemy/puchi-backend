package biz

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Error sentinels for token operations.
var (
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token invalid")
)

// TokenConfig holds configuration for token issuance and verification.
type TokenConfig struct {
	PrivateKey      *rsa.PrivateKey
	PublicKey       *rsa.PublicKey
	KeyID           string
	Issuer          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// AccessTokenClaims represents the claims embedded in an access token.
type AccessTokenClaims struct {
	UserID        uuid.UUID `json:"sub"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	Roles         []string  `json:"roles"`
	PermVersion   int64     `json:"perm_version"`
	SessionID     uuid.UUID `json:"sid"`
}

// TokenUsecase handles JWT token issuance and verification.
type TokenUsecase struct {
	cfg    TokenConfig
	method jwt.SigningMethod
}

// NewTokenUsecase creates a new TokenUsecase with the given configuration.
func NewTokenUsecase(cfg TokenConfig) *TokenUsecase {
	return &TokenUsecase{
		cfg:    cfg,
		method: jwt.SigningMethodRS256,
	}
}

// IssueAccessToken signs a new RS256 JWT with the provided claims.
func (uc *TokenUsecase) IssueAccessToken(claims AccessTokenClaims) (string, error) {
	jti, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate jti: %w", err)
	}

	now := time.Now()
	token := jwt.NewWithClaims(uc.method, jwt.MapClaims{
		"iss":            uc.cfg.Issuer,
		"sub":            claims.UserID.String(),
		"email":          claims.Email,
		"email_verified": claims.EmailVerified,
		"roles":          claims.Roles,
		"perm_version":   claims.PermVersion,
		"jti":            jti.String(),
		"sid":            claims.SessionID.String(),
		"iat":            jwt.NewNumericDate(now),
		"nbf":            jwt.NewNumericDate(now.Add(-30 * time.Second)),
		"exp":            jwt.NewNumericDate(now.Add(uc.cfg.AccessTokenTTL)),
	})
	token.Header["kid"] = uc.cfg.KeyID

	return token.SignedString(uc.cfg.PrivateKey)
}

// VerifyAccessToken parses and validates a JWT access token, returning the claims.
func (uc *TokenUsecase) VerifyAccessToken(tokenStr string) (*AccessTokenClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return uc.cfg.PublicKey, nil
	},
		jwt.WithIssuer(uc.cfg.Issuer),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	sub, err := claims.GetSubject()
	if err != nil {
		return nil, fmt.Errorf("invalid sub claim: %w", err)
	}
	userID, err := uuid.Parse(sub)
	if err != nil {
		return nil, fmt.Errorf("invalid sub claim: %w", err)
	}

	email, _ := claims["email"].(string)
	emailVerified, _ := claims["email_verified"].(bool)

	roles := []string{}
	if r, ok := claims["roles"].([]any); ok {
		for _, v := range r {
			if s, ok := v.(string); ok {
				roles = append(roles, s)
			}
		}
	}

	permVersion := int64(0)
	if pv, ok := claims["perm_version"].(float64); ok {
		permVersion = int64(pv)
	}

	sidStr, _ := claims["sid"].(string)
	sid, _ := uuid.Parse(sidStr)

	return &AccessTokenClaims{
		UserID:        userID,
		Email:         email,
		EmailVerified: emailVerified,
		Roles:         roles,
		PermVersion:   permVersion,
		SessionID:     sid,
	}, nil
}

// GenerateRefreshToken creates a cryptographically random refresh token.
// Returns the raw base64url-encoded token and its SHA-256 hash (also base64url-encoded).
func (uc *TokenUsecase) GenerateRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(b)

	h := sha256.Sum256([]byte(raw))
	hash = base64.RawURLEncoding.EncodeToString(h[:])

	return raw, hash, nil
}

// HashRefreshToken returns the SHA-256 hash of a raw refresh token as base64url.
func (uc *TokenUsecase) HashRefreshToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// PublicKey returns the RSA public key used for token verification.
func (uc *TokenUsecase) PublicKey() *rsa.PublicKey {
	return uc.cfg.PublicKey
}

// KeyID returns the key ID configured for JWT headers.
func (uc *TokenUsecase) KeyID() string {
	return uc.cfg.KeyID
}
