package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/puchidemy/puchi-backend/pkg/apierr"
)

// MiddlewareConfig holds configuration for the auth middleware.
type MiddlewareConfig struct {
	PublicPaths []string
	// IsPublic optionally marks request-specific public routes that cannot be
	// expressed as PublicPaths prefixes (e.g. GET /v1/profile/{username}).
	IsPublic  func(*http.Request) bool
	Validator *SessionValidator
}

// Middleware returns an HTTP middleware that verifies opaque Limen session
// tokens via auth-service introspect. Accepts Authorization: Bearer or the
// limen_session cookie (same opaque token). PublicPaths / IsPublic skip
// required auth; if a Bearer/cookie is present on a public route, the session
// is still validated and user context is populated (optional auth).
func Middleware(cfg MiddlewareConfig) func(http.Handler) http.Handler {
	public := make([]string, len(cfg.PublicPaths))
	copy(public, cfg.PublicPaths)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isPublic := false
			for _, p := range public {
				if strings.HasPrefix(r.URL.Path, p) {
					isPublic = true
					break
				}
			}
			if !isPublic && cfg.IsPublic != nil && cfg.IsPublic(r) {
				isPublic = true
			}

			tokenStr := SessionTokenFromRequest(r)
			if isPublic {
				if tokenStr != "" {
					info, err := cfg.Validator.ParseAndValidate(r.Context(), tokenStr)
					if err == nil {
						ctx := r.Context()
						ctx = NewContextWithUserID(ctx, info.UserID)
						ctx = NewContextWithEmail(ctx, info.Email)
						ctx = NewContextWithRoles(ctx, info.Roles)
						r = r.WithContext(ctx)
					}
					// Invalid token on public route: continue unauthenticated.
				}
				next.ServeHTTP(w, r)
				return
			}

			if tokenStr == "" {
				writeUnauthorized(w, ErrNoSession)
				return
			}

			info, err := cfg.Validator.ParseAndValidate(r.Context(), tokenStr)
			if err != nil {
				writeUnauthorized(w, err)
				return
			}

			ctx := r.Context()
			ctx = NewContextWithUserID(ctx, info.UserID)
			ctx = NewContextWithEmail(ctx, info.Email)
			ctx = NewContextWithRoles(ctx, info.Roles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter, err error) {
	reason := "NO_SESSION"
	if errors.Is(err, ErrSessionExpired) {
		reason = "SESSION_EXPIRED"
	}
	apierr.Unauthorized(w, reason)
}
