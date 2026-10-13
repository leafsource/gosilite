package role

import (
	"context"
	"sort"
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
	repo    repository.ISystemRoleRepository
	rbac    repository.IRoleRBACRepository
	menuRepo repository.ISystemMenuRepository
}

// NewService builds a role service with explicit dependencies.
func NewService(db *gorm.DB) *Service {
	return &Service{
		repo:    repository.NewSystemRoleRepository(db),
		rbac:    repository.NewRoleRBACRepository(db),
		menuRepo: repository.NewSystemMenuRepository(db),
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

// ===== 授权：菜单权限 + 数据权限（快照 / 整体保存） =====

// AuthorizePerm 角色授权权限项（menuType=F 按钮提升而来）
type AuthorizePerm struct {
	ID       int64
	MenuName string
	Perms    string
	Checked  bool
}

// AuthorizeMenuNode 角色授权菜单行（F 不成行）
type AuthorizeMenuNode struct {
	ID       int64
	ParentID int64
	MenuName string
	MenuType string
	Icon     string
	OrderNum int32
	Checked  bool
	Perms    []*AuthorizePerm
	Children []*AuthorizeMenuNode
}

// AuthorizeSnapshot 授权快照：菜单树勾选态 + 数据范围 + 自定义部门
type AuthorizeSnapshot struct {
	DataScope int32
	Menus     []*AuthorizeMenuNode
	DeptIDs   []int64
}

// Authorize 组装角色授权快照（服务端完成 F 提升与勾选态归一化，前端零拼装）
func (s *Service) Authorize(ctx context.Context, roleID int64) (*AuthorizeSnapshot, error) {
	r, err := s.repo.GetByID(ctx, roleID)
	if err != nil {
		return nil, errorx.New(errno.ErrSystemRoleNotFound, errorx.KV("id", formatID(roleID)))
	}
	menuIDs, err := s.rbac.ListMenuIDs(ctx, roleID)
	if err != nil {
		return nil, err
	}
	deptIDs, err := s.rbac.ListDeptIDs(ctx, roleID)
	if err != nil {
		return nil, err
	}
	all, err := s.menuRepo.Find(ctx)
	if err != nil {
		return nil, err
	}

	checked := make(map[int64]bool, len(menuIDs))
	for _, id := range menuIDs {
		checked[id] = true
	}

	// 索引非删除菜单（父链 + 子表按 sort,id 排序）
	type entry struct {
		menu     *model.SystemMenu
		children []*model.SystemMenu
	}
	entries := map[int64]*entry{}
	var roots []*model.SystemMenu
	for _, m := range all {
		if m.IsRemove == 1 {
			continue
		}
		entries[m.ID] = &entry{menu: m}
	}
	for _, m := range all {
		if m.IsRemove == 1 {
			continue
		}
		p, ok := entries[m.ParentID]
		if !ok || m.ParentID == m.ID {
			roots = append(roots, m)
			continue
		}
		p.children = append(p.children, m)
	}
	sortMenus := func(list []*model.SystemMenu) {
		sort.Slice(list, func(i, j int) bool {
			if list[i].Sort != list[j].Sort {
				return list[i].Sort < list[j].Sort
			}
			return list[i].ID < list[j].ID
		})
	}
	sortMenus(roots)
	for _, e := range entries {
		sortMenus(e.children)
	}

	snapshot := &AuthorizeSnapshot{
		DataScope: int32(r.DataScope),
		Menus:     make([]*AuthorizeMenuNode, 0, len(roots)),
		DeptIDs:   deptIDs,
	}

	var walk func(m *model.SystemMenu, owner *AuthorizeMenuNode) *AuthorizeMenuNode
	walk = func(m *model.SystemMenu, owner *AuthorizeMenuNode) *AuthorizeMenuNode {
		nextOwner := owner
		if m.MenuType == "F" {
			if owner != nil {
				p := &AuthorizePerm{ID: m.ID, MenuName: m.Title, Perms: m.Perms, Checked: checked[m.ID]}
				if p.Checked {
					// 归一化：权限勾选 ⇒ 所属行勾选
					owner.Checked = true
				}
				owner.Perms = append(owner.Perms, p)
			}
		} else {
			n := &AuthorizeMenuNode{
				ID: m.ID, ParentID: m.ParentID, MenuName: m.Title, MenuType: m.MenuType,
				Icon: m.Icon, OrderNum: m.Sort, Checked: checked[m.ID],
			}
			if owner == nil {
				snapshot.Menus = append(snapshot.Menus, n)
			} else {
				owner.Children = append(owner.Children, n)
			}
			nextOwner = n
		}
		for _, c := range entries[m.ID].children {
			walk(c, nextOwner)
		}
		return nextOwner
	}
	for _, root := range roots {
		walk(root, nil)
	}
	return snapshot, nil
}

// SaveAuthorize 一次保存：菜单（服务端补祖先闭包）+ 数据范围 + 自定义部门
func (s *Service) SaveAuthorize(ctx context.Context, roleID int64, menuIDs, deptIDs []int64, dataScope int32) error {
	r, err := s.repo.GetByID(ctx, roleID)
	if err != nil {
		return errorx.New(errno.ErrSystemRoleNotFound, errorx.KV("id", formatID(roleID)))
	}
	if dataScope < 1 || dataScope > 5 {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "dataScope 必须在 1-5"))
	}
	all, err := s.menuRepo.Find(ctx)
	if err != nil {
		return err
	}
	parent := make(map[int64]int64, len(all))
	for _, m := range all {
		if m.IsRemove != 1 {
			parent[m.ID] = m.ParentID
		}
	}
	// 祖先闭包：选中行/权限沿 parentId 上溯（仅限非删除菜单链）
	seen := make(map[int64]bool, len(menuIDs))
	expanded := make([]int64, 0, len(menuIDs))
	for _, id := range menuIDs {
		for {
			if seen[id] {
				break
			}
			seen[id] = true
			expanded = append(expanded, id)
			p, ok := parent[id]
			if !ok || p == id {
				break
			}
			id = p
		}
	}
	if err := s.rbac.ReplaceMenuIDs(ctx, roleID, expanded); err != nil {
		return err
	}
	r.DataScope = int16(dataScope)
	if err := s.repo.Update(ctx, r); err != nil {
		return err
	}
	// 非自定义范围时清空自定义部门
	if dataScope == 5 {
		return s.rbac.ReplaceDeptIDs(ctx, roleID, deptIDs)
	}
	return s.rbac.ReplaceDeptIDs(ctx, roleID, nil)
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
