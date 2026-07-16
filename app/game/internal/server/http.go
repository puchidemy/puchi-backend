package server

import (
	v1 "github.com/puchidemy/puchi-backend/app/game/api/todo/v1"
	"github.com/puchidemy/puchi-backend/app/game/internal/conf"
	"github.com/puchidemy/puchi-backend/app/game/internal/service"
	authpkg "github.com/puchidemy/puchi-backend/pkg/auth"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/middleware/validate"
	"github.com/go-kratos/kratos/v3/transport/http"

	"go.einride.tech/aip/fieldbehavior"
	"google.golang.org/protobuf/proto"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, authCfg *conf.Auth, jwtValidator *authpkg.JWTValidator, todo *service.TodoService) *http.Server {
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
		http.Filter(authpkg.Middleware(authpkg.MiddlewareConfig{
			PublicPaths: authCfg.PublicPaths,
			Validator:   jwtValidator,
		})),
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
	v1.RegisterTodoServiceHTTPServer(srv, todo)
	return srv
}
