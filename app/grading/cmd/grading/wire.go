//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"log/slog"

	"github.com/puchidemy/puchi-backend/app/grading/internal/biz"
	"github.com/puchidemy/puchi-backend/app/grading/internal/conf"
	"github.com/puchidemy/puchi-backend/app/grading/internal/data"
	"github.com/puchidemy/puchi-backend/app/grading/internal/server"
	"github.com/puchidemy/puchi-backend/app/grading/internal/service"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
