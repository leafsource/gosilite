package user

import (
	"context"
	"strconv"
	"strings"

	"gosilite/application/base/ctxutil"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	"gosilite/application/framework/service/dept"
	"gosilite/pkg/bcrypt"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

// AdminListItem 管理端用户行
type AdminListItem struct {
	User       *model.SystemUser
	DeptName   string
	RoleIDs    []int64
	PostIDs    []int64
}

// AdminPage 管理端分页（按登录者数据范围过滤部门；默认 fail-closed）
func (s *Service) AdminPage(ctx context.Context, page, pageSize int, keyword string, deptID int64, state string) ([]*AdminListItem, int64, error) {
	login := ctxutil.GetLoginUser(ctx)
	if login == nil {
		return nil, 0, errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("reason", "session is required"))
	}

	// fail-closed：默认不放行
	allDept := false
	var visible []int64

	// 内置管理员角色码放行全部
	for _, c := range login.Roles {
		if c == "admin" || c == "system" {
			allDept = true
			break
		}
	}

	if !allDept {
		roles, err := s.findRoles(ctx, login.ID)
		if err != nil {
			return nil, 0, err
		}
		if len(roles) == 0 {
			// 无角色：仅本人
			return s.adminPageSelfOnly(ctx, page, pageSize, keyword, state)
		}
		// 多角色取最宽（数值越小越宽：1 全部）
		scope := int16(5)
		var custom []int64
		for _, r := range roles {
			if r.DataScope == 1 {
				scope = 1
				break
			}
			// 取更宽的 scope（更小的值）
			if r.DataScope > 0 && r.DataScope < scope {
				scope = r.DataScope
			}
			if r.DataScope == 5 {
				ids, _ := s.roleDeptIDs(ctx, r.ID)
				custom = append(custom, ids...)
			}
		}
		if dept.SystemDeptSVC == nil {
			return nil, 0, errorx.New(errno.ErrSystemForbidden, errorx.KV("msg", "dept service unavailable"))
		}
		allDept, visible, err = dept.SystemDeptSVC.VisibleDeptIDs(ctx, scope, login.DeptID, custom)
		if err != nil {
			return nil, 0, err
		}
		// scope=4 仅本人：走 self-only
		if scope == 4 {
			return s.adminPageSelfOnly(ctx, page, pageSize, keyword, state)
		}
		// 可见集为空且非全部：返回空列表
		if !allDept && len(visible) == 0 {
			return []*AdminListItem{}, 0, nil
		}
	}

	users, total, err := s.repo.AdminList(ctx, page, pageSize, keyword, deptID, state, visible, allDept)
	if err != nil {
		return nil, 0, err
	}
	return s.fillAdminItems(ctx, users, total)
}

// adminPageSelfOnly 仅返回自己
func (s *Service) adminPageSelfOnly(ctx context.Context, page, pageSize int, keyword, state string) ([]*AdminListItem, int64, error) {
	login := ctxutil.GetLoginUser(ctx)
	if login == nil {
		return nil, 0, errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("reason", "session is required"))
	}
	u, err := s.repo.GetByID(ctx, login.ID)
	if err != nil {
		return nil, 0, err
	}
	if state != "" && u.State != state {
		return []*AdminListItem{}, 0, nil
	}
	if keyword != "" && !strings.Contains(u.Account, keyword) && !strings.Contains(u.Username, keyword) {
		return []*AdminListItem{}, 0, nil
	}
	return s.fillAdminItems(ctx, []*model.SystemUser{u}, 1)
}

func (s *Service) fillAdminItems(ctx context.Context, users []*model.SystemUser, total int64) ([]*AdminListItem, int64, error) {
	deptNames := map[int64]string{}
	if dept.SystemDeptSVC != nil {
		if list, err := dept.SystemDeptSVC.List(ctx); err == nil {
			for _, d := range list {
				deptNames[d.ID] = d.DeptName
			}
		}
	}
	out := make([]*AdminListItem, 0, len(users))
	for _, u := range users {
		roleIDs, _ := s.repo.RoleIDs(ctx, u.ID)
		postIDs := []int64{}
		if s.posts != nil {
			postIDs, _ = s.posts.ListPostIDsByUser(ctx, u.ID)
		}
		out = append(out, &AdminListItem{
			User: u, DeptName: deptNames[u.DeptID],
			RoleIDs: roleIDs, PostIDs: postIDs,
		})
	}
	return out, total, nil
}

// AdminCreate 创建用户 + 角色 + 岗位
func (s *Service) AdminCreate(ctx context.Context, u *model.SystemUser, roleIDs, postIDs []int64, password string) error {
	if strings.TrimSpace(u.Account) == "" || strings.TrimSpace(u.Username) == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "账号/用户名不能为空"))
	}
	exists, err := s.repo.ExistsAccount(ctx, u.Account, 0)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrSystemUserAccountExists, errorx.KV("account", u.Account))
	}
	if password == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "创建用户必须设置密码"))
	}
	hashed, err := bcrypt.HashPassword(password)
	if err != nil {
		return err
	}
	u.Password = hashed
	if u.State == "" {
		u.State = "enabled"
	}
	u.IsRemove = 0
	if err := s.repo.Create(ctx, u); err != nil {
		return err
	}
	if err := s.repo.ReplaceRoles(ctx, u.ID, roleIDs); err != nil {
		return err
	}
	if s.posts != nil && len(postIDs) > 0 {
		return s.posts.ReplaceUserPosts(ctx, u.ID, postIDs)
	}
	return nil
}

// AdminUpdate 更新用户资料 + 角色岗位（可选改密）
func (s *Service) AdminUpdate(ctx context.Context, u *model.SystemUser, roleIDs, postIDs []int64, password string) error {
	old, err := s.repo.GetByID(ctx, u.ID)
	if err != nil {
		return errorx.New(errno.ErrSystemUserNotFound, errorx.KV("id", formatID(u.ID)))
	}
	exists, err := s.repo.ExistsAccount(ctx, u.Account, u.ID)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrSystemUserAccountExists, errorx.KV("account", u.Account))
	}
	if password != "" {
		hashed, err := bcrypt.HashPassword(password)
		if err != nil {
			return err
		}
		old.Password = hashed
	}
	old.Account = u.Account
	old.Username = u.Username
	old.MobilePhone = u.MobilePhone
	old.Email = u.Email
	old.DeptID = u.DeptID
	if u.State != "" {
		old.State = u.State
	}
	if err := s.repo.Update(ctx, old); err != nil {
		return err
	}
	if err := s.repo.ReplaceRoles(ctx, u.ID, roleIDs); err != nil {
		return err
	}
	if s.posts != nil {
		return s.posts.ReplaceUserPosts(ctx, u.ID, postIDs)
	}
	return nil
}

// AdminRemove 删除用户角色岗位
func (s *Service) AdminRemove(ctx context.Context, id int64) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return errorx.New(errno.ErrSystemUserNotFound, errorx.KV("id", formatID(id)))
	}
	if err := s.repo.ReplaceRoles(ctx, id, nil); err != nil {
		return err
	}
	if s.posts != nil {
		_ = s.posts.ReplaceUserPosts(ctx, id, nil)
	}
	return s.repo.Delete(ctx, id)
}

// AdminState 启用停用
func (s *Service) AdminState(ctx context.Context, id int64, state string) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return errorx.New(errno.ErrSystemUserNotFound, errorx.KV("id", formatID(id)))
	}
	return s.repo.UpdateState(ctx, id, state)
}

// AdminRoleIDs 用户已选角色
func (s *Service) AdminRoleIDs(ctx context.Context, id int64) ([]int64, error) {
	return s.repo.RoleIDs(ctx, id)
}

// findRoles 用户角色实体
func (s *Service) findRoles(ctx context.Context, userID int64) ([]*model.SystemRole, error) {
	if s.roleRBAC == nil {
		return nil, nil
	}
	return s.roleRBAC.FindRolesByUser(ctx, userID)
}

// roleDeptIDs 角色自定义部门
func (s *Service) roleDeptIDs(ctx context.Context, roleID int64) ([]int64, error) {
	if s.roleRBAC == nil {
		return nil, nil
	}
	return s.roleRBAC.ListDeptIDs(ctx, roleID)
}

// withRoleRBAC 注入角色关联仓储（数据权限）
func (s *Service) withRoleRBAC(r repository.IRoleRBACRepository) *Service {
	s.roleRBAC = r
	return s
}

// Ensure posts repo injected
func (s *Service) withPosts(p repository.IPostRepository) *Service {
	s.posts = p
	return s
}

func formatID(v int64) string {
	return strconv.FormatInt(v, 10)
}
