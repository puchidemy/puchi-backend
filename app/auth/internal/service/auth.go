package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/google/uuid"

	authv1 "github.com/puchidemy/puchi-backend/app/auth/api/auth/v1"
	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"

	khttp "github.com/go-kratos/kratos/v3/transport/http"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthService implements the generated AuthServiceHTTPServer interface.
type AuthService struct {
	uc      *biz.AuthUsecase
	tokenUC *biz.TokenUsecase
}

var _ authv1.AuthServiceHTTPServer = (*AuthService)(nil)

// NewAuthService creates a new AuthService.
func NewAuthService(uc *biz.AuthUsecase, tokenUC *biz.TokenUsecase) *AuthService {
	return &AuthService{uc: uc, tokenUC: tokenUC}
}

// Register handles user registration.
func (s *AuthService) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	user, err := s.uc.Register(ctx, biz.RegisterInput{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
		IP:          getClientIP(ctx),
		UserAgent:   getUserAgent(ctx),
	})
	if err != nil {
		return nil, mapAuthError(err)
	}
	return &authv1.RegisterResponse{
		UserId:  user.ID.String(),
		Email:   user.Email,
		Message: "Registration successful.",
	}, nil
}

// Login handles user authentication and returns tokens.
func (s *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	pair, err := s.uc.Login(ctx, biz.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		IP:        getClientIP(ctx),
		UserAgent: getUserAgent(ctx),
	})
	if err != nil {
		return nil, mapAuthError(err)
	}

	// Set refresh token as HttpOnly cookie.
	setRefreshTokenCookie(ctx, pair.RefreshToken)

	return &authv1.LoginResponse{
		AccessToken: pair.AccessToken,
		ExpiresIn:   pair.ExpiresIn,
		User: &authv1.UserInfo{
			Id:            pair.User.ID.String(),
			Email:         pair.User.Email,
			DisplayName:   pair.User.DisplayName,
			EmailVerified: pair.User.EmailVerified,
		},
	}, nil
}

// HandleJWKS serves the JWKS public key set for token verification.
func (s *AuthService) HandleJWKS(w http.ResponseWriter, r *http.Request) {
	jwks, err := s.tokenUC.JWKSJSON()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to generate JWKS"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=86400, stale-while-revalidate=3600")
	w.WriteHeader(http.StatusOK)
	w.Write(jwks)
}

// HandleRefresh handles refresh token rotation.
func (s *AuthService) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "no refresh token"})
		return
	}

	pair, err := s.uc.RefreshTokens(r.Context(), cookie.Value, getClientIPFromReq(r), getUserAgentFromReq(r))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid refresh token"})
		return
	}

	setRefreshTokenCookie(r.Context(), pair.RefreshToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"access_token": pair.AccessToken,
		"expires_in":   pair.ExpiresIn,
	})
}

// mapAuthError maps biz errors to gRPC status codes (converted to HTTP by Kratos).
func mapAuthError(err error) error {
	switch {
	case errors.Is(err, biz.ErrEmailAlreadyExists):
		return status.Error(codes.AlreadyExists, "email already exists")
	case errors.Is(err, biz.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid email or password")
	case errors.Is(err, biz.ErrUserNotActive):
		return status.Error(codes.PermissionDenied, "account is not active")
	case errors.Is(err, biz.ErrPasswordTooShort):
		return status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	case errors.Is(err, biz.ErrSessionExpired):
		return status.Error(codes.Unauthenticated, "session expired")
	default:
		return status.Error(codes.Internal, "an internal error occurred")
	}
}

// getClientIP extracts the client IP from the HTTP transport context.
func getClientIP(ctx context.Context) string {
	if hctx, ok := ctx.(khttp.Context); ok {
		host, _, err := net.SplitHostPort(hctx.Request().RemoteAddr)
		if err != nil {
			return hctx.Request().RemoteAddr
		}
		return host
	}
	return ""
}

// getClientIPFromReq extracts the client IP from an http.Request.
func getClientIPFromReq(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// getUserAgent extracts the User-Agent from the HTTP transport context.
func getUserAgent(ctx context.Context) string {
	if hctx, ok := ctx.(khttp.Context); ok {
		return hctx.Request().UserAgent()
	}
	return ""
}

// getUserAgentFromReq extracts the User-Agent from an http.Request.
func getUserAgentFromReq(r *http.Request) string {
	return r.UserAgent()
}

// setRefreshTokenCookie sets the refresh token as an HttpOnly, Secure, cross-origin cookie.
func setRefreshTokenCookie(ctx context.Context, token string) {
	hctx, ok := ctx.(khttp.Context)
	if !ok {
		return
	}
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode, // cross-origin (puchi.io.vn → api.puchi.io.vn)
		MaxAge:   30 * 24 * 60 * 60,    // 30 days
	}
	http.SetCookie(hctx.Response(), cookie)
}

// clearRefreshTokenCookie clears the refresh_token cookie by setting MaxAge to -1.
func clearRefreshTokenCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}

// extractJWTClaims parses and validates the JWT access token from the
// Authorization header, returning the user ID, session ID, and JTI.
func (s *AuthService) extractJWTClaims(r *http.Request) (userID uuid.UUID, sessionID uuid.UUID, jti string, err error) {
	authHeader := r.Header.Get("Authorization")
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == authHeader {
		return uuid.Nil, uuid.Nil, "", fmt.Errorf("no bearer token")
	}
	claims, err := s.tokenUC.VerifyAccessToken(tokenStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, "", fmt.Errorf("verify token: %w", err)
	}
	return claims.UserID, claims.SessionID, claims.JTI, nil
}

// HandleLogout revokes the current session and clears the refresh token cookie.
func (s *AuthService) HandleLogout(w http.ResponseWriter, r *http.Request) {
	userID, sessionID, jti, err := s.extractJWTClaims(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	if err := s.uc.Logout(r.Context(), sessionID, userID, getClientIPFromReq(r), getUserAgentFromReq(r), jti); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to logout"})
		return
	}

	clearRefreshTokenCookie(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// HandleChangePassword changes the user's password after verifying the old one.
func (s *AuthService) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	userID, sessionID, jti, err := s.extractJWTClaims(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if err := s.uc.ChangePassword(r.Context(), userID, sessionID, req.OldPassword, req.NewPassword, getClientIPFromReq(r), getUserAgentFromReq(r), jti); err != nil {
		w.Header().Set("Content-Type", "application/json")
		status := http.StatusBadRequest
		switch {
		case errors.Is(err, biz.ErrInvalidCredentials):
			status = http.StatusUnauthorized
		case errors.Is(err, biz.ErrPasswordTooShort):
			status = http.StatusBadRequest
		default:
			status = http.StatusInternalServerError
		}
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Clear refresh_token cookie since all other sessions were revoked
	clearRefreshTokenCookie(w)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// HandleResetRequest initiates a password reset (token generation, no email yet).
func (s *AuthService) HandleResetRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	// Always return success (prevent email enumeration)
	_ = s.uc.RequestPasswordReset(r.Context(), req.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "if the email exists, a reset link has been sent"})
}

// HandleResetComplete completes the password reset.
func (s *AuthService) HandleResetComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if err := s.uc.CompletePasswordReset(r.Context(), req.Token, req.NewPassword, getClientIPFromReq(r), getUserAgentFromReq(r)); err != nil {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case errors.Is(err, biz.ErrInvalidCredentials):
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid or expired reset token"})
		case errors.Is(err, biz.ErrPasswordTooShort):
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "password must be at least 8 characters"})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "password has been reset"})
}
