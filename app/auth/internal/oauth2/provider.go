package oauth2

import "context"

// ProviderUser represents the normalized user info returned by an OAuth2 provider.
type ProviderUser struct {
	ProviderID    string
	Email         string
	EmailVerified bool
	Name          string
	AvatarURL     string
}

// OAuth2Provider defines the interface for social login providers.
type OAuth2Provider interface {
	// Name returns the provider name (e.g. "google", "facebook", "tiktok").
	Name() string
	// AuthURL builds the authorization URL with the given state and PKCE code challenge.
	AuthURL(state string, codeChallenge string) string
	// Exchange exchanges the authorization code for a token and returns the user info.
	Exchange(ctx context.Context, code string, codeVerifier string) (*ProviderUser, error)
}
