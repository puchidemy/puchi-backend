package oauth2

import (
	"context"
	"fmt"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// GoogleProvider implements OAuth2Provider for Google Sign-In (OIDC).
type GoogleProvider struct {
	config  *oauth2.Config
	jwks    *keyfunc.JWKS
	jwksURL string
}

// NewGoogleProvider creates a new GoogleProvider.
func NewGoogleProvider(clientID, clientSecret, redirectURL string) (*GoogleProvider, error) {
	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}
	jwks, err := keyfunc.Get("https://www.googleapis.com/oauth2/v3/certs", keyfunc.Options{})
	if err != nil {
		return nil, fmt.Errorf("get google JWKS: %w", err)
	}
	return &GoogleProvider{config: cfg, jwks: jwks, jwksURL: "https://www.googleapis.com/oauth2/v3/certs"}, nil
}

// Name returns "google".
func (p *GoogleProvider) Name() string { return "google" }

// AuthURL builds the Google OAuth2 authorization URL with PKCE.
func (p *GoogleProvider) AuthURL(state string, codeChallenge string) string {
	return p.config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
}

// Exchange exchanges the authorization code for an id_token and extracts user info.
func (p *GoogleProvider) Exchange(ctx context.Context, code string, codeVerifier string) (*ProviderUser, error) {
	token, err := p.config.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		return nil, fmt.Errorf("google token exchange: %w", err)
	}

	// Extract id_token
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in google response")
	}

	// Verify id_token
	parsed, err := jwt.Parse(idToken, p.jwks.Keyfunc,
		jwt.WithIssuer("https://accounts.google.com"),
		jwt.WithAudience(p.config.ClientID),
		jwt.WithLeeway(30),
	)
	if err != nil {
		return nil, fmt.Errorf("verify google id_token: %w", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid google id_token claims")
	}

	user := &ProviderUser{}
	if sub, _ := claims.GetSubject(); sub != "" {
		user.ProviderID = sub
	}
	if email, _ := claims["email"].(string); email != "" {
		user.Email = email
	}
	if ev, _ := claims["email_verified"].(bool); ev {
		user.EmailVerified = true
	}
	if name, _ := claims["name"].(string); name != "" {
		user.Name = name
	}
	if pic, _ := claims["picture"].(string); pic != "" {
		user.AvatarURL = pic
	}

	return user, nil
}
