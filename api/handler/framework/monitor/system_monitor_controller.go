package monitor

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	monitor "gosilite/api/model/framework/monitor"
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
