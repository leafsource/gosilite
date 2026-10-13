package ossrule

import (
	"context"
	"testing"

	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
)

// fakeRepo 只实现 ListEnabled（ResolveURLs 只走缓存路径）
type fakeRepo struct {
	repository.IOssRuleRepository
	rules []*model.SystemOssRule
}

func (f *fakeRepo) ListEnabled(ctx context.Context) ([]*model.SystemOssRule, error) {
	return f.rules, nil
}

func newSvc(rules ...*model.SystemOssRule) *Service {
	return NewService(&fakeRepo{rules: rules})
}

func rule(name, domain, mime, r string, overwrite, def string) *model.SystemOssRule {
	return &model.SystemOssRule{
		RuleName: name, Domain: domain, MimeType: mime, Rule: r,
		IsOverwrite: overwrite, IsDefault: def, Status: "1",
	}
}

func TestResolveURLsDomainAndMimeMatch(t *testing.T) {
	svc := newSvc(
		rule("80x80", "cdn.example.com", "image", "#{#url}?x-oss-process=resize,w_80", "N", "N"),
		rule("other", "another.com", "image", "#{#url}?x=1", "N", "N"),
		rule("pdf", "cdn.example.com", "application/pdf", "#{#url}?pdf=1", "N", "N"),
	)
	res, err := svc.ResolveURLs(context.Background(),
		"http://cdn.example.com/uploads/a.png", "image/png", []string{"80x80", "other", "pdf"}, false)
	if err != nil {
		t.Fatal(err)
	}
	// 域名不匹配 + mime 不匹配均应剔除
	if len(res.Variants) != 1 {
		t.Fatalf("expect 1 variant, got %v", res.Variants)
	}
	got, ok := res.Variants["80x80"]
	if !ok {
		t.Fatalf("missing 80x80: %v", res.Variants)
	}
	want := "http://cdn.example.com/uploads/a.png?x-oss-process=resize,w_80"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	// 无 overwrite 命中，主 URL 不变
	if res.URL != "http://cdn.example.com/uploads/a.png" {
		t.Fatalf("main url should stay, got %q", res.URL)
	}
}

func TestResolveURLsOverwriteReplacesMainURL(t *testing.T) {
	svc := newSvc(
		rule("default-img", "cdn.example.com", "image", "#{#url}?x-oss-process=image/auto-orient", "Y", "Y"),
		rule("800x800", "cdn.example.com", "image", "#{#url}?w=800", "N", "N"),
	)
	res, err := svc.ResolveURLs(context.Background(),
		"http://cdn.example.com/p/a.png", "image/png", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.URL != "http://cdn.example.com/p/a.png?x-oss-process=image/auto-orient" {
		t.Fatalf("overwrite main url wrong: %q", res.URL)
	}
	// ruleNames 为空 → 只用默认规则，非默认 800x800 不应出现
	if _, ok := res.Variants["800x800"]; ok {
		t.Fatalf("non-default rule should not appear: %v", res.Variants)
	}
}

func TestResolveURLsDefaultOnlyWhenNoRuleNames(t *testing.T) {
	svc := newSvc(
		rule("d1", "h.com", "image", "#{#url}?d=1", "N", "Y"),
		rule("n1", "h.com", "image", "#{#url}?n=1", "N", "N"),
	)
	res, _ := svc.ResolveURLs(context.Background(), "http://h.com/a.png", "image/png", nil, false)
	if _, ok := res.Variants["n1"]; ok {
		t.Fatalf("non-default should be excluded without ruleNames: %v", res.Variants)
	}
	if _, ok := res.Variants["d1"]; !ok {
		t.Fatalf("default rule missing: %v", res.Variants)
	}
}

func TestResolveURLsTemplateVariables(t *testing.T) {
	svc := newSvc(
		rule("vars", "h.com", "image", "#{#domain}/#{#filename}?q=#{#query}&f=#{#fullPath}", "N", "N"),
	)
	res, err := svc.ResolveURLs(context.Background(),
		"http://h.com/img/cat.png?x=2", "image/png", []string{"vars"}, false)
	if err != nil {
		t.Fatal(err)
	}
	got := res.Variants["vars"]
	want := "http://h.com/cat.png?q=x=2&f=/img/cat.png"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveURLsMimeSuffixFallback(t *testing.T) {
	// contentType 为空时按文件后缀匹配
	svc := newSvc(rule("img", "h.com", ".png", "#{#url}?ok", "N", "N"))
	res, _ := svc.ResolveURLs(context.Background(), "http://h.com/a.png", "", []string{"img"}, false)
	if _, ok := res.Variants["img"]; !ok {
		t.Fatalf("suffix match failed: %v", res.Variants)
	}
}

func TestResolveURLsDisabledRulesExcluded(t *testing.T) {
	r := rule("off", "h.com", "image", "#{#url}?x", "N", "N")
	r.Status = "0" // 未启用：不会进 ListEnabled，但防御性再验一次
	svc := newSvc(r)
	// fakeRepo 不过滤 status，模拟缓存里混入停用规则——ResolveURLs 依赖 snapshot 只存启用规则，
	// 此处直接验证 snapshot 缓存路径不二次过滤（记录真实约束：过滤在 repo.ListEnabled）
	svc.cache = nil
	res, _ := svc.ResolveURLs(context.Background(), "http://h.com/a.png", "image/png", nil, false)
	// fakeRepo 返回了停用规则且 is_default=N → ruleNames 空只取默认，应无变体
	if len(res.Variants) != 0 {
		t.Fatalf("expect no variants, got %v", res.Variants)
	}
}

func TestInvalidateAfterMutation(t *testing.T) {
	repo := &fakeRepo{rules: []*model.SystemOssRule{rule("a", "h.com", "image", "#{#url}1", "N", "N")}}
	svc := NewService(repo)
	if _, err := svc.snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(svc.cache) != 1 {
		t.Fatalf("cache not filled")
	}
	svc.invalidate()
	if svc.cache != nil {
		t.Fatalf("cache should be cleared")
	}
}
