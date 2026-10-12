package message

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

// SystemMessageSVC 门面
var SystemMessageSVC *Service

// Service 消息管理（配置/常量/模板/发送记录 + 发送服务）
type Service struct {
	repo repository.IMessageRepository
}

// NewService 构造
func NewService(repo repository.IMessageRepository) *Service {
	return &Service{repo: repo}
}

// InitService 安装门面
func InitService(repo repository.IMessageRepository) *Service {
	SystemMessageSVC = NewService(repo)
	return SystemMessageSVC
}

// ========== 消息配置 ==========

// ConfigPage 分页（messageType/status 为 nil 或空时不过滤）
func (s *Service) ConfigPage(ctx context.Context, page, pageSize int, keyword string, messageType *string, status *int16) ([]*model.SystemMessageConfig, int64, error) {
	return s.repo.ConfigPage(ctx, page, pageSize, keyword, messageType, status)
}

// ConfigGet 详情
func (s *Service) ConfigGet(ctx context.Context, id int64) (*model.SystemMessageConfig, error) {
	m, err := s.repo.ConfigGet(ctx, id)
	if err != nil {
		return nil, errorx.New(errno.ErrMessageConfigNotFound, errorx.KV("id", formatID(id)))
	}
	return m, nil
}

// ConfigCreate 创建
func (s *Service) ConfigCreate(ctx context.Context, m *model.SystemMessageConfig) error {
	if err := validateConfig(m); err != nil {
		return err
	}
	return s.repo.ConfigCreate(ctx, m)
}

// ConfigUpdate 更新
func (s *Service) ConfigUpdate(ctx context.Context, m *model.SystemMessageConfig) error {
	if _, err := s.repo.ConfigGet(ctx, m.ID); err != nil {
		return errorx.New(errno.ErrMessageConfigNotFound, errorx.KV("id", formatID(m.ID)))
	}
	if err := validateConfig(m); err != nil {
		return err
	}
	return s.repo.ConfigSave(ctx, m)
}

// ConfigDelete 删除；返回影响行数
func (s *Service) ConfigDelete(ctx context.Context, id int64) (int64, error) {
	n, err := s.repo.ConfigDelete(ctx, id)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, errorx.New(errno.ErrMessageConfigNotFound, errorx.KV("id", formatID(id)))
	}
	return n, nil
}

func validateConfig(m *model.SystemMessageConfig) error {
	if strings.TrimSpace(m.Title) == "" {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "标题不能为空"))
	}
	m.Title = strings.TrimSpace(m.Title)
	m.MessageType = strings.ToUpper(strings.TrimSpace(m.MessageType))
	if m.MessageType != "MAIL" && m.MessageType != "SMS" {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "消息类型必须为 MAIL 或 SMS"))
	}
	if strings.TrimSpace(m.SupplierType) == "" {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "供应商标识不能为空"))
	}
	m.SupplierType = strings.ToUpper(strings.TrimSpace(m.SupplierType))
	if m.Status != 0 && m.Status != 1 {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "状态必须为 0 或 1"))
	}
	if strings.TrimSpace(m.ConfigJson) != "" && !json.Valid([]byte(m.ConfigJson)) {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "configJson 不是合法 JSON"))
	}
	return nil
}

// ========== 消息常量 ==========

// KeyPage 分页
func (s *Service) KeyPage(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemMessageKey, int64, error) {
	return s.repo.KeyPage(ctx, page, pageSize, keyword)
}

// KeyGet 详情
func (s *Service) KeyGet(ctx context.Context, id int64) (*model.SystemMessageKey, error) {
	m, err := s.repo.KeyGet(ctx, id)
	if err != nil {
		return nil, errorx.New(errno.ErrMessageKeyNotFound, errorx.KV("id", formatID(id)))
	}
	return m, nil
}

// KeyCreate 创建（message_key 全局唯一）
func (s *Service) KeyCreate(ctx context.Context, m *model.SystemMessageKey) error {
	if err := validateKey(m, 0); err != nil {
		return err
	}
	exists, err := s.repo.KeyExists(ctx, m.MessageKey, 0)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrMessageKeyExists, errorx.KV("key", m.MessageKey))
	}
	return s.repo.KeyCreate(ctx, m)
}

// KeyUpdate 更新
func (s *Service) KeyUpdate(ctx context.Context, m *model.SystemMessageKey) error {
	if _, err := s.repo.KeyGet(ctx, m.ID); err != nil {
		return errorx.New(errno.ErrMessageKeyNotFound, errorx.KV("id", formatID(m.ID)))
	}
	if err := validateKey(m, m.ID); err != nil {
		return err
	}
	exists, err := s.repo.KeyExists(ctx, m.MessageKey, m.ID)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrMessageKeyExists, errorx.KV("key", m.MessageKey))
	}
	return s.repo.KeySave(ctx, m)
}

// KeyDelete 删除；返回影响行数
func (s *Service) KeyDelete(ctx context.Context, id int64) (int64, error) {
	n, err := s.repo.KeyDelete(ctx, id)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, errorx.New(errno.ErrMessageKeyNotFound, errorx.KV("id", formatID(id)))
	}
	return n, nil
}

func validateKey(m *model.SystemMessageKey, _ int64) error {
	if strings.TrimSpace(m.Name) == "" {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "名称不能为空"))
	}
	m.Name = strings.TrimSpace(m.Name)
	if strings.TrimSpace(m.MessageKey) == "" {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "消息KEY不能为空"))
	}
	m.MessageKey = strings.TrimSpace(m.MessageKey)
	return nil
}

// ========== 消息模板 ==========

// TemplatePage 分页
func (s *Service) TemplatePage(ctx context.Context, page, pageSize int, keyword string, messageType *string, messageKeyID *int64, status *int16) ([]*model.SystemMessageTemplate, int64, error) {
	return s.repo.TemplatePage(ctx, page, pageSize, keyword, messageType, messageKeyID, status)
}

// TemplateGet 详情
func (s *Service) TemplateGet(ctx context.Context, id int64) (*model.SystemMessageTemplate, error) {
	m, err := s.repo.TemplateGet(ctx, id)
	if err != nil {
		return nil, errorx.New(errno.ErrMessageTemplateNotFound, errorx.KV("id", formatID(id)))
	}
	return m, nil
}

// TemplateCreate 创建
func (s *Service) TemplateCreate(ctx context.Context, m *model.SystemMessageTemplate) error {
	if err := s.deriveTemplateKey(ctx, m); err != nil {
		return err
	}
	if err := validateTemplate(m); err != nil {
		return err
	}
	return s.repo.TemplateCreate(ctx, m)
}

// TemplateUpdate 更新
func (s *Service) TemplateUpdate(ctx context.Context, m *model.SystemMessageTemplate) error {
	if _, err := s.repo.TemplateGet(ctx, m.ID); err != nil {
		return errorx.New(errno.ErrMessageTemplateNotFound, errorx.KV("id", formatID(m.ID)))
	}
	if err := s.deriveTemplateKey(ctx, m); err != nil {
		return err
	}
	if err := validateTemplate(m); err != nil {
		return err
	}
	return s.repo.TemplateSave(ctx, m)
}

// deriveTemplateKey 按 messageKeyId 从消息常量派生 messageKey（对齐 ruoyi 保存时后端补字段，
// 前端只提交选中的常量 id）
func (s *Service) deriveTemplateKey(ctx context.Context, m *model.SystemMessageTemplate) error {
	if m.MessageKeyID <= 0 {
		return nil
	}
	key, err := s.repo.KeyGet(ctx, m.MessageKeyID)
	if err != nil || key == nil {
		return errorx.New(errno.ErrMessageKeyNotFound, errorx.KV("id", formatID(m.MessageKeyID)))
	}
	m.MessageKey = key.MessageKey
	return nil
}

// TemplateDelete 删除；返回影响行数
func (s *Service) TemplateDelete(ctx context.Context, id int64) (int64, error) {
	n, err := s.repo.TemplateDelete(ctx, id)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, errorx.New(errno.ErrMessageTemplateNotFound, errorx.KV("id", formatID(id)))
	}
	return n, nil
}

func validateTemplate(m *model.SystemMessageTemplate) error {
	if strings.TrimSpace(m.TemplateName) == "" {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "模板名称不能为空"))
	}
	m.TemplateName = strings.TrimSpace(m.TemplateName)
	if m.MessageConfigID <= 0 {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "消息配置不能为空"))
	}
	if m.MessageKeyID <= 0 || strings.TrimSpace(m.MessageKey) == "" {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "消息常量不能为空"))
	}
	m.MessageKey = strings.TrimSpace(m.MessageKey)
	m.MessageType = strings.ToUpper(strings.TrimSpace(m.MessageType))
	if m.MessageType != "MAIL" && m.MessageType != "SMS" {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "消息类型必须为 MAIL 或 SMS"))
	}
	m.TemplateMode = strings.ToUpper(strings.TrimSpace(m.TemplateMode))
	if m.TemplateMode != "TEMPLATE_ID" && m.TemplateMode != "TEMPLATE_CONTENT" {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "模板模式必须为 TEMPLATE_ID 或 TEMPLATE_CONTENT"))
	}
	if m.Status != 0 && m.Status != 1 {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "状态必须为 0 或 1"))
	}
	if strings.TrimSpace(m.VarsJson) != "" && !json.Valid([]byte(m.VarsJson)) {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "varsJson 不是合法 JSON"))
	}
	return nil
}

// ========== 发送记录 ==========

// LogPage 分页
func (s *Service) LogPage(ctx context.Context, page, pageSize int, keyword string, messageType, messageKey *string, isSuccess *int16) ([]*model.SystemMessageLog, int64, error) {
	return s.repo.LogPage(ctx, page, pageSize, keyword, messageType, messageKey, isSuccess)
}

// LogGet 详情
func (s *Service) LogGet(ctx context.Context, id int64) (*model.SystemMessageLog, error) {
	m, err := s.repo.LogGet(ctx, id)
	if err != nil {
		return nil, errorx.New(errno.ErrMessageLogNotFound, errorx.KV("id", formatID(id)))
	}
	return m, nil
}

// LogRemove 删除单条发送记录；返回影响行数
func (s *Service) LogRemove(ctx context.Context, id int64) (int64, error) {
	n, err := s.repo.LogRemove(ctx, id)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, errorx.New(errno.ErrMessageLogNotFound, errorx.KV("id", formatID(id)))
	}
	return n, nil
}

// LogClear 清空发送记录；返回删除条数
func (s *Service) LogClear(ctx context.Context) (int64, error) {
	return s.repo.LogClear(ctx)
}

func formatID(v int64) string {
	return strconv.FormatInt(v, 10)
}
