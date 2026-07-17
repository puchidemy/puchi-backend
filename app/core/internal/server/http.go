package server

import (
	nethttp "net/http"
	"strings"

	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/middleware/validate"
	"github.com/go-kratos/kratos/v3/transport/http"
	profilepb "github.com/puchidemy/puchi-backend/app/core/api/profile/v1"
	socialpb "github.com/puchidemy/puchi-backend/app/core/api/social/v1"
	"github.com/puchidemy/puchi-backend/app/core/internal/biz"
	"github.com/puchidemy/puchi-backend/app/core/internal/conf"
	"github.com/puchidemy/puchi-backend/app/core/internal/service"
	authpkg "github.com/puchidemy/puchi-backend/pkg/auth"

	"go.einride.tech/aip/fieldbehavior"
	"google.golang.org/protobuf/proto"
)

var reservedProfileSegments = map[string]struct{}{
	"stats":           {},
	"achievements":    {},
	"linked-accounts": {},
	"merge-guest":     {},
	"avatar":          {},
}

// isPublicProfilePath allows unauthenticated GET /v1/profile/{username}
// (single segment, not a reserved route).
func isPublicProfilePath(r *nethttp.Request) bool {
	if r.Method != nethttp.MethodGet {
		return false
	}
	path := strings.TrimSuffix(r.URL.Path, "/")
	const prefix = "/v1/profile/"
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	seg := strings.TrimPrefix(path, prefix)
	if seg == "" || strings.Contains(seg, "/") {
		return false
	}
	_, reserved := reservedProfileSegments[seg]
	return !reserved
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, authCfg *conf.Auth, jwtValidator *authpkg.JWTValidator, profileService *service.ProfileService, socialService *service.SocialService, _ *biz.ProfileUsecase) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			authpkg.KratosMiddleware(authpkg.MiddlewareConfig{
				PublicPaths: authCfg.PublicPaths,
				IsPublic:    isPublicProfilePath,
				Validator:   jwtValidator,
			}),
			validate.Validator(func(req any) error {
				if msg, ok := req.(proto.Message); ok {
					if err := fieldbehavior.ValidateRequiredFields(msg); err != nil {
						return err
					}
				}
				return nil
			}),
		),
		http.Filter(corsFilter),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	profilepb.RegisterProfileServiceHTTPServer(srv, profileService)
	socialpb.RegisterSocialServiceHTTPServer(srv, socialService)
	srv.HandleFunc("/v1/profile/merge-guest", profileService.HandleMergeGuest)
	return srv
}
