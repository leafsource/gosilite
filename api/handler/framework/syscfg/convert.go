package syscfg

import (
	syscfg "gosilite/api/model/framework/syscfg"
	"gosilite/application/framework/dal/model"
)

// ========== 字典 ==========

func toDictTypeDTO(m *model.SystemDictType) *syscfg.DictTypeDTO {
	return &syscfg.DictTypeDTO{
		ID: m.ID, DictName: m.DictName, DictType: m.DictType,
		Status: int32(m.Status), OrderNum: m.OrderNum, Remark: m.Remark, GroupId: m.GroupID,
	}
}

func toDictTypeDTOs(list []*model.SystemDictType) []*syscfg.DictTypeDTO {
	out := make([]*syscfg.DictTypeDTO, 0, len(list))
	for _, m := range list {
		out = append(out, toDictTypeDTO(m))
	}
	return out
}

func fromDictTypeSaveRequest(req *syscfg.DictTypeSaveRequest) *model.SystemDictType {
	m := &model.SystemDictType{
		DictName: req.DictName, DictType: req.DictType,
		Status: int16(req.Status), OrderNum: req.OrderNum, GroupID: req.GroupId,
	}
	if req.Remark != nil {
		m.Remark = *req.Remark
	}
	return m
}

func toDictGroupDTO(m *model.SystemDictGroup) *syscfg.DictGroupDTO {
	return &syscfg.DictGroupDTO{
		ID: m.ID, GroupName: m.GroupName, OrderNum: m.OrderNum, Remark: m.Remark,
	}
}

func toDictGroupDTOs(list []*model.SystemDictGroup) []*syscfg.DictGroupDTO {
	out := make([]*syscfg.DictGroupDTO, 0, len(list))
	for _, m := range list {
		out = append(out, toDictGroupDTO(m))
	}
	return out
}

func fromDictGroupSaveRequest(req *syscfg.DictGroupSaveRequest) *model.SystemDictGroup {
	m := &model.SystemDictGroup{GroupName: req.GroupName, OrderNum: req.OrderNum}
	if req.Remark != nil {
		m.Remark = *req.Remark
	}
	return m
}

func toDictDataDTO(m *model.SystemDictData) *syscfg.DictDataDTO {
	return &syscfg.DictDataDTO{
		ID: m.ID, DictType: m.DictType, DictLabel: m.DictLabel, DictValue: m.DictValue,
		CssClass: m.CSSClass, ListOrder: m.ListOrder, Status: int32(m.Status), IsDefault: int32(m.IsDefault),
		Theme: m.Theme, Color: m.Color,
	}
}

func toDictDataDTOs(list []*model.SystemDictData) []*syscfg.DictDataDTO {
	out := make([]*syscfg.DictDataDTO, 0, len(list))
	for _, m := range list {
		out = append(out, toDictDataDTO(m))
	}
	return out
}

func fromDictDataSaveRequest(req *syscfg.DictDataSaveRequest) *model.SystemDictData {
	m := &model.SystemDictData{
		DictType: req.DictType, DictLabel: req.DictLabel, DictValue: req.DictValue,
		ListOrder: req.ListOrder, Status: int16(req.Status), IsDefault: int16(req.IsDefault),
	}
	if req.CssClass != nil {
		m.CSSClass = *req.CssClass
	}
	if req.Theme != nil {
		m.Theme = *req.Theme
	}
	if req.Color != nil {
		m.Color = *req.Color
	}
	return m
}

func toDictOptionDTOs(list []*model.SystemDictData) []*syscfg.DictOptionDTO {
	out := make([]*syscfg.DictOptionDTO, 0, len(list))
	for _, m := range list {
		out = append(out, &syscfg.DictOptionDTO{Label: m.DictLabel, Value: m.DictValue, Theme: m.Theme, Color: m.Color})
	}
	return out
}

// ========== 参数配置 ==========

func toConfigDTO(m *model.SystemConfig) *syscfg.ConfigDTO {
	return &syscfg.ConfigDTO{
		ID: m.ID, ConfigName: m.ConfigName, ConfigKey: m.ConfigKey,
		ConfigValue: m.ConfigValue, ConfigType: m.ConfigType,
		Remark: m.Remark, Status: int32(m.Status),
	}
}

func toConfigDTOs(list []*model.SystemConfig) []*syscfg.ConfigDTO {
	out := make([]*syscfg.ConfigDTO, 0, len(list))
	for _, m := range list {
		out = append(out, toConfigDTO(m))
	}
	return out
}

func fromConfigSaveRequest(req *syscfg.ConfigSaveRequest) *model.SystemConfig {
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
	return m
}
