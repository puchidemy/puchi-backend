package auth

import (
	"net/http"
	"strings"
)

// DefaultLimenSessionCookie is Limen's default session cookie name.
// The cookie value is the same opaque token accepted as Authorization: Bearer.
const DefaultLimenSessionCookie = "limen_session"

// SessionTokenFromRequest returns the opaque Limen session token from the
// Authorization Bearer header, or falls back to the limen_session cookie.
// Browser clients often send only the HttpOnly cookie (JS cannot read it to
// set Authorization); Go services must accept both.
func SessionTokenFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		token := strings.TrimSpace(authHeader[len("Bearer "):])
		if token != "" {
			return token
		}
	}
	if c, err := r.Cookie(DefaultLimenSessionCookie); err == nil && c != nil {
		if token := strings.TrimSpace(c.Value); token != "" {
			return token
		}
	}
	return ""
}
