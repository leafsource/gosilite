package user

import (
	"gosilite/application/base/ctxutil"
	"gosilite/application/framework/dal/model"
)

// toLoginUser 用户模型 → 登录用户上下文（头像需传入已解析的资源 URL）
func toLoginUser(u *model.SystemUser, avatarURL string, roles, permissions []string) *ctxutil.LoginUser {
	return &ctxutil.LoginUser{
		ID:          u.ID,
		Username:    u.Username,
		Account:     u.Account,
		AvatarURL:   avatarURL,
		Roles:       roles,
		Permissions: permissions,
		DeptID:      u.DeptID,
	}
}

// toAdminListItem 用户模型 + 关联信息 → 管理端用户行
func toAdminListItem(u *model.SystemUser, deptName string, roleIDs, postIDs []int64) *AdminListItem {
	return &AdminListItem{
		User: u, DeptName: deptName,
		RoleIDs: roleIDs, PostIDs: postIDs,
	}
}
