package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware_PublicPathSkipsAuth(t *testing.T) {
	called := false
	h := Middleware(MiddlewareConfig{
		PublicPaths: []string{"/v1/health"},
		Validator:   &SessionValidator{},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/healthz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called || rr.Code != http.StatusOK {
		t.Fatalf("public path should pass, code=%d called=%v", rr.Code, called)
	}
}

func TestMiddleware_IsPublicOptionalAuth(t *testing.T) {
	called := false
	h := Middleware(MiddlewareConfig{
		IsPublic: func(r *http.Request) bool {
			return r.URL.Path == "/v1/profile/alice"
		},
		Validator: &SessionValidator{},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if _, ok := UserIDFromContext(r.Context()); ok {
			t.Fatal("expected no user without bearer")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/profile/alice", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called || rr.Code != http.StatusOK {
		t.Fatalf("IsPublic should pass without token, code=%d called=%v", rr.Code, called)
	}
}

func TestMiddleware_ProtectedRequiresAuth(t *testing.T) {
	h := Middleware(MiddlewareConfig{
		Validator: &SessionValidator{},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/profile", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want 401", rr.Code)
	}
}
