package framework

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"

	frameworkCommonFileRouter "gosilite/api/router/framework/commonfile"
	frameworkLogRouter "gosilite/api/router/framework/log"
	frameworkMenuRouter "gosilite/api/router/framework/menu"
	frameworkMessageRouter "gosilite/api/router/framework/message"
	frameworkMonitorRouter "gosilite/api/router/framework/monitor"
	frameworkNoticeRouter "gosilite/api/router/framework/notice"
	frameworkOssRouter "gosilite/api/router/framework/oss"
	frameworkOssConfigRouter "gosilite/api/router/framework/ossconfig"
	frameworkOssRuleRouter "gosilite/api/router/framework/ossrule"
	frameworkSchedulerRouter "gosilite/api/router/framework/scheduler"
	frameworkSyscfgRouter "gosilite/api/router/framework/syscfg"
	frameworkSyscfgGlobalRouter "gosilite/api/router/framework/syscfgglobal"
	frameworkSystemRouter "gosilite/api/router/framework/system"
	frameworkUserRouter "gosilite/api/router/framework/user"
	"gosilite/application/framework/dal/repository"
	"gosilite/application/framework/scheduler"
	"gosilite/application/framework/service/config"
	"gosilite/application/framework/service/dept"
	"gosilite/application/framework/service/dict"
	"gosilite/application/framework/service/menu"
	"gosilite/application/framework/service/message"
	"gosilite/application/framework/service/monitor"
	"gosilite/application/framework/service/notice"
	"gosilite/application/framework/service/ossconfig"
	"gosilite/application/framework/service/ossfile"
	"gosilite/application/framework/service/ossrule"
	"gosilite/application/framework/service/post"
	"gosilite/application/framework/service/role"
	schedulerSvc "gosilite/application/framework/service/scheduler"
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
	config.InitService(repository.NewConfigRepository(deps.DB))
	user.InitService(ctx, deps.DB, deps.ImageOSS, deps.Cache)
	role.InitService(ctx, deps.DB)
	menu.InitService(ctx, deps.DB)
	dept.InitService(ctx, deps.DB)
	post.InitService(ctx, deps.DB)
	dict.InitService(repository.NewDictRepository(deps.DB))
	notice.InitService(repository.NewNoticeRepository(deps.DB))
	message.InitService(repository.NewMessageRepository(deps.DB))
	syslog.InitService(repository.NewSysLogRepository(deps.DB))
	monitor.InitService(deps.DB, deps.Cache)
	monitor.InitCacheService(deps.Cache)
	schedulerSvc.InitService(repository.NewSchedulerRepository(deps.DB), scheduler.NewCacheLeaseStore(deps.Cache))
	// 对象存储三件套（spec system-oss）：
	// ossconfig 变更 → StorageCtl.Invalidate 使默认配置即时生效（S2.4）
	ossconfig.InitService(repository.NewOssConfigRepository(deps.DB), deps.StorageCtl.Invalidate)
	ossrule.InitService(repository.NewOssRuleRepository(deps.DB))
	// ossfile 依赖默认配置客户端 + 规则消费（装饰 VO 的 ruleUrls）
	ossfile.InitService(repository.NewOssFileRepository(deps.DB), deps.StorageCtl, ossrule.SystemOssRuleSVC)
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
	// 统一文件服务（spec system-oss S2.11）：登录即可的通用上传/下载
	frameworkCommonFileRouter.Register(r)
	frameworkLogRouter.Register(r)
	frameworkMonitorRouter.Register(r)
	frameworkSchedulerRouter.Register(r)
	// 消息管理（spec system-message）：配置/常量/模板/记录 + 发送
	frameworkMessageRouter.Register(r)
	frameworkOssConfigRouter.Register(r)
	frameworkOssRuleRouter.Register(r)
	frameworkOssRouter.Register(r)
}
