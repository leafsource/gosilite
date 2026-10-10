package user

import (
	"context"

	"gosilite/application/framework/dal/repository"
	"gosilite/application/infra/cache"
	"gosilite/application/infra/imagex"

	"gorm.io/gorm"
)

// SystemUserSVC is the package facade for handlers and middleware; set once by InitService.
var SystemUserSVC *Service

// Service holds user and session business operations.
type Service struct {
	repo     repository.ISystemUserRepository
	imageOSS imagex.ImageX
	sessions *SessionStore
	posts    repository.IPostRepository
	roleRBAC repository.IRoleRBACRepository
}

// NewService builds a user service with explicit dependencies.
func NewService(db *gorm.DB, imageOSS imagex.ImageX, cacheClient cache.CacheClient) *Service {
	return &Service{
		repo:     repository.NewSystemUserRepository(db),
		imageOSS: imageOSS,
		sessions: NewSessionStore(cacheClient),
		posts:    repository.NewPostRepository(db),
		roleRBAC: repository.NewRoleRBACRepository(db),
	}
}

// InitService installs the package facade used by handlers and middleware.
func InitService(ctx context.Context, db *gorm.DB, imageOSS imagex.ImageX, cacheClient cache.CacheClient) *Service {
	SystemUserSVC = NewService(db, imageOSS, cacheClient)
	return SystemUserSVC
}

// ProfileVO 个人资料（AvatarURL 已转为可访问地址，CreatedAt 为展示格式）。
type ProfileVO struct {
	ID          int64
	Account     string
	Username    string
	AvatarURL   string
	Email       string
	MobilePhone string
	State       string
	CreatedAt   string
}
