package repository

import (
	"context"

	"gosilite/application/framework/dal/model"

	"gorm.io/gorm"
)

// IRoleRBACRepository 角色菜单/数据范围仓库
type IRoleRBACRepository interface {
	ReplaceMenuIDs(ctx context.Context, roleID int64, menuIDs []int64) error
	ListMenuIDs(ctx context.Context, roleID int64) ([]int64, error)
	ReplaceDeptIDs(ctx context.Context, roleID int64, deptIDs []int64) error
	ListDeptIDs(ctx context.Context, roleID int64) ([]int64, error)
	ListMenuIDsByUser(ctx context.Context, userID int64) ([]int64, error)
	FindRolesByUser(ctx context.Context, userID int64) ([]*model.SystemRole, error)
	ListRoleDepts(ctx context.Context, roleIDs []int64) (map[int64][]int64, error)
}

type roleRBACRepository struct {
	db *gorm.DB
}

// NewRoleRBACRepository 创建角色 RBAC 关联仓库
func NewRoleRBACRepository(db *gorm.DB) IRoleRBACRepository {
	return &roleRBACRepository{db: db}
}

func (r *roleRBACRepository) ReplaceMenuIDs(ctx context.Context, roleID int64, menuIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.SystemRoleMenu{}).Error; err != nil {
			return err
		}
		for _, mid := range menuIDs {
			if err := tx.Create(&model.SystemRoleMenu{RoleID: roleID, MenuID: mid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *roleRBACRepository) ListMenuIDs(ctx context.Context, roleID int64) ([]int64, error) {
	var ids []int64
	err := r.db.WithContext(ctx).Model(&model.SystemRoleMenu{}).
		Where("role_id = ?", roleID).Pluck("menu_id", &ids).Error
	return ids, err
}

func (r *roleRBACRepository) ReplaceDeptIDs(ctx context.Context, roleID int64, deptIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.SystemRoleDept{}).Error; err != nil {
			return err
		}
		for _, did := range deptIDs {
			if err := tx.Create(&model.SystemRoleDept{RoleID: roleID, DeptID: did}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *roleRBACRepository) ListDeptIDs(ctx context.Context, roleID int64) ([]int64, error) {
	var ids []int64
	err := r.db.WithContext(ctx).Model(&model.SystemRoleDept{}).
		Where("role_id = ?", roleID).Pluck("dept_id", &ids).Error
	return ids, err
}

func (r *roleRBACRepository) ListMenuIDsByUser(ctx context.Context, userID int64) ([]int64, error) {
	var ids []int64
	err := r.db.WithContext(ctx).
		Model(&model.SystemRoleMenu{}).
		Where("role_id IN (?)", r.db.Model(&model.SystemUserRole{}).
			Select("role_id").Where("user_id = ? AND is_remove = 0", userID)).
		Pluck("menu_id", &ids).Error
	return ids, err
}

func (r *roleRBACRepository) FindRolesByUser(ctx context.Context, userID int64) ([]*model.SystemRole, error) {
	var roles []*model.SystemRole
	err := r.db.WithContext(ctx).
		Model(&model.SystemRole{}).
		Where("id IN (?) AND is_remove = 0",
			r.db.Model(&model.SystemUserRole{}).Select("role_id").Where("user_id = ? AND is_remove = 0", userID)).
		Find(&roles).Error
	return roles, err
}

func (r *roleRBACRepository) ListRoleDepts(ctx context.Context, roleIDs []int64) (map[int64][]int64, error) {
	result := map[int64][]int64{}
	if len(roleIDs) == 0 {
		return result, nil
	}
	type row struct {
		RoleID int64
		DeptID int64
	}
	var rows []row
	err := r.db.WithContext(ctx).Model(&model.SystemRoleDept{}).
		Where("role_id IN ?", roleIDs).
		Select("role_id, dept_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, v := range rows {
		result[v.RoleID] = append(result[v.RoleID], v.DeptID)
	}
	return result, nil
}
