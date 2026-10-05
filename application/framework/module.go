package framework

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"

	frameworkMenuRouter "gosilite/api/router/framework/menu"
	frameworkUserRouter "gosilite/api/router/framework/user"
	"gosilite/application/framework/service/menu"
	"gosilite/application/framework/service/role"
	"gosilite/application/framework/service/user"
	"gosilite/application/infra"
)

// Module is the framework domain plugin (user / role / menu).
type Module struct{}

// NewModule constructs the framework plugin.
func NewModule() *Module { return &Module{} }

// Name implements application.Module.
func (m *Module) Name() string { return "framework" }

// Init implements application.Module.
func (m *Module) Init(ctx context.Context, deps *infra.AppDependencies) error {
	if deps == nil || deps.DB == nil {
		return fmt.Errorf("framework module: infra dependencies are required")
	}
	user.InitService(ctx, deps.DB, deps.ImageOSS, deps.Cache)
	role.InitService(ctx, deps.DB)
	menu.InitService(ctx, deps.DB)
	return nil
}

// RegisterRoutes implements application.Module.
func (m *Module) RegisterRoutes(r *server.Hertz) {
	frameworkUserRouter.Register(r)
	frameworkMenuRouter.Register(r)
}
