package message

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	message "gosilite/api/model/framework/message"
	messageSvc "gosilite/application/framework/service/message"
	R "gosilite/pkg/utils/response"
)

// 路由与请求/响应 DTO 均由 hz 从 idl/framework/system_message.thrift 生成
// （api/model/framework/message + api/router/framework/message），
// 本文件只保留 Handler 实现（绑定 → 调 Service → 回包）。
// spec: docs/compose/spec/system-message.md S2.3

// ===== 消息配置 =====

// ConfigList 消息配置分页
// @Summary 消息配置列表
// @Description 消息配置分页列表接口
// @Tags 消息管理
// @Produce json
// @Param data query message.MessageConfigListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/config/list [GET]
func ConfigList(ctx context.Context, c *app.RequestContext) {
	var req message.MessageConfigListRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	list, total, err := messageSvc.SystemMessageSVC.ConfigPage(ctx, int(req.Page), int(req.PageSize), req.Keyword, req.MessageType, optI32ToInt16(req.Status))
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, &message.MessageConfigListResponse{List: toConfigDTOs(list), Total: total})
}

// ConfigGet 消息配置详情
// @Summary 消息配置详情
// @Description 消息配置详情接口
// @Tags 消息管理
// @Produce json
// @Param data query message.MessageIdRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/config [GET]
func ConfigGet(ctx context.Context, c *app.RequestContext) {
	var req message.MessageIdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m, err := messageSvc.SystemMessageSVC.ConfigGet(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toConfigDTO(m))
}

// ConfigCreate 新增消息配置
// @Summary 新增消息配置
// @Description 新增消息配置接口
// @Tags 消息管理
// @Produce json
// @Param data body message.MessageConfigSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/config [POST]
func ConfigCreate(ctx context.Context, c *app.RequestContext) {
	var req message.MessageConfigSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m := fromConfigSaveRequest(&req)
	if err := messageSvc.SystemMessageSVC.ConfigCreate(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toConfigDTO(m))
}

// ConfigUpdate 修改消息配置
// @Summary 修改消息配置
// @Description 修改消息配置接口
// @Tags 消息管理
// @Produce json
// @Param data body message.MessageConfigSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/config [PUT]
func ConfigUpdate(ctx context.Context, c *app.RequestContext) {
	var req message.MessageConfigSaveRequest
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
	if err := messageSvc.SystemMessageSVC.ConfigUpdate(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	saved, err := messageSvc.SystemMessageSVC.ConfigGet(ctx, m.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toConfigDTO(saved))
}

// ConfigRemove 删除消息配置
// @Summary 删除消息配置
// @Description 删除消息配置接口
// @Tags 消息管理
// @Produce json
// @Param data body message.MessageIdRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/config [DELETE]
func ConfigRemove(ctx context.Context, c *app.RequestContext) {
	var req message.MessageIdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	n, err := messageSvc.SystemMessageSVC.ConfigDelete(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(n))
}

// ===== 消息常量 =====

// KeyList 消息常量分页
// @Summary 消息常量列表
// @Description 消息常量分页列表接口
// @Tags 消息管理
// @Produce json
// @Param data query message.MessageKeyListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/key/list [GET]
func KeyList(ctx context.Context, c *app.RequestContext) {
	var req message.MessageKeyListRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	list, total, err := messageSvc.SystemMessageSVC.KeyPage(ctx, int(req.Page), int(req.PageSize), req.Keyword)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, &message.MessageKeyListResponse{List: toKeyDTOs(list), Total: total})
}

// KeyGet 消息常量详情
// @Summary 消息常量详情
// @Description 消息常量详情接口
// @Tags 消息管理
// @Produce json
// @Param data query message.MessageIdRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/key [GET]
func KeyGet(ctx context.Context, c *app.RequestContext) {
	var req message.MessageIdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m, err := messageSvc.SystemMessageSVC.KeyGet(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toKeyDTO(m))
}

// KeyCreate 新增消息常量
// @Summary 新增消息常量
// @Description 新增消息常量接口
// @Tags 消息管理
// @Produce json
// @Param data body message.MessageKeySaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/key [POST]
func KeyCreate(ctx context.Context, c *app.RequestContext) {
	var req message.MessageKeySaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m := fromKeySaveRequest(&req)
	if err := messageSvc.SystemMessageSVC.KeyCreate(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toKeyDTO(m))
}

// KeyUpdate 修改消息常量
// @Summary 修改消息常量
// @Description 修改消息常量接口
// @Tags 消息管理
// @Produce json
// @Param data body message.MessageKeySaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/key [PUT]
func KeyUpdate(ctx context.Context, c *app.RequestContext) {
	var req message.MessageKeySaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if req.ID == nil {
		R.InvalidParamRequestResponse(c, "id required")
		return
	}
	m := fromKeySaveRequest(&req)
	m.ID = *req.ID
	if err := messageSvc.SystemMessageSVC.KeyUpdate(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	saved, err := messageSvc.SystemMessageSVC.KeyGet(ctx, m.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toKeyDTO(saved))
}

// KeyRemove 删除消息常量
// @Summary 删除消息常量
// @Description 删除消息常量接口
// @Tags 消息管理
// @Produce json
// @Param data body message.MessageIdRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/key [DELETE]
func KeyRemove(ctx context.Context, c *app.RequestContext) {
	var req message.MessageIdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	n, err := messageSvc.SystemMessageSVC.KeyDelete(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(n))
}

// ===== 消息模板 =====

// TemplateList 消息模板分页
// @Summary 消息模板列表
// @Description 消息模板分页列表接口
// @Tags 消息管理
// @Produce json
// @Param data query message.MessageTemplateListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/template/list [GET]
func TemplateList(ctx context.Context, c *app.RequestContext) {
	var req message.MessageTemplateListRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	list, total, err := messageSvc.SystemMessageSVC.TemplatePage(ctx, int(req.Page), int(req.PageSize), req.Keyword, req.MessageType, req.MessageKeyId, optI32ToInt16(req.Status))
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, &message.MessageTemplateListResponse{List: toTemplateDTOs(list), Total: total})
}

// TemplateGet 消息模板详情
// @Summary 消息模板详情
// @Description 消息模板详情接口
// @Tags 消息管理
// @Produce json
// @Param data query message.MessageIdRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/template [GET]
func TemplateGet(ctx context.Context, c *app.RequestContext) {
	var req message.MessageIdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m, err := messageSvc.SystemMessageSVC.TemplateGet(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toTemplateDTO(m))
}

// TemplateCreate 新增消息模板
// @Summary 新增消息模板
// @Description 新增消息模板接口
// @Tags 消息管理
// @Produce json
// @Param data body message.MessageTemplateSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/template [POST]
func TemplateCreate(ctx context.Context, c *app.RequestContext) {
	var req message.MessageTemplateSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m := fromTemplateSaveRequest(&req)
	if err := messageSvc.SystemMessageSVC.TemplateCreate(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toTemplateDTO(m))
}

// TemplateUpdate 修改消息模板
// @Summary 修改消息模板
// @Description 修改消息模板接口
// @Tags 消息管理
// @Produce json
// @Param data body message.MessageTemplateSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/template [PUT]
func TemplateUpdate(ctx context.Context, c *app.RequestContext) {
	var req message.MessageTemplateSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if req.ID == nil {
		R.InvalidParamRequestResponse(c, "id required")
		return
	}
	m := fromTemplateSaveRequest(&req)
	m.ID = *req.ID
	if err := messageSvc.SystemMessageSVC.TemplateUpdate(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	saved, err := messageSvc.SystemMessageSVC.TemplateGet(ctx, m.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toTemplateDTO(saved))
}

// TemplateRemove 删除消息模板
// @Summary 删除消息模板
// @Description 删除消息模板接口
// @Tags 消息管理
// @Produce json
// @Param data body message.MessageIdRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/template [DELETE]
func TemplateRemove(ctx context.Context, c *app.RequestContext) {
	var req message.MessageIdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	n, err := messageSvc.SystemMessageSVC.TemplateDelete(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(n))
}

// TemplateSendTest 测试发送（无论成败都落发送记录）
// @Summary 测试发送
// @Description 消息模板测试发送接口（走发送服务，无论成败都落发送记录）
// @Tags 消息管理
// @Produce json
// @Param data body message.MessageSendTestRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/template/send-test [POST]
func TemplateSendTest(ctx context.Context, c *app.RequestContext) {
	var req message.MessageSendTestRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if strings.TrimSpace(req.Account) == "" {
		R.InvalidParamRequestResponse(c, "account required")
		return
	}
	results, err := messageSvc.SystemMessageSVC.Send(ctx, messageSvc.SendReq{
		TemplateID: req.MessageTemplateId,
		Accounts:   []string{req.Account},
		Vars:       req.Vars,
	})
	if err != nil {
		// 失败路径：发送记录已落库（若已进入发送阶段），错误信息原样返回给前端
		R.Error(ctx, c, err)
		return
	}
	if len(results) == 0 {
		R.InvalidParamRequestResponse(c, "发送结果为空")
		return
	}
	r0 := results[0]
	R.SuccessWithData(ctx, c, &message.MessageSendTestResponse{
		IsSuccess:    r0.IsSuccess,
		ResponseBody: r0.ResponseBody,
		CostTime:     r0.CostTime,
		LogId:        r0.LogID,
	})
}

// ===== 发送记录 =====

// LogList 发送记录分页
// @Summary 发送记录列表
// @Description 发送记录分页列表接口
// @Tags 消息管理
// @Produce json
// @Param data query message.MessageLogListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/log/list [GET]
func LogList(ctx context.Context, c *app.RequestContext) {
	var req message.MessageLogListRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	list, total, err := messageSvc.SystemMessageSVC.LogPage(ctx, int(req.Page), int(req.PageSize), req.Keyword, req.MessageType, req.MessageKey, optI32ToInt16(req.IsSuccess))
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, &message.MessageLogListResponse{List: toLogDTOs(list), Total: total})
}

// LogGet 发送记录详情
// @Summary 发送记录详情
// @Description 发送记录详情接口
// @Tags 消息管理
// @Produce json
// @Param data query message.MessageIdRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/log [GET]
func LogGet(ctx context.Context, c *app.RequestContext) {
	var req message.MessageIdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m, err := messageSvc.SystemMessageSVC.LogGet(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toLogDTO(m))
}

// LogRemove 删除单条发送记录
// @Summary 删除发送记录
// @Description 删除单条发送记录接口（返回删除条数）
// @Tags 消息管理
// @Produce json
// @Param data body message.MessageIdRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/log [DELETE]
func LogRemove(ctx context.Context, c *app.RequestContext) {
	var req message.MessageIdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	n, err := messageSvc.SystemMessageSVC.LogRemove(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, n)
}

// LogClear 清空发送记录
// @Summary 清空发送记录
// @Description 清空发送记录接口（返回删除条数）
// @Tags 消息管理
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/message/log/clear [DELETE]
func LogClear(ctx context.Context, c *app.RequestContext) {
	n, err := messageSvc.SystemMessageSVC.LogClear(ctx)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, n)
}

// ===== 元数据 =====

// Suppliers 消息类型可用供应商
// @Summary 消息供应商列表
// @Description 按消息类型返回可用供应商列表接口
// @Tags 消息管理
// @Produce json
// @Param data query message.MessageSuppliersRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/message/suppliers [GET]
func Suppliers(ctx context.Context, c *app.RequestContext) {
	var req message.MessageSuppliersRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	opts := messageSvc.Suppliers(req.MessageType)
	list := make([]*message.MessageSupplierOption, 0, len(opts))
	for _, o := range opts {
		list = append(list, &message.MessageSupplierOption{Code: o.Code, Label: o.Label})
	}
	R.SuccessWithData(ctx, c, &message.MessageSuppliersResponse{List: list})
}

// FieldConfigs supplier 动态表单字段定义
// @Summary 消息配置动态字段
// @Description supplier → 动态表单字段定义 JSON 接口（前端 evalFieldConfigs 消费）
// @Tags 消息管理
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/message/field-configs [GET]
func FieldConfigs(ctx context.Context, c *app.RequestContext) {
	R.SuccessWithData(ctx, c, &message.MessageFieldConfigsResponse{JSON: messageSvc.FieldConfigsJSON()})
}

// optI32ToInt16 可选过滤条件转换（未设置返回 nil）
func optI32ToInt16(p *int32) *int16 {
	if p == nil {
		return nil
	}
	v := int16(*p)
	return &v
}
