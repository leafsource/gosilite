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

func toDeptDTO(m *model.SystemDept) *system.DeptDTO {
	return &system.DeptDTO{
		ID: m.ID, ParentId: m.ParentID, Ancestors: m.Ancestors,
		DeptName: m.DeptName, OrderNum: m.OrderNum, Status: int32(m.Status),
	}
}

func toDeptNode(n *dept.Node) *system.DeptNode {
	node := &system.DeptNode{
		ID: n.ID, ParentId: n.ParentID, Label: n.Label,
		OrderNum: n.OrderNum, Status: int32(n.Status),
	}
	for _, c := range n.Children {
		node.Children = append(node.Children, toDeptNode(c))
	}
	return node
}

// DeptTree .
// @router /system/dept/tree [GET]
func DeptTree(ctx context.Context, c *app.RequestContext) {
	nodes, err := dept.SystemDeptSVC.Tree(ctx)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	resp := &system.DeptTreeResponse{List: make([]*system.DeptNode, 0, len(nodes))}
	for _, n := range nodes {
		resp.List = append(resp.List, toDeptNode(n))
	}
	R.SuccessWithData(ctx, c, resp)
}

// DeptList .
// @router /system/dept/list [GET]
func DeptList(ctx context.Context, c *app.RequestContext) {
	list, err := dept.SystemDeptSVC.List(ctx)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	resp := &system.DeptListResponse{List: make([]*system.DeptDTO, 0, len(list)), Total: int64(len(list))}
	for _, d := range list {
		resp.List = append(resp.List, toDeptDTO(d))
	}
	R.SuccessWithData(ctx, c, resp)
}

// DeptCreate .
// @router /system/dept [POST]
func DeptCreate(ctx context.Context, c *app.RequestContext) {
	var req system.DeptSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	d := &model.SystemDept{
		ParentID: req.ParentId, DeptName: req.DeptName,
		OrderNum: req.OrderNum, Status: int16(req.Status),
	}
	if err := dept.SystemDeptSVC.Create(ctx, d); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDeptDTO(d))
}

// DeptUpdate .
// @router /system/dept [PUT]
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
	d := &model.SystemDept{
		ID: *req.ID, ParentID: req.ParentId, DeptName: req.DeptName,
		OrderNum: req.OrderNum, Status: int16(req.Status),
	}
	if err := dept.SystemDeptSVC.Update(ctx, d); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDeptDTO(d))
}

// DeptRemove .
// @router /system/dept [DELETE]
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

func toPostDTO(m *model.SystemPost) *system.PostDTO {
	return &system.PostDTO{
		ID: m.ID, PostCode: m.PostCode, PostName: m.PostName,
		OrderNum: m.OrderNum, Status: int32(m.Status),
	}
}

// PostList .
// @router /system/post/list [GET]
func PostList(ctx context.Context, c *app.RequestContext) {
	var req system.PageQuery
	_ = c.BindAndValidate(&req)
	list, total, err := post.SystemPostSVC.Page(ctx, int(req.Page), int(req.PageSize), req.Keyword)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	resp := &system.PostListResponse{List: make([]*system.PostDTO, 0, len(list)), Total: total}
	for _, p := range list {
		resp.List = append(resp.List, toPostDTO(p))
	}
	R.SuccessWithData(ctx, c, resp)
}

// PostCreate .
// @router /system/post [POST]
func PostCreate(ctx context.Context, c *app.RequestContext) {
	var req system.PostSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	p := &model.SystemPost{
		PostCode: req.PostCode, PostName: req.PostName,
		OrderNum: req.OrderNum, Status: int16(req.Status),
	}
	if err := post.SystemPostSVC.Create(ctx, p); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toPostDTO(p))
}

// PostUpdate .
// @router /system/post [PUT]
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
	p := &model.SystemPost{
		ID: *req.ID, PostCode: req.PostCode, PostName: req.PostName,
		OrderNum: req.OrderNum, Status: int16(req.Status),
	}
	if err := post.SystemPostSVC.Update(ctx, p); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toPostDTO(p))
}

// PostRemove .
// @router /system/post [DELETE]
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

func toUserDTO(item *user.AdminListItem) *system.AdminUserDTO {
	u := item.User
	return &system.AdminUserDTO{
		ID: u.ID, Account: u.Account, Username: u.Username,
		MobilePhone: u.MobilePhone, Email: u.Email, State: u.State,
		DeptId: u.DeptID, DeptName: item.DeptName,
		RoleIds: item.RoleIDs, PostIds: item.PostIDs, AvatarUrl: u.AvatarURL,
	}
}

// UserList .
// @router /system/user/list [GET]
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
	resp := &system.AdminUserListResponse{List: make([]*system.AdminUserDTO, 0, len(items)), Total: total}
	for _, it := range items {
		resp.List = append(resp.List, toUserDTO(it))
	}
	R.SuccessWithData(ctx, c, resp)
}

// UserCreate .
// @router /system/user [POST]
func UserCreate(ctx context.Context, c *app.RequestContext) {
	var req system.AdminUserSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	u := &model.SystemUser{
		Account: req.Account, Username: req.Username,
		DeptID: req.DeptId, State: req.State,
	}
	if req.MobilePhone != nil {
		u.MobilePhone = *req.MobilePhone
	}
	if req.Email != nil {
		u.Email = *req.Email
	}
	password := ""
	if req.Password != nil {
		password = *req.Password
	}
	if err := user.SystemUserSVC.AdminCreate(ctx, u, req.RoleIds, req.PostIds, password); err != nil {
		R.Error(ctx, c, err)
		return
	}
	roleIDs, _ := user.SystemUserSVC.AdminRoleIDs(ctx, u.ID)
	R.SuccessWithData(ctx, c, &system.AdminUserDTO{
		ID: u.ID, Account: u.Account, Username: u.Username,
		MobilePhone: u.MobilePhone, Email: u.Email, State: u.State,
		DeptId: u.DeptID, RoleIds: roleIDs, PostIds: req.PostIds, AvatarUrl: u.AvatarURL,
	})
}

// UserUpdate .
// @router /system/user [PUT]
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
	u := &model.SystemUser{
		ID: *req.ID, Account: req.Account, Username: req.Username,
		DeptID: req.DeptId, State: req.State,
	}
	if req.MobilePhone != nil {
		u.MobilePhone = *req.MobilePhone
	}
	if req.Email != nil {
		u.Email = *req.Email
	}
	password := ""
	if req.Password != nil {
		password = *req.Password
	}
	if err := user.SystemUserSVC.AdminUpdate(ctx, u, req.RoleIds, req.PostIds, password); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// UserRemove .
// @router /system/user [DELETE]
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
// @router /system/user/status [PUT]
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
// @router /system/user/roles [GET]
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
	R.SuccessWithData(ctx, c, &system.IdsResponse{Ids: ids})
}

// ========== 菜单 ==========

func toMenuDTO(m *model.SystemMenu) *system.MenuDTO {
	return &system.MenuDTO{
		ID: m.ID, ParentId: m.ParentID, MenuName: m.Title, MenuType: m.MenuType,
		Path: m.Path, Component: m.Component, Perms: m.Perms, Icon: m.Icon,
		OrderNum: m.Sort, Visible: m.Visible != "hide" && m.Visible != "0",
		Status: int32(m.Status),
	}
}

func toMenuNode(n *menu.TreeNode) *system.MenuNode {
	node := &system.MenuNode{
		ID: n.ID, ParentId: n.ParentID, Label: n.Label, MenuType: n.MenuType,
		Perms: n.Perms, OrderNum: n.OrderNum, Visible: n.Visible, Status: int32(n.Status),
		Path: n.Path, Component: n.Component, Icon: n.Icon,
	}
	for _, ch := range n.Children {
		node.Children = append(node.Children, toMenuNode(ch))
	}
	return node
}

// MenuTree .
// @router /system/menu/tree [GET]
func MenuTree(ctx context.Context, c *app.RequestContext) {
	nodes, err := menu.SystemMenuSVC.Tree(ctx)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	resp := &system.MenuTreeResponse{List: make([]*system.MenuNode, 0, len(nodes))}
	for _, n := range nodes {
		resp.List = append(resp.List, toMenuNode(n))
	}
	R.SuccessWithData(ctx, c, resp)
}

// MenuList .
// @router /system/menu/list [GET]
func MenuList(ctx context.Context, c *app.RequestContext) {
	list, total, err := menu.SystemMenuSVC.ListPage(ctx)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	resp := &system.MenuListResponse{List: make([]*system.MenuDTO, 0, len(list)), Total: total}
	for _, m := range list {
		if m.IsRemove == 1 {
			continue
		}
		resp.List = append(resp.List, toMenuDTO(m))
	}
	R.SuccessWithData(ctx, c, resp)
}

func menuFromReq(req *system.MenuSaveRequest) *model.SystemMenu {
	m := &model.SystemMenu{
		ParentID: req.ParentId, Title: req.MenuName, MenuType: req.MenuType,
		Sort: req.OrderNum, Status: int16(req.Status),
	}
	if req.Path != nil {
		m.Path = *req.Path
	}
	if req.Component != nil {
		m.Component = *req.Component
	}
	if req.Perms != nil {
		m.Perms = *req.Perms
		m.PermissionCode = *req.Perms
	}
	if req.Icon != nil {
		m.Icon = *req.Icon
	}
	if req.Visible {
		m.Visible = "show"
	} else {
		m.Visible = "hide"
	}
	return m
}

// MenuCreate .
// @router /system/menu [POST]
func MenuCreate(ctx context.Context, c *app.RequestContext) {
	var req system.MenuSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m := menuFromReq(&req)
	if err := menu.SystemMenuSVC.CreateMenu(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toMenuDTO(m))
}

// MenuUpdate .
// @router /system/menu [PUT]
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
	m := menuFromReq(&req)
	m.ID = *req.ID
	if err := menu.SystemMenuSVC.UpdateMenu(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toMenuDTO(m))
}

// MenuRemove .
// @router /system/menu [DELETE]
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

func toRoleDTO(r *model.SystemRole) *system.RoleDTO {
	return &system.RoleDTO{
		ID: r.ID, Name: r.Name, Code: r.Code, Description: r.Description,
		State: r.State, DataScope: int32(r.DataScope), OrderNum: r.OrderNum,
	}
}

// RoleList .
// @router /system/role/list [GET]
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
	resp := &system.RoleListResponse{List: make([]*system.RoleDTO, 0, len(list)), Total: total}
	for _, r := range list {
		resp.List = append(resp.List, toRoleDTO(r))
	}
	R.SuccessWithData(ctx, c, resp)
}

// RoleCreate .
// @router /system/role [POST]
func RoleCreate(ctx context.Context, c *app.RequestContext) {
	var req system.RoleSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	r := &model.SystemRole{
		Name: req.Name, Code: req.Code, State: req.State,
		DataScope: int16(req.DataScope), OrderNum: req.OrderNum,
	}
	if req.Description != nil {
		r.Description = *req.Description
	}
	if err := role.SystemRoleSVC.CreateRole(ctx, r); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toRoleDTO(r))
}

// RoleUpdate .
// @router /system/role [PUT]
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
	r := &model.SystemRole{
		ID: *req.ID, Name: req.Name, Code: req.Code, State: req.State,
		DataScope: int16(req.DataScope), OrderNum: req.OrderNum,
	}
	if req.Description != nil {
		r.Description = *req.Description
	}
	if err := role.SystemRoleSVC.UpdateRole(ctx, r); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toRoleDTO(r))
}

// RoleRemove .
// @router /system/role [DELETE]
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

// RoleAssignMenu .
// @router /system/role/menu [PUT]
func RoleAssignMenu(ctx context.Context, c *app.RequestContext) {
	var req system.RoleMenuRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := role.SystemRoleSVC.AssignMenus(ctx, req.RoleId, req.MenuIds); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// RoleAssignDept .
// @router /system/role/dept [PUT]
func RoleAssignDept(ctx context.Context, c *app.RequestContext) {
	var req system.RoleDeptRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := role.SystemRoleSVC.AssignDepts(ctx, req.RoleId, req.DeptIds); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// RoleMenus .
// @router /system/role/menus [GET]
func RoleMenus(ctx context.Context, c *app.RequestContext) {
	var req system.IdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	ids, err := role.SystemRoleSVC.MenuIDs(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, &system.IdsResponse{Ids: ids})
}

// RoleDepts .
// @router /system/role/depts [GET]
func RoleDepts(ctx context.Context, c *app.RequestContext) {
	var req system.IdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	ids, err := role.SystemRoleSVC.DeptIDs(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, &system.IdsResponse{Ids: ids})
}
