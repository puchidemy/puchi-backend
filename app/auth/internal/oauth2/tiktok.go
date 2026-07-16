package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

// TikTokProvider implements OAuth2Provider for TikTok Login.
// TikTok does NOT support OIDC; user info is fetched via the userinfo API.
type TikTokProvider struct {
	config *oauth2.Config
	http   *http.Client
}

// NewTikTokProvider creates a new TikTokProvider.
func NewTikTokProvider(clientKey, clientSecret, redirectURL string) *TikTokProvider {
	cfg := &oauth2.Config{
		ClientID:     clientKey,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"user.info.basic"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.tiktok.com/v2/auth/authorize/",
			TokenURL: "https://open.tiktokapis.com/v2/oauth/token/",
		},
	}
	return &TikTokProvider{config: cfg, http: http.DefaultClient}
}

// Name returns "tiktok".
func (p *TikTokProvider) Name() string { return "tiktok" }

// AuthURL builds the TikTok OAuth2 authorization URL with PKCE.
func (p *TikTokProvider) AuthURL(state string, codeChallenge string) string {
	return p.config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
}

// tiktokUserInfoResponse represents the TikTok /v2/user/info/ API response.
type tiktokUserInfoResponse struct {
	Data struct {
		User struct {
			OpenID     string `json:"open_id"`
			UnionID    string `json:"union_id,omitempty"`
			DisplayName string `json:"display_name"`
			AvatarURL  string `json:"avatar_url"`
		} `json:"user"`
	} `json:"data"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Exchange exchanges the authorization code for an access token, then fetches user info.
func (p *TikTokProvider) Exchange(ctx context.Context, code string, codeVerifier string) (*ProviderUser, error) {
	token, err := p.config.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		return nil, fmt.Errorf("tiktok token exchange: %w", err)
	}

	// Fetch user info from TikTok userinfo API
	user, err := p.fetchUserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("tiktok fetch user info: %w", err)
	}

	return user, nil
}

// fetchUserInfo calls the TikTok userinfo API and extracts user details.
// TikTok does NOT return email; only open_id, display_name, and avatar_url are available.
func (p *TikTokProvider) fetchUserInfo(ctx context.Context, accessToken string) (*ProviderUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://open.tiktokapis.com/v2/user/info/?fields=open_id,display_name,avatar_url", nil)
	if err != nil {
		return nil, fmt.Errorf("create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read userinfo response: %w", err)
	}

	var info tiktokUserInfoResponse
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse userinfo response: %w", err)
	}

	if info.Error.Code != "" && info.Error.Code != "ok" {
		return nil, fmt.Errorf("tiktok API error: code=%s message=%s", info.Error.Code, info.Error.Message)
	}

	user := &ProviderUser{
		ProviderID: info.Data.User.OpenID,
		Name:       info.Data.User.DisplayName,
		AvatarURL:  info.Data.User.AvatarURL,
	}
	// TikTok does not provide email in the userinfo API.

	return user, nil
}
