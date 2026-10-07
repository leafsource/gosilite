package menu

import (
	"context"
	"sort"
	"strings"

	"gosilite/application/base/ctxutil"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"

	"gorm.io/gorm"
)

// SystemMenuSVC is the package facade for handlers; set once by InitService.
var SystemMenuSVC *Service

// Service holds menu business operations.
type Service struct {
	repo repository.ISystemMenuRepository
}

// NewService builds a menu service with explicit dependencies.
func NewService(db *gorm.DB) *Service {
	return newService(repository.NewSystemMenuRepository(db))
}

// newService 便于单测注入假仓储
func newService(repo repository.ISystemMenuRepository) *Service {
	return &Service{repo: repo}
}

// InitService installs the package facade used by handlers.
func InitService(ctx context.Context, db *gorm.DB) *Service {
	SystemMenuSVC = NewService(db)
	return SystemMenuSVC
}

// TreeNode 管理用菜单树
type TreeNode struct {
	ID        int64       `json:"id"`
	ParentID  int64       `json:"parentId"`
	Label     string      `json:"label"`
	MenuType  string      `json:"menuType"`
	Perms     string      `json:"perms"`
	OrderNum  int32       `json:"orderNum"`
	Visible   bool        `json:"visible"`
	Status    int16       `json:"status"`
	Path      string      `json:"path"`
	Component string      `json:"component"`
	Icon      string      `json:"icon"`
	Children  []*TreeNode `json:"children,omitempty"`
}

// Tree 全量菜单树（管理端）
func (s *Service) Tree(ctx context.Context) ([]*TreeNode, error) {
	all, err := s.repo.Find(ctx)
	if err != nil {
		return nil, err
	}
	nodes := map[int64]*TreeNode{}
	for _, m := range all {
		if m.IsRemove == 1 {
			continue
		}
		nodes[m.ID] = toTreeNode(m)
	}
	var roots []*TreeNode
	for id, n := range nodes {
		if p, ok := nodes[n.ParentID]; ok && id != n.ParentID {
			p.Children = append(p.Children, n)
		} else {
			roots = append(roots, n)
		}
	}
	sortNodes(roots)
	return roots, nil
}

func (s *Service) CreateMenu(ctx context.Context, m *model.SystemMenu) error {
	if strings.TrimSpace(m.Title) == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "菜单名称不能为空"))
	}
	if m.MenuType == "" {
		m.MenuType = "C"
	}
	m.IsRemove = 0
	if m.Visible == "" {
		m.Visible = "show"
	}
	if m.State == "" {
		m.State = "enabled"
	}
	return s.repo.Create(ctx, m)
}

// UpdateMenu 更新
func (s *Service) UpdateMenu(ctx context.Context, m *model.SystemMenu) error {
	if _, err := s.repo.GetByID(ctx, m.ID); err != nil {
		return errorx.New(errno.ErrSystemMenuNotFound, errorx.KV("id", formatID(m.ID)))
	}
	return s.repo.Update(ctx, m)
}

// RemoveMenu 删除（有子菜单则拒绝）
func (s *Service) RemoveMenu(ctx context.Context, id int64) error {
	all, err := s.repo.Find(ctx)
	if err != nil {
		return err
	}
	for _, m := range all {
		if m.ParentID == id && m.IsRemove != 1 {
			return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "存在子菜单，无法删除"))
		}
	}
	return s.repo.Delete(ctx, id)
}

// ListPage 管理端列表（不分页，树表全量）
func (s *Service) ListPage(ctx context.Context) ([]*model.SystemMenu, int64, error) {
	return s.repo.List(ctx, 1, 500)
}

func (s *Service) Create(ctx context.Context, entity *model.SystemMenu) error {
	return s.repo.Create(ctx, entity)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*model.SystemMenu, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, entity *model.SystemMenu) error {
	return s.repo.Update(ctx, entity)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, page, pageSize int) ([]*model.SystemMenu, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}

// AllMenu returns all menus.
func (s *Service) AllMenu(ctx context.Context) ([]*model.SystemMenu, error) {
	return s.repo.Find(ctx)
}

// VisibleMenus 当前登录用户可见菜单（/menu/routes 用）：
// admin 角色返回全量，其余按 user_role -> role_menu 过滤；剔除按钮(F)与已删行。
func (s *Service) VisibleMenus(ctx context.Context) ([]*model.SystemMenu, error) {
	loginUser := ctxutil.GetLoginUser(ctx)
	if loginUser == nil {
		return nil, errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("reason", "not login"))
	}
	all, err := s.repo.Find(ctx)
	if err != nil {
		return nil, err
	}
	if ctxutil.IsSuperRole(loginUser.Roles) {
		return filterRouteMenus(all, nil), nil
	}
	ids, err := s.repo.FindMenuIDsByUser(ctx, loginUser.ID)
	if err != nil {
		return nil, err
	}
	allow := make(map[int64]bool, len(ids))
	for _, id := range ids {
		allow[id] = true
	}
	return filterRouteMenus(all, allow), nil
}

// filterRouteMenus 过滤路由用菜单；allow 为 nil 表示不过滤 id（超管）
func filterRouteMenus(menus []*model.SystemMenu, allow map[int64]bool) []*model.SystemMenu {
	out := make([]*model.SystemMenu, 0, len(menus))
	for _, m := range menus {
		if m.IsRemove == 1 || m.MenuType == "F" {
			continue
		}
		if allow != nil && !allow[m.ID] {
			continue
		}
		out = append(out, m)
	}
	return out
}

// MenusByIDs 按给定菜单 id 集合过滤（剔除已删行）
func (s *Service) MenusByIDs(ctx context.Context, ids []int64) ([]*model.SystemMenu, error) {
	all, err := s.repo.Find(ctx)
	if err != nil {
		return nil, err
	}
	set := map[int64]bool{}
	for _, id := range ids {
		set[id] = true
	}
	var out []*model.SystemMenu
	for _, m := range all {
		if m.IsRemove != 1 && set[m.ID] {
			out = append(out, m)
		}
	}
	return out, nil
}

func sortNodes(ns []*TreeNode) {
	sort.Slice(ns, func(i, j int) bool {
		if ns[i].OrderNum != ns[j].OrderNum {
			return ns[i].OrderNum < ns[j].OrderNum
		}
		return ns[i].ID < ns[j].ID
	})
	for _, n := range ns {
		sortNodes(n.Children)
	}
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
