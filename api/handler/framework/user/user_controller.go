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
	"io"
	"strings"
	"time"

	user "gosilite/api/model/framework/user"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// Login godoc
// @Summary 用户登录
// @Description 用户登录接口
// @Tags 用户
// @Accept json
// @Produce json
// @Param login body user.LoginRequest true "登录参数"
// @Success 200 {object} response.Resp
// @router /user/login [POST]
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
// @router /user/info [GET]
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
// @router user/profile [GET]
func GetProfile(ctx context.Context, c *app.RequestContext) {
	loginUser := ctxutil.GetLoginUser(ctx)
	R.SuccessWithData(ctx, c, toLoginUserResponse(loginUser))
}

// UpdateAvatar .
// @router /user/update/upload_avatar [POST]
func UpdateAvatar(ctx context.Context, c *app.RequestContext) {
	var err error
	var req user.UserUpdateAvatarRequest

	// Get the uploaded file
	file, err := c.FormFile("avatar")
	if err != nil {
		hlog.CtxErrorf(ctx, "Get Avatar Fail failed, err=%v", err)
		R.InvalidParamRequestResponse(c, "missing avatar file")
		return
	}
	// Check file type
	if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		R.InvalidParamRequestResponse(c, "invalid file type, only image allowed")
		return
	}
	// Read file content
	src, err := file.Open()
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	defer src.Close()

	fileContent, err := io.ReadAll(src)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}

	req.Avatar = fileContent
	mimeType := file.Header.Get("Content-Type")
	ext := strings.Split(mimeType, "/")[1]

	avatarURL, err := userService.SystemUserSVC.UpdateAvatar(ctx, ctxutil.GetLoginUserID(ctx), ext, req.Avatar)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}

	R.SuccessWithData(ctx, c, &user.UserUpdateAvatarResponse{
		WebURI: avatarURL,
	})
}

// Refresh .
// @router /user/refresh [POST]
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
// @router /user/logout [POST]
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
// @router /user/password [POST]
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
