package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

// Middleware returns an HTTP middleware that verifies Zitadel JWTs from the
// Authorization header. Requests matching publicPaths skip verification.
// If syncer is non-nil, it performs lazy user creation after JWT verification.
func Middleware(publicPaths []string, jwtValidator *JWTValidator, syncer *UserSyncer) func(http.Handler) http.Handler {
	public := make([]string, len(publicPaths))
	copy(public, publicPaths)

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
				writeUnauthorized(w, nil)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == authHeader {
				writeUnauthorized(w, nil)
				return
			}

			claims, err := jwtValidator.ParseAndValidate(tokenStr)
			if err != nil {
				writeUnauthorized(w, err)
				return
			}

			userID, ok := claims["sub"].(string)
			if !ok || userID == "" {
				writeUnauthorized(w, nil)
				return
			}

			// Extract email from JWT claims if available
			email, _ := claims["email"].(string)

			ctx := NewContextWithUserID(r.Context(), userID)

			// Lazy creation: ensure user exists in DB after JWT verification
			if syncer != nil {
				if err := syncer.EnsureUserExists(ctx, userID, email); err != nil {
					slog.Warn("auth sync: failed to ensure user exists",
						"user_id", userID,
						"error", err,
					)
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	reason := "NO_SESSION"
	if err != nil && strings.Contains(err.Error(), "expired") {
		reason = "SESSION_EXPIRED"
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    401,
		"message": "unauthorized",
		"reason":  reason,
	})
}
