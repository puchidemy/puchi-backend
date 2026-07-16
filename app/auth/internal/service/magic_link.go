package service

import (
	"encoding/json"
	"net/http"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
)

// MagicLinkService handles HTTP endpoints for magic link passwordless login.
type MagicLinkService struct {
	uc *biz.MagicLinkUsecase
}

// NewMagicLinkService creates a new MagicLinkService.
func NewMagicLinkService(uc *biz.MagicLinkUsecase) *MagicLinkService {
	return &MagicLinkService{uc: uc}
}

// HandleSend processes magic link send requests.
func (s *MagicLinkService) HandleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email      string `json:"email"`
		RedirectTo string `json:"redirect_to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	_, err := s.uc.Send(r.Context(), req.Email, req.RedirectTo)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
		return
	}

	// Always return success (prevent email enumeration)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// HandleVerify processes magic link verification.
func (s *MagicLinkService) HandleVerify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	pair, err := s.uc.Verify(r.Context(), token)
	if err != nil {
		http.Redirect(w, r, "/auth/sign-in?error=invalid_link", http.StatusFound)
		return
	}

	setRefreshTokenCookie(r.Context(), pair.RefreshToken)

	// Redirect to frontend
	redirectTo := "/learn"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"access_token": pair.AccessToken,
		"expires_in":   pair.ExpiresIn,
		"redirect_to":  redirectTo,
	})
}
