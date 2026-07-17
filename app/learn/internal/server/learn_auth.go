package server

import (
	"net/http"
	"strings"

	"github.com/puchidemy/puchi-backend/pkg/apierr"
	authpkg "github.com/puchidemy/puchi-backend/pkg/auth"
)

// learnOptionalAuthFilter validates Limen session (Bearer or limen_session
// cookie) on curriculum read paths when present, injecting user context
// without requiring auth for guest-only cookies.
func learnOptionalAuthFilter(validator *authpkg.SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isLearnGuestOrUserPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			tokenStr := authpkg.SessionTokenFromRequest(r)
			if tokenStr == "" {
				// Anonymous / guest cookie path
				next.ServeHTTP(w, r)
				return
			}

			info, err := validator.ParseAndValidate(r.Context(), tokenStr)
			if err != nil {
				// Stale Bearer must not block guest trial reads — fall through as guest.
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			ctx = authpkg.NewContextWithUserID(ctx, info.UserID)
			ctx = authpkg.NewContextWithEmail(ctx, info.Email)
			ctx = authpkg.NewContextWithRoles(ctx, info.Roles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func isLearnGuestOrUserPath(path string) bool {
	return strings.HasPrefix(path, "/v1/learn/units/") ||
		strings.HasPrefix(path, "/v1/learn/lessons/") ||
		strings.HasPrefix(path, "/v1/learn/attempts/")
}

func writeLearnUnauthorized(w http.ResponseWriter) {
	apierr.Unauthorized(w, "NO_SESSION")
}
