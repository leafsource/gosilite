package ossfile

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	"gosilite/application/framework/service/ossrule"
	infraStorage "gosilite/application/infra/storage"
	"gosilite/application/infra/storage/dynamic"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

// --- fakes ---

type fakeRepo struct {
	rows     map[int64]*model.SystemOss
	pageRows []*model.SystemOss
	total    int64
	pageErr  error

	created   *model.SystemOss
	deleted   []int64
	saved     *model.SystemOss
	createdOK bool
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{rows: map[int64]*model.SystemOss{}}
}

func (f *fakeRepo) Page(_ context.Context, filter repository.OssFileFilter) ([]*model.SystemOss, int64, error) {
	if f.pageErr != nil {
		return nil, 0, f.pageErr
	}
	if f.pageRows != nil {
		return f.pageRows, f.total, nil
	}
	return nil, 0, nil
}
func (f *fakeRepo) GetByID(_ context.Context, id int64) (*model.SystemOss, error) {
	if m, ok := f.rows[id]; ok {
		return m, nil
	}
	return nil, errors.New("not found")
}
func (f *fakeRepo) GetByIDs(_ context.Context, ids []int64) ([]*model.SystemOss, error) {
	var out []*model.SystemOss
	for _, id := range ids {
		if m, ok := f.rows[id]; ok {
			out = append(out, m)
		}
	}
	return out, nil
}
func (f *fakeRepo) Create(_ context.Context, m *model.SystemOss) error {
	f.created = m
	f.createdOK = true
	m.ID = 1
	f.rows[1] = m
	return nil
}
func (f *fakeRepo) SaveEditable(_ context.Context, m *model.SystemOss) error {
	f.saved = m
	if row, ok := f.rows[m.ID]; ok {
		row.OriginalName = m.OriginalName
		row.IsLock = m.IsLock
	}
	return nil
}
func (f *fakeRepo) SoftDelete(_ context.Context, ids []int64) error {
	f.deleted = append(f.deleted, ids...)
	return nil
}

// fakeStore 最小存储
type fakeStore struct {
	putKey     string
	deletedKey string
	putErr     error
	getData    []byte
	url        string
}

func (f *fakeStore) PutObject(_ context.Context, key string, _ []byte, _ ...infraStorage.PutOptFn) error {
	if f.putErr != nil {
		return f.putErr
	}
	f.putKey = key
	return nil
}
func (f *fakeStore) PutObjectWithReader(context.Context, string, io.Reader, ...infraStorage.PutOptFn) error {
	return nil
}
func (f *fakeStore) GetObject(context.Context, string) ([]byte, error) { return f.getData, nil }
func (f *fakeStore) DeleteObject(_ context.Context, key string) error {
	f.deletedKey = key
	return nil
}
func (f *fakeStore) GetObjectUrl(_ context.Context, key string, _ ...infraStorage.GetOptFn) (string, error) {
	if f.url != "" {
		return f.url, nil
	}
	return "http://fake/" + key, nil
}
func (f *fakeStore) HeadObject(context.Context, string, ...infraStorage.GetOptFn) (*infraStorage.FileInfo, error) {
	return nil, nil
}
func (f *fakeStore) ListAllObjects(context.Context, string, ...infraStorage.GetOptFn) ([]*infraStorage.FileInfo, error) {
	return nil, nil
}
func (f *fakeStore) ListObjectsPaginated(context.Context, *infraStorage.ListObjectsPaginatedInput, ...infraStorage.GetOptFn) (*infraStorage.ListObjectsPaginatedOutput, error) {
	return &infraStorage.ListObjectsPaginatedOutput{}, nil
}

// fakeCtl
type fakeCtl struct {
	store  *fakeStore
	meta   *dynamic.Meta
	byKey  map[string]infraStorage.Storage
	keyErr error
}

func (f *fakeCtl) Resolve(context.Context) (infraStorage.Storage, *dynamic.Meta, error) {
	return f.store, f.meta, nil
}
func (f *fakeCtl) ByKey(_ context.Context, key string) (infraStorage.Storage, error) {
	if f.keyErr != nil {
		return nil, f.keyErr
	}
	if s, ok := f.byKey[key]; ok {
		return s, nil
	}
	return nil, errors.New("storage config not found")
}

// fakeRules
type fakeRules struct {
	result ossrule.ResolveResult
	err    error
	calls  int
}

func (f *fakeRules) RuleNames(context.Context) ([]string, error) { return nil, nil }

func (f *fakeRules) ResolveURLs(_ context.Context, u, _ string, _ []string, _ bool) (ossrule.ResolveResult, error) {
	f.calls++
	if f.err != nil {
		return ossrule.ResolveResult{}, f.err
	}
	res := f.result
	// 与真实实现一致：URL 缺省回显入参
	if res.URL == "" {
		res.URL = u
	}
	if res.Variants == nil {
		res.Variants = map[string]string{}
	}
	return res, nil
}

func errCode(err error) int32 {
	var se errorx.StatusError
	if errors.As(err, &se) {
		return se.Code()
	}
	return 0
}

func defaultMeta() *dynamic.Meta {
	return &dynamic.Meta{
		ConfigKey: "local", Prefix: "", Endpoint: "http://127.0.0.1:19000",
		BucketName: "skyx", Domain: "127.0.0.1:19000", IsHTTPS: false, AccessPolicy: "1",
	}
}

// --- tests ---

func TestUploadWritesKeyAndRecord(t *testing.T) {
	repo := newFakeRepo()
	store := &fakeStore{}
	svc := NewService(repo, &fakeCtl{store: store, meta: defaultMeta()}, &fakeRules{})

	vo, err := svc.Upload(context.Background(), "photo.PNG", "image/png", []byte("data"))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if !strings.HasPrefix(store.putKey, "uploads/") {
		t.Fatalf("key should start uploads/, got %q", store.putKey)
	}
	if !strings.HasSuffix(store.putKey, ".png") {
		t.Fatalf("key should end .png, got %q", store.putKey)
	}
	if repo.created == nil {
		t.Fatal("record not created")
	}
	if repo.created.Service != "local" {
		t.Fatalf("service should be config key, got %q", repo.created.Service)
	}
	if repo.created.FileSuffix != "png" {
		t.Fatalf("suffix should be lowercase no-dot, got %q", repo.created.FileSuffix)
	}
	// URL 按 domain 拼（http）
	if !strings.HasPrefix(vo.URL, "http://127.0.0.1:19000/") {
		t.Fatalf("access url wrong: %q", vo.URL)
	}
}

func TestUploadPrefixApplied(t *testing.T) {
	repo := newFakeRepo()
	store := &fakeStore{}
	meta := defaultMeta()
	meta.Prefix = "app1"
	svc := NewService(repo, &fakeCtl{store: store, meta: meta}, &fakeRules{})
	if _, err := svc.Upload(context.Background(), "a.txt", "text/plain", []byte("x")); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(store.putKey, "app1/uploads/") {
		t.Fatalf("prefix not applied: %q", store.putKey)
	}
}

func TestUploadSizeLimit(t *testing.T) {
	svc := NewService(newFakeRepo(), &fakeCtl{store: &fakeStore{}, meta: defaultMeta()}, &fakeRules{})
	_, err := svc.Upload(context.Background(), "big.bin", "application/octet-stream", make([]byte, MaxUploadBytes+1))
	if errCode(err) != int32(errno.ErrOssFileInvalidParam) {
		t.Fatalf("expect size limit err, got %v", err)
	}
}

func TestUploadWithoutDefaultConfig(t *testing.T) {
	svc := NewService(newFakeRepo(), &fakeCtl{store: &fakeStore{}, meta: nil}, &fakeRules{})
	_, err := svc.Upload(context.Background(), "a.txt", "text/plain", []byte("x"))
	if errCode(err) != int32(errno.ErrOssFileInvalidParam) {
		t.Fatalf("expect invalid param (no default config), got %v", err)
	}
}

func TestDeleteRejectsLocked(t *testing.T) {
	repo := newFakeRepo()
	repo.rows[1] = &model.SystemOss{ID: 1, FileName: "k", IsLock: true}
	svc := NewService(repo, &fakeCtl{store: &fakeStore{}, meta: defaultMeta()}, &fakeRules{})

	err := svc.Delete(context.Background(), []int64{1})
	if errCode(err) != int32(errno.ErrOssFileLocked) {
		t.Fatalf("expect locked error, got %v", err)
	}
	if len(repo.deleted) != 0 {
		t.Fatal("locked file must not be deleted")
	}
}

func TestDeleteUnlockedRemovesRecordAndObject(t *testing.T) {
	repo := newFakeRepo()
	repo.rows[1] = &model.SystemOss{ID: 1, FileName: "uploads/a.png", Service: "local", IsLock: false}
	store := &fakeStore{}
	ctl := &fakeCtl{store: store, meta: defaultMeta(), byKey: map[string]infraStorage.Storage{"local": store}}
	svc := NewService(repo, ctl, &fakeRules{})

	if err := svc.Delete(context.Background(), []int64{1}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(repo.deleted) != 1 || repo.deleted[0] != 1 {
		t.Fatalf("record not soft deleted: %v", repo.deleted)
	}
	if store.deletedKey != "uploads/a.png" {
		t.Fatalf("object not deleted: %q", store.deletedKey)
	}
}

func TestDownloadByServiceKey(t *testing.T) {
	repo := newFakeRepo()
	repo.rows[7] = &model.SystemOss{ID: 7, FileName: "k.bin", OriginalName: "原始 文件.bin", ContentType: "application/octet-stream", Service: "local"}
	store := &fakeStore{getData: []byte("payload")}
	ctl := &fakeCtl{store: store, meta: defaultMeta(), byKey: map[string]infraStorage.Storage{"local": store}}
	svc := NewService(repo, ctl, &fakeRules{})

	name, ct, data, err := svc.Download(context.Background(), 7)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	if name != "原始 文件.bin" || ct != "application/octet-stream" || string(data) != "payload" {
		t.Fatalf("unexpected: %q %q %q", name, ct, data)
	}
}

func TestDownloadServiceMissing(t *testing.T) {
	repo := newFakeRepo()
	repo.rows[7] = &model.SystemOss{ID: 7, FileName: "k", Service: "gone"}
	svc := NewService(repo, &fakeCtl{store: &fakeStore{}, meta: defaultMeta()}, &fakeRules{})

	_, _, _, err := svc.Download(context.Background(), 7)
	if errCode(err) != int32(errno.ErrOssFileServiceMissing) {
		t.Fatalf("expect service missing, got %v", err)
	}
}

func TestPageDecoratesRuleVariants(t *testing.T) {
	repo := newFakeRepo()
	repo.pageRows = []*model.SystemOss{{ID: 1, FileName: "a.png", URL: "http://cdn/x.png", ContentType: "image/png", Service: "local"}}
	repo.total = 1
	rules := &fakeRules{result: ossrule.ResolveResult{
		URL:      "http://cdn/x.png?ow=1",
		Variants: map[string]string{"80x80": "http://cdn/x.png?w=80"},
	}}
	svc := NewService(repo, &fakeCtl{store: &fakeStore{}, meta: defaultMeta()}, rules)

	vos, total, err := svc.Page(context.Background(), repository.OssFileFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(vos) != 1 {
		t.Fatalf("total=%d len=%d", total, len(vos))
	}
	if rules.calls != 1 {
		t.Fatalf("rules not consumed: %d", rules.calls)
	}
	if vos[0].URL != "http://cdn/x.png?ow=1" {
		t.Fatalf("overwrite not applied: %q", vos[0].URL)
	}
	if vos[0].RuleUrls["80x80"] == "" {
		t.Fatalf("variants missing: %v", vos[0].RuleUrls)
	}
}

func TestPagePrivateBucketPresigns(t *testing.T) {
	repo := newFakeRepo()
	repo.pageRows = []*model.SystemOss{{ID: 1, FileName: "a.png", URL: "http://old/a.png", Service: "local"}}
	repo.total = 1
	meta := defaultMeta()
	meta.AccessPolicy = "0" // private
	svc := NewService(repo, &fakeCtl{store: &fakeStore{}, meta: meta}, &fakeRules{})

	vos, _, err := svc.Page(context.Background(), repository.OssFileFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(vos[0].URL, "http://fake/") {
		t.Fatalf("private bucket should presign, got %q", vos[0].URL)
	}
}

func TestUpdateValidatesAndSaves(t *testing.T) {
	repo := newFakeRepo()
	repo.rows[3] = &model.SystemOss{ID: 3, OriginalName: "old"}
	svc := NewService(repo, &fakeCtl{store: &fakeStore{}, meta: defaultMeta()}, &fakeRules{})

	if err := svc.Update(context.Background(), &model.SystemOss{ID: 3, OriginalName: "新名", IsLock: true}); err != nil {
		t.Fatal(err)
	}
	if repo.saved == nil || repo.saved.OriginalName != "新名" || !repo.saved.IsLock {
		t.Fatalf("save not applied: %+v", repo.saved)
	}
	if err := svc.Update(context.Background(), &model.SystemOss{ID: 3, OriginalName: "  "}); errCode(err) != int32(errno.ErrOssFileInvalidParam) {
		t.Fatalf("empty original name should be rejected, got %v", err)
	}
	if err := svc.Update(context.Background(), &model.SystemOss{ID: 99, OriginalName: "x"}); errCode(err) != int32(errno.ErrOssFileNotFound) {
		t.Fatalf("missing id should be not found, got %v", err)
	}
}

func TestBuildUploadKey(t *testing.T) {
	key := buildUploadKey("", "a.BC12") // 非法扩展名字符被剔除
	if !strings.HasPrefix(key, "uploads/") || !strings.HasSuffix(key, ".bc12") {
		t.Fatalf("key=%q", key)
	}
	key = buildUploadKey("pre/", "x.txt")
	if !strings.HasPrefix(key, "pre/uploads/") {
		t.Fatalf("prefix key=%q", key)
	}
	// 无扩展名
	key = buildUploadKey("", "README")
	if strings.HasSuffix(key, ".") {
		t.Fatalf("no ext should not end with dot: %q", key)
	}
}
