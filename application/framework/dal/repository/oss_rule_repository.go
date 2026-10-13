package repository

import (
	"context"
	"time"

	"gosilite/application/framework/dal/model"

	"gorm.io/gorm"
)

// IOssRuleRepository OSS处理规则仓储
type IOssRuleRepository interface {
	Page(ctx context.Context, page, pageSize int, ruleName, domain string, status, isDefault *string) ([]*model.SystemOssRule, int64, error)
	GetByID(ctx context.Context, id int64) (*model.SystemOssRule, error)
	// ExistsDomainName 同域名下规则名唯一（excludeID 用于编辑自身）
	ExistsDomainName(ctx context.Context, domain, ruleName string, excludeID int64) (bool, error)
	Create(ctx context.Context, m *model.SystemOssRule) error
	// Save 仅更新编辑允许变更的列
	Save(ctx context.Context, m *model.SystemOssRule) error
	SoftDelete(ctx context.Context, ids []int64) error

	// ListEnabled 启用规则快照（rule_sort, created_at 升序），供 ResolveURLs 缓存
	ListEnabled(ctx context.Context) ([]*model.SystemOssRule, error)
	// DisableOverwriteForDomain 同域名默认规则覆盖置 N（changeOverwrite 联动）
	DisableOverwriteForDomain(ctx context.Context, domain string) error
	// UpdateOverwrite 单规则覆盖开关
	UpdateOverwrite(ctx context.Context, id int64, isOverwrite string) error
}

type ossRuleRepository struct{ db *gorm.DB }

// NewOssRuleRepository 创建规则仓储
func NewOssRuleRepository(db *gorm.DB) IOssRuleRepository {
	return &ossRuleRepository{db: db}
}

func (r *ossRuleRepository) Page(ctx context.Context, page, pageSize int, ruleName, domain string, status, isDefault *string) ([]*model.SystemOssRule, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}
	q := r.db.WithContext(ctx).Model(&model.SystemOssRule{}).Where("del_flag = 0")
	if ruleName != "" {
		q = q.Where("rule_name LIKE ?", "%"+ruleName+"%")
	}
	if domain != "" {
		q = q.Where("domain LIKE ?", "%"+domain+"%")
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if isDefault != nil {
		q = q.Where("is_default = ?", *isDefault)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.SystemOssRule
	err := q.Order("rule_sort ASC, created_at ASC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *ossRuleRepository) GetByID(ctx context.Context, id int64) (*model.SystemOssRule, error) {
	var m model.SystemOssRule
	err := r.db.WithContext(ctx).Where("id = ? AND del_flag = 0", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ossRuleRepository) ExistsDomainName(ctx context.Context, domain, ruleName string, excludeID int64) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.SystemOssRule{}).
		Where("del_flag = 0 AND domain = ? AND rule_name = ?", domain, ruleName)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var count int64
	err := q.Count(&count).Error
	return count > 0, err
}

func (r *ossRuleRepository) Create(ctx context.Context, m *model.SystemOssRule) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *ossRuleRepository) Save(ctx context.Context, m *model.SystemOssRule) error {
	// 不触碰 is_overwrite（由 ChangeOverwrite 单独管）与审计列
	return r.db.WithContext(ctx).Model(&model.SystemOssRule{}).Where("id = ? AND del_flag = 0", m.ID).
		Updates(map[string]any{
			"rule_name":  m.RuleName,
			"domain":     m.Domain,
			"mime_type":  m.MimeType,
			"rule":       m.Rule,
			"is_default": m.IsDefault,
			"status":     m.Status,
			"rule_sort":  m.RuleSort,
			"remark":     m.Remark,
			"updated_at": time.Now(),
		}).Error
}

func (r *ossRuleRepository) SoftDelete(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Model(&model.SystemOssRule{}).
		Where("id IN ?", ids).
		Update("del_flag", 2).Error
}

func (r *ossRuleRepository) ListEnabled(ctx context.Context) ([]*model.SystemOssRule, error) {
	var list []*model.SystemOssRule
	err := r.db.WithContext(ctx).
		Where("del_flag = 0 AND status = '1'").
		Order("rule_sort ASC, created_at ASC").
		Find(&list).Error
	return list, err
}

func (r *ossRuleRepository) DisableOverwriteForDomain(ctx context.Context, domain string) error {
	return r.db.WithContext(ctx).Model(&model.SystemOssRule{}).
		Where("del_flag = 0 AND domain = ? AND is_default = 'Y'", domain).
		Update("is_overwrite", "N").Error
}

func (r *ossRuleRepository) UpdateOverwrite(ctx context.Context, id int64, isOverwrite string) error {
	return r.db.WithContext(ctx).Model(&model.SystemOssRule{}).
		Where("id = ? AND del_flag = 0", id).
		Updates(map[string]any{"is_overwrite": isOverwrite, "updated_at": time.Now()}).Error
}
