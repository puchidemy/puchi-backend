package server

import (
	"net/http"

	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/middleware/validate"
	kratoshttp "github.com/go-kratos/kratos/v3/transport/http"
	pb "github.com/puchidemy/puchi-backend/app/learn/api/learn/v1"
	"github.com/puchidemy/puchi-backend/app/learn/internal/conf"
	"github.com/puchidemy/puchi-backend/app/learn/internal/service"
	authpkg "github.com/puchidemy/puchi-backend/pkg/auth"

	"go.einride.tech/aip/fieldbehavior"
	"google.golang.org/protobuf/proto"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, authCfg *conf.Auth, learnCfg *conf.Learn, sessionValidator *authpkg.SessionValidator, learnService *service.LearnService) *kratoshttp.Server {
	publicPaths := append([]string(nil), authCfg.PublicPaths...)
	if learnCfg != nil {
		publicPaths = ensurePublicPath(publicPaths, "/v1/learn/guest/session")
		publicPaths = ensurePublicPath(publicPaths, "/v1/learn/units/")
		publicPaths = ensurePublicPath(publicPaths, "/v1/learn/lessons/")
		publicPaths = ensurePublicPath(publicPaths, "/v1/learn/attempts/")
	}

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
		kratoshttp.Filter(
			learnOptionalAuthFilter(sessionValidator),
			authpkg.Middleware(authpkg.MiddlewareConfig{
				PublicPaths: publicPaths,
				Validator:   sessionValidator,
			}),
		),
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
	pb.RegisterLearnServiceHTTPServer(srv, learnService)
	srv.HandleFunc("/v1/healthz", handleHealthz)
	return srv
}

func ensurePublicPath(paths []string, path string) []string {
	for _, p := range paths {
		if p == path {
			return paths
		}
	}
	return append(paths, path)
}

// handleHealthz reports service liveness for k8s probes and smoke tests.
func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
