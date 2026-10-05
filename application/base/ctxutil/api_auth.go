package ctxutil

import (
	"context"
	"gosilite/pkg/ctxcache"
	"gosilite/types/consts"
)

type LoginUser struct {
	ID          int64    `json:"id"`
	Account     string   `json:"account"`
	Username    string   `json:"username"`
	AvatarURL   string   `json:"avatar_url"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

func GetLoginUser(ctx context.Context) *LoginUser {
	data, ok := ctxcache.Get[*LoginUser](ctx, consts.LoginUserDataCtx)
	if !ok {
		return nil
	}
	return data
}

func GetLoginUserID(ctx context.Context) int64 {
	if loginUser := GetLoginUser(ctx); loginUser != nil {

		return loginUser.ID
	}
	panic("login user is nil")
}

func SetLoginUser(ctx context.Context, loginUser *LoginUser) {
	ctxcache.Store(ctx, consts.LoginUserDataCtx, loginUser)
}
