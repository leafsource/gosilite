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
