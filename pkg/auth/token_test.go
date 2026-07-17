package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionTokenFromRequest_Bearer(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer tok-abc")
	if got := SessionTokenFromRequest(r); got != "tok-abc" {
		t.Fatalf("got %q", got)
	}
}

func TestSessionTokenFromRequest_CookieFallback(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: DefaultLimenSessionCookie, Value: "cookie-tok"})
	if got := SessionTokenFromRequest(r); got != "cookie-tok" {
		t.Fatalf("got %q", got)
	}
}

func TestSessionTokenFromRequest_BearerPrefersOverCookie(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer header-tok")
	r.AddCookie(&http.Cookie{Name: DefaultLimenSessionCookie, Value: "cookie-tok"})
	if got := SessionTokenFromRequest(r); got != "header-tok" {
		t.Fatalf("got %q", got)
	}
}

func TestSessionTokenFromRequest_Empty(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := SessionTokenFromRequest(r); got != "" {
		t.Fatalf("got %q", got)
	}
}
