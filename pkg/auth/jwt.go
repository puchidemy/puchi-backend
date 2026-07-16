package auth

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrNoSession      = errors.New("no session found")
	ErrSessionExpired = errors.New("session expired")
)

// JWTValidator validates JWT tokens using JWKS from the auth-service.
type JWTValidator struct {
	jwksURL   string
	issuer    string
	cacheTTL  time.Duration
	mu        sync.RWMutex
	jwks      *keyfunc.JWKS
	lastFetch time.Time
}

// Claims represents the JWT claims issued by the auth-service.
type Claims struct {
	UserID        string   `json:"sub"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Roles         []string `json:"roles"`
	PermVersion   int64    `json:"perm_version"`
	SessionID     string   `json:"sid"`
}

// ParseAndValidate parses and validates a JWT token string.
// It fetches the JWKS from the auth-service (with caching), validates
// issuer and expiration (30s leeway), and returns typed Claims.
func (v *JWTValidator) ParseAndValidate(ctx context.Context, tokenStr string) (*Claims, error) {
	jwks, err := v.getJWKS(ctx)
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenStr, jwks.Keyfunc,
		jwt.WithIssuer(v.issuer),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrSessionExpired
		}
		return nil, ErrNoSession
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrNoSession
	}

	c := &Claims{}

	if sub, err := mapClaims.GetSubject(); err == nil {
		c.UserID = sub
	}
	if e, ok := mapClaims["email"].(string); ok {
		c.Email = e
	}
	if ev, ok := mapClaims["email_verified"].(bool); ok {
		c.EmailVerified = ev
	}
	if r, ok := mapClaims["roles"].([]any); ok {
		for _, v := range r {
			if s, ok := v.(string); ok {
				c.Roles = append(c.Roles, s)
			}
		}
	}
	if pv, ok := mapClaims["perm_version"].(float64); ok {
		c.PermVersion = int64(pv)
	}
	if sid, ok := mapClaims["sid"].(string); ok {
		c.SessionID = sid
	}

	return c, nil
}

// getJWKS returns the JWKS, fetching from the auth-service with double-checked
// locking cache if the TTL has expired.
func (v *JWTValidator) getJWKS(ctx context.Context) (*keyfunc.JWKS, error) {
	v.mu.RLock()
	if v.jwks != nil && time.Since(v.lastFetch) < v.cacheTTL {
		jwks := v.jwks
		v.mu.RUnlock()
		return jwks, nil
	}
	v.mu.RUnlock()

	v.mu.Lock()
	defer v.mu.Unlock()

	if v.jwks != nil && time.Since(v.lastFetch) < v.cacheTTL {
		return v.jwks, nil
	}

	jwks, err := keyfunc.Get(v.jwksURL, keyfunc.Options{RefreshInterval: v.cacheTTL})
	if err != nil {
		return nil, err
	}

	v.jwks = jwks
	v.lastFetch = time.Now()
	return jwks, nil
}
