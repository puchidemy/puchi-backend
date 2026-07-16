package auth

import authpkg "github.com/puchidemy/puchi-backend/pkg/auth"

// JWTValidator validates JWT tokens using JWKS from Zitadel.
type JWTValidator = authpkg.JWTValidator

// NewJWTValidator creates a new JWTValidator with the given issuer and JWKS URL.
func NewJWTValidator(issuerURL, jwksURL string) (*JWTValidator, error) {
	return authpkg.NewJWTValidator(issuerURL, jwksURL)
}
