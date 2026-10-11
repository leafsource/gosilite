// Package dynamic 提供按数据库配置动态构建/切换的 S3 存储客户端。
//
// 设计（spec: docs/compose/spec/system-oss.md S2.3）：
//   - Controller 实现 storage.Storage，所有调用委托给当前客户端，对消费方透明；
//   - 当前客户端来自 resolver（读 system_oss_config 的默认行）；resolver 返回空/错误
//     或构建失败时回退 fallback（conf.yaml 构造的客户端），既有功能永不断；
//   - 配置变更后调 Invalidate()，下次调用重建。
package dynamic

import (
	"context"
	"fmt"
	"io"
	"sync"

	"gosilite/application/infra/storage"
	"gosilite/application/infra/storage/impl/s3"
	"gosilite/conf"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

// Config system_oss_config 默认行（status='1'）。
type Config struct {
	ConfigKey    string
	AccessKey    string
	SecretKey    string
	BucketName   string
	Prefix       string
	Endpoint     string
	Domain       string
	IsHTTPS      string
	Region       string
	AccessPolicy string // 0=private 1=public 2=custom
}

// Meta 当前生效配置的展示/拼 URL 元信息。
type Meta struct {
	ConfigKey    string
	Prefix       string
	Endpoint     string
	BucketName   string
	Domain       string // URL 主机（domain 为空时用 endpoint）
	IsHTTPS      bool
	AccessPolicy string
}

// Resolver 返回默认配置；无默认行时返回 (nil, nil)。
type Resolver func(ctx context.Context) (*Config, error)

// KeyResolver 按 config_key 查配置；不存在返回 (nil, nil)。
type KeyResolver func(ctx context.Context, configKey string) (*Config, error)

// Builder 由配置构建客户端；默认走 s3.New（含 CheckAndCreateBucket）。
type Builder func(ctx context.Context, cfg *Config) (storage.Storage, error)

// Controller 动态存储控制器。
type Controller struct {
	resolver Resolver
	keyRes   KeyResolver
	builder  Builder
	fallback storage.Storage

	mu sync.Mutex
	// store/meta 默认客户端（Resolve）
	store storage.Storage
	meta  *Meta
	built bool
	// byKey 非默认配置的客户端缓存（文件台账按 service 字段回溯）
	byKey map[string]storage.Storage
}

// New 创建控制器。fallback 必须非 nil（conf 构造，启动期已校验）。
func New(resolver Resolver, fallback storage.Storage) *Controller {
	return NewWithBuilder(resolver, fallback, defaultBuilder)
}

// NewWithBuilder 供测试注入构建器。
func NewWithBuilder(resolver Resolver, fallback storage.Storage, b Builder) *Controller {
	return &Controller{resolver: resolver, builder: b, fallback: fallback}
}

func defaultBuilder(ctx context.Context, cfg *Config) (storage.Storage, error) {
	return s3.New(ctx, cfg.AccessKey, cfg.SecretKey, cfg.BucketName, cfg.Endpoint, cfg.Region)
}

// Resolve 确保当前客户端就绪，返回客户端与生效配置元信息（回退 conf 时 meta 取 conf 值）。
func (c *Controller) Resolve(ctx context.Context) (storage.Storage, *Meta, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.built {
		return c.store, c.meta, nil
	}

	cfg, err := c.resolver(ctx)
	if err != nil {
		hlog.CtxWarnf(ctx, "storage dynamic: resolve default config failed, fallback to conf: %v", err)
		s, m := c.useFallback()
		return s, m, nil
	}
	if cfg == nil {
		s, m := c.useFallback()
		return s, m, nil
	}
	store, err := c.builder(ctx, cfg)
	if err != nil {
		hlog.CtxWarnf(ctx, "storage dynamic: build client for %q failed, fallback to conf: %v", cfg.ConfigKey, err)
		s, m := c.useFallback()
		return s, m, nil
	}
	c.store = store
	c.meta = metaOf(cfg)
	c.built = true
	return c.store, c.meta, nil
}

func (c *Controller) useFallback() (storage.Storage, *Meta) {
	c.store = c.fallback
	c.meta = fallbackMeta()
	c.built = true
	return c.store, c.meta
}

// Invalidate 清缓存，下次 Resolve 重建。
func (c *Controller) Invalidate() {
	c.mu.Lock()
	c.built = false
	c.store = nil
	c.meta = nil
	c.byKey = nil
	c.mu.Unlock()
}

// WithKeyResolver 注入按 config_key 的配置查询（文件台账按 service 回溯客户端）。
func (c *Controller) WithKeyResolver(kr KeyResolver) *Controller {
	c.mu.Lock()
	c.keyRes = kr
	c.mu.Unlock()
	return c
}

// ByKey 按 config_key 取客户端；配置不存在返回错误（不回退 conf——台账记录了确切的服务商）。
// "conf" 是回退客户端的标识（fallbackMeta.ConfigKey），上传时走了 conf 保底的记录下载/删除走回退客户端。
func (c *Controller) ByKey(ctx context.Context, configKey string) (storage.Storage, error) {
	if configKey == "conf" || configKey == "" {
		return c.fallback, nil
	}
	c.mu.Lock()
	if c.byKey != nil {
		if s, ok := c.byKey[configKey]; ok {
			c.mu.Unlock()
			return s, nil
		}
	}
	kr := c.keyRes
	c.mu.Unlock()

	if kr == nil {
		return nil, fmt.Errorf("storage config %q not found: key resolver not configured", configKey)
	}
	cfg, err := kr(ctx, configKey)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, fmt.Errorf("storage config %q not found", configKey)
	}
	s, err := c.builder(ctx, cfg)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	if c.byKey == nil {
		c.byKey = map[string]storage.Storage{}
	}
	c.byKey[configKey] = s
	c.mu.Unlock()
	return s, nil
}

// --- storage.Storage 委托 ---

func (c *Controller) current(ctx context.Context) (storage.Storage, error) {
	s, _, err := c.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (c *Controller) PutObject(ctx context.Context, objectKey string, content []byte, opts ...storage.PutOptFn) error {
	s, err := c.current(ctx)
	if err != nil {
		return err
	}
	return s.PutObject(ctx, objectKey, content, opts...)
}

func (c *Controller) PutObjectWithReader(ctx context.Context, objectKey string, content io.Reader, opts ...storage.PutOptFn) error {
	s, err := c.current(ctx)
	if err != nil {
		return err
	}
	return s.PutObjectWithReader(ctx, objectKey, content, opts...)
}

func (c *Controller) GetObject(ctx context.Context, objectKey string) ([]byte, error) {
	s, err := c.current(ctx)
	if err != nil {
		return nil, err
	}
	return s.GetObject(ctx, objectKey)
}

func (c *Controller) DeleteObject(ctx context.Context, objectKey string) error {
	s, err := c.current(ctx)
	if err != nil {
		return err
	}
	return s.DeleteObject(ctx, objectKey)
}

func (c *Controller) GetObjectUrl(ctx context.Context, objectKey string, opts ...storage.GetOptFn) (string, error) {
	s, err := c.current(ctx)
	if err != nil {
		return "", err
	}
	return s.GetObjectUrl(ctx, objectKey, opts...)
}

func (c *Controller) HeadObject(ctx context.Context, objectKey string, opts ...storage.GetOptFn) (*storage.FileInfo, error) {
	s, err := c.current(ctx)
	if err != nil {
		return nil, err
	}
	return s.HeadObject(ctx, objectKey, opts...)
}

func (c *Controller) ListAllObjects(ctx context.Context, prefix string, opts ...storage.GetOptFn) ([]*storage.FileInfo, error) {
	s, err := c.current(ctx)
	if err != nil {
		return nil, err
	}
	return s.ListAllObjects(ctx, prefix, opts...)
}

func (c *Controller) ListObjectsPaginated(ctx context.Context, input *storage.ListObjectsPaginatedInput, opts ...storage.GetOptFn) (*storage.ListObjectsPaginatedOutput, error) {
	s, err := c.current(ctx)
	if err != nil {
		return nil, err
	}
	return s.ListObjectsPaginated(ctx, input, opts...)
}

// --- resolver / meta ---

// DBResolver 从 system_oss_config 读取默认行（status='1'）。无行返回 (nil, nil)。
func DBResolver(db *gorm.DB) Resolver {
	return func(ctx context.Context) (*Config, error) {
		return queryConfig(ctx, db, "status = '1'")
	}
}

// DBKeyResolver 按 config_key 读配置。无行返回 (nil, nil)。
func DBKeyResolver(db *gorm.DB) KeyResolver {
	return func(ctx context.Context, configKey string) (*Config, error) {
		return queryConfig(ctx, db, "config_key = ?", configKey)
	}
}

func queryConfig(ctx context.Context, db *gorm.DB, where string, args ...any) (*Config, error) {
	var row struct {
		ConfigKey    string `gorm:"column:config_key"`
		AccessKey    string `gorm:"column:access_key"`
		SecretKey    string `gorm:"column:secret_key"`
		BucketName   string `gorm:"column:bucket_name"`
		Prefix       string `gorm:"column:prefix"`
		Endpoint     string `gorm:"column:endpoint"`
		Domain       string `gorm:"column:domain"`
		IsHTTPS      string `gorm:"column:is_https"`
		Region       string `gorm:"column:region"`
		AccessPolicy string `gorm:"column:access_policy"`
	}
	err := db.WithContext(ctx).
		Table("system_oss_config").
		Select("config_key, access_key, secret_key, bucket_name, prefix, endpoint, domain, is_https, region, access_policy").
		Where(where, args...).
		Where("del_flag = 0").
		First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &Config{
		ConfigKey: row.ConfigKey, AccessKey: row.AccessKey, SecretKey: row.SecretKey,
		BucketName: row.BucketName, Prefix: row.Prefix, Endpoint: row.Endpoint,
		Domain: row.Domain, IsHTTPS: row.IsHTTPS, Region: row.Region, AccessPolicy: row.AccessPolicy,
	}, nil
}

func metaOf(cfg *Config) *Meta {
	domain := cfg.Domain
	if domain == "" {
		domain = cfg.Endpoint
	}
	return &Meta{
		ConfigKey:    cfg.ConfigKey,
		Prefix:       cfg.Prefix,
		Endpoint:     cfg.Endpoint,
		BucketName:   cfg.BucketName,
		Domain:       domain,
		IsHTTPS:      cfg.IsHTTPS == "Y" || cfg.IsHTTPS == "1",
		AccessPolicy: cfg.AccessPolicy,
	}
}

func fallbackMeta() *Meta {
	sc := conf.GetConf().Storage
	return &Meta{
		ConfigKey:    "conf",
		Endpoint:     sc.Endpoint,
		BucketName:   sc.BucketName,
		Domain:       firstNonEmpty(sc.Domain, sc.Endpoint),
		IsHTTPS:      false,
		AccessPolicy: "1",
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

var _ storage.Storage = (*Controller)(nil)

// String 便于日志辨识。
func (c *Controller) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.built {
		return "dynamic<unresolved>"
	}
	return fmt.Sprintf("dynamic<%s>", c.meta.ConfigKey)
}
