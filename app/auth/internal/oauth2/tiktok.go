package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// TikTokProvider implements OAuth2Provider for TikTok Login.
// TikTok does NOT support OIDC; user info is fetched via the userinfo API.
// Important: TikTok v2 OAuth uses client_key parameter naming (not standard client_id).
type TikTokProvider struct {
	clientID     string
	clientSecret string
	redirectURL  string
	scopes       []string
	authURL      string
	tokenURL     string
	http         *http.Client
}

// NewTikTokProvider creates a new TikTokProvider.
func NewTikTokProvider(clientKey, clientSecret, redirectURL string) *TikTokProvider {
	return &TikTokProvider{
		clientID:     clientKey,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		scopes:       []string{"user.info.basic"},
		authURL:      "https://www.tiktok.com/v2/auth/authorize/",
		tokenURL:     "https://open.tiktokapis.com/v2/oauth/token/",
		http:         http.DefaultClient,
	}
}

// Name returns "tiktok".
func (p *TikTokProvider) Name() string { return "tiktok" }

// AuthURL builds the TikTok OAuth2 authorization URL with PKCE.
// Note: TikTok v2 OAuth requires client_key parameter, NOT client_id
// (which Go's oauth2.AuthCodeURL would produce).
func (p *TikTokProvider) AuthURL(state string, codeChallenge string) string {
	vals := url.Values{}
	vals.Add("client_key", p.clientID)
	vals.Add("response_type", "code")
	vals.Add("redirect_uri", p.redirectURL)
	vals.Add("scope", strings.Join(p.scopes, ","))
	vals.Add("state", state)
	vals.Add("code_challenge", codeChallenge)
	vals.Add("code_challenge_method", "S256")
	return p.authURL + "?" + vals.Encode()
}

// tiktokUserInfoResponse represents the TikTok /v2/user/info/ API response.
type tiktokUserInfoResponse struct {
	Data struct {
		User struct {
			OpenID      string `json:"open_id"`
			UnionID     string `json:"union_id,omitempty"`
			DisplayName string `json:"display_name"`
			AvatarURL   string `json:"avatar_url"`
		} `json:"user"`
	} `json:"data"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// tiktokTokenResponse represents the TikTok /v2/oauth/token/ API response.
type tiktokTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
	RefreshExpiresIn int `json:"refresh_expires_in"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

// Exchange exchanges the authorization code for an access token, then fetches user info.
// Uses a direct HTTP POST instead of Go's oauth2.Config.Exchange because TikTok
// requires client_key/client_secret in the POST body (not client_id).
func (p *TikTokProvider) Exchange(ctx context.Context, code string, codeVerifier string) (*ProviderUser, error) {
	// Build token exchange request body
	vals := url.Values{}
	vals.Set("client_key", p.clientID)
	vals.Set("client_secret", p.clientSecret)
	vals.Set("code", code)
	vals.Set("grant_type", "authorization_code")
	vals.Set("redirect_uri", p.redirectURL)
	vals.Set("code_verifier", codeVerifier)

	body := vals.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}

	var tokenResp tiktokTokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("parse token response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("tiktok token error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("tiktok token response missing access_token")
	}

	// Fetch user info with the access token
	return p.fetchUserInfo(ctx, tokenResp.AccessToken)
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
