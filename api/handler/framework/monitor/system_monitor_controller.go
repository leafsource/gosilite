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
// @Summary 在线用户列表
// @Description 在线用户列表接口
// @Tags 监控
// @Produce json
// @Param data query monitor.OnlineListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/online/list [GET]
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
// @Summary 在线会话强退
// @Description 在线会话强退接口
// @Tags 监控
// @Produce json
// @Param data body monitor.OnlineForceRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/online/force [POST]
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
// @Summary 数据源监控
// @Description 数据源监控接口
// @Tags 监控
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/monitor/datasource [GET]
func Datasource(ctx context.Context, c *app.RequestContext) {
	vo, err := monitorService.SystemMonitorSVC.Datasource(ctx)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDatasourceDTO(vo))
}

// Server Web 应用 runtime 快照。
// @Summary 运行时监控
// @Description 运行时监控接口
// @Tags 监控
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/monitor/server [GET]
func Server(ctx context.Context, c *app.RequestContext) {
	R.SuccessWithData(ctx, c, toServerDTO(monitorService.SystemMonitorSVC.Server()))
}

// SqlMonitor SQL 模板聚合与慢 SQL。
// @Summary SQL 监控
// @Description SQL 监控接口
// @Tags 监控
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/monitor/sql [GET]
func SqlMonitor(ctx context.Context, c *app.RequestContext) {
	R.SuccessWithData(ctx, c, toSqlMonitorDTO(monitorService.SystemMonitorSVC.SqlMonitor()))
}

// Firewall SQL 防火墙观察统计。
// @Summary SQL 防火墙状态
// @Description SQL 防火墙状态接口
// @Tags 监控
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/monitor/firewall [GET]
func Firewall(ctx context.Context, c *app.RequestContext) {
	R.SuccessWithData(ctx, c, toFirewallDTO(monitorService.SystemMonitorSVC.Firewall()))
}

// UriList URI 维度访问统计。
// @Summary URI 访问统计
// @Description URI 访问统计接口
// @Tags 监控
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/monitor/uri [GET]
func UriList(ctx context.Context, c *app.RequestContext) {
	R.SuccessWithData(ctx, c, toUriListDTO(monitorService.SystemMonitorSVC.UriList()))
}

// HttpMonitor Hertz method/status/时延摘要。
// @Summary HTTP 请求统计
// @Description HTTP 请求统计接口
// @Tags 监控
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/monitor/http [GET]
func HttpMonitor(ctx context.Context, c *app.RequestContext) {
	R.SuccessWithData(ctx, c, toHttpMonitorDTO(monitorService.SystemMonitorSVC.HttpMonitor()))
}
