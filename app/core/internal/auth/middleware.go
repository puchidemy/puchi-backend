package auth

import (
	"net/http"

	"github.com/puchidemy/puchi-backend/app/core/internal/conf"
	authpkg "github.com/puchidemy/puchi-backend/pkg/auth"
)

// Middleware returns an HTTP middleware that verifies Zitadel JWTs from the
// Authorization header. Requests matching public_paths skip verification.
// If syncer is non-nil, it performs lazy user creation after JWT verification.
func Middleware(cfg *conf.Auth, jwtValidator *JWTValidator, syncer *UserSyncer) func(http.Handler) http.Handler {
	return authpkg.Middleware(cfg.PublicPaths, jwtValidator, syncer)
}
