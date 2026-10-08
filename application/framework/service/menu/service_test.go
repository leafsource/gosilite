package menu

import (
	"context"
	"errors"
	"testing"

	"gosilite/application/base/ctxutil"
	"gosilite/application/framework/dal/model"
	"gosilite/pkg/ctxcache"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

type fakeMenuRepo struct {
	menus     []*model.SystemMenu
	idsByUser map[int64][]int64
}

func (f *fakeMenuRepo) Create(ctx context.Context, m *model.SystemMenu) error { return nil }
func (f *fakeMenuRepo) GetByID(ctx context.Context, id int64) (*model.SystemMenu, error) {
	return nil, nil
}
func (f *fakeMenuRepo) Update(ctx context.Context, m *model.SystemMenu) error { return nil }
func (f *fakeMenuRepo) Delete(ctx context.Context, id int64) error            { return nil }
func (f *fakeMenuRepo) List(ctx context.Context, page, pageSize int) ([]*model.SystemMenu, int64, error) {
	return nil, 0, nil
}
func (f *fakeMenuRepo) Find(ctx context.Context) ([]*model.SystemMenu, error) {
	return f.menus, nil
}
func (f *fakeMenuRepo) FindMenuIDsByUser(ctx context.Context, userID int64) ([]int64, error) {
	return f.idsByUser[userID], nil
}

func errCode(err error) int32 {
	var se errorx.StatusError
	if errors.As(err, &se) {
		return se.Code()
	}
	return 0
}

func loginCtx(id int64, roles ...string) context.Context {
	ctx := ctxcache.Init(context.Background())
	ctxutil.SetLoginUser(ctx, &ctxutil.LoginUser{ID: id, Roles: roles})
	return ctx
}

func fixtureMenus() []*model.SystemMenu {
	return []*model.SystemMenu{
		{ID: 1, ParentID: 0, MenuType: "M", Path: "system"},
		{ID: 2, ParentID: 1, MenuType: "C", Path: "system/user"},
		{ID: 3, ParentID: 1, MenuType: "C", Path: "system/menu"},
		{ID: 4, ParentID: 2, MenuType: "F", Path: ""}, // 按钮行不进路由
		{ID: 5, ParentID: 1, MenuType: "C", Path: "system/removed", IsRemove: 1},
	}
}

func TestVisibleMenusAdminGetsAllExceptButtons(t *testing.T) {
	svc := newService(&fakeMenuRepo{menus: fixtureMenus()})
	menus, err := svc.VisibleMenus(loginCtx(1, "admin"))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(menus) != 3 {
		t.Fatalf("want 3 route menus (no F, no removed), got %d", len(menus))
	}
	for _, m := range menus {
		if m.MenuType == "F" || m.IsRemove == 1 {
			t.Fatalf("unexpected row in routes: id=%d type=%s", m.ID, m.MenuType)
		}
	}
}

func TestVisibleMenusNormalRoleGetsSubset(t *testing.T) {
	svc := newService(&fakeMenuRepo{
		menus:     fixtureMenus(),
		idsByUser: map[int64][]int64{50: {1, 2}},
	})
	menus, err := svc.VisibleMenus(loginCtx(50, "system_default"))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(menus) != 2 {
		t.Fatalf("want 2 menus for subset, got %d", len(menus))
	}
	for _, m := range menus {
		if m.ID == 3 {
			t.Fatalf("menu 3 must be filtered out for this role")
		}
	}
}

func TestVisibleMenusNoRoleGetsEmpty(t *testing.T) {
	svc := newService(&fakeMenuRepo{menus: fixtureMenus(), idsByUser: map[int64][]int64{}})
	menus, err := svc.VisibleMenus(loginCtx(50))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(menus) != 0 {
		t.Fatalf("want empty routes for role-less user, got %d", len(menus))
	}
}

func TestVisibleMenusRequiresLogin(t *testing.T) {
	svc := newService(&fakeMenuRepo{menus: fixtureMenus()})
	ctx := ctxcache.Init(context.Background())
	_, err := svc.VisibleMenus(ctx)
	if errCode(err) != errno.ErrUserAuthenticationFailed {
		t.Fatalf("want auth failed, got %v", err)
	}
}

func TestValidateMenuPerType(t *testing.T) {
	cases := []struct {
		name    string
		in      *model.SystemMenu
		wantErr bool
		check   func(t *testing.T, m *model.SystemMenu)
	}{
		{
			name:    "目录缺路由路径",
			in:      &model.SystemMenu{Title: "系统", MenuType: "M", Component: "layout"},
			wantErr: true,
		},
		{
			name:    "目录缺组件路径",
			in:      &model.SystemMenu{Title: "系统", MenuType: "M", Path: "/system"},
			wantErr: true,
		},
		{
			name: "目录合法并清空权限标识",
			in: &model.SystemMenu{
				Title: " 系统 ", MenuType: "M", Path: " /system ", Component: "layout",
				Perms: "system:all", PermissionCode: "system:all",
			},
			check: func(t *testing.T, m *model.SystemMenu) {
				if m.Title != "系统" || m.Path != "/system" {
					t.Fatalf("want trimmed title/path, got title=%q path=%q", m.Title, m.Path)
				}
				if m.Perms != "" || m.PermissionCode != "" {
					t.Fatalf("directory perms should be cleared, got %q/%q", m.Perms, m.PermissionCode)
				}
			},
		},
		{
			name:    "菜单缺组件路径",
			in:      &model.SystemMenu{Title: "用户管理", MenuType: "C", Path: "/system/user"},
			wantErr: true,
		},
		{
			name: "菜单合法且权限标识选填",
			in: &model.SystemMenu{
				Title: "用户管理", MenuType: "C", Path: "/system/user",
				Component: "system/user/index", Perms: "system:user:list",
			},
			check: func(t *testing.T, m *model.SystemMenu) {
				if m.Perms != "system:user:list" {
					t.Fatalf("menu perms should be kept, got %q", m.Perms)
				}
			},
		},
		{
			name:    "按钮缺权限标识",
			in:      &model.SystemMenu{Title: "新增用户", MenuType: "F"},
			wantErr: true,
		},
		{
			name: "按钮缺操作类型被拒绝",
			in: &model.SystemMenu{
				Title: "新增用户", MenuType: "F", Perms: "system:user:add",
			},
			wantErr: true,
		},
		{
			name: "按钮合法并归零路由字段",
			in: &model.SystemMenu{
				Title: "新增用户", MenuType: "F", Perms: "system:user:add", ActionType: " add ",
				Path: "/system/user", Component: "system/user/index", Icon: "add",
			},
			check: func(t *testing.T, m *model.SystemMenu) {
				if m.Path != "" || m.Component != "" || m.Icon != "" {
					t.Fatalf("button route fields should be cleared, got %q/%q/%q", m.Path, m.Component, m.Icon)
				}
				if m.PermissionCode != "system:user:add" {
					t.Fatalf("permission code should sync perms, got %q", m.PermissionCode)
				}
				if m.ActionType != "add" {
					t.Fatalf("action type should be trimmed, got %q", m.ActionType)
				}
			},
		},
		{
			name: "非按钮类型清空操作类型",
			in: &model.SystemMenu{
				Title: "用户管理", MenuType: "C", Path: "/system/user",
				Component: "system/user/index", ActionType: "add",
			},
			check: func(t *testing.T, m *model.SystemMenu) {
				if m.ActionType != "" {
					t.Fatalf("action type should be cleared for non-button, got %q", m.ActionType)
				}
			},
		},
		{
			name:    "空名称被拒绝",
			in:      &model.SystemMenu{Title: "  ", MenuType: "C", Path: "/a", Component: "a/index"},
			wantErr: true,
		},
		{
			name:    "非法类型被拒绝",
			in:      &model.SystemMenu{Title: "X", MenuType: "X", Path: "/a", Component: "a/index"},
			wantErr: true,
		},
		{
			name: "空类型默认菜单并按 C 校验",
			in:   &model.SystemMenu{Title: "用户管理", Path: "/system/user", Component: "system/user/index"},
			check: func(t *testing.T, m *model.SystemMenu) {
				if m.MenuType != "C" {
					t.Fatalf("want default type C, got %q", m.MenuType)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMenu(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatal("want error, got nil")
				}
				if errCode(err) != errno.ErrSystemInvalidParam {
					t.Fatalf("want invalid param code, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if tc.check != nil {
				tc.check(t, tc.in)
			}
		})
	}
}
