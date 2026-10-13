package menu

import (
	"gosilite/application/framework/dal/model"
)

// toTreeNode 菜单模型 → 管理端树节点
func toTreeNode(m *model.SystemMenu) *TreeNode {
	return &TreeNode{
		ID: m.ID, ParentID: m.ParentID, Label: m.Title,
		MenuType: m.MenuType, Perms: m.Perms, OrderNum: m.Sort,
		Visible: m.Visible != "hide" && m.Visible != "0", Status: m.Status,
		Path: m.Path, Component: m.Component, Icon: m.Icon, ActionType: m.ActionType,
		Redirect: m.Redirect,
	}
}
