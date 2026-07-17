package server

import (
	"log/slog"
	nethttp "net/http"

	"github.com/go-kratos/kratos/v3/transport/http"

	authv1 "github.com/puchidemy/puchi-backend/app/auth/api/auth/v1"
	"github.com/puchidemy/puchi-backend/app/auth/internal/conf"
	"github.com/puchidemy/puchi-backend/app/auth/internal/oauth2"
	"github.com/puchidemy/puchi-backend/app/auth/internal/service"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, authCfg *conf.Auth, authService *service.AuthService, magicLinkService *service.MagicLinkService, mfaService *service.MFAService, adminService *service.AdminService, socialService *service.SocialService, emailVerificationService *service.EmailVerificationService) *http.Server {
	var opts = []http.ServerOption{}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}

	srv := http.NewServer(opts...)
	authv1.RegisterAuthServiceHTTPServer(srv, authService)

	// Register JWKS endpoint (before CORS wrapper so JWKS needs no CORS)
	srv.HandleFunc("/.well-known/jwks.json", authService.HandleJWKS)

	// Register auth REST handlers
	srv.HandleFunc("/auth/logout", authService.HandleLogout)
	srv.HandleFunc("/auth/password/change", authService.HandleChangePassword)
	srv.HandleFunc("/auth/password/reset/request", authService.HandleResetRequest)
	srv.HandleFunc("/auth/password/reset", authService.HandleResetComplete)

	// Register session management endpoints
	srv.HandleFunc("/auth/sessions", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		switch r.Method {
		case nethttp.MethodGet:
			authService.HandleListSessions(w, r)
		case nethttp.MethodDelete:
			authService.HandleRevokeAllSessions(w, r)
		default:
			w.WriteHeader(nethttp.StatusMethodNotAllowed)
		}
	})
	srv.HandleFunc("/auth/sessions/", authService.HandleRevokeSession)

	// Register refresh endpoint (raw handler, not via proto-defined route)
	srv.HandleFunc("/auth/refresh", authService.HandleRefresh)

	// Register magic link endpoints
	srv.HandleFunc("/auth/magic-link/send", magicLinkService.HandleSend)
	srv.HandleFunc("/auth/magic-link/verify", magicLinkService.HandleVerify)

	// Register MFA endpoints
	srv.HandleFunc("/auth/mfa/enroll", mfaService.HandleEnroll)
	srv.HandleFunc("/auth/mfa/verify", mfaService.HandleVerify)
	srv.HandleFunc("/auth/mfa/disable", mfaService.HandleDisable)

	// Register admin RBAC endpoints
	srv.HandleFunc("/admin/roles", adminService.HandleListRoles)
	srv.HandleFunc("/admin/permissions", adminService.HandleListPermissions)
	srv.HandleFunc("/admin/users/{id}/roles", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		switch r.Method {
		case nethttp.MethodGet:
			adminService.HandleGetUserRoles(w, r)
		case nethttp.MethodPost:
			adminService.HandleAssignRole(w, r)
		case nethttp.MethodDelete:
			adminService.HandleRemoveRole(w, r)
		default:
			w.WriteHeader(nethttp.StatusMethodNotAllowed)
		}
	})
	srv.HandleFunc("/admin/users/{id}/permissions", adminService.HandleGetUserPermissions)

	// ---- Email Verification Routes ----
	srv.HandleFunc("/auth/email/verify/send", emailVerificationService.HandleSend)
	srv.HandleFunc("/auth/email/verify", emailVerificationService.HandleVerify)

	// ---- Social OAuth2 Routes ----
	initSocialProviders(socialService, authCfg)

	srv.HandleFunc("/auth/social/{provider}", socialService.HandleSocialLogin)
	srv.HandleFunc("/auth/callback/{provider}", socialService.HandleOAuthCallback)
	srv.HandleFunc("/auth/social/link", socialService.HandleLink)
	srv.HandleFunc("/auth/social/unlink", socialService.HandleUnlink)
	srv.HandleFunc("/auth/social/connections", socialService.HandleListConnections)

	// Wrap with CORS middleware
	if len(authCfg.CorsAllowedOrigins) > 0 {
		corsOpts := CORSOptions{
			AllowedOrigins:   authCfg.CorsAllowedOrigins,
			AllowCredentials: true,
		}
		srv.Handler = CORSHandler(srv.Handler, corsOpts)
	}

	return srv
}

// initSocialProviders initialises OAuth2 providers from config and sets them on the SocialService.
// It also sets the frontend URL from the first allowed origin for OAuth callback redirects.
func initSocialProviders(socialService *service.SocialService, authCfg *conf.Auth) {
	if authCfg.Social == nil {
		return
	}

	providers := make(map[string]oauth2.OAuth2Provider)

	if gp := authCfg.Social.Google; gp != nil && gp.ClientId != "" && gp.ClientSecret != "" && gp.RedirectUrl != "" {
		p, err := oauth2.NewGoogleProvider(gp.ClientId, gp.ClientSecret, gp.RedirectUrl)
		if err != nil {
			slog.Warn("failed to init Google OAuth2 provider", slog.Any("error", err))
		} else {
			providers["google"] = p
		}
	}

	if fp := authCfg.Social.Facebook; fp != nil && fp.ClientId != "" && fp.ClientSecret != "" && fp.RedirectUrl != "" {
		p, err := oauth2.NewFacebookProvider(fp.ClientId, fp.ClientSecret, fp.RedirectUrl)
		if err != nil {
			slog.Warn("failed to init Facebook OAuth2 provider", slog.Any("error", err))
		} else {
			providers["facebook"] = p
		}
	}

	if tp := authCfg.Social.Tiktok; tp != nil && tp.ClientId != "" && tp.ClientSecret != "" && tp.RedirectUrl != "" {
		p := oauth2.NewTikTokProvider(tp.ClientId, tp.ClientSecret, tp.RedirectUrl)
		providers["tiktok"] = p
	}

	socialService.SetProviders(providers)

	// Use first allowed origin as the frontend URL for OAuth callback redirects
	if len(authCfg.CorsAllowedOrigins) > 0 {
		socialService.SetFrontendURL(authCfg.CorsAllowedOrigins[0])
	}
}
