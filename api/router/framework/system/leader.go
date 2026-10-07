package system

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	handler "gosilite/api/handler/framework/system"
	"gosilite/api/middleware/permission"
)

// RegisterLeader 注册部门负责人设置接口。
// 按 AGENTS「纯手写系统管理路由可跳过 hz」执行：避免 hz 重生成冲掉
// middleware.go 里已挂的 perms 中间件（spec: config-consumption 同一处理方式）。
func RegisterLeader(r *server.Hertz) {
	r.PUT("/system/dept/leader", append(permission.HPermissions("system:dept:edit"), handler.DeptLeader)...)
}
