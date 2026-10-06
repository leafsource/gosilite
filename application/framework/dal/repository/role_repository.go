package repository

import (
	"context"

	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/query"

	"gorm.io/gorm"
)

// SystemRoleRepository 仓库接口
type ISystemRoleRepository interface {
	Create(ctx context.Context, SystemRole *model.SystemRole) error
	GetByID(ctx context.Context, id int64) (*model.SystemRole, error)
	Update(ctx context.Context, SystemRole *model.SystemRole) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, pageSize int) ([]*model.SystemRole, int64, error)
	ExistsCode(ctx context.Context, code string, excludeID int64) (bool, error)
	ListPage(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemRole, int64, error)
}

// SystemRoleRepositoryImpl 仓库实现
type SystemRoleRepositoryImpl struct {
	query *query.Query
	db    *gorm.DB
}

// NewSystemRoleRepository 创建仓库实例
func NewSystemRoleRepository(db *gorm.DB) ISystemRoleRepository {
	return &SystemRoleRepositoryImpl{
		query: query.Use(db),
		db:    db,
	}
}

func (r *SystemRoleRepositoryImpl) Create(ctx context.Context, SystemRole *model.SystemRole) error {
	return r.query.SystemRole.WithContext(ctx).Create(SystemRole)
}

func (r *SystemRoleRepositoryImpl) GetByID(ctx context.Context, id int64) (*model.SystemRole, error) {
	entity, err := r.query.SystemRole.Where(r.query.SystemRole.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *SystemRoleRepositoryImpl) Update(ctx context.Context, SystemRole *model.SystemRole) error {
	return r.query.SystemRole.WithContext(ctx).Save(SystemRole)
}

func (r *SystemRoleRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.query.SystemRole.WithContext(ctx).Where(r.query.SystemRole.ID.Eq(id)).Delete()
	return err
}

func (r *SystemRoleRepositoryImpl) List(ctx context.Context, page, pageSize int) ([]*model.SystemRole, int64, error) {
	total, err := r.query.SystemRole.WithContext(ctx).Count()
	if err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	entities, err := r.query.SystemRole.WithContext(ctx).Offset(offset).Limit(pageSize).Find()
	if err != nil {
		return nil, 0, err
	}
	return entities, total, nil
}

func (r *SystemRoleRepositoryImpl) ExistsCode(ctx context.Context, code string, excludeID int64) (bool, error) {
	total, err := r.query.SystemRole.WithContext(ctx).
		Where(r.query.SystemRole.Code.Eq(code), r.query.SystemRole.ID.Neq(excludeID)).
		Count()
	if err != nil {
		return false, err
	}
	return total > 0, nil
}

func (r *SystemRoleRepositoryImpl) ListPage(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemRole, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}
	q := r.db.WithContext(ctx).Model(&model.SystemRole{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("name LIKE ? OR code LIKE ?", kw, kw)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var entities []*model.SystemRole
	err := q.Order("order_num, id").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&entities).Error
	if err != nil {
		return nil, 0, err
	}
	return entities, total, nil
}
