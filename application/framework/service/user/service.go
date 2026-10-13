package user

import (
	"context"

	"gosilite/application/framework/dal/repository"
	"gosilite/application/infra/cache"
	"gosilite/application/infra/imagex"
	"gosilite/conf"

	"gorm.io/gorm"
)

// SystemUserSVC is the package facade for handlers and middleware; set once by InitService.
var SystemUserSVC *Service

// Service holds user and session business operations.
type Service struct {
	repo                  repository.ISystemUserRepository
	imageOSS              imagex.ImageX
	sessions              *SessionStore
	posts                 repository.IPostRepository
	roleRBAC              repository.IRoleRBACRepository
	cache                 cache.CacheClient
	loginSecurityDefaults conf.LoginSecurity
	// ossFileRepo 台账仓储：按 fileId 回查对象 key 更新头像（spec S2.11 统一文件服务）
	ossFileRepo repository.IOssFileRepository
}

// NewService builds a user service with explicit dependencies.
func NewService(db *gorm.DB, imageOSS imagex.ImageX, cacheClient cache.CacheClient) *Service {
	return &Service{
		repo:                  repository.NewSystemUserRepository(db),
		ossFileRepo:           repository.NewOssFileRepository(db),
		imageOSS:              imageOSS,
		cache:                 cacheClient,
		loginSecurityDefaults: conf.GetConf().LoginSecurity,
		sessions:              NewSessionStore(cacheClient),
		posts:                 repository.NewPostRepository(db),
		roleRBAC:              repository.NewRoleRBACRepository(db),
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
