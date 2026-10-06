package post

import (
	"context"
	"strconv"
	"strings"

	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"

	"gorm.io/gorm"
)

// SystemPostSVC 门面
var SystemPostSVC *Service

// Service 岗位业务
type Service struct {
	repo repository.IPostRepository
}

// NewService 构造岗位服务
func NewService(db *gorm.DB) *Service {
	return &Service{repo: repository.NewPostRepository(db)}
}

// InitService 安装门面
func InitService(ctx context.Context, db *gorm.DB) *Service {
	SystemPostSVC = NewService(db)
	return SystemPostSVC
}

// Page 分页
func (s *Service) Page(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemPost, int64, error) {
	return s.repo.List(ctx, page, pageSize, keyword)
}

// Create 新建
func (s *Service) Create(ctx context.Context, p *model.SystemPost) error {
	if strings.TrimSpace(p.PostCode) == "" || strings.TrimSpace(p.PostName) == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "岗位编码/名称不能为空"))
	}
	exists, err := s.repo.ExistsCode(ctx, p.PostCode, 0)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrSystemPostCodeExists, errorx.KV("code", p.PostCode))
	}
	p.DelFlag = 0
	return s.repo.Create(ctx, p)
}

// Update 更新
func (s *Service) Update(ctx context.Context, p *model.SystemPost) error {
	if _, err := s.repo.GetByID(ctx, p.ID); err != nil {
		return errorx.New(errno.ErrSystemPostNotFound, errorx.KV("id", strconv.FormatInt(p.ID,10)))
	}
	exists, err := s.repo.ExistsCode(ctx, p.PostCode, p.ID)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrSystemPostCodeExists, errorx.KV("code", p.PostCode))
	}
	return s.repo.Update(ctx, p)
}

// Delete 软删
func (s *Service) Delete(ctx context.Context, id int64) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return errorx.New(errno.ErrSystemPostNotFound, errorx.KV("id", strconv.FormatInt(id, 10)))
	}
	return s.repo.SoftDelete(ctx, id)
}

