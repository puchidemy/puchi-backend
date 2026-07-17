//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"log/slog"

	"github.com/puchidemy/puchi-backend/app/core/internal/biz"
	"github.com/puchidemy/puchi-backend/app/core/internal/conf"
	"github.com/puchidemy/puchi-backend/app/core/internal/data"
	"github.com/puchidemy/puchi-backend/app/core/internal/events"
	"github.com/puchidemy/puchi-backend/app/core/internal/server"
	"github.com/puchidemy/puchi-backend/app/core/internal/service"
	authpkg "github.com/puchidemy/puchi-backend/pkg/auth"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Auth, *authpkg.JWTValidator, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, events.ProviderSet, newApp))
}
