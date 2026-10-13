package syscfg

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	syscfg "gosilite/api/model/framework/syscfg"
	"gosilite/application/framework/service/config"
	R "gosilite/pkg/utils/response"
)

// List .
// @Summary 配置列表
// @Description 配置列表接口
// @Tags 系统配置
// @Produce json
// @Param data query syscfg.PageQuery false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/config/list [GET]
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
// @Summary 配置
// @Description 配置接口
// @Tags 系统配置
// @Produce json
// @Param data body syscfg.ConfigSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/config [POST]
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
// @Summary 配置
// @Description 配置接口
// @Tags 系统配置
// @Produce json
// @Param data body syscfg.ConfigSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/config [PUT]
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
// @Summary 配置
// @Description 配置接口
// @Tags 系统配置
// @Produce json
// @Param data body syscfg.IdRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/config [DELETE]
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
// @Summary 键
// @Description 键接口
// @Tags 系统配置
// @Produce json
// @Param data query syscfg.ConfigKeyRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/config/key [GET]
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
