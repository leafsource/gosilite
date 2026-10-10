package user

import (
	user "gosilite/api/model/framework/user"
	"gosilite/application/base/ctxutil"
	userService "gosilite/application/framework/service/user"
)

// toLoginUserResponse 登录用户 → 个人信息响应
func toLoginUserResponse(loginUser *ctxutil.LoginUser) *user.LoginUserResponse {
	return &user.LoginUserResponse{
		User: &user.UserProfile{
			ID:        loginUser.ID,
			Account:   loginUser.Account,
			Username:  loginUser.Username,
			AvatarURL: loginUser.AvatarURL,
		},
		Roles:       loginUser.Roles,
		Permissions: loginUser.Permissions,
	}
}

// toSessionTokenResponse 会话令牌 → 令牌响应
func toSessionTokenResponse(accessToken, refreshToken string, expiresIn int64) *user.SessionTokenResponse {
	return &user.SessionTokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		RefreshToken: refreshToken,
	}
}

// toProfileResponse 个人资料 VO → 资料响应
func toProfileResponse(vo *userService.ProfileVO) *user.ProfileResponse {
	return &user.ProfileResponse{
		ID:          vo.ID,
		Account:     vo.Account,
		Username:    vo.Username,
		AvatarURL:   vo.AvatarURL,
		Email:       vo.Email,
		MobilePhone: vo.MobilePhone,
		State:       vo.State,
		CreatedAt:   vo.CreatedAt,
	}
}
