package auth

import "time"

// InitAuth creates a JWTValidator configured to validate tokens from
// a self-hosted auth-service. The validator fetches JWKS from
// {authServiceURL}/.well-known/jwks.json and uses authServiceURL
// as the expected issuer.
func InitAuth(authServiceURL string) *JWTValidator {
	return &JWTValidator{
		jwksURL:  authServiceURL + "/.well-known/jwks.json",
		issuer:   authServiceURL,
		cacheTTL: 15 * time.Minute,
	}
}
