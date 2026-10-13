package user

import (
	"context"
	"fmt"
	"time"

	"gosilite/application/base/ctxutil"
	"gosilite/pkg/bcrypt"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

func (s *Service) Login(ctx context.Context, account, password string) (*int64, error) {
	return s.login(ctx, account, password)
}

func (s *Service) Authorizator(ctx context.Context, userId int64) (*ctxutil.LoginUser, error) {
	user, err := s.repo.GetByID(ctx, userId)
	if err != nil {
		return nil, err
	}

	resource, err := s.imageOSS.GetResourceURL(ctx, user.AvatarURL)
	if err != nil {
		return nil, err
	}
	return toLoginUser(user, resource.URL,
		s.repo.FindRoleCodes(ctx, user.ID),
		s.repo.FindPermissionCodes(ctx, user.ID),
	), nil
}

func (s *Service) GetProfile(ctx context.Context) (*ProfileVO, error) {
	userId := ctxutil.GetLoginUserID(ctx)
	user, err := s.repo.GetByID(ctx, userId)
	if err != nil {
		return nil, err
	}
	avatarURL := user.AvatarURL
	if avatarURL != "" {
		resource, err := s.imageOSS.GetResourceURL(ctx, avatarURL)
		if err != nil {
			return nil, err
		}
		avatarURL = resource.URL
	}
	return &ProfileVO{
		ID:          user.ID,
		Account:     user.Account,
		Username:    user.Username,
		AvatarURL:   avatarURL,
		Email:       user.Email,
		MobilePhone: user.MobilePhone,
		State:       user.State,
		CreatedAt:   user.CreateTime.Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateAvatarByFileId 统一文件服务两步链路第二步（spec S2.11）：
// 按 /common/file/upload 返回的台账 fileId 回查对象 key 写入 avatar_url
// （保持相对 uri 语义，GetProfile 现有 GetResourceURL 转换零改动；服务端回查防客户端塞任意 url）。
func (s *Service) UpdateAvatarByFileId(ctx context.Context, userId, fileId int64) (string, error) {
	if s.ossFileRepo == nil {
		return "", errorx.New(errno.ErrUserInvalidParamCode, errorx.KV("msg", "file repo not ready"))
	}
	file, err := s.ossFileRepo.GetByID(ctx, fileId)
	if err != nil {
		return "", errorx.New(errno.ErrUserInvalidParamCode, errorx.KV("msg", "file not found"))
	}
	if err := s.repo.UpdateAvatar(ctx, userId, file.FileName); err != nil {
		return "", err
	}
	resource, err := s.imageOSS.GetResourceURL(ctx, file.FileName)
	if err != nil {
		// 台账 key 已写入，URL 转换失败返回 key 兜底
		return file.FileName, nil
	}
	return resource.URL, nil
}

func (s *Service) ChangePassword(ctx context.Context, userId int64, currentPassword, newPassword string) error {
	if newPassword == "" || len(newPassword) < 8 {
		return errorx.New(errno.ErrUserInvalidParamCode, errorx.KV("msg", "new password too short"))
	}

	user, err := s.repo.GetByID(ctx, userId)
	if err != nil {
		return err
	}
	if !bcrypt.CheckPassword(user.Password, currentPassword) {
		return errorx.New(errno.ErrUserInfoInvalidateCode)
	}

	hashed, err := bcrypt.HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.Password = hashed
	if err := s.repo.Update(ctx, user); err != nil {
		return err
	}
	return s.sessions.RevokeUserRefreshTokens(ctx, userId)
}

func (s *Service) IssueSessionTokens(ctx context.Context, userId int64) (*SessionTokens, error) {
	refreshToken, err := s.sessions.IssueRefreshToken(ctx, userId)
	if err != nil {
		return nil, err
	}
	return &SessionTokens{
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) RefreshSession(ctx context.Context, refreshToken string) (*SessionTokens, int64, error) {
	userId, err := s.sessions.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, 0, errorx.New(errno.ErrUserSessionInvalidateCode, errorx.KV("reason", err.Error()))
	}
	if err := s.sessions.RevokeRefreshToken(ctx, refreshToken); err != nil {
		return nil, 0, err
	}
	newRefresh, err := s.sessions.IssueRefreshToken(ctx, userId)
	if err != nil {
		return nil, 0, err
	}
	return &SessionTokens{
		RefreshToken: newRefresh,
	}, userId, nil
}

func (s *Service) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	return s.sessions.RevokeRefreshToken(ctx, refreshToken)
}

func (s *Service) RevokeUserSessions(ctx context.Context, userId int64) error {
	return s.sessions.RevokeUserRefreshTokens(ctx, userId)
}

func (s *Service) DenyAccessToken(ctx context.Context, accessToken string, expireUnix int64) error {
	expire := time.Unix(expireUnix, 0)
	if expireUnix <= 0 {
		expire = time.Now().Add(time.Hour)
	}
	return s.sessions.DenyAccessToken(ctx, accessToken, expire)
}

func (s *Service) IsAccessTokenDenied(ctx context.Context, accessToken string) (bool, error) {
	return s.sessions.IsAccessTokenDenied(ctx, accessToken)
}

var _ = fmt.Sprintf
