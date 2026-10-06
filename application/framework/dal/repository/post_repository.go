package repository

import (
	"context"

	"gosilite/application/framework/dal/model"

	"gorm.io/gorm"
)

// IPostRepository 岗位仓库
type IPostRepository interface {
	Create(ctx context.Context, post *model.SystemPost) error
	GetByID(ctx context.Context, id int64) (*model.SystemPost, error)
	Update(ctx context.Context, post *model.SystemPost) error
	SoftDelete(ctx context.Context, id int64) error
	List(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemPost, int64, error)
	ExistsCode(ctx context.Context, code string, excludeID int64) (bool, error)
	FindByIDs(ctx context.Context, ids []int64) ([]*model.SystemPost, error)
	ReplaceUserPosts(ctx context.Context, userID int64, postIDs []int64) error
	ListPostIDsByUser(ctx context.Context, userID int64) ([]int64, error)
}

type postRepository struct {
	db *gorm.DB
}

// NewPostRepository 创建岗位仓库
func NewPostRepository(db *gorm.DB) IPostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *model.SystemPost) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *postRepository) GetByID(ctx context.Context, id int64) (*model.SystemPost, error) {
	var post model.SystemPost
	err := r.db.WithContext(ctx).Where("id = ? AND del_flag = 0", id).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) Update(ctx context.Context, post *model.SystemPost) error {
	return r.db.WithContext(ctx).Save(post).Error
}

func (r *postRepository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.SystemPost{}).Where("id = ?", id).Update("del_flag", 2).Error
}

func (r *postRepository) List(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemPost, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}
	q := r.db.WithContext(ctx).Model(&model.SystemPost{}).Where("del_flag = 0")
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("post_name LIKE ? OR post_code LIKE ?", kw, kw)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.SystemPost
	err := q.Order("order_num, id").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *postRepository) ExistsCode(ctx context.Context, code string, excludeID int64) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.SystemPost{}).
		Where("post_code = ? AND del_flag = 0 AND id <> ?", code, excludeID).
		Count(&n).Error
	return n > 0, err
}

func (r *postRepository) FindByIDs(ctx context.Context, ids []int64) ([]*model.SystemPost, error) {
	var list []*model.SystemPost
	if len(ids) == 0 {
		return list, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ? AND del_flag = 0", ids).Find(&list).Error
	return list, err
}

func (r *postRepository) ReplaceUserPosts(ctx context.Context, userID int64, postIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.SystemUserPost{}).Error; err != nil {
			return err
		}
		for _, pid := range postIDs {
			if err := tx.Create(&model.SystemUserPost{UserID: userID, PostID: pid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *postRepository) ListPostIDsByUser(ctx context.Context, userID int64) ([]int64, error) {
	var ids []int64
	err := r.db.WithContext(ctx).Model(&model.SystemUserPost{}).
		Where("user_id = ?", userID).Pluck("post_id", &ids).Error
	return ids, err
}
