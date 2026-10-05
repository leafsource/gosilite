package menu

import (
	"context"

	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"

	"gorm.io/gorm"
)

// SystemMenuSVC is the package facade for handlers; set once by InitService.
var SystemMenuSVC *Service

// Service holds menu business operations.
type Service struct {
	repo repository.ISystemMenuRepository
}

// NewService builds a menu service with explicit dependencies.
func NewService(db *gorm.DB) *Service {
	return &Service{repo: repository.NewSystemMenuRepository(db)}
}

// InitService installs the package facade used by handlers.
func InitService(ctx context.Context, db *gorm.DB) *Service {
	SystemMenuSVC = NewService(db)
	return SystemMenuSVC
}

func (s *Service) Create(ctx context.Context, entity *model.SystemMenu) error {
	return s.repo.Create(ctx, entity)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*model.SystemMenu, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, entity *model.SystemMenu) error {
	return s.repo.Update(ctx, entity)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, page, pageSize int) ([]*model.SystemMenu, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}

// AllMenu returns all menus.
func (s *Service) AllMenu(ctx context.Context) ([]*model.SystemMenu, error) {
	return s.repo.Find(ctx)
}
