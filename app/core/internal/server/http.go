package server

import (
	pb "github.com/puchidemy/puchi-backend/app/core/api/profile/v1"
	"github.com/puchidemy/puchi-backend/app/core/internal/auth"
	"github.com/puchidemy/puchi-backend/app/core/internal/biz"
	"github.com/puchidemy/puchi-backend/app/core/internal/conf"
	"github.com/puchidemy/puchi-backend/app/core/internal/service"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/middleware/validate"
	"github.com/go-kratos/kratos/v3/transport/http"

	"go.einride.tech/aip/fieldbehavior"
	"google.golang.org/protobuf/proto"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, authCfg *conf.Auth, jwtValidator *auth.JWTValidator, profileService *service.ProfileService, profileUC *biz.ProfileUsecase) *http.Server {
	syncer := auth.NewUserSyncerFromUsecase(profileUC)

	var opts = []http.ServerOption{
		http.Middleware(
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
		http.Filter(corsFilter),
		http.Filter(auth.Middleware(authCfg, jwtValidator, syncer)),
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
	pb.RegisterProfileServiceHTTPServer(srv, profileService)
	return srv
}
