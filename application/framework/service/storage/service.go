package storage

import (
	"context"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"

	infraStorage "gosilite/application/infra/storage"
	"gosilite/conf"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

const (
	// MaxUploadBytes 上传大小上限 10MB
	MaxUploadBytes = 10 << 20

	listFetchPageSize = 1000
	// 遍历上限（50 万对象），防游标不收敛
	maxListPages = 500
	maxPageSize  = 100
)

// Service 存储管理：infra storage 原语的探活/聚合/增删查，不写业务表。
type Service struct {
	store infraStorage.Storage
}

// SystemStorageSVC 包级门面。
var SystemStorageSVC *Service

// NewService 构造。
func NewService(store infraStorage.Storage) *Service {
	return &Service{store: store}
}

// InitService 安装门面。
func InitService(store infraStorage.Storage) *Service {
	SystemStorageSVC = NewService(store)
	return SystemStorageSVC
}

// Info 探活并聚合对象数与总大小；存储不可达时返回 connected=false + message（HTTP 200 业务信封）。
func (s *Service) Info(ctx context.Context) (*InfoVO, error) {
	cfg := conf.GetConf().Storage
	out := &InfoVO{
		Endpoint: cfg.Endpoint,
		Bucket:   cfg.BucketName,
		Domain:   cfg.Domain,
	}
	cursor := ""
	for i := 0; i < maxListPages; i++ {
		res, err := s.store.ListObjectsPaginated(ctx, &infraStorage.ListObjectsPaginatedInput{
			PageSize: listFetchPageSize,
			Cursor:   cursor,
		})
		if err != nil {
			out.Message = err.Error()
			return out, nil
		}
		for _, f := range res.Files {
			out.ObjectCount++
			out.TotalSize += f.Size
		}
		if !res.IsTruncated {
			break
		}
		cursor = res.Cursor
	}
	out.Connected = true
	return out, nil
}

// List 分页拉取后按 prefix 前缀与 keyword 包含过滤，返回过滤后 total。
func (s *Service) List(ctx context.Context, page, pageSize int, prefix, keyword string) ([]*ObjectVO, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	prefix = strings.TrimSpace(prefix)
	keyword = strings.TrimSpace(keyword)

	var matched []*ObjectVO
	cursor := ""
	for i := 0; i < maxListPages; i++ {
		res, err := s.store.ListObjectsPaginated(ctx, &infraStorage.ListObjectsPaginatedInput{
			Prefix:   prefix,
			PageSize: listFetchPageSize,
			Cursor:   cursor,
		})
		if err != nil {
			return nil, 0, errorx.New(errno.ErrStorageListFailed, errorx.KV("msg", err.Error()))
		}
		for _, f := range res.Files {
			if prefix != "" && !strings.HasPrefix(f.Key, prefix) {
				continue
			}
			if keyword != "" && !strings.Contains(f.Key, keyword) {
				continue
			}
			matched = append(matched, toObjectVO(f))
		}
		if !res.IsTruncated {
			break
		}
		cursor = res.Cursor
	}

	total := int64(len(matched))
	start := (page - 1) * pageSize
	if start >= len(matched) {
		return []*ObjectVO{}, total, nil
	}
	end := start + pageSize
	if end > len(matched) {
		end = len(matched)
	}
	return matched[start:end], total, nil
}

// Delete 删除对象。
func (s *Service) Delete(ctx context.Context, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "key is required"))
	}
	if err := s.store.DeleteObject(ctx, key); err != nil {
		return errorx.New(errno.ErrStorageDeleteFailed, errorx.KV("msg", err.Error()))
	}
	return nil
}

// Upload 随机 key（uploads/<yyyyMMdd>/<uuid>.<ext>）上传并返回可访问地址。
func (s *Service) Upload(ctx context.Context, filename, contentType string, content []byte) (*UploadVO, error) {
	if len(content) == 0 {
		return nil, errorx.New(errno.ErrUploadInvalidParamCode, errorx.KV("msg", "file is empty"))
	}
	if len(content) > MaxUploadBytes {
		return nil, errorx.New(errno.ErrUploadInvalidFileSizeCode)
	}
	key := buildUploadKey(filename)
	var opts []infraStorage.PutOptFn
	if contentType != "" {
		opts = append(opts, infraStorage.WithContentType(contentType))
	}
	if err := s.store.PutObject(ctx, key, content, opts...); err != nil {
		return nil, errorx.New(errno.ErrUploadSystemErrorCode, errorx.KV("msg", err.Error()))
	}
	uri, err := s.store.GetObjectUrl(ctx, key)
	if err != nil {
		return nil, errorx.New(errno.ErrUploadSystemErrorCode, errorx.KV("msg", err.Error()))
	}
	return &UploadVO{WebURI: uri, Key: key}, nil
}

// buildUploadKey uploads/<yyyyMMdd>/<uuid>.<ext>；扩展名只保留字母数字
func buildUploadKey(filename string) string {
	ext := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, strings.ToLower(strings.TrimPrefix(path.Ext(filename), ".")))

	key := "uploads/" + time.Now().Format("20060102") + "/" + uuid.NewString()
	if ext != "" {
		key += "." + ext
	}
	return key
}
