package repository

import (
	"context"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/query"

	"gorm.io/gorm"
)

// SystemRoleRepository 仓库接口
type ISystemRoleRepository interface {
	// Create 创建SystemRole
	Create(ctx context.Context, SystemRole *model.SystemRole) error

	// GetByID 根据ID获取SystemRole
	GetByID(ctx context.Context, id int64) (*model.SystemRole, error)

	// Update 更新SystemRole
	Update(ctx context.Context, SystemRole *model.SystemRole) error

	// Delete 删除SystemRole
	Delete(ctx context.Context, id int64) error

	// List 获取SystemRole列表
	List(ctx context.Context, page, pageSize int) ([]*model.SystemRole, int64, error)
}

// SystemRoleRepositoryImpl 仓库实现
type SystemRoleRepositoryImpl struct {
	query *query.Query
}

// NewSystemRoleRepository 创建仓库实例
func NewSystemRoleRepository(db *gorm.DB) ISystemRoleRepository {
	return &SystemRoleRepositoryImpl{
		query: query.Use(db),
	}
}

// Create 创建SystemRole
func (r *SystemRoleRepositoryImpl) Create(ctx context.Context, SystemRole *model.SystemRole) error {
	return r.query.SystemRole.WithContext(ctx).Create(SystemRole)
}

// GetByID 根据ID获取SystemRole
func (r *SystemRoleRepositoryImpl) GetByID(ctx context.Context, id int64) (*model.SystemRole, error) {
	entity, err := r.query.SystemRole.Where(r.query.SystemRole.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// Update 更新SystemRole
func (r *SystemRoleRepositoryImpl) Update(ctx context.Context, SystemRole *model.SystemRole) error {
	return r.query.SystemRole.WithContext(ctx).Save(SystemRole)
}

// Delete 删除SystemRole
func (r *SystemRoleRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.query.SystemRole.WithContext(ctx).Where(r.query.SystemRole.ID.Eq(id)).Delete()
	return err
}

// List 获取SystemRole列表
func (r *SystemRoleRepositoryImpl) List(ctx context.Context, page, pageSize int) ([]*model.SystemRole, int64, error) {
	// 计算总数
	total, err := r.query.SystemRole.WithContext(ctx).Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	entities, err := r.query.SystemRole.WithContext(ctx).Offset(offset).Limit(pageSize).Find()
	if err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}
