package auth

import (
	"github.com/puchidemy/puchi-backend/app/core/internal/conf"
	authpkg "github.com/puchidemy/puchi-backend/pkg/auth"
)

// InitZitadel creates a JWTValidator from auth config.
// Call this in main.go before wireApp.
func InitZitadel(cfg *conf.Auth) (*JWTValidator, error) {
	return authpkg.NewJWTValidator(cfg.Zitadel.IssuerUrl, cfg.Zitadel.JwksUrl)
}
