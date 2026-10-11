package ossrule

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	ossrule "gosilite/api/model/framework/ossrule"
	"gosilite/application/framework/dal/model"
	ossruleSvc "gosilite/application/framework/service/ossrule"
	R "gosilite/pkg/utils/response"
)

// 路由与请求/响应 DTO 均由 hz 从 idl/framework/system_oss_rule.thrift 生成
// （api/model/framework/ossrule + api/router/framework/ossrule），
// 本文件只保留 Handler 实现（绑定 → 调 Service → 回包）。
// spec: docs/compose/spec/system-oss.md S2.6

const timeLayout = "2006-01-02 15:04:05"

// List 规则分页列表
// @Summary OSS处理规则列表
// @Description OSS处理规则分页列表接口
// @Tags 系统管理
// @Produce json
// @Param data query ossrule.OssRuleListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss/rule/list [GET]
func List(ctx context.Context, c *app.RequestContext) {
	var req ossrule.OssRuleListRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	list, total, err := ossruleSvc.SystemOssRuleSVC.Page(ctx, int(req.Page), int(req.PageSize), req.RuleName, req.Domain, req.Status, req.IsDefault)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, &ossrule.OssRuleListResponse{List: toDTOs(list), Total: total})
}

// Get 规则详情
// @Summary OSS处理规则详情
// @Description OSS处理规则详情接口
// @Tags 系统管理
// @Produce json
// @Param data query ossrule.OssRuleIdRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss/rule [GET]
func Get(ctx context.Context, c *app.RequestContext) {
	var req ossrule.OssRuleIdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m, err := ossruleSvc.SystemOssRuleSVC.Get(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDTO(m))
}

// Create 新增规则
// @Summary 新增OSS处理规则
// @Description 新增OSS处理规则接口
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param data body ossrule.OssRuleSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss/rule [POST]
func Create(ctx context.Context, c *app.RequestContext) {
	var req ossrule.OssRuleSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m := fromSaveRequest(&req, 0)
	if err := ossruleSvc.SystemOssRuleSVC.Create(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDTO(m))
}

// Update 编辑规则
// @Summary 修改OSS处理规则
// @Description 修改OSS处理规则接口
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param data body ossrule.OssRuleSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss/rule [PUT]
func Update(ctx context.Context, c *app.RequestContext) {
	var req ossrule.OssRuleSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if req.ID == nil {
		R.InvalidParamRequestResponse(c, "id required")
		return
	}
	m := fromSaveRequest(&req, *req.ID)
	if err := ossruleSvc.SystemOssRuleSVC.Update(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	saved, err := ossruleSvc.SystemOssRuleSVC.Get(ctx, m.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDTO(saved))
}

// Remove 批量删除规则
// @Summary 删除OSS处理规则
// @Description 删除OSS处理规则接口
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param data body ossrule.OssRuleIdsRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss/rule [DELETE]
func Remove(ctx context.Context, c *app.RequestContext) {
	var req ossrule.OssRuleIdsRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := ossruleSvc.SystemOssRuleSVC.Delete(ctx, req.Ids); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(len(req.Ids)))
}

// ChangeOverwrite 覆盖字段值开关
// @Summary 规则覆盖字段值修改
// @Description 规则覆盖字段值修改接口
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param data body ossrule.OssRuleOverwriteRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss/rule/overwrite [PUT]
func ChangeOverwrite(ctx context.Context, c *app.RequestContext) {
	var req ossrule.OssRuleOverwriteRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m, err := ossruleSvc.SystemOssRuleSVC.ChangeOverwrite(ctx, req.ID, req.IsOverwrite)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDTO(m))
}

// RefreshCache 刷新规则缓存
// @Summary 刷新OSS处理规则缓存
// @Description 刷新OSS处理规则缓存接口
// @Tags 系统管理
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/oss/rule/cache [DELETE]
func RefreshCache(ctx context.Context, c *app.RequestContext) {
	if err := ossruleSvc.SystemOssRuleSVC.RefreshCache(ctx); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

func toDTO(m *model.SystemOssRule) *ossrule.OssRuleDTO {
	dto := &ossrule.OssRuleDTO{
		ID:          m.ID,
		RuleName:    m.RuleName,
		Domain:      m.Domain,
		MimeType:    m.MimeType,
		Rule:        m.Rule,
		IsOverwrite: m.IsOverwrite,
		IsDefault:   m.IsDefault,
		Status:      m.Status,
		RuleSort:    int32(m.RuleSort),
		Remark:      m.Remark,
	}
	if !m.CreatedAt.IsZero() {
		dto.CreatedAt = m.CreatedAt.Format(timeLayout)
	}
	if !m.UpdatedAt.IsZero() {
		dto.UpdatedAt = m.UpdatedAt.Format(timeLayout)
	}
	return dto
}

func toDTOs(list []*model.SystemOssRule) []*ossrule.OssRuleDTO {
	out := make([]*ossrule.OssRuleDTO, 0, len(list))
	for _, m := range list {
		out = append(out, toDTO(m))
	}
	return out
}

// fromSaveRequest is_overwrite 不在 Save 契约中（由 ChangeOverwrite 单独管），落库保留零值由服务层处理
func fromSaveRequest(req *ossrule.OssRuleSaveRequest, id int64) *model.SystemOssRule {
	return &model.SystemOssRule{
		ID:         id,
		RuleName:   req.RuleName,
		Domain:     req.Domain,
		MimeType:   req.MimeType,
		Rule:       req.Rule,
		IsDefault:  req.IsDefault,
		Status:     req.Status,
		RuleSort:   req.RuleSort,
		Remark:     req.Remark,
		DelFlag:    0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}
