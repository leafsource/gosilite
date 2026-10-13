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
	// UpdateLeader 设置/清除负责人（leaderUserID=0 清除）
	UpdateLeader(ctx context.Context, deptID, leaderUserID int64) error
	// UserExists 校验负责人用户存在
	UserExists(ctx context.Context, userID int64) (bool, error)
	// LeaderInfo 全量负责人映射（dept_id → leader 信息）
	LeaderInfo(ctx context.Context) ([]DeptLeader, error)
}

// DeptLeader 部门负责人信息（含用户名，join system_user）
type DeptLeader struct {
	DeptID     int64  `gorm:"column:dept_id"`
	LeaderID   int64  `gorm:"column:leader_user_id"`
	LeaderName string `gorm:"column:leader_name"`
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

// UpdateLeader 设置/清除部门负责人（leaderUserID=0 清除）
func (r *departmentRepository) UpdateLeader(ctx context.Context, deptID, leaderUserID int64) error {
	return r.db.WithContext(ctx).
		Model(&model.SystemDept{}).
		Where("id = ?", deptID).
		Update("leader_user_id", leaderUserID).Error
}

// UserExists 校验负责人用户存在
func (r *departmentRepository) UserExists(ctx context.Context, userID int64) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&model.SystemUser{}).
		Where("id = ?", userID).
		Count(&n).Error
	return n > 0, err
}

// LeaderInfo 全量负责人映射数据（dept → leader 用户名）
// 注：裸表名 system_user 在本环境解析异常，必须 public. 限定（与迁移同因，见 spec journey）
func (r *departmentRepository) LeaderInfo(ctx context.Context) ([]DeptLeader, error) {
	var rows []DeptLeader
	err := r.db.WithContext(ctx).
		Raw(`SELECT d.id AS dept_id, d.leader_user_id, COALESCE(u.username, '') AS leader_name
FROM public.system_dept d
LEFT JOIN public.system_user u ON u.id = d.leader_user_id
WHERE d.del_flag = 0`).
		Scan(&rows).Error
	return rows, err
}
