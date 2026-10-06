package syscfg

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	syscfg "gosilite/api/model/framework/syscfg"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/service/dict"
	R "gosilite/pkg/utils/response"
)

func typeDTO(m *model.SystemDictType) *syscfg.DictTypeDTO {
	return &syscfg.DictTypeDTO{
		ID: m.ID, DictName: m.DictName, DictType: m.DictType,
		Status: int32(m.Status), OrderNum: m.OrderNum, Remark: m.Remark,
	}
}

func dataDTO(m *model.SystemDictData) *syscfg.DictDataDTO {
	return &syscfg.DictDataDTO{
		ID: m.ID, DictType: m.DictType, DictLabel: m.DictLabel, DictValue: m.DictValue,
		CssClass: m.CSSClass, ListOrder: m.ListOrder, Status: int32(m.Status), IsDefault: int32(m.IsDefault),
	}
}

// TypeList .
// @router /system/dict/type/list [GET]
func TypeList(ctx context.Context, c *app.RequestContext) {
	var req syscfg.PageQuery
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	list, total, err := dict.SystemDictTypeSVC.PageType(ctx, int(req.Page), int(req.PageSize), req.Keyword)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	resp := &syscfg.DictTypeListResponse{List: make([]*syscfg.DictTypeDTO, 0, len(list)), Total: total}
	for _, m := range list {
		resp.List = append(resp.List, typeDTO(m))
	}
	R.SuccessWithData(ctx, c, resp)
}

// TypeCreate .
// @router /system/dict/type [POST]
func TypeCreate(ctx context.Context, c *app.RequestContext) {
	var req syscfg.DictTypeSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m := &model.SystemDictType{
		DictName: req.DictName, DictType: req.DictType,
		Status: int16(req.Status), OrderNum: req.OrderNum,
	}
	if req.Remark != nil {
		m.Remark = *req.Remark
	}
	if err := dict.SystemDictTypeSVC.CreateType(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, typeDTO(m))
}

// TypeUpdate .
// @router /system/dict/type [PUT]
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
	m := &model.SystemDictType{
		ID: *req.ID, DictName: req.DictName, DictType: req.DictType,
		Status: int16(req.Status), OrderNum: req.OrderNum,
	}
	if req.Remark != nil {
		m.Remark = *req.Remark
	}
	if err := dict.SystemDictTypeSVC.UpdateType(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, typeDTO(m))
}

// TypeRemove .
// @router /system/dict/type [DELETE]
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
// @router /system/dict/data/list [GET]
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
	resp := &syscfg.DictDataListResponse{List: make([]*syscfg.DictDataDTO, 0, len(list)), Total: total}
	for _, m := range list {
		resp.List = append(resp.List, dataDTO(m))
	}
	R.SuccessWithData(ctx, c, resp)
}

// DataCreate .
// @router /system/dict/data [POST]
func DataCreate(ctx context.Context, c *app.RequestContext) {
	var req syscfg.DictDataSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m := &model.SystemDictData{
		DictType: req.DictType, DictLabel: req.DictLabel, DictValue: req.DictValue,
		ListOrder: req.ListOrder, Status: int16(req.Status), IsDefault: int16(req.IsDefault),
	}
	if req.CssClass != nil {
		m.CSSClass = *req.CssClass
	}
	if err := dict.SystemDictDataSVC.CreateData(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, dataDTO(m))
}

// DataUpdate .
// @router /system/dict/data [PUT]
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
	m := &model.SystemDictData{
		ID: *req.ID, DictType: req.DictType, DictLabel: req.DictLabel, DictValue: req.DictValue,
		ListOrder: req.ListOrder, Status: int16(req.Status), IsDefault: int16(req.IsDefault),
	}
	if req.CssClass != nil {
		m.CSSClass = *req.CssClass
	}
	if err := dict.SystemDictDataSVC.UpdateData(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, dataDTO(m))
}

// DataRemove .
// @router /system/dict/data [DELETE]
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
// @router /system/dict/data/options [GET]
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
	resp := &syscfg.DictOptionsResponse{List: make([]*syscfg.DictOptionDTO, 0, len(list))}
	for _, m := range list {
		resp.List = append(resp.List, &syscfg.DictOptionDTO{Label: m.DictLabel, Value: m.DictValue})
	}
	R.SuccessWithData(ctx, c, resp)
}
