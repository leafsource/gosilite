package repository

import (
	"context"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/query"

	"gorm.io/gorm"
)

// SystemMenuRepository 仓库接口
type ISystemMenuRepository interface {
	// Create 创建SystemMenu
	Create(ctx context.Context, SystemMenu *model.SystemMenu) error

	// GetByID 根据ID获取SystemMenu
	GetByID(ctx context.Context, id int64) (*model.SystemMenu, error)

	// Update 更新SystemMenu
	Update(ctx context.Context, SystemMenu *model.SystemMenu) error

	// Delete 删除SystemMenu
	Delete(ctx context.Context, id int64) error

	// List 获取SystemMenu列表
	List(ctx context.Context, page, pageSize int) ([]*model.SystemMenu, int64, error)

	// Find 获取SystemMenu列表
	Find(ctx context.Context) ([]*model.SystemMenu, error)

	// FindMenuIDsByUser 用户经 user_role -> role_menu 关联的菜单 id 集合
	FindMenuIDsByUser(ctx context.Context, userID int64) ([]int64, error)
}

// SystemMenuRepositoryImpl 仓库实现
type SystemMenuRepositoryImpl struct {
	query *query.Query
}

// NewSystemMenuRepository 创建仓库实例
func NewSystemMenuRepository(db *gorm.DB) ISystemMenuRepository {
	return &SystemMenuRepositoryImpl{
		query: query.Use(db),
	}
}

// Create 创建SystemMenu
func (r *SystemMenuRepositoryImpl) Create(ctx context.Context, SystemMenu *model.SystemMenu) error {
	return r.query.SystemMenu.WithContext(ctx).Create(SystemMenu)
}

// GetByID 根据ID获取SystemMenu
func (r *SystemMenuRepositoryImpl) GetByID(ctx context.Context, id int64) (*model.SystemMenu, error) {
	entity, err := r.query.SystemMenu.Where(r.query.SystemMenu.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// Update 更新SystemMenu
func (r *SystemMenuRepositoryImpl) Update(ctx context.Context, SystemMenu *model.SystemMenu) error {
	return r.query.SystemMenu.WithContext(ctx).Save(SystemMenu)
}

// Delete 删除SystemMenu
func (r *SystemMenuRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.query.SystemMenu.WithContext(ctx).Where(r.query.SystemMenu.ID.Eq(id)).Delete()
	return err
}

// List 获取SystemMenu列表
func (r *SystemMenuRepositoryImpl) List(ctx context.Context, page, pageSize int) ([]*model.SystemMenu, int64, error) {
	// 计算总数
	total, err := r.query.SystemMenu.WithContext(ctx).Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	entities, err := r.query.SystemMenu.WithContext(ctx).Offset(offset).Limit(pageSize).Find()
	if err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// Find 获取SystemMenu列表
func (r *SystemMenuRepositoryImpl) Find(ctx context.Context) ([]*model.SystemMenu, error) {
	entities, err := r.query.SystemMenu.WithContext(ctx).Find()
	if err != nil {
		return nil, err
	}
	return entities, nil
}

// FindMenuIDsByUser 用户经 user_role -> role_menu 关联的菜单 id 集合
func (r *SystemMenuRepositoryImpl) FindMenuIDsByUser(ctx context.Context, userID int64) ([]int64, error) {
	var ids []int64

	roleMenu := r.query.SystemRoleMenu
	userRole := r.query.SystemUserRole

	err := roleMenu.WithContext(ctx).
		LeftJoin(userRole, roleMenu.RoleID.EqCol(userRole.RoleID)).
		Where(userRole.UserID.Eq(userID)).
		Distinct().
		Pluck(roleMenu.MenuID, &ids)
	if err != nil {
		return nil, err
	}
	return ids, nil
}
