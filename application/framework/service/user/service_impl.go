package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gosilite/application/base/ctxutil"
	"gosilite/application/framework/dal/model"
	"gosilite/application/infra/imagex"
	"gosilite/pkg/bcrypt"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

func (s *Service) Login(ctx context.Context, account, password string) (*int64, error) {
	user, err := s.repo.GetByAccount(ctx, account)
	if err != nil {
		return nil, err
	}

	if !bcrypt.CheckPassword(user.Password, password) {
		return nil, errors.New("密码错误")
	}

	return &user.ID, nil
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

func (s *Service) GetProfile(ctx context.Context) (*model.SystemUser, error) {
	userId := ctxutil.GetLoginUserID(ctx)
	user, err := s.repo.GetByID(ctx, userId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) UpdateAvatar(ctx context.Context, userId int64, ext string, image []byte) (url string, err error) {
	avatarKey := "user_avatar/" + strconv.FormatInt(userId, 10) + "." + ext
	imageResult, err := s.imageOSS.Upload(ctx, image, imagex.WithStoreKey(avatarKey))
	if err != nil {
		return "", err
	}

	err = s.repo.UpdateAvatar(ctx, userId, imageResult.FileInfo.Uri)
	if err != nil {
		return "", err
	}

	resource, err := s.imageOSS.GetResourceURL(ctx, imageResult.FileInfo.Uri)
	if err != nil {
		return "", err
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
