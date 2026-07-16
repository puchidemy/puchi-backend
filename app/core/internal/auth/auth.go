package auth

import (
	"github.com/puchidemy/puchi-backend/app/core/internal/conf"
)

// InitZitadel creates a JWTValidator from auth config.
// Call this in main.go before wireApp.
func InitZitadel(cfg *conf.Auth) (*JWTValidator, error) {
	return NewJWTValidator(cfg.Zitadel.IssuerUrl, cfg.Zitadel.JwksUrl)
}
