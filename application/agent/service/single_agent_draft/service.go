package single_agent_draft

import (
	"context"

	"gosilite/application/agent/dal/model"
	"gosilite/application/agent/dal/repository"

	"gorm.io/gorm"
)

// SingleAgentDraftSVC is the package facade for handlers; set once by InitService.
var SingleAgentDraftSVC *Service

// Service holds single-agent draft business operations.
type Service struct {
	repo repository.ISingleAgentDraftRepository
}

// NewService builds the draft service with explicit dependencies.
func NewService(db *gorm.DB) *Service {
	return &Service{repo: repository.NewSingleAgentDraftRepository(db)}
}

// InitService installs the package facade used by handlers.
func InitService(ctx context.Context, db *gorm.DB) *Service {
	SingleAgentDraftSVC = NewService(db)
	return SingleAgentDraftSVC
}

func (s *Service) Create(ctx context.Context, entity *model.SingleAgentDraft) error {
	return s.repo.Create(ctx, entity)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*model.SingleAgentDraft, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, entity *model.SingleAgentDraft) error {
	return s.repo.Update(ctx, entity)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, page, pageSize int) ([]*model.SingleAgentDraft, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}
