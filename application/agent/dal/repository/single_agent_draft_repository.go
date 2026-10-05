package repository

import (
	"context"
	"gosilite/application/agent/dal/model"
	"gosilite/application/agent/dal/query"
	"gorm.io/gorm"
)

// SingleAgentDraftRepository 仓库接口
type ISingleAgentDraftRepository interface {
	// Create 创建SingleAgentDraft
	Create(ctx context.Context, SingleAgentDraft *model.SingleAgentDraft) error

	// GetByID 根据ID获取SingleAgentDraft
	GetByID(ctx context.Context, id int64) (*model.SingleAgentDraft, error)

	// Update 更新SingleAgentDraft
	Update(ctx context.Context, SingleAgentDraft *model.SingleAgentDraft) error

	// Delete 删除SingleAgentDraft
	Delete(ctx context.Context, id int64) error

	// List 获取SingleAgentDraft列表
	List(ctx context.Context, page, pageSize int) ([]*model.SingleAgentDraft, int64, error)
}

// SingleAgentDraftRepositoryImpl 仓库实现
type SingleAgentDraftRepositoryImpl struct {
	query *query.Query
}

// NewSingleAgentDraftRepository 创建仓库实例
func NewSingleAgentDraftRepository(db *gorm.DB) ISingleAgentDraftRepository {
	return &SingleAgentDraftRepositoryImpl{
		query: query.Use(db),
	}
}

// Create 创建SingleAgentDraft
func (r *SingleAgentDraftRepositoryImpl) Create(ctx context.Context, SingleAgentDraft *model.SingleAgentDraft) error {
	return r.query.SingleAgentDraft.WithContext(ctx).Create(SingleAgentDraft)
}

// GetByID 根据ID获取SingleAgentDraft
func (r *SingleAgentDraftRepositoryImpl) GetByID(ctx context.Context, id int64) (*model.SingleAgentDraft, error) {
	entity, err := r.query.SingleAgentDraft.Where(r.query.SingleAgentDraft.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// Update 更新SingleAgentDraft
func (r *SingleAgentDraftRepositoryImpl) Update(ctx context.Context, SingleAgentDraft *model.SingleAgentDraft) error {
	return r.query.SingleAgentDraft.WithContext(ctx).Save(SingleAgentDraft)
}

// Delete 删除SingleAgentDraft
func (r *SingleAgentDraftRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.query.SingleAgentDraft.WithContext(ctx).Where(r.query.SingleAgentDraft.ID.Eq(id)).Delete()
	return err
}

// List 获取SingleAgentDraft列表
func (r *SingleAgentDraftRepositoryImpl) List(ctx context.Context, page, pageSize int) ([]*model.SingleAgentDraft, int64, error) {
	// 计算总数
	total, err := r.query.SingleAgentDraft.WithContext(ctx).Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	entities, err := r.query.SingleAgentDraft.WithContext(ctx).Offset(offset).Limit(pageSize).Find()
	if err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}
