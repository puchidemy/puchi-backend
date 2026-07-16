//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"log/slog"

	"github.com/puchidemy/puchi-backend/app/user/internal/biz"
	"github.com/puchidemy/puchi-backend/app/user/internal/conf"
	"github.com/puchidemy/puchi-backend/app/user/internal/data"
	"github.com/puchidemy/puchi-backend/app/user/internal/server"
	"github.com/puchidemy/puchi-backend/app/user/internal/service"
	authpkg "github.com/puchidemy/puchi-backend/pkg/auth"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Auth, *authpkg.JWTValidator, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
