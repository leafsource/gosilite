package repository

import (
	"context"

	"gosilite/application/framework/dal/model"

	"gorm.io/gorm"
)

// IDictRepository 字典仓储
type IDictRepository interface {
	PageTypes(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemDictType, int64, error)
	GetTypeByID(ctx context.Context, id int64) (*model.SystemDictType, error)
	ExistsType(ctx context.Context, dictType string, excludeID int64) (bool, error)
	CountDataByType(ctx context.Context, dictType string) (int64, error)
	CreateType(ctx context.Context, m *model.SystemDictType) error
	SaveType(ctx context.Context, m *model.SystemDictType) error
	SoftDeleteType(ctx context.Context, id int64) error

	PageData(ctx context.Context, dictType string, page, pageSize int) ([]*model.SystemDictData, int64, error)
	Options(ctx context.Context, dictType string) ([]*model.SystemDictData, error)
	GetDataByID(ctx context.Context, id int64) (*model.SystemDictData, error)
	CreateData(ctx context.Context, m *model.SystemDictData) error
	SaveData(ctx context.Context, m *model.SystemDictData) error
	SoftDeleteData(ctx context.Context, id int64) error
}

type dictRepository struct{ db *gorm.DB }

// NewDictRepository 创建字典仓储
func NewDictRepository(db *gorm.DB) IDictRepository {
	return &dictRepository{db: db}
}

func (r *dictRepository) PageTypes(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemDictType, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}
	q := r.db.WithContext(ctx).Model(&model.SystemDictType{}).Where("del_flag = 0")
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("dict_name LIKE ? OR dict_type LIKE ?", kw, kw)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.SystemDictType
	err := q.Order("order_num, id").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *dictRepository) GetTypeByID(ctx context.Context, id int64) (*model.SystemDictType, error) {
	var m model.SystemDictType
	err := r.db.WithContext(ctx).Where("id = ? AND del_flag = 0", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *dictRepository) ExistsType(ctx context.Context, dictType string, excludeID int64) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.SystemDictType{}).
		Where("dict_type = ? AND del_flag = 0 AND id <> ?", dictType, excludeID).Count(&n).Error
	return n > 0, err
}

func (r *dictRepository) CountDataByType(ctx context.Context, dictType string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.SystemDictData{}).
		Where("dict_type = ? AND del_flag = 0", dictType).Count(&n).Error
	return n, err
}

func (r *dictRepository) CreateType(ctx context.Context, m *model.SystemDictType) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *dictRepository) SaveType(ctx context.Context, m *model.SystemDictType) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *dictRepository) SoftDeleteType(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.SystemDictType{}).
		Where("id = ?", id).Update("del_flag", 2).Error
}

func (r *dictRepository) PageData(ctx context.Context, dictType string, page, pageSize int) ([]*model.SystemDictData, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 500 {
		pageSize = 50
	}
	q := r.db.WithContext(ctx).Model(&model.SystemDictData{}).Where("del_flag = 0")
	if dictType != "" {
		q = q.Where("dict_type = ?", dictType)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.SystemDictData
	err := q.Order("list_order, id").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *dictRepository) Options(ctx context.Context, dictType string) ([]*model.SystemDictData, error) {
	var list []*model.SystemDictData
	err := r.db.WithContext(ctx).
		Where("dict_type = ? AND del_flag = 0 AND status = 0", dictType).
		Order("list_order, id").
		Find(&list).Error
	return list, err
}

func (r *dictRepository) GetDataByID(ctx context.Context, id int64) (*model.SystemDictData, error) {
	var m model.SystemDictData
	err := r.db.WithContext(ctx).Where("id = ? AND del_flag = 0", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *dictRepository) CreateData(ctx context.Context, m *model.SystemDictData) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *dictRepository) SaveData(ctx context.Context, m *model.SystemDictData) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *dictRepository) SoftDeleteData(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.SystemDictData{}).
		Where("id = ?", id).Update("del_flag", 2).Error
}
