package permission

import (
	"context"
	"testing"

	"gosilite/application/base/ctxutil"
)

func TestPermissionConditionAdminBypass(t *testing.T) {
	cond := Permissions("system:user:list")
	for _, roles := range [][]string{{"admin"}, {"system"}} {
		login := &ctxutil.LoginUser{ID: 1, Roles: roles}
		if !cond.Check(context.Background(), nil, login) {
			t.Fatalf("super role %v must bypass perms check", roles)
		}
	}
}

func TestPermissionConditionGranted(t *testing.T) {
	cond := Permissions("system:user:list")
	login := &ctxutil.LoginUser{ID: 2, Roles: []string{"system_default"}, Permissions: []string{"system:user:list"}}
	if !cond.Check(context.Background(), nil, login) {
		t.Fatal("user with exact perm must pass")
	}
}

func TestPermissionConditionDenied(t *testing.T) {
	cond := Permissions("system:user:list")
	login := &ctxutil.LoginUser{ID: 2, Roles: []string{"system_default"}, Permissions: []string{"system:menu:list"}}
	if cond.Check(context.Background(), nil, login) {
		t.Fatal("user without perm must be denied")
	}
}

func TestPermissionConditionWildcard(t *testing.T) {
	cond := Permissions("system:user:add")
	login := &ctxutil.LoginUser{ID: 2, Roles: []string{"system_default"}, Permissions: []string{"system:user:*"}}
	if !cond.Check(context.Background(), nil, login) {
		t.Fatal("wildcard perm must match")
	}
}

func TestPermissionConditionNotLogin(t *testing.T) {
	cond := Permissions("system:user:list")
	if cond.Check(context.Background(), nil, nil) {
		t.Fatal("anonymous must be denied")
	}
}
