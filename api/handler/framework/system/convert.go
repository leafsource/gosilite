package system

import (
	system "gosilite/api/model/framework/system"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/service/dept"
	"gosilite/application/framework/service/menu"
	"gosilite/application/framework/service/user"
)

// ========== 部门 ==========

func toDeptDTO(m *model.SystemDept) *system.DeptDTO {
	return &system.DeptDTO{
		ID: m.ID, ParentId: m.ParentID, Ancestors: m.Ancestors,
		DeptName: m.DeptName, OrderNum: m.OrderNum, Status: int32(m.Status),
	}
}

func toDeptDTOs(list []*model.SystemDept) []*system.DeptDTO {
	out := make([]*system.DeptDTO, 0, len(list))
	for _, m := range list {
		out = append(out, toDeptDTO(m))
	}
	return out
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

func fromDeptSaveRequest(req *system.DeptSaveRequest) *model.SystemDept {
	return &model.SystemDept{
		ParentID: req.ParentId, DeptName: req.DeptName,
		OrderNum: req.OrderNum, Status: int16(req.Status),
	}
}

// ========== 岗位 ==========

func toPostDTO(m *model.SystemPost) *system.PostDTO {
	return &system.PostDTO{
		ID: m.ID, PostCode: m.PostCode, PostName: m.PostName,
		OrderNum: m.OrderNum, Status: int32(m.Status),
	}
}

func toPostDTOs(list []*model.SystemPost) []*system.PostDTO {
	out := make([]*system.PostDTO, 0, len(list))
	for _, m := range list {
		out = append(out, toPostDTO(m))
	}
	return out
}

func fromPostSaveRequest(req *system.PostSaveRequest) *model.SystemPost {
	return &model.SystemPost{
		PostCode: req.PostCode, PostName: req.PostName,
		OrderNum: req.OrderNum, Status: int16(req.Status),
	}
}

// ========== 用户 ==========

// adminUserDTO SystemUser + 关联信息 → AdminUserDTO 的公共映射
func adminUserDTO(u *model.SystemUser, deptName string, roleIDs, postIDs []int64) *system.AdminUserDTO {
	return &system.AdminUserDTO{
		ID: u.ID, Account: u.Account, Username: u.Username,
		MobilePhone: u.MobilePhone, Email: u.Email, State: u.State,
		DeptId: u.DeptID, DeptName: deptName,
		RoleIds: roleIDs, PostIds: postIDs, AvatarUrl: u.AvatarURL,
	}
}

func toUserDTO(item *user.AdminListItem) *system.AdminUserDTO {
	return adminUserDTO(item.User, item.DeptName, item.RoleIDs, item.PostIDs)
}

func toUserDTOs(items []*user.AdminListItem) []*system.AdminUserDTO {
	out := make([]*system.AdminUserDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toUserDTO(item))
	}
	return out
}

func fromAdminUserSaveRequest(req *system.AdminUserSaveRequest) *model.SystemUser {
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
	return u
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

func toMenuDTOs(list []*model.SystemMenu) []*system.MenuDTO {
	out := make([]*system.MenuDTO, 0, len(list))
	for _, m := range list {
		out = append(out, toMenuDTO(m))
	}
	return out
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

func toMenuNodes(nodes []*menu.TreeNode) []*system.MenuNode {
	out := make([]*system.MenuNode, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, toMenuNode(n))
	}
	return out
}

func fromMenuSaveRequest(req *system.MenuSaveRequest) *model.SystemMenu {
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

// ========== 角色 ==========

func toRoleDTO(r *model.SystemRole) *system.RoleDTO {
	return &system.RoleDTO{
		ID: r.ID, Name: r.Name, Code: r.Code, Description: r.Description,
		State: r.State, DataScope: int32(r.DataScope), OrderNum: r.OrderNum,
	}
}

func toRoleDTOs(list []*model.SystemRole) []*system.RoleDTO {
	out := make([]*system.RoleDTO, 0, len(list))
	for _, r := range list {
		out = append(out, toRoleDTO(r))
	}
	return out
}

func fromRoleSaveRequest(req *system.RoleSaveRequest) *model.SystemRole {
	r := &model.SystemRole{
		Name: req.Name, Code: req.Code, State: req.State,
		DataScope: int16(req.DataScope), OrderNum: req.OrderNum,
	}
	if req.Description != nil {
		r.Description = *req.Description
	}
	return r
}
