package auth

import "time"

// InitAuth creates a SessionValidator that introspects opaque Bearer tokens
// against the auth-service (Limen) GET /internal/session endpoint.
func InitAuth(authServiceURL string) *SessionValidator {
	return &SessionValidator{
		authServiceURL: authServiceURL,
		httpClient:     defaultHTTPClient,
		cacheTTL:       60 * time.Second,
	}
}
