package message

import (
	message "gosilite/api/model/framework/message"
	"gosilite/application/framework/dal/model"
)

const timeLayout = "2006-01-02 15:04:05"

// ===== 消息配置 =====

func toConfigDTO(m *model.SystemMessageConfig) *message.MessageConfigDTO {
	dto := &message.MessageConfigDTO{
		ID:           m.ID,
		Title:        m.Title,
		MessageType:  m.MessageType,
		SupplierType: m.SupplierType,
		ConfigJson:   m.ConfigJson,
		Status:       int32(m.Status),
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

func toConfigDTOs(list []*model.SystemMessageConfig) []*message.MessageConfigDTO {
	out := make([]*message.MessageConfigDTO, 0, len(list))
	for _, m := range list {
		out = append(out, toConfigDTO(m))
	}
	return out
}

func fromConfigSaveRequest(req *message.MessageConfigSaveRequest) *model.SystemMessageConfig {
	return &model.SystemMessageConfig{
		Title:        req.Title,
		MessageType:  req.MessageType,
		SupplierType: req.SupplierType,
		ConfigJson:   req.ConfigJson,
		Status:       int16(req.Status),
		Remark:       req.Remark,
	}
}

// ===== 消息常量 =====

func toKeyDTO(m *model.SystemMessageKey) *message.MessageKeyDTO {
	dto := &message.MessageKeyDTO{
		ID:         m.ID,
		Name:       m.Name,
		MessageKey: m.MessageKey,
		Remark:     m.Remark,
	}
	if !m.CreatedAt.IsZero() {
		dto.CreatedAt = m.CreatedAt.Format(timeLayout)
	}
	if !m.UpdatedAt.IsZero() {
		dto.UpdatedAt = m.UpdatedAt.Format(timeLayout)
	}
	return dto
}

func toKeyDTOs(list []*model.SystemMessageKey) []*message.MessageKeyDTO {
	out := make([]*message.MessageKeyDTO, 0, len(list))
	for _, m := range list {
		out = append(out, toKeyDTO(m))
	}
	return out
}

func fromKeySaveRequest(req *message.MessageKeySaveRequest) *model.SystemMessageKey {
	return &model.SystemMessageKey{
		Name:       req.Name,
		MessageKey: req.MessageKey,
		Remark:     req.Remark,
	}
}

// ===== 消息模板 =====

func toTemplateDTO(m *model.SystemMessageTemplate) *message.MessageTemplateDTO {
	dto := &message.MessageTemplateDTO{
		ID:              m.ID,
		TemplateName:    m.TemplateName,
		MessageConfigId: m.MessageConfigID,
		MessageKeyId:    m.MessageKeyID,
		MessageKey:      m.MessageKey,
		MessageType:     m.MessageType,
		TemplateMode:    m.TemplateMode,
		Title:           m.Title,
		Signature:       m.Signature,
		TemplateId:      m.TemplateID,
		Content:         m.Content,
		VarsJson:        m.VarsJson,
		Status:          int32(m.Status),
		Remark:          m.Remark,
	}
	if !m.CreatedAt.IsZero() {
		dto.CreatedAt = m.CreatedAt.Format(timeLayout)
	}
	if !m.UpdatedAt.IsZero() {
		dto.UpdatedAt = m.UpdatedAt.Format(timeLayout)
	}
	return dto
}

func toTemplateDTOs(list []*model.SystemMessageTemplate) []*message.MessageTemplateDTO {
	out := make([]*message.MessageTemplateDTO, 0, len(list))
	for _, m := range list {
		out = append(out, toTemplateDTO(m))
	}
	return out
}

func fromTemplateSaveRequest(req *message.MessageTemplateSaveRequest) *model.SystemMessageTemplate {
	return &model.SystemMessageTemplate{
		TemplateName:    req.TemplateName,
		MessageConfigID: req.MessageConfigId,
		MessageKeyID:    req.MessageKeyId,
		MessageKey:      req.MessageKey,
		MessageType:     req.MessageType,
		TemplateMode:    req.TemplateMode,
		Title:           req.Title,
		Signature:       req.Signature,
		TemplateID:      req.TemplateId,
		Content:         req.Content,
		VarsJson:        req.VarsJson,
		Status:          int16(req.Status),
		Remark:          req.Remark,
	}
}

// ===== 发送记录 =====

func toLogDTO(m *model.SystemMessageLog) *message.MessageLogDTO {
	dto := &message.MessageLogDTO{
		ID:                  m.ID,
		MessageTemplateId:   m.MessageTemplateID,
		MessageKey:          m.MessageKey,
		MessageTemplateName: m.MessageTemplateName,
		MessageType:         m.MessageType,
		TemplateMode:        m.TemplateMode,
		Account:             m.Account,
		Title:               m.Title,
		TemplateId:          m.TemplateID,
		Content:             m.Content,
		MessageConfigTitle:  m.MessageConfigTitle,
		SupplierType:        m.SupplierType,
		ResponseBody:        m.ResponseBody,
	}
	if m.IsSuccess != nil {
		dto.IsSuccess = int32(*m.IsSuccess)
	}
	if m.CostTime != nil {
		dto.CostTime = *m.CostTime
	}
	if !m.LogTime.IsZero() {
		dto.LogTime = m.LogTime.Format(timeLayout)
	}
	return dto
}

func toLogDTOs(list []*model.SystemMessageLog) []*message.MessageLogDTO {
	out := make([]*message.MessageLogDTO, 0, len(list))
	for _, m := range list {
		out = append(out, toLogDTO(m))
	}
	return out
}
