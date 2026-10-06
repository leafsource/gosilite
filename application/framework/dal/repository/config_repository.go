package repository

import (
	"context"

	"gosilite/application/framework/dal/model"

	"gorm.io/gorm"
)

// IConfigRepository 参数配置仓储
type IConfigRepository interface {
	Page(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemConfig, int64, error)
	GetByID(ctx context.Context, id int64) (*model.SystemConfig, error)
	GetByKey(ctx context.Context, key string) (*model.SystemConfig, error)
	ListByType(ctx context.Context, configType string) ([]*model.SystemConfig, error)
	ExistsKey(ctx context.Context, key string, excludeID int64) (bool, error)
	Create(ctx context.Context, m *model.SystemConfig) error
	Save(ctx context.Context, m *model.SystemConfig) error
	SoftDelete(ctx context.Context, id int64) error
}

type configRepository struct{ db *gorm.DB }

// NewConfigRepository 创建参数仓储
func NewConfigRepository(db *gorm.DB) IConfigRepository {
	return &configRepository{db: db}
}

func (r *configRepository) Page(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemConfig, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}
	q := r.db.WithContext(ctx).Model(&model.SystemConfig{}).Where("del_flag = 0")
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("config_name LIKE ? OR config_key LIKE ?", kw, kw)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.SystemConfig
	err := q.Order("id").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *configRepository) GetByID(ctx context.Context, id int64) (*model.SystemConfig, error) {
	var m model.SystemConfig
	err := r.db.WithContext(ctx).Where("id = ? AND del_flag = 0", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *configRepository) GetByKey(ctx context.Context, key string) (*model.SystemConfig, error) {
	var m model.SystemConfig
	err := r.db.WithContext(ctx).Where("config_key = ? AND del_flag = 0", key).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// ListByType 按类型取启用的参数（config_type=global 等）
func (r *configRepository) ListByType(ctx context.Context, configType string) ([]*model.SystemConfig, error) {
	var list []*model.SystemConfig
	err := r.db.WithContext(ctx).
		Where("config_type = ? AND status = 0 AND del_flag = 0", configType).
		Order("id").
		Find(&list).Error
	return list, err
}

func (r *configRepository) ExistsKey(ctx context.Context, key string, excludeID int64) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.SystemConfig{}).
		Where("config_key = ? AND del_flag = 0 AND id <> ?", key, excludeID).Count(&n).Error
	return n > 0, err
}

func (r *configRepository) Create(ctx context.Context, m *model.SystemConfig) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *configRepository) Save(ctx context.Context, m *model.SystemConfig) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *configRepository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.SystemConfig{}).
		Where("id = ?", id).Update("del_flag", 2).Error
}
