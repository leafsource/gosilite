package log

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
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
// @router /system/operlog/list [GET]
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
// @router /system/operlog [GET]
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
// @router /system/operlog/clean [POST]
func OperLogClean(ctx context.Context, c *app.RequestContext) {
	var err error
	var req log.LogCleanRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	n, err := syslog.SysLogSVC.CleanOper(ctx, req.Ids, req.All)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}

	R.SuccessWithData(ctx, c, int32(n))
}

// ErrorLogList .
// @router /system/errorlog/list [GET]
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
// @router /system/errorlog [GET]
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
// @router /system/errorlog/handle [POST]
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
// @router /system/errorlog/clean [POST]
func ErrorLogClean(ctx context.Context, c *app.RequestContext) {
	var err error
	var req log.LogCleanRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	n, err := syslog.SysLogSVC.CleanError(ctx, req.Ids, req.All)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}

	R.SuccessWithData(ctx, c, int32(n))
}

// LoginLogList .
// @router /system/loginlog/list [GET]
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
