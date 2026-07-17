package oauthtiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/thecodearcher/limen/plugins/oauth"
)

const (
	authURL  = "https://www.tiktok.com/v2/auth/authorize/"
	tokenURL = "https://open.tiktokapis.com/v2/oauth/token/"
	userURL  = "https://open.tiktokapis.com/v2/user/info/"
)

// New returns a Limen OAuth provider for TikTok Login Kit (Web).
// Credentials: TIKTOK_CLIENT_KEY (or TIKTOK_CLIENT_ID) + TIKTOK_CLIENT_SECRET.
func New() oauth.Provider {
	key := firstNonEmpty(os.Getenv("TIKTOK_CLIENT_KEY"), os.Getenv("TIKTOK_CLIENT_ID"))
	secret := os.Getenv("TIKTOK_CLIENT_SECRET")
	return &provider{
		clientKey:    key,
		clientSecret: secret,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
	}
}

type provider struct {
	clientKey    string
	clientSecret string
	httpClient   *http.Client
}

func (p *provider) Name() string { return "tiktok" }

// OAuth2Config is unused for TikTok (custom authorize + token exchange), but required by interface.
func (p *provider) OAuth2Config() (*oauth2.Config, []oauth2.AuthCodeOption) {
	return &oauth2.Config{
		ClientID:     p.clientKey,
		ClientSecret: p.clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		Scopes: []string{"user.info.basic"},
	}, nil
}

// PKCEEnabled: TikTok Web Login Kit does not use PKCE.
func (p *provider) PKCEEnabled() bool { return false }

func (p *provider) BuildAuthorizationURL(_ context.Context, state, _, callbackRedirectURI string) (string, error) {
	if p.clientKey == "" {
		return "", fmt.Errorf("tiktok: missing TIKTOK_CLIENT_KEY")
	}
	q := url.Values{}
	q.Set("client_key", p.clientKey)
	q.Set("response_type", "code")
	q.Set("scope", "user.info.basic")
	q.Set("redirect_uri", callbackRedirectURI)
	q.Set("state", state)
	return authURL + "?" + q.Encode(), nil
}

func (p *provider) ExchangeAuthorizationCode(ctx context.Context, code, _, redirectURI string) (*oauth.TokenResponse, error) {
	form := url.Values{}
	form.Set("client_key", p.clientKey)
	form.Set("client_secret", p.clientSecret)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", redirectURI)
	return p.postToken(ctx, form)
}

func (p *provider) RefreshToken(ctx context.Context, refreshToken string) (*oauth.TokenResponse, error) {
	form := url.Values{}
	form.Set("client_key", p.clientKey)
	form.Set("client_secret", p.clientSecret)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	return p.postToken(ctx, form)
}

type tokenAPIResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	OpenID           string `json:"open_id"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorCode        int    `json:"error_code"`
	Description      string `json:"description"`
}

func (p *provider) postToken(ctx context.Context, form url.Values) (*oauth.TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tr tokenAPIResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("tiktok: decode token response: %w", err)
	}
	if tr.Error != "" && tr.Error != "ok" {
		return nil, fmt.Errorf("tiktok: token error %s: %s", tr.Error, firstNonEmpty(tr.ErrorDescription, tr.Description))
	}
	if tr.ErrorCode != 0 {
		return nil, fmt.Errorf("tiktok: token error_code=%d: %s", tr.ErrorCode, tr.Description)
	}
	if tr.AccessToken == "" {
		return nil, fmt.Errorf("tiktok: empty access_token in response: %s", string(body))
	}

	out := &oauth.TokenResponse{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		TokenType:    firstNonEmpty(tr.TokenType, "Bearer"),
		Scope:        tr.Scope,
		AdditionalData: map[string]any{
			"open_id": tr.OpenID,
		},
	}
	if tr.ExpiresIn > 0 {
		out.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	}
	return out, nil
}

func (p *provider) GetUserInfo(ctx context.Context, token *oauth.TokenResponse) (*oauth.ProviderUserInfo, error) {
	fields := "open_id,union_id,avatar_url,display_name"
	u := userURL + "?fields=" + url.QueryEscape(fields)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tiktok: user info status %d: %s", resp.StatusCode, string(body))
	}

	var payload struct {
		Data struct {
			User map[string]any `json:"user"`
		} `json:"data"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("tiktok: decode user info: %w", err)
	}
	if payload.Error.Code != "" && payload.Error.Code != "ok" {
		return nil, fmt.Errorf("tiktok: user info error %s: %s", payload.Error.Code, payload.Error.Message)
	}

	user := payload.Data.User
	openID, _ := user["open_id"].(string)
	if openID == "" {
		if token.AdditionalData != nil {
			if oid, ok := token.AdditionalData["open_id"].(string); ok {
				openID = oid
			}
		}
	}
	if openID == "" {
		return nil, fmt.Errorf("tiktok: missing open_id")
	}
	name, _ := user["display_name"].(string)
	avatar, _ := user["avatar_url"].(string)

	// TikTok Login Kit (basic) does not return email — use stable synthetic address.
	email := fmt.Sprintf("%s@tiktok.oauth.puchi.local", openID)

	return &oauth.ProviderUserInfo{
		ID:            openID,
		Email:         email,
		EmailVerified: false,
		Name:          name,
		AvatarURL:     avatar,
		Raw:           user,
	}, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
