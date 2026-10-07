package framework

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"

	frameworkMenuRouter "gosilite/api/router/framework/menu"
	frameworkSyscfgRouter "gosilite/api/router/framework/syscfg"
	frameworkSystemRouter "gosilite/api/router/framework/system"
	frameworkUserRouter "gosilite/api/router/framework/user"
	"gosilite/application/framework/dal/repository"
	"gosilite/application/framework/service/config"
	"gosilite/application/framework/service/dept"
	"gosilite/application/framework/service/dict"
	"gosilite/application/framework/service/menu"
	"gosilite/application/framework/service/post"
	"gosilite/application/framework/service/role"
	"gosilite/application/framework/service/user"
	"gosilite/application/infra"
)

// Module is the framework domain plugin (user / role / menu / dept / post).
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
	dept.InitService(ctx, deps.DB)
	post.InitService(ctx, deps.DB)
	dict.InitService(repository.NewDictRepository(deps.DB))
	config.InitService(repository.NewConfigRepository(deps.DB))
	return nil
}

// RegisterRoutes implements application.Module.
func (m *Module) RegisterRoutes(r *server.Hertz) {
	frameworkUserRouter.Register(r)
	frameworkMenuRouter.Register(r)
	frameworkSystemRouter.Register(r)
	frameworkSystemRouter.RegisterLeader(r)
	frameworkSyscfgRouter.Register(r)
	frameworkSyscfgRouter.RegisterGlobal(r)
}
