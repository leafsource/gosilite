package syscfg

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	syscfg "gosilite/api/model/framework/syscfg"
	"gosilite/application/framework/service/config"
	R "gosilite/pkg/utils/response"
)

// List .
// @router /system/config/list [GET]
func List(ctx context.Context, c *app.RequestContext) {
	var req syscfg.PageQuery
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	list, total, err := config.SystemConfigSVC.Page(ctx, int(req.Page), int(req.PageSize), req.Keyword)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	resp := &syscfg.ConfigListResponse{List: toConfigDTOs(list), Total: total}
	R.SuccessWithData(ctx, c, resp)
}

// Create .
// @router /system/config [POST]
func Create(ctx context.Context, c *app.RequestContext) {
	var req syscfg.ConfigSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m := fromConfigSaveRequest(&req)
	if err := config.SystemConfigSVC.Create(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toConfigDTO(m))
}

// Update .
// @router /system/config [PUT]
func Update(ctx context.Context, c *app.RequestContext) {
	var req syscfg.ConfigSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if req.ID == nil {
		R.InvalidParamRequestResponse(c, "id required")
		return
	}
	m := fromConfigSaveRequest(&req)
	m.ID = *req.ID
	if err := config.SystemConfigSVC.Update(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toConfigDTO(m))
}

// Remove .
// @router /system/config [DELETE]
func Remove(ctx context.Context, c *app.RequestContext) {
	var req syscfg.IdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := config.SystemConfigSVC.Delete(ctx, req.ID); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// GetByKey .
// @router /system/config/key [GET]
func GetByKey(ctx context.Context, c *app.RequestContext) {
	var req syscfg.ConfigKeyRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m, err := config.SystemConfigSVC.GetByKey(ctx, req.Key)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toConfigDTO(m))
}
