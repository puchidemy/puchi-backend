//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"log/slog"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
	"github.com/puchidemy/puchi-backend/app/auth/internal/conf"
	"github.com/puchidemy/puchi-backend/app/auth/internal/data"
	"github.com/puchidemy/puchi-backend/app/auth/internal/server"
	"github.com/puchidemy/puchi-backend/app/auth/internal/service"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Auth, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		newApp,
		NewTokenConfig,
		NewEncryptionKey,
		NewFrontendURL,
		NewEmailVerificationConfig,
	))
}
