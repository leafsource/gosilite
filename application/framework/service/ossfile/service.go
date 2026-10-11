package ossfile

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"

	"gosilite/application/base/ctxutil"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	"gosilite/application/framework/service/ossrule"
	infraStorage "gosilite/application/infra/storage"
	"gosilite/application/infra/storage/dynamic"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

const (
	// MaxUploadBytes 上传大小上限 20MB（spec S2.5）
	MaxUploadBytes = 20 << 20

	listFetchPageSize = 1000
	maxListPages      = 500 // 遍历上限，防游标不收敛
	maxPageSize       = 100
	// private 桶 URL 有效期 1h（源码 matchingUrl 为 120s，页面停留几分钟图即挂，放宽到 1h 保浏览体验）
	privateURLExpireSeconds = 3600
)

// RuleResolver 规则消费（ossrule.Service 实现）
type RuleResolver interface {
	ResolveURLs(ctx context.Context, originalURL, contentType string, ruleNames []string, useDefault bool) (ossrule.ResolveResult, error)
	// RuleNames 全部启用规则名：限定 List 装饰只收集存在的变体（含非默认规则）
	RuleNames(ctx context.Context) ([]string, error)
}

// StorageCtl 动态存储控制器能力面（*dynamic.Controller 实现）
type StorageCtl interface {
	Resolve(ctx context.Context) (infraStorage.Storage, *dynamic.Meta, error)
	ByKey(ctx context.Context, configKey string) (infraStorage.Storage, error)
}

// Service 文件管理：台账 CRUD + 上传下载 + 桶视图（吸收旧存储管理）。
// 装配：module.Init 传 deps.StorageCtl 与 ossrule.SystemOssRuleSVC。
type Service struct {
	repo  repository.IOssFileRepository
	ctl   StorageCtl
	rules RuleResolver
}

// SystemOssSVC 门面
var SystemOssSVC *Service

// NewService 构造
func NewService(repo repository.IOssFileRepository, ctl StorageCtl, rules RuleResolver) *Service {
	return &Service{repo: repo, ctl: ctl, rules: rules}
}

// InitService 安装门面
func InitService(repo repository.IOssFileRepository, ctl StorageCtl, rules RuleResolver) *Service {
	SystemOssSVC = NewService(repo, ctl, rules)
	return SystemOssSVC
}

// ===== 台账 =====

// FileVO 台账行（含规则变体）
type FileVO struct {
	*model.SystemOss
	// RuleUrls 规则命中的变体 URL（key=ruleName）；overwrite 命中时 URL 字段已被改写
	RuleUrls map[string]string
}

// Page 分页 + 匹配 URL（private 桶转 120s 预签名）+ 规则变体
func (s *Service) Page(ctx context.Context, f repository.OssFileFilter) ([]*FileVO, int64, error) {
	list, total, err := s.repo.Page(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	vos, err := s.decorate(ctx, list)
	if err != nil {
		return nil, 0, err
	}
	return vos, total, nil
}

// Get 详情（同 Page 的 URL/规则装饰）
func (s *Service) Get(ctx context.Context, id int64) (*FileVO, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errorx.New(errno.ErrOssFileNotFound, errorx.KV("id", fmt.Sprintf("%d", id)))
	}
	vos, err := s.decorate(ctx, []*model.SystemOss{m})
	if err != nil || len(vos) == 0 {
		if err != nil {
			return nil, err
		}
		return nil, errorx.New(errno.ErrOssFileNotFound, errorx.KV("id", fmt.Sprintf("%d", id)))
	}
	return vos[0], nil
}

// decorate：private 桶 presign + ResolveURLs 变体
func (s *Service) decorate(ctx context.Context, list []*model.SystemOss) ([]*FileVO, error) {
	out := make([]*FileVO, 0, len(list))
	store, meta, err := s.ctl.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range list {
		vo := &FileVO{SystemOss: m, RuleUrls: map[string]string{}}
		vo.URL = m.URL
		// 仅 private 桶转临时 URL（源码 matchingUrl）
		if meta != nil && meta.AccessPolicy == "0" && store != nil {
			if u, err := store.GetObjectUrl(ctx, m.FileName, infraStorage.WithExpire(privateURLExpireSeconds)); err == nil {
				vo.URL = u
			}
		}
		if s.rules != nil {
			// 显式传全部规则名：主 URL 仍走 overwrite/默认语义，变体收集所有命中的非覆盖规则
			// （ruleNames=nil 时源码语义只取默认规则，会导致非默认规则变体永不出现，功能不闭环）
			names, nErr := s.rules.RuleNames(ctx)
			if nErr != nil {
				names = nil
			}
			if rr, err := s.rules.ResolveURLs(ctx, vo.URL, m.ContentType, names, false); err == nil {
				vo.URL = rr.URL
				if len(rr.Variants) > 0 {
					vo.RuleUrls = rr.Variants
				}
			}
		}
		out = append(out, vo)
	}
	return out, nil
}

// Upload 上传（multipart 内容已由 handler 读出）：写桶 + 插台账。
func (s *Service) Upload(ctx context.Context, filename, contentType string, content []byte) (*FileVO, error) {
	if len(content) == 0 {
		return nil, errorx.New(errno.ErrOssFileInvalidParam, errorx.KV("msg", "上传文件不能为空"))
	}
	if len(content) > MaxUploadBytes {
		return nil, errorx.New(errno.ErrOssFileInvalidParam, errorx.KV("msg", "文件超过 20MB 上限"))
	}
	store, meta, err := s.ctl.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	if meta == nil || store == nil {
		return nil, errorx.New(errno.ErrOssFileInvalidParam, errorx.KV("msg", "存储未配置默认配置，请先在 OSS配置 中设默认"))
	}

	key := buildUploadKey(meta.Prefix, filename)
	var opts []infraStorage.PutOptFn
	if contentType != "" {
		opts = append(opts, infraStorage.WithContentType(contentType))
	}
	if err := store.PutObject(ctx, key, content, opts...); err != nil {
		return nil, errorx.New(errno.ErrOssFileUploadFailed, errorx.KV("msg", err.Error()))
	}
	accessURL := buildAccessURL(meta, key)

	suffix := strings.TrimPrefix(strings.ToLower(path.Ext(filename)), ".")
	m := &model.SystemOss{
		FileName:     key,
		OriginalName: filename,
		FileSuffix:   suffix,
		URL:          accessURL,
		Size:         int64(len(content)),
		ContentType:  contentType,
		IsLock:       false,
		Service:      meta.ConfigKey,
		DelFlag:      0,
	}
	if login := ctxutil.GetLoginUser(ctx); login != nil {
		if login.Username != "" {
			m.CreateBy = login.Username
		} else {
			m.CreateBy = login.Account
		}
	}
	if err := s.repo.Create(ctx, m); err != nil {
		// 台账失败：对象已落桶，删除保一致（尽力而为）
		_ = store.DeleteObject(ctx, key)
		return nil, errorx.New(errno.ErrOssFileUploadFailed, errorx.KV("msg", err.Error()))
	}
	vos, err := s.decorate(ctx, []*model.SystemOss{m})
	if err != nil || len(vos) == 0 {
		return &FileVO{SystemOss: m, RuleUrls: map[string]string{}}, nil
	}
	return vos[0], nil
}

// Download 取下载内容与原始文件名（handler 负责回流附件头）。
func (s *Service) Download(ctx context.Context, id int64) (filename string, contentType string, content []byte, err error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return "", "", nil, errorx.New(errno.ErrOssFileNotFound, errorx.KV("id", fmt.Sprintf("%d", id)))
	}
	store, err := s.ctl.ByKey(ctx, m.Service)
	if err != nil {
		return "", "", nil, errorx.New(errno.ErrOssFileServiceMissing, errorx.KV("key", m.Service))
	}
	data, err := store.GetObject(ctx, m.FileName)
	if err != nil {
		return "", "", nil, errorx.New(errno.ErrOssFileDownloadFailed, errorx.KV("msg", err.Error()))
	}
	return m.OriginalName, m.ContentType, data, nil
}

// Update 编辑：仅 originalName / isLock（源码 updateByBo 语义，无分类字段）
func (s *Service) Update(ctx context.Context, m *model.SystemOss) error {
	if _, err := s.repo.GetByID(ctx, m.ID); err != nil {
		return errorx.New(errno.ErrOssFileNotFound, errorx.KV("id", fmt.Sprintf("%d", m.ID)))
	}
	if strings.TrimSpace(m.OriginalName) == "" {
		return errorx.New(errno.ErrOssFileInvalidParam, errorx.KV("msg", "原名不能为空"))
	}
	return s.repo.SaveEditable(ctx, m)
}

// Delete 台账删除：锁定拒绝 → 软删 → 按各行 service 删桶对象（对象删除失败只记日志，源码语义）
func (s *Service) Delete(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return errorx.New(errno.ErrOssFileInvalidParam, errorx.KV("msg", "ids 不能为空"))
	}
	rows, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return errorx.New(errno.ErrOssFileNotFound, errorx.KV("id", "batch"))
	}
	for _, r := range rows {
		if r.IsLock {
			return errorx.New(errno.ErrOssFileLocked)
		}
	}
	if err := s.repo.SoftDelete(ctx, ids); err != nil {
		return errorx.New(errno.ErrOssFileDeleteFailed, errorx.KV("msg", err.Error()))
	}
	// 台账已删，真实删对象尽力而为（失败不影响业务结果，源码 removeRealOss 语义）
	for _, r := range rows {
		store, err := s.ctl.ByKey(ctx, r.Service)
		if err != nil {
			continue
		}
		_ = store.DeleteObject(ctx, r.FileName)
	}
	return nil
}

// ===== 辅助 =====

// buildUploadKey <prefix>uploads/<yyyyMMdd>/<uuid>.<ext>；扩展名只保留字母数字
func buildUploadKey(prefix, filename string) string {
	ext := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, strings.ToLower(strings.TrimPrefix(path.Ext(filename), ".")))

	prefix = strings.Trim(prefix, "/")
	key := "uploads/" + time.Now().Format("20060102") + "/" + uuid.NewString()
	if ext != "" {
		key += "." + ext
	}
	if prefix != "" {
		key = prefix + "/" + key
	}
	return key
}

// buildAccessURL 按 domain/is_https 拼访问 URL（domain 空则用 Endpoint 回退在 metaOf 已处理）
func buildAccessURL(meta *dynamic.Meta, key string) string {
	scheme := "http"
	if meta.IsHTTPS {
		scheme = "https"
	}
	return scheme + "://" + strings.TrimRight(meta.Domain, "/") + "/" + strings.TrimLeft(key, "/")
}
