package log

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"gosilite/api/handler/idparse"
	log "gosilite/api/model/framework/log"
	syslog "gosilite/application/framework/service/syslog"
	R "gosilite/pkg/utils/response"
)

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// OperLogList .
// @Summary 操作日志列表
// @Description 操作日志列表接口
// @Tags 日志
// @Produce json
// @Param data query log.OperLogListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/operlog/list [GET]
func OperLogList(ctx context.Context, c *app.RequestContext) {
	var err error
	var req log.OperLogListRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	rows, total, err := syslog.SysLogSVC.OperPage(ctx, req.Page, req.PageSize,
		req.TraceId, req.UserName, req.Method, derefStr(req.HttpMethod),
		req.OperType, req.ExecStatus, req.StartTime, req.EndTime)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}

	R.SuccessWithData(ctx, c, &log.LogListResponse{List: toOperDTOs(rows), Total: total})
}

// OperLogDetail .
// @Summary 操作日志
// @Description 操作日志接口
// @Tags 日志
// @Produce json
// @Param data query log.LogIdRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/operlog [GET]
func OperLogDetail(ctx context.Context, c *app.RequestContext) {
	var err error
	var req log.LogIdRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	row, err := syslog.SysLogSVC.OperDetail(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}

	R.SuccessWithData(ctx, c, toOperDetailDTO(row))
}

// OperLogClean .
// @Summary 清空
// @Description 清空接口
// @Tags 日志
// @Produce json
// @Param data body log.LogCleanRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/operlog/clean [POST]
func OperLogClean(ctx context.Context, c *app.RequestContext) {
	var err error
	var req log.LogCleanRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	ids, err := idparse.Int64s(req.Ids)
	if err != nil {
		R.InvalidParamRequestResponse(c, "invalid ids")
		return
	}
	n, err := syslog.SysLogSVC.CleanOper(ctx, ids, req.All)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}

	R.SuccessWithData(ctx, c, int32(n))
}

// ErrorLogList .
// @Summary 错误日志列表
// @Description 错误日志列表接口
// @Tags 日志
// @Produce json
// @Param data query log.ErrorLogListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/errorlog/list [GET]
func ErrorLogList(ctx context.Context, c *app.RequestContext) {
	var err error
	var req log.ErrorLogListRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	rows, total, err := syslog.SysLogSVC.ErrorPage(ctx, req.Page, req.PageSize,
		req.TraceId, req.UserName, req.Keyword, derefStr(req.HttpMethod),
		req.HandleStatus, req.ErrLevel, req.StartTime, req.EndTime)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}

	R.SuccessWithData(ctx, c, &log.ErrorLogListResponse{List: toErrorDTOs(rows), Total: total})
}

// ErrorLogDetail .
// @Summary 错误日志
// @Description 错误日志接口
// @Tags 日志
// @Produce json
// @Param data query log.LogIdRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/errorlog [GET]
func ErrorLogDetail(ctx context.Context, c *app.RequestContext) {
	var err error
	var req log.LogIdRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	row, err := syslog.SysLogSVC.ErrorDetail(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}

	R.SuccessWithData(ctx, c, toErrorDetailDTO(row))
}

// ErrorLogHandle .
// @Summary 处理
// @Description 处理接口
// @Tags 日志
// @Produce json
// @Param data body log.ErrorHandleRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/errorlog/handle [POST]
func ErrorLogHandle(ctx context.Context, c *app.RequestContext) {
	var err error
	var req log.ErrorHandleRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	n, err := syslog.SysLogSVC.HandleError(ctx, req.ID, req.HandleRemark)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}

	R.SuccessWithData(ctx, c, int32(n))
}

// ErrorLogClean .
// @Summary 清空
// @Description 清空接口
// @Tags 日志
// @Produce json
// @Param data body log.LogCleanRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/errorlog/clean [POST]
func ErrorLogClean(ctx context.Context, c *app.RequestContext) {
	var err error
	var req log.LogCleanRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	ids, err := idparse.Int64s(req.Ids)
	if err != nil {
		R.InvalidParamRequestResponse(c, "invalid ids")
		return
	}
	n, err := syslog.SysLogSVC.CleanError(ctx, ids, req.All)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}

	R.SuccessWithData(ctx, c, int32(n))
}

// LoginLogList .
// @Summary 登录日志列表
// @Description 登录日志列表接口
// @Tags 日志
// @Produce json
// @Param data query log.LoginLogListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/loginlog/list [GET]
func LoginLogList(ctx context.Context, c *app.RequestContext) {
	var err error
	var req log.LoginLogListRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	rows, total, err := syslog.SysLogSVC.LoginPage(ctx, req.Page, req.PageSize,
		req.UserName, req.ClientIp, req.Status, req.StartTime, req.EndTime)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}

	R.SuccessWithData(ctx, c, &log.LoginLogListResponse{List: toLoginDTOs(rows), Total: total})
}
