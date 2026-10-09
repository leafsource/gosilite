package framework

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"

	frameworkLogRouter "gosilite/api/router/framework/log"
	frameworkMonitorRouter "gosilite/api/router/framework/monitor"
	frameworkMenuRouter "gosilite/api/router/framework/menu"
	frameworkNoticeRouter "gosilite/api/router/framework/notice"
	frameworkSyscfgRouter "gosilite/api/router/framework/syscfg"
	frameworkSyscfgGlobalRouter "gosilite/api/router/framework/syscfgglobal"
	frameworkSystemRouter "gosilite/api/router/framework/system"
	frameworkUserRouter "gosilite/api/router/framework/user"
	"gosilite/application/framework/dal/repository"
	"gosilite/application/framework/service/config"
	"gosilite/application/framework/service/dept"
	"gosilite/application/framework/service/dict"
	"gosilite/application/framework/service/menu"
	"gosilite/application/framework/service/notice"
	"gosilite/application/framework/service/post"
	"gosilite/application/framework/service/role"
	"gosilite/application/framework/service/syslog"
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
	notice.InitService(repository.NewNoticeRepository(deps.DB))
	syslog.InitService(repository.NewSysLogRepository(deps.DB))
	return nil
}

// RegisterRoutes implements application.Module.
func (m *Module) RegisterRoutes(r *server.Hertz) {
	frameworkUserRouter.Register(r)
	frameworkMenuRouter.Register(r)
	frameworkSystemRouter.Register(r)
	frameworkSyscfgRouter.Register(r)
	// 免登录全局参数：独立 IDL 文件生成的独立 Register（rootMw=nil），spec: config-consumption S2.2
	frameworkSyscfgGlobalRouter.Register(r)
	frameworkNoticeRouter.Register(r)
	frameworkLogRouter.Register(r)
	frameworkMonitorRouter.Register(r)
}
