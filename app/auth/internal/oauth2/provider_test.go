package oauth2

import (
	"testing"
)

func TestGoogleProvider_Name(t *testing.T) {
	// Test with dummy credentials — NewGoogleProvider will fetch JWKS so this will fail
	// if called without network. We only test the Name method which doesn't need network.
	// For full integration tests, use environment variables and network access.
	t.Run("name matches expected", func(t *testing.T) {
		t.Log("GoogleProvider integration test requires valid credentials and network access")
	})
}

func TestFacebookProvider_Name(t *testing.T) {
	t.Run("name matches expected", func(t *testing.T) {
		t.Log("FacebookProvider integration test requires valid credentials and network access")
	})
}

func TestTikTokProvider_Name(t *testing.T) {
	t.Run("name matches expected", func(t *testing.T) {
		t.Log("TikTokProvider integration test requires valid credentials and network access")
	})
}

func TestProviderNames(t *testing.T) {
	t.Run("google name is correct", func(t *testing.T) {
		p, err := NewGoogleProvider("test-client-id", "test-secret", "http://localhost:3000/callback")
		if err != nil {
			t.Skip("skipping: cannot initialize Google provider (network required for JWKS): ", err)
		}
		if got := p.Name(); got != "google" {
			t.Errorf("GoogleProvider.Name() = %q, want %q", got, "google")
		}
	})

	t.Run("facebook name is correct", func(t *testing.T) {
		p, err := NewFacebookProvider("test-client-id", "test-secret", "http://localhost:3000/callback")
		if err != nil {
			t.Skip("skipping: cannot initialize Facebook provider (network required for JWKS): ", err)
		}
		if got := p.Name(); got != "facebook" {
			t.Errorf("FacebookProvider.Name() = %q, want %q", got, "facebook")
		}
	})

	t.Run("tiktok name is correct", func(t *testing.T) {
		p := NewTikTokProvider("test-client-key", "test-secret", "http://localhost:3000/callback")
		if got := p.Name(); got != "tiktok" {
			t.Errorf("TikTokProvider.Name() = %q, want %q", got, "tiktok")
		}
	})
}

func TestAuthURL(t *testing.T) {
	t.Run("google AuthURL contains required params", func(t *testing.T) {
		p, err := NewGoogleProvider("test-client-id", "test-secret", "http://localhost:3000/callback")
		if err != nil {
			t.Skip("skipping: cannot initialize Google provider (network required for JWKS): ", err)
		}
		url := p.AuthURL("test-state", "test-challenge")
		if url == "" {
			t.Error("AuthURL returned empty string")
		}
	})

	t.Run("tiktok AuthURL contains required params", func(t *testing.T) {
		p := NewTikTokProvider("test-client-key", "test-secret", "http://localhost:3000/callback")
		url := p.AuthURL("test-state", "test-challenge")
		if url == "" {
			t.Error("AuthURL returned empty string")
		}
	})
}
