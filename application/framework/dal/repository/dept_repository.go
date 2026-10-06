package repository

import (
	"context"

	"gosilite/application/framework/dal/model"

	"gorm.io/gorm"
)

// IDepartmentRepository 部门仓库
type IDepartmentRepository interface {
	Create(ctx context.Context, dept *model.SystemDept) error
	GetByID(ctx context.Context, id int64) (*model.SystemDept, error)
	Update(ctx context.Context, dept *model.SystemDept) error
	SoftDelete(ctx context.Context, id int64) error
	ListAll(ctx context.Context) ([]*model.SystemDept, error)
	CountChildren(ctx context.Context, parentID int64) (int64, error)
	CountUsers(ctx context.Context, deptID int64) (int64, error)
	SaveAll(ctx context.Context, depts []*model.SystemDept) error
}

type departmentRepository struct {
	db *gorm.DB
}

// NewDepartmentRepository 创建部门仓库
func NewDepartmentRepository(db *gorm.DB) IDepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *departmentRepository) Create(ctx context.Context, dept *model.SystemDept) error {
	return r.db.WithContext(ctx).Create(dept).Error
}

func (r *departmentRepository) GetByID(ctx context.Context, id int64) (*model.SystemDept, error) {
	var dept model.SystemDept
	err := r.db.WithContext(ctx).
		Where("id = ? AND del_flag = 0", id).
		First(&dept).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

func (r *departmentRepository) Update(ctx context.Context, dept *model.SystemDept) error {
	return r.db.WithContext(ctx).Save(dept).Error
}

func (r *departmentRepository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.SystemDept{}).
		Where("id = ?", id).
		Update("del_flag", 2).Error
}

func (r *departmentRepository) ListAll(ctx context.Context) ([]*model.SystemDept, error) {
	var list []*model.SystemDept
	err := r.db.WithContext(ctx).
		Where("del_flag = 0").
		Order("order_num, id").
		Find(&list).Error
	return list, err
}

func (r *departmentRepository) CountChildren(ctx context.Context, parentID int64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&model.SystemDept{}).
		Where("parent_id = ? AND del_flag = 0", parentID).
		Count(&n).Error
	return n, err
}

func (r *departmentRepository) CountUsers(ctx context.Context, deptID int64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&model.SystemUser{}).
		Where("dept_id = ?", deptID).
		Count(&n).Error
	return n, err
}

func (r *departmentRepository) SaveAll(ctx context.Context, depts []*model.SystemDept) error {
	for _, d := range depts {
		if err := r.db.WithContext(ctx).Save(d).Error; err != nil {
			return err
		}
	}
	return nil
}
