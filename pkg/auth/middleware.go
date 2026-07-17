package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// MiddlewareConfig holds configuration for the auth middleware.
type MiddlewareConfig struct {
	PublicPaths []string
	Validator   *JWTValidator
}

// Middleware returns an HTTP middleware that verifies JWT tokens from the
// Authorization header. Requests matching PublicPaths skip verification.
// On success, the user ID and roles are injected into the request context.
func Middleware(cfg MiddlewareConfig) func(http.Handler) http.Handler {
	public := make([]string, len(cfg.PublicPaths))
	copy(public, cfg.PublicPaths)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, p := range public {
				if strings.HasPrefix(r.URL.Path, p) {
					next.ServeHTTP(w, r)
					return
				}
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeUnauthorized(w, ErrNoSession)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == authHeader {
				writeUnauthorized(w, ErrNoSession)
				return
			}

			claims, err := cfg.Validator.ParseAndValidate(r.Context(), tokenStr)
			if err != nil {
				writeUnauthorized(w, err)
				return
			}

			ctx := r.Context()
			ctx = NewContextWithUserID(ctx, claims.UserID)
			ctx = NewContextWithEmail(ctx, claims.Email)
			ctx = NewContextWithRoles(ctx, claims.Roles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	reason := "NO_SESSION"
	if errors.Is(err, ErrSessionExpired) {
		reason = "SESSION_EXPIRED"
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    401,
		"message": "unauthorized",
		"reason":  reason,
	})
}
