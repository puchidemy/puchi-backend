package server

import (
	todov1 "github.com/puchidemy/puchi-backend/app/notification/api/todo/v1"
	notifv1 "github.com/puchidemy/puchi-backend/app/notification/api/notification/v1"
	"github.com/puchidemy/puchi-backend/app/notification/internal/conf"
	"github.com/puchidemy/puchi-backend/app/notification/internal/service"
	authpkg "github.com/puchidemy/puchi-backend/pkg/auth"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, authCfg *conf.Auth, jwtValidator *authpkg.JWTValidator, todo *service.TodoService, notif *service.NotificationService) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			authpkg.KratosMiddleware(authpkg.MiddlewareConfig{
				PublicPaths: authCfg.PublicPaths,
				Validator:   jwtValidator,
			}),
		),
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
	todov1.RegisterTodoServiceHTTPServer(srv, todo)
	notifv1.RegisterNotificationServiceHTTPServer(srv, notif)
	return srv
}
