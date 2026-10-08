package syscfgglobal

import (
	syscfgglobal "gosilite/api/model/framework/syscfgglobal"
	"gosilite/application/framework/dal/model"
)

func toConfigGlobalDTO(m *model.SystemConfig) *syscfgglobal.ConfigGlobalDTO {
	return &syscfgglobal.ConfigGlobalDTO{
		ID:          m.ID,
		ConfigName:  m.ConfigName,
		ConfigKey:   m.ConfigKey,
		ConfigValue: m.ConfigValue,
		ConfigType:  m.ConfigType,
		Remark:      m.Remark,
		Status:      int32(m.Status),
	}
}

func toConfigGlobalDTOs(list []*model.SystemConfig) []*syscfgglobal.ConfigGlobalDTO {
	out := make([]*syscfgglobal.ConfigGlobalDTO, 0, len(list))
	for _, m := range list {
		out = append(out, toConfigGlobalDTO(m))
	}
	return out
}
