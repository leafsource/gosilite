package syscfg

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	syscfg "gosilite/api/model/framework/syscfg"
	"gosilite/application/framework/service/dict"
	R "gosilite/pkg/utils/response"
)

// TypeList .
// @Summary 类型列表
// @Description 类型列表接口
// @Tags 系统配置
// @Produce json
// @Param data query syscfg.DictTypeListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dict/type/list [GET]
func TypeList(ctx context.Context, c *app.RequestContext) {
	var req syscfg.DictTypeListRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	list, total, err := dict.SystemDictTypeSVC.PageType(ctx, int(req.Page), int(req.PageSize), req.Keyword, req.GroupId)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	resp := &syscfg.DictTypeListResponse{List: toDictTypeDTOs(list), Total: total}
	R.SuccessWithData(ctx, c, resp)
}

// TypeCreate .
// @Summary 类型
// @Description 类型接口
// @Tags 系统配置
// @Produce json
// @Param data body syscfg.DictTypeSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dict/type [POST]
func TypeCreate(ctx context.Context, c *app.RequestContext) {
	var req syscfg.DictTypeSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m := fromDictTypeSaveRequest(&req)
	if err := dict.SystemDictTypeSVC.CreateType(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDictTypeDTO(m))
}

// TypeUpdate .
// @Summary 类型
// @Description 类型接口
// @Tags 系统配置
// @Produce json
// @Param data body syscfg.DictTypeSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dict/type [PUT]
func TypeUpdate(ctx context.Context, c *app.RequestContext) {
	var req syscfg.DictTypeSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if req.ID == nil {
		R.InvalidParamRequestResponse(c, "id required")
		return
	}
	m := fromDictTypeSaveRequest(&req)
	m.ID = *req.ID
	if err := dict.SystemDictTypeSVC.UpdateType(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDictTypeDTO(m))
}

// TypeRemove .
// @Summary 类型
// @Description 类型接口
// @Tags 系统配置
// @Produce json
// @Param data body syscfg.IdRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dict/type [DELETE]
func TypeRemove(ctx context.Context, c *app.RequestContext) {
	var req syscfg.IdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := dict.SystemDictTypeSVC.DeleteType(ctx, req.ID); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// DataList .
// @Summary 数据列表
// @Description 数据列表接口
// @Tags 系统配置
// @Produce json
// @Param data query syscfg.DictDataListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dict/data/list [GET]
func DataList(ctx context.Context, c *app.RequestContext) {
	var req syscfg.DictDataListRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	dictType := ""
	if req.DictType != nil {
		dictType = *req.DictType
	}
	list, total, err := dict.SystemDictDataSVC.ListData(ctx, dictType, int(req.Page), int(req.PageSize))
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	resp := &syscfg.DictDataListResponse{List: toDictDataDTOs(list), Total: total}
	R.SuccessWithData(ctx, c, resp)
}

// DataCreate .
// @Summary 数据
// @Description 数据接口
// @Tags 系统配置
// @Produce json
// @Param data body syscfg.DictDataSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dict/data [POST]
func DataCreate(ctx context.Context, c *app.RequestContext) {
	var req syscfg.DictDataSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m := fromDictDataSaveRequest(&req)
	if err := dict.SystemDictDataSVC.CreateData(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDictDataDTO(m))
}

// DataUpdate .
// @Summary 数据
// @Description 数据接口
// @Tags 系统配置
// @Produce json
// @Param data body syscfg.DictDataSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dict/data [PUT]
func DataUpdate(ctx context.Context, c *app.RequestContext) {
	var req syscfg.DictDataSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if req.ID == nil {
		R.InvalidParamRequestResponse(c, "id required")
		return
	}
	m := fromDictDataSaveRequest(&req)
	m.ID = *req.ID
	if err := dict.SystemDictDataSVC.UpdateData(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDictDataDTO(m))
}

// DataRemove .
// @Summary 数据
// @Description 数据接口
// @Tags 系统配置
// @Produce json
// @Param data body syscfg.IdRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dict/data [DELETE]
func DataRemove(ctx context.Context, c *app.RequestContext) {
	var req syscfg.IdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := dict.SystemDictDataSVC.DeleteData(ctx, req.ID); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// DataOptions .
// @Summary 选项
// @Description 选项接口
// @Tags 系统配置
// @Produce json
// @Param data query syscfg.DictOptionsRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dict/data/options [GET]
func DataOptions(ctx context.Context, c *app.RequestContext) {
	var req syscfg.DictOptionsRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	list, err := dict.SystemDictDataSVC.Options(ctx, req.DictType)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	resp := &syscfg.DictOptionsResponse{List: toDictOptionDTOs(list)}
	R.SuccessWithData(ctx, c, resp)
}

// GroupList .
// @Summary 分组列表
// @Description 分组列表接口
// @Tags 系统配置
// @Produce json
// @Param data query syscfg.PageQuery false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dict/group/list [GET]
func GroupList(ctx context.Context, c *app.RequestContext) {
	var req syscfg.PageQuery
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	list, err := dict.SystemDictGroupSVC.List(ctx)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	resp := &syscfg.DictGroupListResponse{List: toDictGroupDTOs(list)}
	R.SuccessWithData(ctx, c, resp)
}

// GroupCreate .
// @Summary 分组
// @Description 分组接口
// @Tags 系统配置
// @Produce json
// @Param data body syscfg.DictGroupSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dict/group [POST]
func GroupCreate(ctx context.Context, c *app.RequestContext) {
	var req syscfg.DictGroupSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m := fromDictGroupSaveRequest(&req)
	if err := dict.SystemDictGroupSVC.Create(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDictGroupDTO(m))
}

// GroupUpdate .
// @Summary 分组
// @Description 分组接口
// @Tags 系统配置
// @Produce json
// @Param data body syscfg.DictGroupSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dict/group [PUT]
func GroupUpdate(ctx context.Context, c *app.RequestContext) {
	var req syscfg.DictGroupSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if req.ID == nil {
		R.InvalidParamRequestResponse(c, "id required")
		return
	}
	m := fromDictGroupSaveRequest(&req)
	m.ID = *req.ID
	if err := dict.SystemDictGroupSVC.Update(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDictGroupDTO(m))
}

// GroupRemove .
// @Summary 分组
// @Description 分组接口
// @Tags 系统配置
// @Produce json
// @Param data body syscfg.IdRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dict/group [DELETE]
func GroupRemove(ctx context.Context, c *app.RequestContext) {
	var req syscfg.IdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := dict.SystemDictGroupSVC.Delete(ctx, req.ID); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}
