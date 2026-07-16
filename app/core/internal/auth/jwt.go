package auth

import (
	"sync"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTValidator validates JWT tokens using JWKS from Zitadel.
type JWTValidator struct {
	issuerURL string
	jwksURL   string
	jwks      *keyfunc.JWKS
	mu        sync.RWMutex
	lastFetch time.Time
	ttl       time.Duration
}

// NewJWTValidator creates a new JWTValidator with the given issuer and JWKS URL.
func NewJWTValidator(issuerURL, jwksURL string) (*JWTValidator, error) {
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{})
	if err != nil {
		return nil, err
	}
	return &JWTValidator{
		issuerURL: issuerURL,
		jwksURL:   jwksURL,
		jwks:      jwks,
		lastFetch: time.Now(),
		ttl:       15 * time.Minute,
	}, nil
}

// refreshJWKS re-fetches the JWKS if the TTL has expired.
func (v *JWTValidator) refreshJWKS() error {
	v.mu.RLock()
	age := time.Since(v.lastFetch)
	v.mu.RUnlock()

	if age < v.ttl {
		return nil
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	// Double-check after acquiring lock
	if time.Since(v.lastFetch) < v.ttl {
		return nil
	}

	jwks, err := keyfunc.Get(v.jwksURL, keyfunc.Options{})
	if err != nil {
		return err
	}
	v.jwks = jwks
	v.lastFetch = time.Now()
	return nil
}

// ParseAndValidate parses and validates a JWT token string.
// Returns the token claims if valid, or an error if invalid.
func (v *JWTValidator) ParseAndValidate(tokenStr string) (jwt.MapClaims, error) {
	if err := v.refreshJWKS(); err != nil {
		return nil, err
	}

	v.mu.RLock()
	jwks := v.jwks
	v.mu.RUnlock()

	token, err := jwt.Parse(tokenStr, jwks.Keyfunc)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	// Verify issuer
	if iss, ok := claims["iss"].(string); ok && v.issuerURL != "" && iss != v.issuerURL {
		return nil, jwt.ErrTokenInvalidIssuer
	}

	return claims, nil
}
