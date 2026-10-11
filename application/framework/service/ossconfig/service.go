package ossconfig

import (
	"context"
	"fmt"
	"strings"

	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	infraStorage "gosilite/application/infra/storage"
	"gosilite/application/infra/storage/impl/s3"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

// SystemOssConfigSVC 门面
var SystemOssConfigSVC *Service

// Service OSS配置：DB 多配置 + 默认切换，变更后触发动态客户端失效。
// 装配：module.Init 传 deps.StorageCtl.Invalidate（spec system-oss S2.4）。
type Service struct {
	repo      repository.IOssConfigRepository
	invalidate func()
}

// NewService 构造（invalidate 可为 nil，测试用）
func NewService(repo repository.IOssConfigRepository, invalidate func()) *Service {
	return &Service{repo: repo, invalidate: invalidate}
}

// InitService 安装门面
func InitService(repo repository.IOssConfigRepository, invalidate func()) *Service {
	SystemOssConfigSVC = NewService(repo, invalidate)
	return SystemOssConfigSVC
}

func (s *Service) fireInvalidate() {
	if s.invalidate != nil {
		s.invalidate()
	}
}

// Page 分页（keyword 匹配 config_key/endpoint/bucket_name）
func (s *Service) Page(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemOssConfig, int64, error) {
	return s.repo.Page(ctx, page, pageSize, keyword)
}

// Get 详情
func (s *Service) Get(ctx context.Context, id int64) (*model.SystemOssConfig, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errorx.New(errno.ErrOssConfigNotFound, errorx.KV("id", fmt.Sprintf("%d", id)))
	}
	return m, nil
}

// Create 新增：config_key 唯一；status 强制 '0'（不能建即默认）
func (s *Service) Create(ctx context.Context, m *model.SystemOssConfig) error {
	if err := s.validate(m); err != nil {
		return err
	}
	exists, err := s.repo.ExistsKey(ctx, m.ConfigKey, 0)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrOssConfigKeyExists, errorx.KV("key", m.ConfigKey))
	}
	m.Status = "0"
	m.DelFlag = 0
	if err := s.repo.Create(ctx, m); err != nil {
		return errorx.New(errno.ErrOssConfigSaveFailed, errorx.KV("msg", err.Error()))
	}
	s.fireInvalidate()
	return nil
}

// Update 编辑：config_key 建库后不可改（台账 service 字段引用它，忽略请求值保留原值）
func (s *Service) Update(ctx context.Context, m *model.SystemOssConfig) error {
	old, err := s.repo.GetByID(ctx, m.ID)
	if err != nil {
		return errorx.New(errno.ErrOssConfigNotFound, errorx.KV("id", fmt.Sprintf("%d", m.ID)))
	}
	if err := s.validate(m); err != nil {
		return err
	}
	m.ConfigKey = old.ConfigKey // 保留原 key
	m.Status = old.Status       // 状态只由 ChangeStatus 改
	if err := s.repo.SaveEditable(ctx, m); err != nil {
		return errorx.New(errno.ErrOssConfigSaveFailed, errorx.KV("msg", err.Error()))
	}
	s.fireInvalidate()
	return nil
}

// Delete 拒绝删除默认配置；软删
func (s *Service) Delete(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return errorx.New(errno.ErrOssConfigInvalid, errorx.KV("msg", "ids 不能为空"))
	}
	for _, id := range ids {
		m, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return errorx.New(errno.ErrOssConfigNotFound, errorx.KV("id", fmt.Sprintf("%d", id)))
		}
		if m.Status == "1" {
			return errorx.New(errno.ErrOssConfigIsDefault)
		}
	}
	if err := s.repo.SoftDelete(ctx, ids); err != nil {
		return errorx.New(errno.ErrOssConfigSaveFailed, errorx.KV("msg", err.Error()))
	}
	s.fireInvalidate()
	return nil
}

// ChangeStatus 设默认：事务内全部置 0 → 目标置 1，即时生效
func (s *Service) ChangeStatus(ctx context.Context, id int64) (*model.SystemOssConfig, error) {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return nil, errorx.New(errno.ErrOssConfigNotFound, errorx.KV("id", fmt.Sprintf("%d", id)))
	}
	if err := s.repo.ChangeDefaultTx(ctx, id); err != nil {
		return nil, errorx.New(errno.ErrOssConfigSaveFailed, errorx.KV("msg", err.Error()))
	}
	s.fireInvalidate()
	return s.repo.GetByID(ctx, id)
}

// Test 连通探活：临时构建客户端 + List(limit=1)；错误不抛 HTTP 错，返回 connected=false（探活语义）
func (s *Service) Test(ctx context.Context, ak, sk, bucket, endpoint, region string) (bool, string) {
	if strings.TrimSpace(endpoint) == "" || strings.TrimSpace(bucket) == "" {
		return false, "endpoint / bucketName 不能为空"
	}
	store, err := s3.New(ctx, ak, sk, bucket, endpoint, region)
	if err != nil {
		return false, fmt.Sprintf("连接失败: %v", err)
	}
	if _, err := store.ListObjectsPaginated(ctx, &infraStorage.ListObjectsPaginatedInput{PageSize: 1}); err != nil {
		return false, fmt.Sprintf("列举对象失败: %v", err)
	}
	return true, "连接正常"
}

func (s *Service) validate(m *model.SystemOssConfig) error {
	if strings.TrimSpace(m.ConfigKey) == "" {
		return errorx.New(errno.ErrOssConfigInvalid, errorx.KV("msg", "configKey 不能为空"))
	}
	if strings.TrimSpace(m.BucketName) == "" {
		return errorx.New(errno.ErrOssConfigInvalid, errorx.KV("msg", "桶名称不能为空"))
	}
	if strings.TrimSpace(m.Endpoint) == "" {
		return errorx.New(errno.ErrOssConfigInvalid, errorx.KV("msg", "访问站点不能为空"))
	}
	if m.AccessPolicy != "0" && m.AccessPolicy != "1" && m.AccessPolicy != "2" {
		m.AccessPolicy = "1"
	}
	if m.IsHTTPS != "Y" && m.IsHTTPS != "N" {
		m.IsHTTPS = "N"
	}
	return nil
}
