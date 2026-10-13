package permission

import (
	"context"
	"gosilite/application/base/ctxutil"
	"gosilite/pkg/errorx"
	R "gosilite/pkg/utils/response"
	"gosilite/types/errno"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

// Condition 权限条件接口
type Condition interface {
	Check(ctx context.Context, c *app.RequestContext, loginUser *ctxutil.LoginUser) bool
}

// AndCondition 逻辑与条件
type AndCondition struct {
	conditions []Condition
}

func (a *AndCondition) Check(ctx context.Context, c *app.RequestContext, loginUser *ctxutil.LoginUser) bool {
	for _, condition := range a.conditions {
		if !condition.Check(ctx, c, loginUser) {
			return false
		}
	}
	return true
}

// OrCondition 逻辑或条件
type OrCondition struct {
	conditions []Condition
}

func (o *OrCondition) Check(ctx context.Context, c *app.RequestContext, loginUser *ctxutil.LoginUser) bool {
	for _, condition := range o.conditions {
		if condition.Check(ctx, c, loginUser) {
			return true
		}
	}
	return false
}

// RoleCondition 角色条件
type RoleCondition struct {
	roles []string
}

func (r *RoleCondition) Check(ctx context.Context, c *app.RequestContext, loginUser *ctxutil.LoginUser) bool {
	if loginUser == nil {
		return false
	}

	roleSet := make(map[string]bool)
	for _, role := range loginUser.Roles {
		roleSet[role] = true
	}

	for _, requiredRole := range r.roles {
		if roleSet[requiredRole] {
			return true
		}
	}
	return false
}

// PermissionCondition 权限条件
type PermissionCondition struct {
	permissions []string
}

func (p *PermissionCondition) Check(ctx context.Context, c *app.RequestContext, loginUser *ctxutil.LoginUser) bool {
	if loginUser == nil {
		return false
	}

	// 超管（admin/system）豁免 perms 校验
	if ctxutil.IsSuperRole(loginUser.Roles) {
		return true
	}

	if len(p.permissions) == 0 {
		return true
	}

	permissionSet := make(map[string]bool)
	for _, perm := range loginUser.Permissions {
		permissionSet[perm] = true
	}

	// 检查直接权限
	for _, perm := range p.permissions {
		if permissionSet[perm] {
			return true
		}
	}

	// 检查通配符权限
	for userPerm := range permissionSet {
		for _, requiredPerm := range p.permissions {
			if checkWildcardPermission(userPerm, requiredPerm) {
				return true
			}
		}
	}

	return false
}

// 工厂函数
func And(conditions ...Condition) Condition {
	return &AndCondition{conditions: conditions}
}

func Or(conditions ...Condition) Condition {
	return &OrCondition{conditions: conditions}
}

func Roles(roles ...string) Condition {
	return &RoleCondition{roles: roles}
}

func Permissions(permissions ...string) Condition {
	return &PermissionCondition{permissions: permissions}
}

// New 创建权限检查中间件
func New(conditions ...Condition) []app.HandlerFunc {
	return append(NewSecurity(), func(ctx context.Context, c *app.RequestContext) {

		loginUser := ctxutil.GetLoginUser(ctx)

		// 如果没有条件，则只需要登录
		if len(conditions) == 0 {
			if loginUser == nil {
				R.Error(ctx, c, errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("reason", "not login")))
				c.Abort()
				return
			}
			c.Next(ctx)
			return
		}

		// 如果有多个条件，默认使用 And 逻辑
		var condition Condition
		if len(conditions) == 1 {
			condition = conditions[0]
		} else {
			condition = And(conditions...)
		}

		// 执行权限检查
		if !condition.Check(ctx, c, loginUser) {
			// 获取错误信息
			var msg strings.Builder

			var buildMsg func(conditions []Condition, operator string)
			buildMsg = func(conditions []Condition, operator string) {
				for i, cond := range conditions {
					if i > 0 {
						msg.WriteString(" " + operator + " ")
					}
					switch c := cond.(type) {
					case *AndCondition:
						buildMsg(c.conditions, "and")
					case *OrCondition:
						buildMsg(c.conditions, "or")
					case *RoleCondition:
						msg.WriteString("Roles(" + strings.Join(c.roles, ", ") + ")")
					case *PermissionCondition:
						msg.WriteString("Permissions(" + strings.Join(c.permissions, ", ") + ")")

					}
				}
			}
			buildMsg(conditions, "and")

			if loginUser == nil {
				R.Error(ctx, c, errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("reason", "not login")))
			} else {
				R.Error(ctx, c, errorx.New(errno.ErrUserPermissionCode, errorx.KV("msg", msg.String())))
			}
			c.Abort()
			return
		}

		c.Next(ctx)
	})

}

// 兼容性函数
func Default() []app.HandlerFunc {
	return New()
}

func HasRoles(roleCode ...string) []app.HandlerFunc {
	return New(Roles(roleCode...))
}

func HPermissions(permissionCode ...string) []app.HandlerFunc {
	return New(Permissions(permissionCode...))
}

// 辅助函数
func AndRolesAndPermissions(roles []string, permissions []string) []app.HandlerFunc {
	var conditions []Condition
	if len(roles) > 0 {
		conditions = append(conditions, Roles(roles...))
	}
	if len(permissions) > 0 {
		conditions = append(conditions, Permissions(permissions...))
	}
	return New(conditions...)
}

func OrRolesOrPermissions(roles []string, permissions []string) []app.HandlerFunc {
	var conditions []Condition
	if len(roles) > 0 {
		conditions = append(conditions, Roles(roles...))
	}
	if len(permissions) > 0 {
		conditions = append(conditions, Permissions(permissions...))
	}
	return New(Or(conditions...))
}

// checkWildcardPermission 检查通配符权限
func checkWildcardPermission(userPerm, requiredPerm string) bool {
	if len(userPerm) > 2 && userPerm[len(userPerm)-2:] == ":*" {
		prefix := userPerm[:len(userPerm)-2]
		return len(requiredPerm) > len(prefix)+1 && requiredPerm[:len(prefix)+1] == prefix+":"
	}
	return false
}

// 示例使用方式：
// 1. permission.And(
//        permission.Roles("admin"),
//        permission.Permissions("user:write"),
//    )
//
// 2. permission.Or(
//        permission.Roles("admin"),
//        permission.Roles("manager"),
//    )
//
// 3. permission.And(
//        permission.Roles("admin"),
//        permission.And(
//            permission.Permissions("user:write", "user:read"),
//            permission.Roles("editor"),
//        ),
//    )
//
// 4. permission.New(permission.Roles("admin")) // 等同于 HasRoles("admin")
// 5. permission.New(permission.Permissions("user:write")) // 等同于 HPermissions("user:write")
