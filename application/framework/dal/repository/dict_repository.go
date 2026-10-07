package repository

import (
	"context"

	"gosilite/application/framework/dal/model"

	"gorm.io/gorm"
)

// IDictRepository 字典仓储
type IDictRepository interface {
	// PageTypes groupID：nil=全部，0=未分组，>0=指定分组
	PageTypes(ctx context.Context, page, pageSize int, keyword string, groupID *int64) ([]*model.SystemDictType, int64, error)
	GetTypeByID(ctx context.Context, id int64) (*model.SystemDictType, error)
	ExistsType(ctx context.Context, dictType string, excludeID int64) (bool, error)
	CountDataByType(ctx context.Context, dictType string) (int64, error)
	CreateType(ctx context.Context, m *model.SystemDictType) error
	SaveType(ctx context.Context, m *model.SystemDictType) error
	SoftDeleteType(ctx context.Context, id int64) error

	ListGroups(ctx context.Context) ([]*model.SystemDictGroup, error)
	GetGroupByID(ctx context.Context, id int64) (*model.SystemDictGroup, error)
	ExistsGroupName(ctx context.Context, groupName string, excludeID int64) (bool, error)
	CountTypesByGroup(ctx context.Context, groupID int64) (int64, error)
	CreateGroup(ctx context.Context, m *model.SystemDictGroup) error
	SaveGroup(ctx context.Context, m *model.SystemDictGroup) error
	SoftDeleteGroup(ctx context.Context, id int64) error

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

func (r *dictRepository) PageTypes(ctx context.Context, page, pageSize int, keyword string, groupID *int64) ([]*model.SystemDictType, int64, error) {
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
	if groupID != nil {
		q = q.Where("group_id = ?", *groupID)
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

func (r *dictRepository) ListGroups(ctx context.Context) ([]*model.SystemDictGroup, error) {
	var list []*model.SystemDictGroup
	err := r.db.WithContext(ctx).Where("del_flag = 0").
		Order("order_num, id").Find(&list).Error
	return list, err
}

func (r *dictRepository) GetGroupByID(ctx context.Context, id int64) (*model.SystemDictGroup, error) {
	var m model.SystemDictGroup
	err := r.db.WithContext(ctx).Where("id = ? AND del_flag = 0", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *dictRepository) ExistsGroupName(ctx context.Context, groupName string, excludeID int64) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.SystemDictGroup{}).
		Where("group_name = ? AND del_flag = 0 AND id <> ?", groupName, excludeID).Count(&n).Error
	return n > 0, err
}

func (r *dictRepository) CountTypesByGroup(ctx context.Context, groupID int64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.SystemDictType{}).
		Where("group_id = ? AND del_flag = 0", groupID).Count(&n).Error
	return n, err
}

func (r *dictRepository) CreateGroup(ctx context.Context, m *model.SystemDictGroup) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *dictRepository) SaveGroup(ctx context.Context, m *model.SystemDictGroup) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *dictRepository) SoftDeleteGroup(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.SystemDictGroup{}).
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
