package config

import (
	"context"
	"strconv"
	"strings"

	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

// SystemConfigSVC 门面
var SystemConfigSVC *Service

// Service 参数配置
type Service struct {
	repo repository.IConfigRepository
}

// NewService 构造
func NewService(repo repository.IConfigRepository) *Service {
	return &Service{repo: repo}
}

// InitService 安装门面
func InitService(repo repository.IConfigRepository) *Service {
	SystemConfigSVC = NewService(repo)
	return SystemConfigSVC
}

// Page 分页
func (s *Service) Page(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemConfig, int64, error) {
	return s.repo.Page(ctx, page, pageSize, keyword)
}

// Create 创建（key 唯一）
func (s *Service) Create(ctx context.Context, c *model.SystemConfig) error {
	if strings.TrimSpace(c.ConfigKey) == "" || strings.TrimSpace(c.ConfigName) == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "参数名称/键不能为空"))
	}
	exists, err := s.repo.ExistsKey(ctx, c.ConfigKey, 0)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrSystemConfigKeyExists, errorx.KV("key", c.ConfigKey))
	}
	if c.ConfigType == "" {
		c.ConfigType = "system"
	}
	c.DelFlag = 0
	return s.repo.Create(ctx, c)
}

// Update 更新
func (s *Service) Update(ctx context.Context, c *model.SystemConfig) error {
	if _, err := s.repo.GetByID(ctx, c.ID); err != nil {
		return errorx.New(errno.ErrSystemConfigNotFound, errorx.KV("id", formatID(c.ID)))
	}
	exists, err := s.repo.ExistsKey(ctx, c.ConfigKey, c.ID)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrSystemConfigKeyExists, errorx.KV("key", c.ConfigKey))
	}
	return s.repo.Save(ctx, c)
}

// Delete 软删
func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.SoftDelete(ctx, id)
}

// ListGlobal 匿名可读的全局参数（config_type=global 且启用）
func (s *Service) ListGlobal(ctx context.Context) ([]*model.SystemConfig, error) {
	return s.repo.ListByType(ctx, "global")
}

// GetByKey 运行时按键取值
func (s *Service) GetByKey(ctx context.Context, key string) (*model.SystemConfig, error) {
	m, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, errorx.New(errno.ErrSystemConfigNotFound, errorx.KV("key", key))
	}
	return m, nil
}

func formatID(v int64) string {
	return strconv.FormatInt(v, 10)
}
