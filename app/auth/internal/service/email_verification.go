package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
)

// EmailVerificationService handles email verification HTTP endpoints.
type EmailVerificationService struct {
	uc      *biz.EmailVerificationUsecase
	tokenUC *biz.TokenUsecase
}

// NewEmailVerificationService creates a new EmailVerificationService.
func NewEmailVerificationService(uc *biz.EmailVerificationUsecase, tokenUC *biz.TokenUsecase) *EmailVerificationService {
	return &EmailVerificationService{uc: uc, tokenUC: tokenUC}
}

// HandleSend - POST /auth/email/verify/send
// Resends the email verification code for the authenticated user.
func (s *EmailVerificationService) HandleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	claims, err := s.extractJWTClaims(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	if claims.EmailVerified {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "email already verified"})
		return
	}

	if _, err := s.uc.Send(r.Context(), claims.UserID, claims.Email); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to send verification email"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "verification email sent"})
}

// HandleVerify - POST /auth/email/verify
// Verifies the email using the 6-digit code sent to the user's email.
func (s *EmailVerificationService) HandleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	claims, err := s.extractJWTClaims(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	if claims.EmailVerified {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "email already verified"})
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

	if req.Code == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "code is required"})
		return
	}

	if err := s.uc.Verify(r.Context(), claims.UserID, req.Code); err != nil {
		w.Header().Set("Content-Type", "application/json")
		status := http.StatusBadRequest
		switch {
		case errors.Is(err, biz.ErrInvalidVerificationCode):
			status = http.StatusBadRequest
		default:
			status = http.StatusInternalServerError
		}
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"message":        "email verified successfully",
		"email_verified": true,
	})
}

// extractJWTClaims parses and validates the JWT access token from the
// Authorization header, returning the claims.
func (s *EmailVerificationService) extractJWTClaims(r *http.Request) (*biz.AccessTokenClaims, error) {
	authHeader := r.Header.Get("Authorization")
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == authHeader {
		return nil, errors.New("no bearer token")
	}
	claims, err := s.tokenUC.VerifyAccessToken(tokenStr)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// extractUserID parses the JWT and returns just the user ID.
func (s *EmailVerificationService) extractUserID(r *http.Request) (uuid.UUID, error) {
	claims, err := s.extractJWTClaims(r)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}
