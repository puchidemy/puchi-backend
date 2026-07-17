package oauth2

import (
	"context"
	"fmt"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// FacebookProvider implements OAuth2Provider for Facebook Login (OIDC).
type FacebookProvider struct {
	config  *oauth2.Config
	jwks    *keyfunc.JWKS
	jwksURL string
}

// NewFacebookProvider creates a new FacebookProvider.
func NewFacebookProvider(clientID, clientSecret, redirectURL string) (*FacebookProvider, error) {
	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email", "public_profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.facebook.com/v25.0/dialog/oauth",
			TokenURL: "https://graph.facebook.com/v25.0/oauth/access_token",
		},
	}
	jwks, err := keyfunc.Get("https://www.facebook.com/.well-known/oauth/openid/jwks/", keyfunc.Options{})
	if err != nil {
		return nil, fmt.Errorf("get facebook JWKS: %w", err)
	}
	return &FacebookProvider{config: cfg, jwks: jwks, jwksURL: "https://www.facebook.com/.well-known/oauth/openid/jwks/"}, nil
}

// Name returns "facebook".
func (p *FacebookProvider) Name() string { return "facebook" }

// AuthURL builds the Facebook OAuth2 authorization URL with PKCE.
func (p *FacebookProvider) AuthURL(state string, codeChallenge string) string {
	return p.config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
}

// Exchange exchanges the authorization code for an id_token and extracts user info.
// Facebook requires redirect_uri to be passed again during token exchange (byte-for-byte match).
func (p *FacebookProvider) Exchange(ctx context.Context, code string, codeVerifier string) (*ProviderUser, error) {
	token, err := p.config.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
		oauth2.SetAuthURLParam("redirect_uri", p.config.RedirectURL),
	)
	if err != nil {
		return nil, fmt.Errorf("facebook token exchange: %w", err)
	}

	// Extract id_token
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in facebook response")
	}

	// Verify id_token
	parsed, err := jwt.Parse(idToken, p.jwks.Keyfunc,
		jwt.WithIssuer("https://www.facebook.com"),
		jwt.WithAudience(p.config.ClientID),
		jwt.WithLeeway(30),
	)
	if err != nil {
		return nil, fmt.Errorf("verify facebook id_token: %w", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid facebook id_token claims")
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
