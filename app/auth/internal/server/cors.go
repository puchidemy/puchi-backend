package server

import (
	"net/http"
	"strings"
)

// CORSOptions holds configuration for the CORS middleware.
type CORSOptions struct {
	AllowedOrigins   []string
	AllowCredentials bool
}

// CORSHandler wraps an http.Handler with CORS headers.
// Preflight (OPTIONS) requests are handled directly.
func CORSHandler(next http.Handler, opts CORSOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Set CORS headers for allowed origins
		if origin != "" && isOriginAllowed(origin, opts.AllowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")

			if opts.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isOriginAllowed checks if an origin matches any pattern in the allowed list.
func isOriginAllowed(origin string, allowed []string) bool {
	if len(allowed) == 0 {
		return false
	}
	for _, a := range allowed {
		if a == "*" || strings.EqualFold(a, origin) {
			return true
		}
	}
	return false
}
