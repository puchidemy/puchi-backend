package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
)

// MFAService handles MFA-related HTTP requests.
type MFAService struct {
	uc       *biz.MFAUsecase
	tokenUC  *biz.TokenUsecase
}

// NewMFAService creates a new MFAService.
func NewMFAService(uc *biz.MFAUsecase, tokenUC *biz.TokenUsecase) *MFAService {
	return &MFAService{uc: uc, tokenUC: tokenUC}
}

// HandleEnroll starts TOTP enrollment for the authenticated user.
func (s *MFAService) HandleEnroll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _, email, err := extractJWTClaims(r, s.tokenUC)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	enrollment, err := s.uc.Enroll(r.Context(), userID, email)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to enroll"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enrollment)
}

// HandleVerify confirms a TOTP code and enables MFA.
func (s *MFAService) HandleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _, _, err := extractJWTClaims(r, s.tokenUC)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if err := s.uc.VerifyCode(r.Context(), userID, req.Code); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// HandleDisable removes MFA for the authenticated user.
func (s *MFAService) HandleDisable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _, _, err := extractJWTClaims(r, s.tokenUC)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	if err := s.uc.Disable(r.Context(), userID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// extractJWTClaims parses and validates the JWT access token from the
// Authorization header, returning the user ID, session ID, and email.
func extractJWTClaims(r *http.Request, tokenUC *biz.TokenUsecase) (userID uuid.UUID, sessionID uuid.UUID, email string, err error) {
	authHeader := r.Header.Get("Authorization")
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == authHeader {
		return uuid.Nil, uuid.Nil, "", fmt.Errorf("no bearer token")
	}
	claims, err := tokenUC.VerifyAccessToken(tokenStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, "", fmt.Errorf("verify token: %w", err)
	}
	return claims.UserID, claims.SessionID, claims.Email, nil
}
