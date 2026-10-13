package repository

import (
	"context"
	"time"

	"gosilite/application/framework/dal/model"

	"gorm.io/gorm"
)

// OssFileFilter 文件台账查询条件
type OssFileFilter struct {
	Page         int
	PageSize     int
	// FileName 对象 key 模糊
	FileName     string
	OriginalName string
	// 小写后缀列表（服务层已归一小写）
	FileSuffix []string
	CreateBy   string
	// Service 服务商 config_key 精确
	Service   string
	BeginTime string // yyyy-MM-dd HH:mm:ss，空=不限
	EndTime   string
}

// IOssFileRepository OSS文件台账仓储
type IOssFileRepository interface {
	Page(ctx context.Context, f OssFileFilter) ([]*model.SystemOss, int64, error)
	GetByID(ctx context.Context, id int64) (*model.SystemOss, error)
	// GetByIDs 批量取（删除前校验锁）
	GetByIDs(ctx context.Context, ids []int64) ([]*model.SystemOss, error)
	Create(ctx context.Context, m *model.SystemOss) error
	// SaveEditable 仅更新 original_name / is_lock
	SaveEditable(ctx context.Context, m *model.SystemOss) error
	SoftDelete(ctx context.Context, ids []int64) error
}

type ossFileRepository struct{ db *gorm.DB }

// NewOssFileRepository 创建文件台账仓储
func NewOssFileRepository(db *gorm.DB) IOssFileRepository {
	return &ossFileRepository{db: db}
}

func (r *ossFileRepository) buildFilter(ctx context.Context, f OssFileFilter) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&model.SystemOss{}).Where("del_flag = 0")
	if f.FileName != "" {
		q = q.Where("file_name LIKE ?", "%"+f.FileName+"%")
	}
	if f.OriginalName != "" {
		q = q.Where("original_name LIKE ?", "%"+f.OriginalName+"%")
	}
	if f.Service != "" {
		q = q.Where("service = ?", f.Service)
	}
	if len(f.FileSuffix) > 0 {
		q = q.Where("file_suffix IN ?", f.FileSuffix)
	}
	if f.CreateBy != "" {
		q = q.Where("create_by LIKE ?", "%"+f.CreateBy+"%")
	}
	if f.BeginTime != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", f.BeginTime, time.Local); err == nil {
			q = q.Where("created_at >= ?", t)
		}
	}
	if f.EndTime != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", f.EndTime, time.Local); err == nil {
			q = q.Where("created_at <= ?", t)
		}
	}
	return q
}

func (r *ossFileRepository) Page(ctx context.Context, f OssFileFilter) ([]*model.SystemOss, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 200 {
		f.PageSize = 10
	}
	q := r.buildFilter(ctx, f)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.SystemOss
	err := q.Order("id DESC").Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

func (r *ossFileRepository) GetByID(ctx context.Context, id int64) (*model.SystemOss, error) {
	var m model.SystemOss
	err := r.db.WithContext(ctx).Where("id = ? AND del_flag = 0", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ossFileRepository) GetByIDs(ctx context.Context, ids []int64) ([]*model.SystemOss, error) {
	var list []*model.SystemOss
	err := r.db.WithContext(ctx).Where("id IN ? AND del_flag = 0", ids).Find(&list).Error
	return list, err
}

func (r *ossFileRepository) Create(ctx context.Context, m *model.SystemOss) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *ossFileRepository) SaveEditable(ctx context.Context, m *model.SystemOss) error {
	return r.db.WithContext(ctx).Model(&model.SystemOss{}).Where("id = ? AND del_flag = 0", m.ID).
		Updates(map[string]any{
			"original_name": m.OriginalName,
			"is_lock":       m.IsLock,
			"updated_at":    time.Now(),
		}).Error
}

func (r *ossFileRepository) SoftDelete(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Model(&model.SystemOss{}).
		Where("id IN ?", ids).
		Update("del_flag", 2).Error
}
