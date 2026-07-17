package server

import (
	nethttp "net/http"

	"github.com/go-kratos/kratos/v3/transport/http"

	authv1 "github.com/puchidemy/puchi-backend/app/auth/api/auth/v1"
	"github.com/puchidemy/puchi-backend/app/auth/internal/conf"
	"github.com/puchidemy/puchi-backend/app/auth/internal/service"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, authCfg *conf.Auth, authService *service.AuthService, magicLinkService *service.MagicLinkService, mfaService *service.MFAService, adminService *service.AdminService) *http.Server {
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
	srv.HandleFunc("/admin/users/{id}/roles", func(w http.ResponseWriter, r *nethttp.Request) {
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
