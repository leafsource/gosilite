package user

import (
	"context"
	"errors"

	"gosilite/api/middleware/permission"
	"gosilite/application/base/ctxutil"
	userService "gosilite/application/framework/service/user"
	"gosilite/pkg/errorx"
	R "gosilite/pkg/utils/response"
	"gosilite/types/errno"
	"time"

	user "gosilite/api/model/framework/user"

	"github.com/cloudwego/hertz/pkg/app"
)

// Login godoc
// @Summary 用户登录
// @Summary 用户登录
// @Description 用户登录接口
// @Tags 用户
// @Produce json
// @Param data body user.LoginRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /user/login [POST]
func Login(ctx context.Context, c *app.RequestContext) {
	var err error
	var req user.LoginRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	// 鉴权中间件
	permission.Security.LoginHandler(ctx, c)
}

// GetInfo .
// @Summary 获取当前登录用户
// @Description 获取当前登录用户接口
// @Tags 用户
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /user/info [GET]
func GetInfo(ctx context.Context, c *app.RequestContext) {
	loginUser := ctxutil.GetLoginUser(ctx)
	if loginUser == nil {
		R.Error(ctx, c, errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("reason", "session is required")))
		return
	}
	resp := toLoginUserResponse(loginUser)
	R.SuccessWithData(ctx, c, resp)
}

// GetProfile .
// @Summary 获取个人信息
// @Description 获取个人信息接口
// @Tags 用户
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /user/profile [GET]
func GetProfile(ctx context.Context, c *app.RequestContext) {
	profile, err := userService.SystemUserSVC.GetProfile(ctx)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toProfileResponse(profile))
}


// Refresh .
// @Summary 刷新访问令牌
// @Description 刷新访问令牌接口
// @Tags 用户
// @Produce json
// @Param data body user.RefreshRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /user/refresh [POST]
func Refresh(ctx context.Context, c *app.RequestContext) {
	var req user.RefreshRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	tokens, userId, err := userService.SystemUserSVC.RefreshSession(ctx, req.RefreshToken)
	if err != nil {
		// 只回业务消息：err.Error() 含 code/stack 全量 dump，不能直接进响应
		var se errorx.StatusError
		msg := err.Error()
		if errors.As(err, &se) {
			msg = se.Msg()
		}
		R.Unauthorized(c, msg)
		return
	}

	accessToken, expire, err := permission.MintAccessToken(userId)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	expiresIn := int64(time.Until(expire).Seconds())
	if expiresIn < 0 {
		expiresIn = 0
	}

	// refresh rotation 后迁移在线登记（内部容错）
	userService.SystemUserSVC.MoveOnline(ctx, userId, permission.BearerToken(c), accessToken,
		tokens.RefreshToken, c.ClientIP(), string(c.GetHeader("User-Agent")), expire)

	R.SuccessWithData(ctx, c, toSessionTokenResponse(accessToken, tokens.RefreshToken, expiresIn))
}

// Logout .
// @Summary 登出
// @Description 登出接口
// @Tags 用户
// @Produce json
// @Param data body user.LogoutRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /user/logout [POST]
func Logout(ctx context.Context, c *app.RequestContext) {
	var req user.LogoutRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	if req.IsSetRefreshToken() {
		if err := userService.SystemUserSVC.RevokeRefreshToken(ctx, req.GetRefreshToken()); err != nil {
			R.Error(ctx, c, err)
			return
		}
	}
	// 摘除在线登记（容错降级）
	userService.SystemUserSVC.DeregisterOnline(ctx, permission.BearerToken(c))
	if err := permission.DenyCurrentAccessToken(ctx, c); err != nil {
		R.Error(ctx, c, err)
		return
	}

	R.SuccessWithData(ctx, c, nil)
}

// ChangePassword .
// @Summary 修改密码
// @Description 修改密码接口
// @Tags 用户
// @Produce json
// @Param data body user.ChangePasswordRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /user/password [POST]
func ChangePassword(ctx context.Context, c *app.RequestContext) {
	var req user.ChangePasswordRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	loginUser := ctxutil.GetLoginUser(ctx)
	if loginUser == nil {
		R.Error(ctx, c, errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("reason", "session is required")))
		return
	}

	if err := userService.SystemUserSVC.ChangePassword(ctx, loginUser.ID, req.CurrentPassword, req.NewPassword); err != nil {
		R.Error(ctx, c, err)
		return
	}

	R.SuccessWithData(ctx, c, nil)
}


// UpdateProfileAvatar 按台账 fileId 更新头像（统一文件服务第二步，spec S2.11）
// @Summary 更新头像
// @Description 按统一文件服务返回的文件 id 更新头像（服务端回查台账防任意 url）
// @Tags 用户
// @Accept json
// @Produce json
// @Param data body user.UpdateProfileAvatarRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /user/profile/avatar [PUT]
func UpdateProfileAvatar(ctx context.Context, c *app.RequestContext) {
	var req user.UpdateProfileAvatarRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	login := ctxutil.GetLoginUser(ctx)
	if login == nil {
		R.InvalidParamRequestResponse(c, "login required")
		return
	}
	avatarURL, err := userService.SystemUserSVC.UpdateAvatarByFileId(ctx, login.ID, req.FileId)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, &user.UpdateProfileAvatarResponse{AvatarURL: avatarURL})
}
