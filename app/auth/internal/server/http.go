package server

import (
	"github.com/go-kratos/kratos/v3/transport/http"

	authv1 "github.com/puchidemy/puchi-backend/app/auth/api/auth/v1"
	"github.com/puchidemy/puchi-backend/app/auth/internal/conf"
	"github.com/puchidemy/puchi-backend/app/auth/internal/service"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, authCfg *conf.Auth, authService *service.AuthService) *http.Server {
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
