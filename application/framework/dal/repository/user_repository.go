package repository

import (
	"context"
	"strconv"

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

	// AdminList 管理端分页
	AdminList(ctx context.Context, page, pageSize int, keyword string, deptID int64, state string, deptIDs []int64, allDept bool) ([]*model.SystemUser, int64, error)

	// ExistsAccount 账号唯一
	ExistsAccount(ctx context.Context, account string, excludeID int64) (bool, error)

	// ReplaceRoles 重绑角色
	ReplaceRoles(ctx context.Context, userID int64, roleIDs []int64) error

	// RoleIDs 用户角色
	RoleIDs(ctx context.Context, userID int64) ([]int64, error)

	// UpdateState 状态
	UpdateState(ctx context.Context, userID int64, state string) error

	UpdateAvatar(ctx context.Context, userId int64, avatarKey string) error
}

// SystemUserRepositoryImpl 仓库实现
type SystemUserRepositoryImpl struct {
	query *query.Query
	db    *gorm.DB
}

// NewSystemUserRepository 创建仓库实例
func NewSystemUserRepository(db *gorm.DB) ISystemUserRepository {
	return &SystemUserRepositoryImpl{
		query: query.Use(db),
		db:    db,
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
	roleMenu := r.query.SystemRoleMenu
	userRole := r.query.SystemUserRole

	menu.WithContext(ctx).Select(menu.Perms).
		LeftJoin(roleMenu, menu.ID.EqCol(roleMenu.MenuID)).
		LeftJoin(userRole, roleMenu.RoleID.EqCol(userRole.RoleID)).
		Where(userRole.UserID.Eq(userId), menu.Perms.Neq("")).
		Distinct().
		Pluck(menu.Perms, &permissionCodes)
	return permissionCodes
}

func (r *SystemUserRepositoryImpl) UpdateAvatar(ctx context.Context, userId int64, avatarKey string) error {
	_, err := r.query.SystemUser.WithContext(ctx).Where(r.query.SystemUser.ID.Eq(userId)).Update(r.query.SystemUser.AvatarURL, avatarKey)
	return err
}

func (r *SystemUserRepositoryImpl) AdminList(ctx context.Context, page, pageSize int, keyword string, deptID int64, state string, deptIDs []int64, allDept bool) ([]*model.SystemUser, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}
	q := r.db.WithContext(ctx).Model(&model.SystemUser{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("account LIKE ? OR username LIKE ?", kw, kw)
	}
	if deptID > 0 {
		// 组织架构树选择：含子部门（ancestors 链匹配）
		deptStr := strconv.FormatInt(deptID, 10)
		q = q.Where(
			"dept_id IN (SELECT id FROM public.system_dept WHERE del_flag = 0 AND (id = ? OR ancestors LIKE ? OR ancestors LIKE ?))",
			deptID, "%,"+deptStr+",%", "%,"+deptStr,
		)
	}
	if state != "" {
		q = q.Where("state = ?", state)
	}
	if !allDept {
		q = q.Where("dept_id IN ?", deptIDs)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.SystemUser
	err := q.Order("id").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *SystemUserRepositoryImpl) ExistsAccount(ctx context.Context, account string, excludeID int64) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.SystemUser{}).
		Where("account = ? AND id <> ?", account, excludeID).Count(&n).Error
	return n > 0, err
}

func (r *SystemUserRepositoryImpl) ReplaceRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.SystemUserRole{}).Error; err != nil {
			return err
		}
		for _, rid := range roleIDs {
			row := &model.SystemUserRole{UserID: userID, RoleID: rid}
			if err := tx.Create(row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SystemUserRepositoryImpl) RoleIDs(ctx context.Context, userID int64) ([]int64, error) {
	var ids []int64
	err := r.db.WithContext(ctx).Model(&model.SystemUserRole{}).
		Where("user_id = ? AND is_remove = 0", userID).
		Distinct("role_id").
		Pluck("role_id", &ids).Error
	return ids, err
}

func (r *SystemUserRepositoryImpl) UpdateState(ctx context.Context, userID int64, state string) error {
	return r.db.WithContext(ctx).Model(&model.SystemUser{}).
		Where("id = ?", userID).Update("state", state).Error
}
