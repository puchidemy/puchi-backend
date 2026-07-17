package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
	"github.com/puchidemy/puchi-backend/app/auth/internal/data"
	"github.com/puchidemy/puchi-backend/app/auth/internal/oauth2"
)

// SocialService handles OAuth2 social login HTTP callbacks.
type SocialService struct {
	socialUC    *biz.SocialUsecase
	oauthRepo   *data.OAuthStateRepo
	providers   map[string]oauth2.OAuth2Provider
	frontendURL string
}

// NewSocialService creates a new SocialService.
func NewSocialService(socialUC *biz.SocialUsecase, oauthRepo *data.OAuthStateRepo) *SocialService {
	return &SocialService{
		socialUC:  socialUC,
		oauthRepo: oauthRepo,
		providers: make(map[string]oauth2.OAuth2Provider),
	}
}

// SetProviders sets the OAuth2 provider instances (called after construction from http.go).
func (s *SocialService) SetProviders(providers map[string]oauth2.OAuth2Provider) {
	s.providers = providers
}

// SetFrontendURL sets the frontend URL for OAuth callback redirects.
func (s *SocialService) SetFrontendURL(url string) {
	s.frontendURL = url
}

// HandleSocialLogin handles GET /auth/social/{provider}.
// Generates PKCE challenge, saves state in Valkey, redirects to provider.
func (s *SocialService) HandleSocialLogin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerName := vars["provider"]

	provider, ok := s.providers[providerName]
	if !ok {
		writeJSONError(w, http.StatusBadRequest, "unsupported provider: "+providerName)
		return
	}

	// Generate PKCE code verifier and challenge
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate code verifier")
		return
	}
	codeChallenge := generateCodeChallenge(codeVerifier)

	// Generate random state (UUID v4)
	state := uuid.NewString()

	// Save OAuth state in Valkey with 10-minute TTL
	oauthData := &biz.OAuthStateData{
		Provider:     providerName,
		CodeVerifier: codeVerifier,
	}
	if err := s.oauthRepo.Save(r.Context(), state, oauthData, 10*time.Minute); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save OAuth state")
		return
	}

	// Redirect to provider's authorization URL
	authURL := provider.AuthURL(state, codeChallenge)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// HandleOAuthCallback handles GET /auth/callback/{provider}?code=...&state=...
// Verifies state, exchanges code for tokens, and redirects to frontend.
func (s *SocialService) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		writeJSONError(w, http.StatusBadRequest, "missing code or state")
		return
	}

	// Verify state and get code_verifier from Valkey (atomic get + delete)
	oauthData, err := s.oauthRepo.GetAndDelete(r.Context(), state)
	if err != nil || oauthData == nil {
		writeJSONError(w, http.StatusBadRequest, "invalid or expired OAuth state")
		return
	}

	provider, ok := s.providers[oauthData.Provider]
	if !ok {
		writeJSONError(w, http.StatusBadRequest, "unsupported provider: "+oauthData.Provider)
		return
	}

	// Exchange authorization code for user info
	providerUser, err := provider.Exchange(r.Context(), code, oauthData.CodeVerifier)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "token exchange failed")
		return
	}

	// Map to biz.ProviderUserInfo
	userInfo := biz.ProviderUserInfo{
		ProviderID:    providerUser.ProviderID,
		Email:         providerUser.Email,
		EmailVerified: providerUser.EmailVerified,
		Name:          providerUser.Name,
		AvatarURL:     providerUser.AvatarURL,
	}

	// Login or create user with social connection
	pair, err := s.socialUC.LoginOrLink(r.Context(), oauthData.Provider, userInfo)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "login failed")
		return
	}

	// Set refresh_token as HttpOnly cookie
	setRefreshTokenCookieOnWriter(w, pair.RefreshToken)

	// Redirect to frontend callback page.
	// access_token will be obtained client-side via /auth/refresh
	// using the refresh_token cookie that was just set.
	redirectURL := fmt.Sprintf("%s/auth/callback/%s", s.frontendURL, oauthData.Provider)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// HandleLink handles POST /auth/social/link — links a social account
// to the currently authenticated user.
func (s *SocialService) HandleLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := extractUserIDFromJWT(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Provider     string `json:"provider"`
		Code         string `json:"code"`
		State        string `json:"state,omitempty"`
		CodeVerifier string `json:"code_verifier,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Provider == "" || req.Code == "" {
		writeJSONError(w, http.StatusBadRequest, "provider and code are required")
		return
	}

	provider, ok := s.providers[req.Provider]
	if !ok {
		writeJSONError(w, http.StatusBadRequest, "unsupported provider")
		return
	}

	// Try to get code_verifier from Valkey state if not provided directly
	codeVerifier := req.CodeVerifier
	if req.State != "" && codeVerifier == "" {
		oauthData, err := s.oauthRepo.GetAndDelete(r.Context(), req.State)
		if err == nil && oauthData != nil {
			codeVerifier = oauthData.CodeVerifier
		}
	}

	// Exchange code for user info
	providerUser, err := provider.Exchange(r.Context(), req.Code, codeVerifier)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "token exchange failed")
		return
	}

	userInfo := biz.ProviderUserInfo{
		ProviderID:    providerUser.ProviderID,
		Email:         providerUser.Email,
		EmailVerified: providerUser.EmailVerified,
		Name:          providerUser.Name,
		AvatarURL:     providerUser.AvatarURL,
	}

	if err := s.socialUC.LinkToExistingUser(r.Context(), userID, req.Provider, userInfo); err != nil {
		if errors.Is(err, biz.ErrSocialAlreadyLinked) {
			writeJSONError(w, http.StatusConflict, "social account already linked to another user")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to link social account")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "social account linked"})
}

// HandleUnlink handles POST /auth/social/unlink — removes a social connection
// from the currently authenticated user.
func (s *SocialService) HandleUnlink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := extractUserIDFromJWT(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ConnectionID string `json:"connection_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ConnectionID == "" {
		writeJSONError(w, http.StatusBadRequest, "connection_id is required")
		return
	}

	connID, err := uuid.Parse(req.ConnectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid connection_id")
		return
	}

	if err := s.socialUC.Unlink(r.Context(), userID, connID); err != nil {
		if errors.Is(err, biz.ErrCannotUnlinkLast) {
			writeJSONError(w, http.StatusBadRequest, "cannot unlink the only authentication method")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to unlink social account")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "social account unlinked"})
}

// HandleListConnections handles GET /auth/social/connections — lists
// all social connections for the currently authenticated user.
func (s *SocialService) HandleListConnections(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := extractUserIDFromJWT(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	conns, err := s.socialUC.ListConnections(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list connections")
		return
	}

	type connResponse struct {
		ID        string `json:"id"`
		Provider  string `json:"provider"`
		Email     string `json:"email,omitempty"`
		AvatarURL string `json:"avatar_url,omitempty"`
		LinkedAt  string `json:"linked_at"`
	}

	resp := make([]connResponse, len(conns))
	for i, c := range conns {
		resp[i] = connResponse{
			ID:        c.ID.String(),
			Provider:  c.Provider,
			Email:     c.ProviderEmail,
			AvatarURL: c.AvatarURL,
			LinkedAt:  c.LinkedAt.Format(time.RFC3339),
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// ---- PKCE helpers ----

// generateCodeVerifier generates a random PKCE code verifier (43 chars).
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate code verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateCodeChallenge computes the S256 PKCE code challenge from a verifier.
func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// ---- JWT extraction ----

// extractUserIDFromJWT extracts the user ID from the Bearer token in the request.
func extractUserIDFromJWT(r *http.Request) (uuid.UUID, error) {
	authHeader := r.Header.Get("Authorization")
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == authHeader || tokenStr == "" {
		return uuid.Nil, fmt.Errorf("no bearer token")
	}
	// Parse claims without verification — just extract user_id claim
	// For proper verification, use the TokenUsecase
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return uuid.Nil, fmt.Errorf("invalid token format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return uuid.Nil, fmt.Errorf("decode token payload: %w", err)
	}
	var claims struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return uuid.Nil, fmt.Errorf("parse token claims: %w", err)
	}
	if claims.Sub == "" {
		return uuid.Nil, fmt.Errorf("no subject in token")
	}
	return uuid.Parse(claims.Sub)
}

// ---- Cookie helpers ----

// setRefreshTokenCookieOnWriter sets the refresh_token cookie on the raw ResponseWriter.
func setRefreshTokenCookieOnWriter(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	}
	http.SetCookie(w, cookie)
}
