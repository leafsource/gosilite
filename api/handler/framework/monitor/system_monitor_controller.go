package monitor

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	monitor "gosilite/api/model/framework/monitor"
	monitorService "gosilite/application/framework/service/monitor"
	userService "gosilite/application/framework/service/user"
	R "gosilite/pkg/utils/response"
)

// OnlineList 在线用户列表（登录时间倒序，内存分页）。
// @router /system/online/list [GET]
func OnlineList(ctx context.Context, c *app.RequestContext) {
	var err error
	var req monitor.OnlineListRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	rows, total, err := userService.SystemUserSVC.OnlinePage(ctx, req.UserName, int(req.Page), int(req.PageSize))
	if err != nil {
		R.Error(ctx, c, err)
		return
	}

	R.SuccessWithData(ctx, c, &monitor.OnlineListResponse{List: toOnlineDTOs(rows), Total: total})
}

// OnlineForce 会话级强退。
// @router /system/online/force [POST]
func OnlineForce(ctx context.Context, c *app.RequestContext) {
	var err error
	var req monitor.OnlineForceRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	if err := userService.SystemUserSVC.ForceOffline(ctx, req.ID); err != nil {
		R.Error(ctx, c, err)
		return
	}

	R.SuccessWithData(ctx, c, int32(1))
}

// Datasource 数据源连接池与 Redis 探活。
// @router /system/monitor/datasource [GET]
func Datasource(ctx context.Context, c *app.RequestContext) {
	vo, err := monitorService.SystemMonitorSVC.Datasource(ctx)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDatasourceDTO(vo))
}

// Server Web 应用 runtime 快照。
// @router /system/monitor/server [GET]
func Server(ctx context.Context, c *app.RequestContext) {
	R.SuccessWithData(ctx, c, toServerDTO(monitorService.SystemMonitorSVC.Server()))
}

// SqlMonitor SQL 模板聚合与慢 SQL。
// @router /system/monitor/sql [GET]
func SqlMonitor(ctx context.Context, c *app.RequestContext) {
	R.SuccessWithData(ctx, c, toSqlMonitorDTO(monitorService.SystemMonitorSVC.SqlMonitor()))
}

// Firewall SQL 防火墙观察统计。
// @router /system/monitor/firewall [GET]
func Firewall(ctx context.Context, c *app.RequestContext) {
	R.SuccessWithData(ctx, c, toFirewallDTO(monitorService.SystemMonitorSVC.Firewall()))
}

// UriList URI 维度访问统计。
// @router /system/monitor/uri [GET]
func UriList(ctx context.Context, c *app.RequestContext) {
	R.SuccessWithData(ctx, c, toUriListDTO(monitorService.SystemMonitorSVC.UriList()))
}

// HttpMonitor Hertz method/status/时延摘要。
// @router /system/monitor/http [GET]
func HttpMonitor(ctx context.Context, c *app.RequestContext) {
	R.SuccessWithData(ctx, c, toHttpMonitorDTO(monitorService.SystemMonitorSVC.HttpMonitor()))
}
