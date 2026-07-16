package auth

// InitZitadel creates a JWTValidator from issuer and JWKS URLs.
// Call this in main.go before wireApp.
func InitZitadel(issuerURL, jwksURL string) (*JWTValidator, error) {
	return NewJWTValidator(issuerURL, jwksURL)
}
