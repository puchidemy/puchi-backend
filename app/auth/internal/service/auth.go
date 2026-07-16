package service

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"

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

// getUserAgent extracts the User-Agent from the HTTP transport context.
func getUserAgent(ctx context.Context) string {
	if hctx, ok := ctx.(khttp.Context); ok {
		return hctx.Request().UserAgent()
	}
	return ""
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
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode, // cross-origin (puchi.io.vn → api.puchi.io.vn)
		MaxAge:   30 * 24 * 60 * 60,    // 30 days
	}
	http.SetCookie(hctx.Response(), cookie)
}
