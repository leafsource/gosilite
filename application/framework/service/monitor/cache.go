package monitor

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"gosilite/application/infra/cache"
	"gosilite/pkg/errorx"
	"gosilite/types/consts"
	"gosilite/types/errno"
)

// 缓存监控：动态前缀聚合 + 键列表 + 读值/删除（只操作 gosilite:*）。

const (
	scanLimit     = 5000
	keyListLimit  = 1000
	clearBatchMax = 200
)

var prefixRemarks = map[string]string{
	consts.AuthRefreshTokenPrefix:          "刷新令牌",
	consts.AuthUserRefreshGenerationPrefix: "用户刷新代数",
	consts.AuthAccessTokenDenyPrefix:       "Access 吊销黑名单",
	consts.AuthOnlineSessionsKey:           "在线会话（Hash）",
}

type CacheService struct {
	cache cache.CacheClient
}

// InitService 由 module 调用；此处直接构造并返回（门面挂在包级）。
func NewCacheService(c cache.CacheClient) *CacheService {
	return &CacheService{cache: c}
}

// SystemCacheSVC 包级门面。
var SystemCacheSVC *CacheService

func InitCacheService(c cache.CacheClient) *CacheService {
	SystemCacheSVC = NewCacheService(c)
	return SystemCacheSVC
}

// PrefixList 扫描 gosilite:* 并按前缀聚合。
func (s *CacheService) PrefixList(ctx context.Context) ([]CachePrefixVO, error) {
	keys, err := s.cache.ScanAll(ctx, "gosilite:*", scanLimit)
	if err != nil {
		return nil, err
	}
	counts := map[string]int64{}
	for _, k := range keys {
		p := prefixOf(k)
		counts[p]++
	}
	out := make([]CachePrefixVO, 0, len(counts))
	for p, c := range counts {
		out = append(out, CachePrefixVO{
			Prefix: p,
			Count:  c,
			Remark: prefixRemarks[p],
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Prefix < out[j].Prefix
	})
	return out, nil
}

// KeyList 列出前缀下真实 key；单 Hash key 则展开 field 作为「键名」行。
func (s *CacheService) KeyList(ctx context.Context, prefix string) ([]CacheKeyVO, error) {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return nil, errorx.New(errno.ErrSystemInvalidParam, errorx.KV("field", "prefix"))
	}
	if !strings.HasPrefix(prefix, "gosilite:") {
		return nil, errorx.New(errno.ErrSystemInvalidParam, errorx.KV("field", "prefix"))
	}
	match := prefix + "*"
	// 单键前缀（如 online_sessions）本身即完整 key
	if !strings.HasSuffix(prefix, ":") {
		if exact, err := s.cache.Exists(ctx, prefix).Result(); err == nil && exact > 0 {
			typ, _ := s.cache.TypeOf(ctx, prefix)
			if typ == "hash" {
				return s.hashFieldRows(ctx, prefix)
			}
			return []CacheKeyVO{{Key: prefix, Type: typ}}, nil
		}
		match = prefix + "*"
	}
	keys, err := s.cache.ScanAll(ctx, match, keyListLimit)
	if err != nil {
		return nil, err
	}
	filtered := keys[:0]
	for _, k := range keys {
		if strings.HasPrefix(k, prefix) {
			filtered = append(filtered, k)
		}
	}
	sort.Strings(filtered)
	out := make([]CacheKeyVO, 0, len(filtered))
	for _, k := range filtered {
		typ, _ := s.cache.TypeOf(ctx, k)
		if typ == "hash" {
			rows, err := s.hashFieldRows(ctx, k)
			if err != nil {
				return nil, err
			}
			out = append(out, rows...)
			continue
		}
		out = append(out, CacheKeyVO{Key: k, Type: typ})
	}
	return out, nil
}

// hashFieldRows hash 的每个 field 一行（键名 = field）。
func (s *CacheService) hashFieldRows(ctx context.Context, key string) ([]CacheKeyVO, error) {
	m, err := s.cache.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	fields := make([]string, 0, len(m))
	for f := range m {
		fields = append(fields, f)
	}
	sort.Strings(fields)
	out := make([]CacheKeyVO, 0, len(fields))
	for _, f := range fields {
		out = append(out, CacheKeyVO{Key: key, Type: "hash", Field: f})
	}
	return out, nil
}

// Value 读单 key / hash field 内容。
func (s *CacheService) Value(ctx context.Context, key, field string) (*CacheValueVO, error) {
	key = strings.TrimSpace(key)
	field = strings.TrimSpace(field)
	if key == "" || !strings.HasPrefix(key, "gosilite:") {
		return nil, errorx.New(errno.ErrSystemInvalidParam, errorx.KV("field", "key"))
	}
	typ, err := s.cache.TypeOf(ctx, key)
	if err != nil {
		return nil, err
	}
	out := &CacheValueVO{Key: key, Type: typ, DisplayKey: key}
	switch typ {
	case "none":
		out.Value = ""
	case "string":
		v, err := s.cache.Get(ctx, key).Result()
		if err != nil && err != cache.EmptyValue {
			return nil, err
		}
		out.Value = v
	case "hash":
		if field != "" {
			v, err := s.cache.HGetAll(ctx, key).Result()
			if err != nil {
				return nil, err
			}
			out.Value = v[field]
			out.DisplayKey = field
			break
		}
		m, err := s.cache.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		b, _ := json.MarshalIndent(m, "", "  ")
		out.Value = string(b)
	default:
		out.Value = "(type not supported for preview: " + typ + ")"
	}
	return out, nil
}

// DeleteKey 删除单 key；field 非空时删除 hash field。
func (s *CacheService) DeleteKey(ctx context.Context, key, field string) error {
	key = strings.TrimSpace(key)
	field = strings.TrimSpace(field)
	if key == "" || !strings.HasPrefix(key, "gosilite:") {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("field", "key"))
	}
	if field != "" {
		return s.cache.HDel(ctx, key, field).Err()
	}
	return s.cache.Delete(ctx, key).Err()
}

// ClearPrefix 清空前缀下所有 key。
func (s *CacheService) ClearPrefix(ctx context.Context, prefix string) (int64, error) {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || !strings.HasPrefix(prefix, "gosilite:") {
		return 0, errorx.New(errno.ErrSystemInvalidParam, errorx.KV("field", "prefix"))
	}
	match := prefix
	if !strings.HasSuffix(prefix, ":") {
		match = prefix + "*"
	}
	keys, err := s.cache.ScanAll(ctx, match, scanLimit)
	if err != nil {
		return 0, err
	}
	var n int64
	for i := 0; i < len(keys); i += clearBatchMax {
		end := i + clearBatchMax
		if end > len(keys) {
			end = len(keys)
		}
		batch := keys[i:end]
		for _, k := range batch {
			if !strings.HasPrefix(k, prefix) && k != prefix {
				continue
			}
			if c, err := s.cache.Delete(ctx, k).Result(); err == nil {
				n += c
			}
		}
	}
	return n, nil
}

// prefixOf 已知完整 key 原样返回；否则截最后一段为动态前缀。
func prefixOf(key string) string {
	if key == consts.AuthOnlineSessionsKey {
		return key
	}
	if i := strings.LastIndex(key, ":"); i > 0 {
		return key[:i+1]
	}
	return key
}
