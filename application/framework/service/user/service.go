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
}

// NewService builds a user service with explicit dependencies.
func NewService(db *gorm.DB, imageOSS imagex.ImageX, cacheClient cache.CacheClient) *Service {
	return &Service{
		repo:     repository.NewSystemUserRepository(db),
		imageOSS: imageOSS,
		sessions: NewSessionStore(cacheClient),
	}
}

// InitService installs the package facade used by handlers and middleware.
func InitService(ctx context.Context, db *gorm.DB, imageOSS imagex.ImageX, cacheClient cache.CacheClient) *Service {
	SystemUserSVC = NewService(db, imageOSS, cacheClient)
	return SystemUserSVC
}
