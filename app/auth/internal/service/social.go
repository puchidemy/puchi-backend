package service

import (
	"encoding/json"
	"net/http"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
)

// SocialService handles OAuth2 social login HTTP callbacks.
type SocialService struct {
	socialUC *biz.SocialUsecase
}

// NewSocialService creates a new SocialService.
func NewSocialService(socialUC *biz.SocialUsecase) *SocialService {
	return &SocialService{socialUC: socialUC}
}

// HandleOAuthCallback handles GET /api/auth/callback/{provider}?code=...&state=...
// This is the OAuth2 redirect endpoint called by the provider after user authorization.
//
// In Phase 3, this will:
//   - Verify state and PKCE code_verifier from Valkey
//   - Exchange the code for tokens via the OAuth2Provider
//   - Call socialUC.LoginOrLink with the returned ProviderUserInfo
//   - Redirect the user to the frontend with the session
func (s *SocialService) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing code or state"})
		return
	}

	// Phase 3: verify state from Valkey, get code_verifier, call provider.Exchange()
	// Phase 3: call s.socialUC.LoginOrLink(ctx, providerName, providerUser)
	// Phase 3: set refresh_token cookie and redirect to frontend

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "callback received"})
}

// HandleLink handles POST /api/auth/social/link — links a social account
// to the currently authenticated user.
//
// Phase 3: authenticated endpoint that receives provider + code,
// performs exchange, and calls socialUC.LinkToExistingUser.
func (s *SocialService) HandleLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"error": "social linking not fully implemented"})
}

// HandleUnlink handles POST /api/auth/social/unlink — removes a social connection
// from the currently authenticated user.
//
// Phase 3: authenticated endpoint that receives connection_id.
func (s *SocialService) HandleUnlink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"error": "social unlink not fully implemented"})
}

// HandleListConnections handles GET /api/auth/social/connections — lists
// all social connections for the currently authenticated user.
//
// Phase 3: authenticated endpoint that returns the user's linked social accounts.
func (s *SocialService) HandleListConnections(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"error": "list social connections not fully implemented"})
}
