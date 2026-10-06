package syscfg

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	syscfg "gosilite/api/handler/framework/syscfg"
)

// RegisterGlobal 注册免登录全局参数端点。
// 生成的 Register 整组继承 rootMw（permission.Default），无法单条豁免登录，
// 故按 AGENTS「纯手写系统管理路由可跳过 hz」另行挂载（spec: config-consumption S2.2）。
func RegisterGlobal(r *server.Hertz) {
	r.GET("/system/config/global", syscfg.GlobalList)
}
