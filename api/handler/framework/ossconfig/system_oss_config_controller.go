package ossconfig

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"gosilite/api/handler/idparse"
	ossconfig "gosilite/api/model/framework/ossconfig"
	"gosilite/application/framework/dal/model"
	ossconfigSvc "gosilite/application/framework/service/ossconfig"
	R "gosilite/pkg/utils/response"
)

// 路由与请求/响应 DTO 均由 hz 从 idl/framework/system_oss_config.thrift 生成
// （api/model/framework/ossconfig + api/router/framework/ossconfig），
// 本文件只保留 Handler 实现（绑定 → 调 Service → 回包）。
// spec: docs/compose/spec/system-oss.md S2.4

const timeLayout = "2006-01-02 15:04:05"

// List 配置分页列表
// @Summary OSS配置列表
// @Description 对象存储配置分页列表接口
// @Tags 系统管理
// @Produce json
// @Param data query ossconfig.OssConfigListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss/config/list [GET]
func List(ctx context.Context, c *app.RequestContext) {
	var req ossconfig.OssConfigListRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	list, total, err := ossconfigSvc.SystemOssConfigSVC.Page(ctx, int(req.Page), int(req.PageSize), req.Keyword)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, &ossconfig.OssConfigListResponse{List: toDTOs(list), Total: total})
}

// Get 配置详情
// @Summary OSS配置详情
// @Description 对象存储配置详情接口
// @Tags 系统管理
// @Produce json
// @Param data query ossconfig.OssConfigIdRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss/config [GET]
func Get(ctx context.Context, c *app.RequestContext) {
	var req ossconfig.OssConfigIdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m, err := ossconfigSvc.SystemOssConfigSVC.Get(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDTO(m))
}

// Create 新增配置
// @Summary 新增OSS配置
// @Description 对象存储配置新增接口
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param data body ossconfig.OssConfigSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss/config [POST]
func Create(ctx context.Context, c *app.RequestContext) {
	var req ossconfig.OssConfigSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m := fromSaveRequest(&req, 0)
	if err := ossconfigSvc.SystemOssConfigSVC.Create(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDTO(m))
}

// Update 编辑配置（config_key 不可改，服务层保留原值）
// @Summary 修改OSS配置
// @Description 对象存储配置修改接口
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param data body ossconfig.OssConfigSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss/config [PUT]
func Update(ctx context.Context, c *app.RequestContext) {
	var req ossconfig.OssConfigSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if req.ID == nil {
		R.InvalidParamRequestResponse(c, "id required")
		return
	}
	m := fromSaveRequest(&req, *req.ID)
	if err := ossconfigSvc.SystemOssConfigSVC.Update(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	saved, err := ossconfigSvc.SystemOssConfigSVC.Get(ctx, m.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDTO(saved))
}

// Remove 批量删除配置（默认配置拒绝）
// @Summary 删除OSS配置
// @Description 对象存储配置删除接口（默认配置不可删）
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param data body ossconfig.OssConfigIdsRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss/config [DELETE]
func Remove(ctx context.Context, c *app.RequestContext) {
	var req ossconfig.OssConfigIdsRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	ids, err := idparse.Int64s(req.Ids)
	if err != nil {
		R.InvalidParamRequestResponse(c, "invalid ids")
		return
	}
	if err := ossconfigSvc.SystemOssConfigSVC.Delete(ctx, ids); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(len(req.Ids)))
}

// ChangeStatus 设为默认（即时生效）
// @Summary 设为默认存储配置
// @Description 对象存储配置设默认接口（切换即时生效）
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param data body ossconfig.OssConfigStatusRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss/config/status [PUT]
func ChangeStatus(ctx context.Context, c *app.RequestContext) {
	var req ossconfig.OssConfigStatusRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m, err := ossconfigSvc.SystemOssConfigSVC.ChangeStatus(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDTO(m))
}

// Test 连通测试（不写库）
// @Summary OSS配置连通测试
// @Description 对象存储配置连通测试接口（临时客户端探活，不写库）
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param data body ossconfig.OssConfigTestRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss/config/test [POST]
func Test(ctx context.Context, c *app.RequestContext) {
	var req ossconfig.OssConfigTestRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	connected, message := ossconfigSvc.SystemOssConfigSVC.Test(ctx, req.AccessKey, req.SecretKey, req.BucketName, req.Endpoint, req.Region)
	R.SuccessWithData(ctx, c, &ossconfig.OssConfigTestResponse{Connected: connected, Message: message})
}

// --- 转换 ---

func toDTO(m *model.SystemOssConfig) *ossconfig.OssConfigDTO {
	dto := &ossconfig.OssConfigDTO{
		ID:           m.ID,
		ConfigKey:    m.ConfigKey,
		AccessKey:    m.AccessKey,
		SecretKey:    m.SecretKey,
		BucketName:   m.BucketName,
		Prefix:       m.Prefix,
		Endpoint:     m.Endpoint,
		Domain:       m.Domain,
		IsHttps:      m.IsHTTPS,
		Region:       m.Region,
		AccessPolicy: m.AccessPolicy,
		Status:       m.Status,
		Ext1:         m.Ext1,
		Remark:       m.Remark,
	}
	if !m.CreatedAt.IsZero() {
		dto.CreatedAt = m.CreatedAt.Format(timeLayout)
	}
	if !m.UpdatedAt.IsZero() {
		dto.UpdatedAt = m.UpdatedAt.Format(timeLayout)
	}
	return dto
}

func toDTOs(list []*model.SystemOssConfig) []*ossconfig.OssConfigDTO {
	out := make([]*ossconfig.OssConfigDTO, 0, len(list))
	for _, m := range list {
		out = append(out, toDTO(m))
	}
	return out
}

func fromSaveRequest(req *ossconfig.OssConfigSaveRequest, id int64) *model.SystemOssConfig {
	now := time.Now()
	return &model.SystemOssConfig{
		ID:           id,
		ConfigKey:    req.ConfigKey,
		AccessKey:    req.AccessKey,
		SecretKey:    req.SecretKey,
		BucketName:   req.BucketName,
		Prefix:       req.Prefix,
		Endpoint:     req.Endpoint,
		Domain:       req.Domain,
		IsHTTPS:      req.IsHttps,
		Region:       req.Region,
		AccessPolicy: req.AccessPolicy,
		Remark:       req.Remark,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
