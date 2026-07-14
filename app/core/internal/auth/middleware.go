package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/puchidemy/puchi-backend/app/core/internal/conf"

	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
)

// Middleware returns an HTTP middleware that verifies Supertokens sessions.
// Requests matching public_paths skip verification.
func Middleware(cfg *conf.Auth) func(http.Handler) http.Handler {
	public := make(map[string]bool, len(cfg.PublicPaths))
	for _, p := range cfg.PublicPaths {
		public[p] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if public[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			wrapper := newResponseWriter(w)

			sessionRequired := true
			sess, err := session.GetSession(r, wrapper, &sessmodels.VerifySessionOptions{
				SessionRequired: &sessionRequired,
			})

			if err != nil {
				writeUnauthorized(w, err)
				return
			}

			if sess == nil {
				writeUnauthorized(w, nil)
				return
			}

			ctx := NewContextWithUserID(r.Context(), sess.GetUserID())
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

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}
