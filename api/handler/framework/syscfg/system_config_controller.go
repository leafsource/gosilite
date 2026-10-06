package syscfg

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	syscfg "gosilite/api/model/framework/syscfg"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/service/config"
	R "gosilite/pkg/utils/response"
)

func configDTO(m *model.SystemConfig) *syscfg.ConfigDTO {
	return &syscfg.ConfigDTO{
		ID: m.ID, ConfigName: m.ConfigName, ConfigKey: m.ConfigKey,
		ConfigValue: m.ConfigValue, ConfigType: m.ConfigType,
		Remark: m.Remark, Status: int32(m.Status),
	}
}

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
	resp := &syscfg.ConfigListResponse{List: make([]*syscfg.ConfigDTO, 0, len(list)), Total: total}
	for _, m := range list {
		resp.List = append(resp.List, configDTO(m))
	}
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
	m := &model.SystemConfig{
		ConfigName: req.ConfigName, ConfigKey: req.ConfigKey,
		ConfigValue: req.ConfigValue, Status: int16(req.Status),
	}
	if req.ConfigType != nil {
		m.ConfigType = *req.ConfigType
	}
	if req.Remark != nil {
		m.Remark = *req.Remark
	}
	if err := config.SystemConfigSVC.Create(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, configDTO(m))
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
	m := &model.SystemConfig{
		ID: *req.ID, ConfigName: req.ConfigName, ConfigKey: req.ConfigKey,
		ConfigValue: req.ConfigValue, Status: int16(req.Status),
	}
	if req.ConfigType != nil {
		m.ConfigType = *req.ConfigType
	}
	if req.Remark != nil {
		m.Remark = *req.Remark
	}
	if err := config.SystemConfigSVC.Update(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, configDTO(m))
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
	R.SuccessWithData(ctx, c, configDTO(m))
}
