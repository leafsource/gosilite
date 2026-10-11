package repository

import (
	"context"
	"time"

	"gosilite/application/framework/dal/model"

	"gorm.io/gorm"
)

// IOssConfigRepository 对象存储配置仓储
type IOssConfigRepository interface {
	Page(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemOssConfig, int64, error)
	GetByID(ctx context.Context, id int64) (*model.SystemOssConfig, error)
	ExistsKey(ctx context.Context, key string, excludeID int64) (bool, error)
	Create(ctx context.Context, m *model.SystemOssConfig) error
	// SaveEditable 除 config_key/status/del_flag 外的可编辑列
	SaveEditable(ctx context.Context, m *model.SystemOssConfig) error
	SoftDelete(ctx context.Context, ids []int64) error
	// ClearDefault 全部置非默认（ChangeStatus 事务第一步）
	ClearDefault(ctx context.Context) error
	// SetDefault 目标置默认
	SetDefault(ctx context.Context, id int64) error
	// ChangeDefaultTx 事务：全部置 0 → 目标置 1
	ChangeDefaultTx(ctx context.Context, id int64) error
}

type ossConfigRepository struct{ db *gorm.DB }

// NewOssConfigRepository 创建对象存储配置仓储
func NewOssConfigRepository(db *gorm.DB) IOssConfigRepository {
	return &ossConfigRepository{db: db}
}

func (r *ossConfigRepository) Page(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemOssConfig, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}
	q := r.db.WithContext(ctx).Model(&model.SystemOssConfig{}).Where("del_flag = 0")
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("config_key LIKE ? OR endpoint LIKE ? OR bucket_name LIKE ?", kw, kw, kw)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.SystemOssConfig
	err := q.Order("status DESC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *ossConfigRepository) GetByID(ctx context.Context, id int64) (*model.SystemOssConfig, error) {
	var m model.SystemOssConfig
	err := r.db.WithContext(ctx).Where("id = ? AND del_flag = 0", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ossConfigRepository) ExistsKey(ctx context.Context, key string, excludeID int64) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.SystemOssConfig{}).
		Where("del_flag = 0 AND config_key = ?", key)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var count int64
	err := q.Count(&count).Error
	return count > 0, err
}

func (r *ossConfigRepository) Create(ctx context.Context, m *model.SystemOssConfig) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *ossConfigRepository) SaveEditable(ctx context.Context, m *model.SystemOssConfig) error {
	return r.db.WithContext(ctx).Model(&model.SystemOssConfig{}).Where("id = ? AND del_flag = 0", m.ID).
		Updates(map[string]any{
			"access_key":    m.AccessKey,
			"secret_key":    m.SecretKey,
			"bucket_name":   m.BucketName,
			"prefix":        m.Prefix,
			"endpoint":      m.Endpoint,
			"domain":        m.Domain,
			"is_https":      m.IsHTTPS,
			"region":        m.Region,
			"access_policy": m.AccessPolicy,
			"ext1":          m.Ext1,
			"remark":        m.Remark,
			"updated_at":    time.Now(),
		}).Error
}

func (r *ossConfigRepository) SoftDelete(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Model(&model.SystemOssConfig{}).
		Where("id IN ?", ids).
		Update("del_flag", 2).Error
}

func (r *ossConfigRepository) ClearDefault(ctx context.Context) error {
	return r.db.WithContext(ctx).Model(&model.SystemOssConfig{}).
		Where("del_flag = 0").
		Update("status", "0").Error
}

func (r *ossConfigRepository) SetDefault(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.SystemOssConfig{}).
		Where("id = ? AND del_flag = 0", id).
		Update("status", "1").Error
}

func (r *ossConfigRepository) ChangeDefaultTx(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.SystemOssConfig{}).
			Where("del_flag = 0").
			Update("status", "0").Error; err != nil {
			return err
		}
		res := tx.Model(&model.SystemOssConfig{}).
			Where("id = ? AND del_flag = 0", id).
			Update("status", "1")
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}
