package repository

import (
	"context"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/query"

	"gorm.io/gorm"
)

// SystemUserRepository 仓库接口
type ISystemUserRepository interface {
	// Create 创建SystemUser
	Create(ctx context.Context, SystemUser *model.SystemUser) error

	// GetByID 根据ID获取SystemUser
	GetByID(ctx context.Context, id int64) (*model.SystemUser, error)

	// Update 更新SystemUser
	Update(ctx context.Context, SystemUser *model.SystemUser) error

	// Delete 删除SystemUser
	Delete(ctx context.Context, id int64) error

	// List 获取SystemUser列表
	List(ctx context.Context, page, pageSize int) ([]*model.SystemUser, int64, error)

	GetByAccount(ctx context.Context, account string) (*model.SystemUser, error)

	FindRoleCodes(ctx context.Context, userId int64) []string

	FindPermissionCodes(ctx context.Context, userId int64) []string

	UpdateAvatar(ctx context.Context, userId int64, avatarKey string) error
}

// SystemUserRepositoryImpl 仓库实现
type SystemUserRepositoryImpl struct {
	query *query.Query
}

// NewSystemUserRepository 创建仓库实例
func NewSystemUserRepository(db *gorm.DB) ISystemUserRepository {
	return &SystemUserRepositoryImpl{
		query: query.Use(db),
	}
}

// Create 创建SystemUser
func (r *SystemUserRepositoryImpl) Create(ctx context.Context, SystemUser *model.SystemUser) error {
	return r.query.SystemUser.WithContext(ctx).Create(SystemUser)
}

// GetByID 根据ID获取SystemUser
func (r *SystemUserRepositoryImpl) GetByID(ctx context.Context, id int64) (*model.SystemUser, error) {
	entity, err := r.query.SystemUser.Where(r.query.SystemUser.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// Update 更新SystemUser
func (r *SystemUserRepositoryImpl) Update(ctx context.Context, SystemUser *model.SystemUser) error {
	return r.query.SystemUser.WithContext(ctx).Save(SystemUser)
}

// Delete 删除SystemUser
func (r *SystemUserRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.query.SystemUser.WithContext(ctx).Where(r.query.SystemUser.ID.Eq(id)).Delete()
	return err
}

// List 获取SystemUser列表
func (r *SystemUserRepositoryImpl) List(ctx context.Context, page, pageSize int) ([]*model.SystemUser, int64, error) {
	// 计算总数
	total, err := r.query.SystemUser.WithContext(ctx).Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	entities, err := r.query.SystemUser.WithContext(ctx).Offset(offset).Limit(pageSize).Find()
	if err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// GetByAccount 根据账号获取SystemUser
func (r *SystemUserRepositoryImpl) GetByAccount(ctx context.Context, account string) (*model.SystemUser, error) {
	entity, err := r.query.SystemUser.Where(r.query.SystemUser.Account.Eq(account)).First()
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *SystemUserRepositoryImpl) FindRoleCodes(ctx context.Context, userId int64) []string {
	var roleCodes []string

	role := r.query.SystemRole
	userRole := r.query.SystemUserRole

	role.WithContext(ctx).LeftJoin(userRole, role.ID.EqCol(userRole.RoleID)).
		Where(userRole.UserID.Eq(userId)).
		Pluck(role.Code, &roleCodes)
	return roleCodes
}

func (r *SystemUserRepositoryImpl) FindPermissionCodes(ctx context.Context, userId int64) []string {
	var permissionCodes []string

	menu := r.query.SystemMenu
	role := r.query.SystemRole
	permission := r.query.SystemRolePermission
	userRole := r.query.SystemUserRole

	menu.WithContext(ctx).Select(menu.PermissionCode).
		LeftJoin(permission, menu.ID.EqCol(permission.PermissionID)).
		LeftJoin(role, permission.RoleID.EqCol(role.ID)).
		LeftJoin(userRole, role.ID.EqCol(userRole.RoleID)).
		Where(userRole.UserID.Eq(userId)).
		Pluck(menu.PermissionCode, &permissionCodes)
	return permissionCodes
}

func (r *SystemUserRepositoryImpl) UpdateAvatar(ctx context.Context, userId int64, avatarKey string) error {
	_, err := r.query.SystemUser.WithContext(ctx).Where(r.query.SystemUser.ID.Eq(userId)).Update(r.query.SystemUser.AvatarURL, avatarKey)
	return err
}
