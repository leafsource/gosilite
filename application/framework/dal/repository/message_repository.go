package repository

import (
	"context"
	"strings"
	"time"

	"gosilite/application/framework/dal/model"

	"gorm.io/gorm"
)

// IMessageRepository 消息管理仓储（配置/常量/模板/发送记录）
type IMessageRepository interface {
	// ===== 消息配置 =====
	ConfigPage(ctx context.Context, page, pageSize int, keyword string, messageType *string, status *int16) ([]*model.SystemMessageConfig, int64, error)
	ConfigGet(ctx context.Context, id int64) (*model.SystemMessageConfig, error)
	ConfigCreate(ctx context.Context, m *model.SystemMessageConfig) error
	ConfigSave(ctx context.Context, m *model.SystemMessageConfig) error
	ConfigDelete(ctx context.Context, id int64) (int64, error)

	// ===== 消息常量 =====
	KeyPage(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemMessageKey, int64, error)
	KeyGet(ctx context.Context, id int64) (*model.SystemMessageKey, error)
	KeyCreate(ctx context.Context, m *model.SystemMessageKey) error
	KeySave(ctx context.Context, m *model.SystemMessageKey) error
	KeyDelete(ctx context.Context, id int64) (int64, error)
	// KeyExists message_key 唯一（excludeID 用于更新时排除自身）
	KeyExists(ctx context.Context, messageKey string, excludeID int64) (bool, error)

	// ===== 消息模板 =====
	TemplatePage(ctx context.Context, page, pageSize int, keyword string, messageType *string, messageKeyID *int64, status *int16) ([]*model.SystemMessageTemplate, int64, error)
	TemplateGet(ctx context.Context, id int64) (*model.SystemMessageTemplate, error)
	TemplateCreate(ctx context.Context, m *model.SystemMessageTemplate) error
	TemplateSave(ctx context.Context, m *model.SystemMessageTemplate) error
	TemplateDelete(ctx context.Context, id int64) (int64, error)
	// TemplateGetEnabled 发送用：status=1
	TemplateGetEnabled(ctx context.Context, id int64) (*model.SystemMessageTemplate, error)
	// TemplateGetEnabledByKey 按常量 key + 类型取启用模板（同 key 多条取最新）
	TemplateGetEnabledByKey(ctx context.Context, messageKey, messageType string) (*model.SystemMessageTemplate, error)

	// ===== 发送记录 =====
	LogPage(ctx context.Context, page, pageSize int, keyword string, messageType, messageKey *string, isSuccess *int16) ([]*model.SystemMessageLog, int64, error)
	LogGet(ctx context.Context, id int64) (*model.SystemMessageLog, error)
	LogInsert(ctx context.Context, m *model.SystemMessageLog) error
	LogRemove(ctx context.Context, id int64) (int64, error)
	LogClear(ctx context.Context) (int64, error)
}

type messageRepository struct{ db *gorm.DB }

// NewMessageRepository 创建消息管理仓储
func NewMessageRepository(db *gorm.DB) IMessageRepository {
	return &messageRepository{db: db}
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}
	return page, pageSize
}

// ===== 消息配置 =====

func (r *messageRepository) ConfigPage(ctx context.Context, page, pageSize int, keyword string, messageType *string, status *int16) ([]*model.SystemMessageConfig, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	q := r.db.WithContext(ctx).Model(&model.SystemMessageConfig{})
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("title LIKE ? OR supplier_type LIKE ? OR remark LIKE ?", kw, kw, kw)
	}
	if messageType != nil && *messageType != "" {
		q = q.Where("message_type = ?", *messageType)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.SystemMessageConfig
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *messageRepository) ConfigGet(ctx context.Context, id int64) (*model.SystemMessageConfig, error) {
	var m model.SystemMessageConfig
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *messageRepository) ConfigCreate(ctx context.Context, m *model.SystemMessageConfig) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *messageRepository) ConfigSave(ctx context.Context, m *model.SystemMessageConfig) error {
	return r.db.WithContext(ctx).Model(&model.SystemMessageConfig{}).
		Where("id = ?", m.ID).
		Updates(map[string]any{
			"title":         m.Title,
			"message_type":  m.MessageType,
			"supplier_type": m.SupplierType,
			"config_json":   m.ConfigJson,
			"status":        m.Status,
			"remark":        m.Remark,
			"updated_at":    time.Now(),
		}).Error
}

func (r *messageRepository) ConfigDelete(ctx context.Context, id int64) (int64, error) {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.SystemMessageConfig{})
	return res.RowsAffected, res.Error
}

// ===== 消息常量 =====

func (r *messageRepository) KeyPage(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemMessageKey, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	q := r.db.WithContext(ctx).Model(&model.SystemMessageKey{})
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("name LIKE ? OR message_key LIKE ? OR remark LIKE ?", kw, kw, kw)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.SystemMessageKey
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *messageRepository) KeyGet(ctx context.Context, id int64) (*model.SystemMessageKey, error) {
	var m model.SystemMessageKey
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *messageRepository) KeyCreate(ctx context.Context, m *model.SystemMessageKey) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *messageRepository) KeySave(ctx context.Context, m *model.SystemMessageKey) error {
	return r.db.WithContext(ctx).Model(&model.SystemMessageKey{}).
		Where("id = ?", m.ID).
		Updates(map[string]any{
			"name":        m.Name,
			"message_key": m.MessageKey,
			"remark":      m.Remark,
			"updated_at":  time.Now(),
		}).Error
}

func (r *messageRepository) KeyDelete(ctx context.Context, id int64) (int64, error) {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.SystemMessageKey{})
	return res.RowsAffected, res.Error
}

func (r *messageRepository) KeyExists(ctx context.Context, messageKey string, excludeID int64) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.SystemMessageKey{}).Where("message_key = ?", messageKey)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

// ===== 消息模板 =====

func (r *messageRepository) TemplatePage(ctx context.Context, page, pageSize int, keyword string, messageType *string, messageKeyID *int64, status *int16) ([]*model.SystemMessageTemplate, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	q := r.db.WithContext(ctx).Model(&model.SystemMessageTemplate{})
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("template_name LIKE ? OR message_key LIKE ? OR title LIKE ? OR content LIKE ?", kw, kw, kw, kw)
	}
	if messageType != nil && *messageType != "" {
		q = q.Where("message_type = ?", *messageType)
	}
	if messageKeyID != nil && *messageKeyID > 0 {
		q = q.Where("message_key_id = ?", *messageKeyID)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.SystemMessageTemplate
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *messageRepository) TemplateGet(ctx context.Context, id int64) (*model.SystemMessageTemplate, error) {
	var m model.SystemMessageTemplate
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *messageRepository) TemplateCreate(ctx context.Context, m *model.SystemMessageTemplate) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *messageRepository) TemplateSave(ctx context.Context, m *model.SystemMessageTemplate) error {
	return r.db.WithContext(ctx).Model(&model.SystemMessageTemplate{}).
		Where("id = ?", m.ID).
		Updates(map[string]any{
			"template_name":     m.TemplateName,
			"message_config_id": m.MessageConfigID,
			"message_key_id":    m.MessageKeyID,
			"message_key":       m.MessageKey,
			"message_type":      m.MessageType,
			"template_mode":     m.TemplateMode,
			"title":             m.Title,
			"signature":         m.Signature,
			"template_id":       m.TemplateID,
			"content":           m.Content,
			"vars_json":         m.VarsJson,
			"status":            m.Status,
			"remark":            m.Remark,
			"updated_at":        time.Now(),
		}).Error
}

func (r *messageRepository) TemplateDelete(ctx context.Context, id int64) (int64, error) {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.SystemMessageTemplate{})
	return res.RowsAffected, res.Error
}

func (r *messageRepository) TemplateGetEnabled(ctx context.Context, id int64) (*model.SystemMessageTemplate, error) {
	var m model.SystemMessageTemplate
	if err := r.db.WithContext(ctx).Where("id = ? AND status = 1", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *messageRepository) TemplateGetEnabledByKey(ctx context.Context, messageKey, messageType string) (*model.SystemMessageTemplate, error) {
	var m model.SystemMessageTemplate
	err := r.db.WithContext(ctx).
		Where("message_key = ? AND message_type = ? AND status = 1", messageKey, messageType).
		Order("id DESC").Limit(1).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// ===== 发送记录 =====

func (r *messageRepository) LogPage(ctx context.Context, page, pageSize int, keyword string, messageType, messageKey *string, isSuccess *int16) ([]*model.SystemMessageLog, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	q := r.db.WithContext(ctx).Model(&model.SystemMessageLog{})
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("account LIKE ? OR message_template_name LIKE ? OR message_key LIKE ? OR content LIKE ?", kw, kw, kw, kw)
	}
	if messageType != nil && *messageType != "" {
		q = q.Where("message_type = ?", *messageType)
	}
	if messageKey != nil && *messageKey != "" {
		q = q.Where("message_key = ?", *messageKey)
	}
	if isSuccess != nil {
		q = q.Where("is_success = ?", *isSuccess)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.SystemMessageLog
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *messageRepository) LogGet(ctx context.Context, id int64) (*model.SystemMessageLog, error) {
	var m model.SystemMessageLog
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *messageRepository) LogInsert(ctx context.Context, m *model.SystemMessageLog) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *messageRepository) LogRemove(ctx context.Context, id int64) (int64, error) {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.SystemMessageLog{})
	return res.RowsAffected, res.Error
}

func (r *messageRepository) LogClear(ctx context.Context) (int64, error) {
	res := r.db.WithContext(ctx).Exec("TRUNCATE TABLE public.system_message_log RESTART IDENTITY")
	if res.Error != nil {
		// TRUNCATE 受权限限制时退化为 DELETE
		res = r.db.WithContext(ctx).Exec("DELETE FROM public.system_message_log")
	}
	return res.RowsAffected, res.Error
}
