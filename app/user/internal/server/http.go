package server

import (
	"net/http"

	pb "github.com/puchidemy/puchi-backend/app/user/api/social/v1"
	"github.com/puchidemy/puchi-backend/app/user/internal/conf"
	"github.com/puchidemy/puchi-backend/app/user/internal/service"
	authpkg "github.com/puchidemy/puchi-backend/pkg/auth"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/middleware/validate"
	kratoshttp "github.com/go-kratos/kratos/v3/transport/http"

	"go.einride.tech/aip/fieldbehavior"
	"google.golang.org/protobuf/proto"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, authCfg *conf.Auth, jwtValidator *authpkg.JWTValidator, socialService *service.SocialService) *kratoshttp.Server {
	var opts = []kratoshttp.ServerOption{
		kratoshttp.Middleware(
			recovery.Recovery(),
			validate.Validator(func(req any) error {
				if msg, ok := req.(proto.Message); ok {
					if err := fieldbehavior.ValidateRequiredFields(msg); err != nil {
						return err
					}
				}
				return nil
			}),
		),
		kratoshttp.Filter(authpkg.Middleware(authpkg.MiddlewareConfig{
			PublicPaths: authCfg.PublicPaths,
			Validator:   jwtValidator,
		})),
		kratoshttp.Filter(corsFilter),
	}
	if c.Http.Network != "" {
		opts = append(opts, kratoshttp.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, kratoshttp.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, kratoshttp.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := kratoshttp.NewServer(opts...)
	pb.RegisterSocialServiceHTTPServer(srv, socialService)
	return srv
}

// corsFilter adds CORS headers for browser-based auth flows.
func corsFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
