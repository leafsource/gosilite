package system

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	system "gosilite/api/model/framework/system"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/service/dept"
	"gosilite/application/framework/service/menu"
	"gosilite/application/framework/service/post"
	"gosilite/application/framework/service/role"
	"gosilite/application/framework/service/user"
	R "gosilite/pkg/utils/response"
)

// ========== 部门 ==========

// DeptTree .
// @Summary 部门树
// @Description 部门树接口
// @Tags 系统管理
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/dept/tree [GET]
func DeptTree(ctx context.Context, c *app.RequestContext) {
	nodes, err := dept.SystemDeptSVC.Tree(ctx)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	// 直接序列化服务层节点（含 leaderId/leaderName）；DeptNode 为 hz 生成 DTO 不含负责人字段，
	// 按 AGENTS 不手改 api/model，故不走 toDeptNode 转换，JSON 键与原 DTO 保持一致。
	resp := &struct {
		List []*dept.Node `json:"list"`
	}{List: nodes}
	R.SuccessWithData(ctx, c, resp)
}

// DeptList .
// @Summary 部门列表
// @Description 部门列表接口
// @Tags 系统管理
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/dept/list [GET]
func DeptList(ctx context.Context, c *app.RequestContext) {
	list, err := dept.SystemDeptSVC.List(ctx)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	resp := &system.DeptListResponse{List: toDeptDTOs(list), Total: int64(len(list))}
	R.SuccessWithData(ctx, c, resp)
}

// DeptCreate .
// @Summary 部门
// @Description 部门接口
// @Tags 系统管理
// @Produce json
// @Param data body system.DeptSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dept [POST]
func DeptCreate(ctx context.Context, c *app.RequestContext) {
	var req system.DeptSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	d := fromDeptSaveRequest(&req)
	if err := dept.SystemDeptSVC.Create(ctx, d); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDeptDTO(d))
}

// DeptUpdate .
// @Summary 部门
// @Description 部门接口
// @Tags 系统管理
// @Produce json
// @Param data body system.DeptSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dept [PUT]
func DeptUpdate(ctx context.Context, c *app.RequestContext) {
	var req system.DeptSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if req.ID == nil {
		R.InvalidParamRequestResponse(c, "id required")
		return
	}
	d := fromDeptSaveRequest(&req)
	d.ID = *req.ID
	if err := dept.SystemDeptSVC.Update(ctx, d); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDeptDTO(d))
}

// DeptRemove .
// @Summary 部门
// @Description 部门接口
// @Tags 系统管理
// @Produce json
// @Param data body system.IdRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dept [DELETE]
func DeptRemove(ctx context.Context, c *app.RequestContext) {
	var req system.IdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := dept.SystemDeptSVC.Delete(ctx, req.ID); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// ========== 岗位 ==========

// PostList .
// @Summary 岗位列表
// @Description 岗位列表接口
// @Tags 系统管理
// @Produce json
// @Param data query system.PageQuery false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/post/list [GET]
func PostList(ctx context.Context, c *app.RequestContext) {
	var req system.PageQuery
	_ = c.BindAndValidate(&req)
	list, total, err := post.SystemPostSVC.Page(ctx, int(req.Page), int(req.PageSize), req.Keyword)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	resp := &system.PostListResponse{List: toPostDTOs(list), Total: total}
	R.SuccessWithData(ctx, c, resp)
}

// PostCreate .
// @Summary 岗位
// @Description 岗位接口
// @Tags 系统管理
// @Produce json
// @Param data body system.PostSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/post [POST]
func PostCreate(ctx context.Context, c *app.RequestContext) {
	var req system.PostSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	p := fromPostSaveRequest(&req)
	if err := post.SystemPostSVC.Create(ctx, p); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toPostDTO(p))
}

// PostUpdate .
// @Summary 岗位
// @Description 岗位接口
// @Tags 系统管理
// @Produce json
// @Param data body system.PostSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/post [PUT]
func PostUpdate(ctx context.Context, c *app.RequestContext) {
	var req system.PostSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if req.ID == nil {
		R.InvalidParamRequestResponse(c, "id required")
		return
	}
	p := fromPostSaveRequest(&req)
	p.ID = *req.ID
	if err := post.SystemPostSVC.Update(ctx, p); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toPostDTO(p))
}

// PostRemove .
// @Summary 岗位
// @Description 岗位接口
// @Tags 系统管理
// @Produce json
// @Param data body system.IdRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/post [DELETE]
func PostRemove(ctx context.Context, c *app.RequestContext) {
	var req system.IdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := post.SystemPostSVC.Delete(ctx, req.ID); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// ========== 用户 ==========

// UserList .
// @Summary 用户列表
// @Description 用户列表接口
// @Tags 系统管理
// @Produce json
// @Param data query system.AdminUserListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/user/list [GET]
func UserList(ctx context.Context, c *app.RequestContext) {
	var req system.AdminUserListRequest
	_ = c.BindAndValidate(&req)
	keyword, state := "", ""
	deptID := int64(0)
	if req.Keyword != nil {
		keyword = *req.Keyword
	}
	if req.State != nil {
		state = *req.State
	}
	if req.DeptId != nil {
		deptID = *req.DeptId
	}
	items, total, err := user.SystemUserSVC.AdminPage(ctx, int(req.Page), int(req.PageSize), keyword, deptID, state)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	resp := &system.AdminUserListResponse{List: toUserDTOs(items), Total: total}
	R.SuccessWithData(ctx, c, resp)
}

// UserCreate .
// @Summary 用户
// @Description 用户接口
// @Tags 系统管理
// @Produce json
// @Param data body system.AdminUserSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/user [POST]
func UserCreate(ctx context.Context, c *app.RequestContext) {
	var req system.AdminUserSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	u := fromAdminUserSaveRequest(&req)
	password := ""
	if req.Password != nil {
		password = *req.Password
	}
	roleIDs, err := parseIDs(req.RoleIds)
	if err != nil {
		R.InvalidParamRequestResponse(c, "invalid roleIds")
		return
	}
	postIDs, err := parseIDs(req.PostIds)
	if err != nil {
		R.InvalidParamRequestResponse(c, "invalid postIds")
		return
	}
	if err := user.SystemUserSVC.AdminCreate(ctx, u, roleIDs, postIDs, password); err != nil {
		R.Error(ctx, c, err)
		return
	}
	savedRoleIDs, _ := user.SystemUserSVC.AdminRoleIDs(ctx, u.ID)
	R.SuccessWithData(ctx, c, adminUserDTO(u, "", savedRoleIDs, postIDs))
}

// UserUpdate .
// @Summary 用户
// @Description 用户接口
// @Tags 系统管理
// @Produce json
// @Param data body system.AdminUserSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/user [PUT]
func UserUpdate(ctx context.Context, c *app.RequestContext) {
	var req system.AdminUserSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if req.ID == nil {
		R.InvalidParamRequestResponse(c, "id required")
		return
	}
	u := fromAdminUserSaveRequest(&req)
	u.ID = *req.ID
	password := ""
	if req.Password != nil {
		password = *req.Password
	}
	roleIDs, err := parseIDs(req.RoleIds)
	if err != nil {
		R.InvalidParamRequestResponse(c, "invalid roleIds")
		return
	}
	postIDs, err := parseIDs(req.PostIds)
	if err != nil {
		R.InvalidParamRequestResponse(c, "invalid postIds")
		return
	}
	if err := user.SystemUserSVC.AdminUpdate(ctx, u, roleIDs, postIDs, password); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// UserRemove .
// @Summary 用户
// @Description 用户接口
// @Tags 系统管理
// @Produce json
// @Param data body system.IdRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/user [DELETE]
func UserRemove(ctx context.Context, c *app.RequestContext) {
	var req system.IdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := user.SystemUserSVC.AdminRemove(ctx, req.ID); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// UserStatus .
// @Summary 状态
// @Description 状态接口
// @Tags 系统管理
// @Produce json
// @Param data body system.UserStatusRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/user/status [PUT]
func UserStatus(ctx context.Context, c *app.RequestContext) {
	var req system.UserStatusRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := user.SystemUserSVC.AdminState(ctx, req.ID, req.State); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// UserRoles .
// @Summary 角色
// @Description 角色接口
// @Tags 系统管理
// @Produce json
// @Param data query system.IdRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/user/roles [GET]
func UserRoles(ctx context.Context, c *app.RequestContext) {
	var req system.IdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	ids, err := user.SystemUserSVC.AdminRoleIDs(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, &system.IdsResponse{Ids: formatIDs(ids)})
}

// ========== 菜单 ==========

// MenuTree .
// @Summary 菜单树
// @Description 菜单树接口
// @Tags 系统管理
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/menu/tree [GET]
func MenuTree(ctx context.Context, c *app.RequestContext) {
	nodes, err := menu.SystemMenuSVC.Tree(ctx)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	resp := &system.MenuTreeResponse{List: toMenuNodes(nodes)}
	R.SuccessWithData(ctx, c, resp)
}

// MenuList .
// @Summary 菜单列表
// @Description 菜单列表接口
// @Tags 系统管理
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/menu/list [GET]
func MenuList(ctx context.Context, c *app.RequestContext) {
	list, total, err := menu.SystemMenuSVC.ListPage(ctx)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	visible := make([]*model.SystemMenu, 0, len(list))
	for _, m := range list {
		if m.IsRemove == 1 {
			continue
		}
		visible = append(visible, m)
	}
	resp := &system.MenuListResponse{List: toMenuDTOs(visible), Total: total}
	R.SuccessWithData(ctx, c, resp)
}

// MenuCreate .
// @Summary 菜单
// @Description 菜单接口
// @Tags 系统管理
// @Produce json
// @Param data body system.MenuSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/menu [POST]
func MenuCreate(ctx context.Context, c *app.RequestContext) {
	var req system.MenuSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m := fromMenuSaveRequest(&req)
	if err := menu.SystemMenuSVC.CreateMenu(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toMenuDTO(m))
}

// MenuUpdate .
// @Summary 菜单
// @Description 菜单接口
// @Tags 系统管理
// @Produce json
// @Param data body system.MenuSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/menu [PUT]
func MenuUpdate(ctx context.Context, c *app.RequestContext) {
	var req system.MenuSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if req.ID == nil {
		R.InvalidParamRequestResponse(c, "id required")
		return
	}
	m := fromMenuSaveRequest(&req)
	m.ID = *req.ID
	if err := menu.SystemMenuSVC.UpdateMenu(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toMenuDTO(m))
}

// MenuRemove .
// @Summary 菜单
// @Description 菜单接口
// @Tags 系统管理
// @Produce json
// @Param data body system.IdRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/menu [DELETE]
func MenuRemove(ctx context.Context, c *app.RequestContext) {
	var req system.IdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := menu.SystemMenuSVC.RemoveMenu(ctx, req.ID); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// ========== 角色 ==========

// RoleList .
// @Summary 角色列表
// @Description 角色列表接口
// @Tags 系统管理
// @Produce json
// @Param data query system.RoleListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/role/list [GET]
func RoleList(ctx context.Context, c *app.RequestContext) {
	var req system.RoleListRequest
	_ = c.BindAndValidate(&req)
	keyword := ""
	if req.Keyword != nil {
		keyword = *req.Keyword
	}
	list, total, err := role.SystemRoleSVC.Page(ctx, int(req.Page), int(req.PageSize), keyword)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	resp := &system.RoleListResponse{List: toRoleDTOs(list), Total: total}
	R.SuccessWithData(ctx, c, resp)
}

// RoleCreate .
// @Summary 角色
// @Description 角色接口
// @Tags 系统管理
// @Produce json
// @Param data body system.RoleSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/role [POST]
func RoleCreate(ctx context.Context, c *app.RequestContext) {
	var req system.RoleSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	r := fromRoleSaveRequest(&req)
	if err := role.SystemRoleSVC.CreateRole(ctx, r); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toRoleDTO(r))
}

// RoleUpdate .
// @Summary 角色
// @Description 角色接口
// @Tags 系统管理
// @Produce json
// @Param data body system.RoleSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/role [PUT]
func RoleUpdate(ctx context.Context, c *app.RequestContext) {
	var req system.RoleSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if req.ID == nil {
		R.InvalidParamRequestResponse(c, "id required")
		return
	}
	r := fromRoleSaveRequest(&req)
	r.ID = *req.ID
	if err := role.SystemRoleSVC.UpdateRole(ctx, r); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toRoleDTO(r))
}

// RoleRemove .
// @Summary 角色
// @Description 角色接口
// @Tags 系统管理
// @Produce json
// @Param data body system.IdRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/role [DELETE]
func RoleRemove(ctx context.Context, c *app.RequestContext) {
	var req system.IdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := role.SystemRoleSVC.RemoveRole(ctx, req.ID); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// RoleAuthorize 角色授权快照（菜单树勾选态 + 数据范围 + 自定义部门）
// @Summary 角色授权
// @Description 角色授权接口
// @Tags 系统管理
// @Produce json
// @Param data query system.IdRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/role/authorize [GET]
func RoleAuthorize(ctx context.Context, c *app.RequestContext) {
	var req system.IdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	snap, err := role.SystemRoleSVC.Authorize(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toRoleAuthorizeResponse(snap))
}

// RoleAuthorizeSave 角色授权整体保存（菜单 + 数据范围 + 自定义部门）
// @Summary 授权
// @Description 授权接口
// @Tags 系统管理
// @Produce json
// @Param data body system.RoleAuthorizeSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/role/authorize [PUT]
func RoleAuthorizeSave(ctx context.Context, c *app.RequestContext) {
	var req system.RoleAuthorizeSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	menuIDs, err := parseIDs(req.MenuIds)
	if err != nil {
		R.InvalidParamRequestResponse(c, "invalid menuIds")
		return
	}
	deptIDs, err := parseIDs(req.DeptIds)
	if err != nil {
		R.InvalidParamRequestResponse(c, "invalid deptIds")
		return
	}
	if err := role.SystemRoleSVC.SaveAuthorize(ctx, req.RoleId, menuIDs, deptIDs, req.DataScope); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// DeptLeader 设置/清除部门负责人（userId=0 清除）
// @Summary 负责人
// @Description 负责人接口
// @Tags 系统管理
// @Produce json
// @Param data body system.DeptLeaderRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/dept/leader [PUT]
func DeptLeader(ctx context.Context, c *app.RequestContext) {
	var req system.DeptLeaderRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := dept.SystemDeptSVC.UpdateLeader(ctx, req.DeptId, req.UserId); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}
