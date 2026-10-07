package system

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"gosilite/application/framework/service/dept"
	R "gosilite/pkg/utils/response"
)

type deptLeaderRequest struct {
	DeptID int64 `json:"deptId"`
	UserID int64 `json:"userId"`
}

// DeptLeader 设置/清除部门负责人（userId=0 清除）
// @router /system/dept/leader [PUT]
func DeptLeader(ctx context.Context, c *app.RequestContext) {
	var req deptLeaderRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := dept.SystemDeptSVC.UpdateLeader(ctx, req.DeptID, req.UserID); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}
