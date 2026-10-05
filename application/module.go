package application

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app/server"

	"gosilite/application/infra"
)

// Module is a business plugin: construct services on Init and mount routes on RegisterRoutes.
type Module interface {
	Name() string
	Init(ctx context.Context, deps *infra.AppDependencies) error
	RegisterRoutes(r *server.Hertz)
}
