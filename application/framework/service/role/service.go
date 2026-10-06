package role

import (
	"context"
	"strings"

	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"

	"gorm.io/gorm"
)

// SystemRoleSVC is the package facade for handlers; set once by InitService.
var SystemRoleSVC *Service

// Service holds role business operations.
type Service struct {
	repo repository.ISystemRoleRepository
	rbac repository.IRoleRBACRepository
}

// NewService builds a role service with explicit dependencies.
func NewService(db *gorm.DB) *Service {
	return &Service{
		repo: repository.NewSystemRoleRepository(db),
		rbac: repository.NewRoleRBACRepository(db),
	}
}

// InitService installs the package facade used by handlers.
func InitService(ctx context.Context, db *gorm.DB) *Service {
	SystemRoleSVC = NewService(db)
	return SystemRoleSVC
}

// Page 分页 + 关键字
func (s *Service) Page(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemRole, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}
	return s.repo.ListPage(ctx, page, pageSize, keyword)
}

// CreateRole 创建（校验 code 唯一）
func (s *Service) CreateRole(ctx context.Context, r *model.SystemRole) error {
	if strings.TrimSpace(r.Name) == "" || strings.TrimSpace(r.Code) == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "角色名称/编码不能为空"))
	}
	exists, err := s.repo.ExistsCode(ctx, r.Code, 0)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrSystemRoleCodeExists, errorx.KV("code", r.Code))
	}
	if r.DataScope < 1 || r.DataScope > 5 {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "dataScope 必须在 1-5"))
	}
	r.IsRemove = 0
	return s.repo.Create(ctx, r)
}

// UpdateRole 更新
func (s *Service) UpdateRole(ctx context.Context, r *model.SystemRole) error {
	if _, err := s.repo.GetByID(ctx, r.ID); err != nil {
		return errorx.New(errno.ErrSystemRoleNotFound, errorx.KV("id", formatID(r.ID)))
	}
	exists, err := s.repo.ExistsCode(ctx, r.Code, r.ID)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrSystemRoleCodeExists, errorx.KV("code", r.Code))
	}
	if r.DataScope < 1 || r.DataScope > 5 {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "dataScope 必须在 1-5"))
	}
	return s.repo.Update(ctx, r)
}

// RemoveRole 删除（内置角色不可删）
func (s *Service) RemoveRole(ctx context.Context, id int64) error {
	old, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errorx.New(errno.ErrSystemRoleNotFound, errorx.KV("id", formatID(id)))
	}
	if old.Type == "built-y" {
		return errorx.New(errno.ErrSystemForbidden, errorx.KV("msg", "内置角色不可删除"))
	}
	return s.repo.Delete(ctx, id)
}

// AssignMenus 分配菜单
func (s *Service) AssignMenus(ctx context.Context, roleID int64, menuIDs []int64) error {
	if _, err := s.repo.GetByID(ctx, roleID); err != nil {
		return errorx.New(errno.ErrSystemRoleNotFound, errorx.KV("id", formatID(roleID)))
	}
	return s.rbac.ReplaceMenuIDs(ctx, roleID, menuIDs)
}

// AssignDepts 分配数据范围部门
func (s *Service) AssignDepts(ctx context.Context, roleID int64, deptIDs []int64) error {
	if _, err := s.repo.GetByID(ctx, roleID); err != nil {
		return errorx.New(errno.ErrSystemRoleNotFound, errorx.KV("id", formatID(roleID)))
	}
	return s.rbac.ReplaceDeptIDs(ctx, roleID, deptIDs)
}

// MenuIDs 角色已选菜单
func (s *Service) MenuIDs(ctx context.Context, roleID int64) ([]int64, error) {
	return s.rbac.ListMenuIDs(ctx, roleID)
}

// DeptIDs 角色自定义数据范围
func (s *Service) DeptIDs(ctx context.Context, roleID int64) ([]int64, error) {
	return s.rbac.ListDeptIDs(ctx, roleID)
}

// RBAC 暴露给用户服务做数据权限
func (s *Service) RBAC() repository.IRoleRBACRepository { return s.rbac }

func (s *Service) Create(ctx context.Context, entity *model.SystemRole) error {
	return s.repo.Create(ctx, entity)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*model.SystemRole, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, entity *model.SystemRole) error {
	return s.repo.Update(ctx, entity)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, page, pageSize int) ([]*model.SystemRole, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}

func formatID(v int64) string {
	if v == 0 {
		return "0"
	}
	var b []byte
	n := uint64(v)
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
