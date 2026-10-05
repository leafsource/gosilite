package role

import (
	"context"

	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"

	"gorm.io/gorm"
)

// SystemRoleSVC is the package facade for handlers; set once by InitService.
var SystemRoleSVC *Service

// Service holds role business operations.
type Service struct {
	repo repository.ISystemRoleRepository
}

// NewService builds a role service with explicit dependencies.
func NewService(db *gorm.DB) *Service {
	return &Service{repo: repository.NewSystemRoleRepository(db)}
}

// InitService installs the package facade used by handlers.
func InitService(ctx context.Context, db *gorm.DB) *Service {
	SystemRoleSVC = NewService(db)
	return SystemRoleSVC
}

func (s *Service) Create(ctx context.Context, entity *model.SystemRole) error {
	return s.repo.Create(ctx, entity)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*model.SystemRole, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, entity *model.SystemRole) error {
	return s.repo.Update(ctx, entity)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, page, pageSize int) ([]*model.SystemRole, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}
