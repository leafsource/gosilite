package ossrule

import (
	"context"
	"net/url"
	"strconv"
	"regexp"
	"strings"
	"sync"

	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

// SystemOssRuleSVC 门面
var SystemOssRuleSVC *Service

// Service OSS处理规则
type Service struct {
	repo repository.IOssRuleRepository

	// 进程内规则快照缓存（单节点）；任何变更/手动刷新清空，下次读时重建
	mu    sync.RWMutex
	cache []*model.SystemOssRule
}

// NewService 构造
func NewService(repo repository.IOssRuleRepository) *Service {
	return &Service{repo: repo}
}

// InitService 安装门面
func InitService(repo repository.IOssRuleRepository) *Service {
	SystemOssRuleSVC = NewService(repo)
	return SystemOssRuleSVC
}

// ===== CRUD =====

// Page 分页列表
func (s *Service) Page(ctx context.Context, page, pageSize int, ruleName, domain string, status, isDefault *string) ([]*model.SystemOssRule, int64, error) {
	return s.repo.Page(ctx, page, pageSize, ruleName, domain, status, isDefault)
}

// Get 详情
func (s *Service) Get(ctx context.Context, id int64) (*model.SystemOssRule, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errorx.New(errno.ErrOssRuleNotFound, errorx.KV("id", itoa(id)))
	}
	return m, nil
}

// Create 新增（同域名规则名唯一）
func (s *Service) Create(ctx context.Context, m *model.SystemOssRule) error {
	if strings.TrimSpace(m.RuleName) == "" || strings.TrimSpace(m.Domain) == "" || strings.TrimSpace(m.Rule) == "" {
		return errorx.New(errno.ErrOssRuleInvalidParam, errorx.KV("msg", "规则名称/匹配域名/规则不能为空"))
	}
	exists, err := s.repo.ExistsDomainName(ctx, m.Domain, m.RuleName, 0)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrOssRuleExists, errorx.KV("domain", m.Domain))
	}
	m.IsOverwrite = "N"
	if m.Status == "" {
		m.Status = "1"
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return err
	}
	s.invalidate()
	return nil
}

// Update 编辑：从非默认改默认时强制关闭覆盖（源码 updateByBo 语义）
func (s *Service) Update(ctx context.Context, m *model.SystemOssRule) error {
	old, err := s.repo.GetByID(ctx, m.ID)
	if err != nil {
		return errorx.New(errno.ErrOssRuleNotFound, errorx.KV("id", itoa(m.ID)))
	}
	exists, err := s.repo.ExistsDomainName(ctx, m.Domain, m.RuleName, m.ID)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrOssRuleExists, errorx.KV("domain", m.Domain))
	}
	if old.IsDefault == "N" && m.IsDefault == "Y" {
		m.IsOverwrite = "N"
	} else {
		// Save 不触碰 is_overwrite，保留原值
		m.IsOverwrite = old.IsOverwrite
	}
	if err := s.repo.Save(ctx, m); err != nil {
		return err
	}
	s.invalidate()
	return nil
}

// Delete 软删
func (s *Service) Delete(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return errorx.New(errno.ErrOssRuleInvalidParam, errorx.KV("msg", "ids 不能为空"))
	}
	if err := s.repo.SoftDelete(ctx, ids); err != nil {
		return err
	}
	s.invalidate()
	return nil
}

// ChangeOverwrite 覆盖开关：目标 Y 且该规则是默认 → 同域名其他默认规则覆盖置 N
func (s *Service) ChangeOverwrite(ctx context.Context, id int64, isOverwrite string) (*model.SystemOssRule, error) {
	if isOverwrite != "Y" && isOverwrite != "N" {
		return nil, errorx.New(errno.ErrOssRuleInvalidParam, errorx.KV("msg", "isOverwrite 只能为 Y/N"))
	}
	rule, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errorx.New(errno.ErrOssRuleNotFound, errorx.KV("id", itoa(id)))
	}
	if isOverwrite == "Y" && rule.IsDefault == "Y" {
		if err := s.repo.DisableOverwriteForDomain(ctx, rule.Domain); err != nil {
			return nil, err
		}
	}
	if err := s.repo.UpdateOverwrite(ctx, id, isOverwrite); err != nil {
		return nil, err
	}
	s.invalidate()
	return s.repo.GetByID(ctx, id)
}

// RefreshCache 手动刷新规则缓存（保留源码交互按钮）
func (s *Service) RefreshCache(ctx context.Context) error {
	s.invalidate()
	// 立即重建，让刷新可观测（DB 故障时把错误抛给调用方）
	_, err := s.snapshot(ctx)
	return err
}

// RuleNames 启用规则的全部名字（供 ossfile 装饰时限定变体收集范围，
// 覆盖源码 @OssRule 注解层的「指定规则名」语义——spec S2.6 已知差异的闭环实现）
func (s *Service) RuleNames(ctx context.Context) ([]string, error) {
	rules, err := s.snapshot(ctx)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(rules))
	for _, r := range rules {
		names = append(names, r.RuleName)
	}
	return names, nil
}

func (s *Service) invalidate() {
	s.mu.Lock()
	s.cache = nil
	s.mu.Unlock()
}

func (s *Service) snapshot(ctx context.Context) ([]*model.SystemOssRule, error) {
	s.mu.RLock()
	cached := s.cache
	s.mu.RUnlock()
	if cached != nil {
		return cached, nil
	}
	list, err := s.repo.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.cache = list
	s.mu.Unlock()
	return list, nil
}

// ===== URL 消费（spec S2.6，照抄 ruoyi SysOssRuleServiceImpl.getUrls）=====

// ResolveResult 规则解析结果
type ResolveResult struct {
	// URL 最终主 URL：命中 overwrite 规则时为渲染结果，否则为原 URL
	URL string
	// Variants 非覆盖规则变体，key=ruleName（如 80x80）
	Variants map[string]string
}

// ResolveURLs 按域名+mime 匹配规则并渲染。
// ruleNames 限定使用的规则名；为空或 useDefault=true 时只用 is_default=Y 的规则。
// 已知差异（spec 记录）：模板只支持 #{#var} 字面变量替换，不支持 SpringEL 表达式。
func (s *Service) ResolveURLs(ctx context.Context, originalURL, contentType string, ruleNames []string, useDefault bool) (ResolveResult, error) {
	res := ResolveResult{URL: originalURL, Variants: map[string]string{}}
	if strings.TrimSpace(originalURL) == "" {
		return res, nil
	}
	rules, err := s.snapshot(ctx)
	if err != nil {
		return res, err
	}
	if len(rules) == 0 {
		return res, nil
	}

	// 源码支持逗号分隔的多 URL（splitUrl），逐段处理后合并
	urls := splitURLs(originalURL)
	overwriteURL := ""
	var variants []struct{ name, u string }

	fileName := lastPathSegment(originalURL)
	for _, u := range urls {
		hits := filterRules(rules, u, contentType, fileName, ruleNames, useDefault)
		vars := urlVariables(u)

		// overwrite：命中集合中第一条覆盖规则渲染并替换主 URL
		for _, r := range hits {
			if r.IsOverwrite == "Y" {
				overwriteURL = renderRule(r.Rule, vars)
				break
			}
		}
		// 非覆盖：生成 ruleName 变体
		for _, r := range hits {
			if r.IsOverwrite == "N" {
				variants = append(variants, struct{ name, u string }{r.RuleName, renderRule(r.Rule, vars)})
			}
		}
	}
	if overwriteURL != "" {
		res.URL = overwriteURL
	}
	for _, v := range variants {
		res.Variants[v.name] = v.u
	}
	return res, nil
}

// filterRules 域名 contains + mime 匹配 + default/规则名限定
func filterRules(rules []*model.SystemOssRule, u, contentType, fileName string, ruleNames []string, useDefault bool) []*model.SystemOssRule {
	var hits []*model.SystemOssRule
	for _, r := range rules {
		if !strings.Contains(u, r.Domain) {
			continue
		}
		if !mimeMatch(r.MimeType, contentType, fileName) {
			continue
		}
		if len(ruleNames) == 0 || useDefault {
			if r.IsDefault != "Y" {
				continue
			}
		} else if !contains(ruleNames, r.RuleName) {
			continue
		}
		hits = append(hits, r)
	}
	return hits
}

// mimeMatch：contentType 含任一 mime 分量，或文件名以分量结尾（源码 StrUtil 语义）
func mimeMatch(mimeType, contentType, fileName string) bool {
	if strings.TrimSpace(mimeType) == "" {
		return false
	}
	segs := strings.Split(mimeType, ",")
	lowerName := strings.ToLower(fileName)
	for _, seg := range segs {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		if contentType != "" && strings.Contains(contentType, seg) {
			return true
		}
		if strings.HasSuffix(lowerName, strings.ToLower(seg)) {
			return true
		}
	}
	return false
}

// ruleTemplateRe 匹配 #{#var}
var ruleTemplateRe = regexp.MustCompile(`\#{#(\w+)}`)

// renderRule 渲染规则串中的 #{#var} 占位（domain/path/fullPath/filename/url/fullUrl/query）
func renderRule(rule string, vars map[string]string) string {
	if !strings.Contains(rule, "#{") {
		return rule
	}
	return ruleTemplateRe.ReplaceAllStringFunc(rule, func(m string) string {
		sub := ruleTemplateRe.FindStringSubmatch(m)
		if len(sub) == 2 {
			if v, ok := vars[sub[1]]; ok {
				return v
			}
		}
		return m
	})
}

// urlVariables 解析 URL 得模板变量（源码 getUrlVariable 语义）
func urlVariables(raw string) map[string]string {
	vars := map[string]string{
		"fullUrl": raw,
		"query":   "",
		"domain":  "",
		"path":    "",
		"fullPath": "",
		"filename": "",
		"url":      raw,
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return vars
	}
	schemeHost := parsed.Scheme
	if parsed.Host != "" {
		if schemeHost != "" {
			schemeHost += "://"
		}
		schemeHost += parsed.Host
	}
	vars["domain"] = schemeHost
	vars["fullPath"] = parsed.EscapedPath()
	vars["query"] = parsed.RawQuery
	segs := strings.Split(strings.Trim(parsed.EscapedPath(), "/"), "/")
	if len(segs) == 1 && segs[0] == "" {
		segs = nil
	}
	if len(segs) > 0 {
		vars["filename"] = segs[len(segs)-1]
		if len(segs) > 1 {
			vars["path"] = "/" + strings.Join(segs[:len(segs)-1], "/")
		} else {
			vars["path"] = ""
		}
	}
	vars["url"] = schemeHost + vars["fullPath"]
	return vars
}

// splitURLs 按 ",http" 边界拆多 URL（源码 splitUrl 用 ,(?=http) 前瞻，Go regexp 不支持，手动扫描等价实现）
func splitURLs(u string) []string {
	if strings.TrimSpace(u) == "" {
		return nil
	}
	var parts []string
	rest := u
	for {
		idx := strings.Index(rest, ",http")
		if idx < 0 {
			parts = append(parts, rest)
			break
		}
		parts = append(parts, rest[:idx])
		rest = rest[idx+1:] // 保留 http 起始
	}
	return parts
}

func lastPathSegment(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	path := parsed.EscapedPath()
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[idx+1:]
	}
	return path
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}
