package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsPublicProfilePath(t *testing.T) {
	cases := []struct {
		method string
		path   string
		want   bool
	}{
		{http.MethodGet, "/v1/profile/alice", true},
		{http.MethodGet, "/v1/profile/alice/", true},
		{http.MethodGet, "/v1/profile/stats", false},
		{http.MethodGet, "/v1/profile/achievements", false},
		{http.MethodGet, "/v1/profile/linked-accounts", false},
		{http.MethodGet, "/v1/profile/merge-guest", false},
		{http.MethodGet, "/v1/profile/avatar", false},
		{http.MethodGet, "/v1/profile/settings", false},
		{http.MethodGet, "/v1/profile/stats/daily-activity", false},
		{http.MethodGet, "/v1/profile", false},
		{http.MethodPut, "/v1/profile/alice", false},
		{http.MethodGet, "/v1/profile/alice/extra", false},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		if got := isPublicProfilePath(req); got != tc.want {
			t.Fatalf("%s %s: got %v want %v", tc.method, tc.path, got, tc.want)
		}
	}
}
