package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	infraStorage "gosilite/application/infra/storage"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

// fakeStore 内存实现 storage.Storage，仅覆盖 service 用到的路径。
type fakeStore struct {
	objects map[string]*infraStorage.FileInfo
	order   []string

	listErr   error
	putErr    error
	urlErr    error
	deleteErr error

	listCalls  int
	putKey     string
	deletedKey string
}

func newFakeStore(keys ...string) *fakeStore {
	f := &fakeStore{objects: map[string]*infraStorage.FileInfo{}}
	for i, k := range keys {
		f.objects[k] = &infraStorage.FileInfo{Key: k, Size: int64(100 + i)}
		f.order = append(f.order, k)
	}
	return f
}

func (f *fakeStore) PutObject(_ context.Context, objectKey string, _ []byte, _ ...infraStorage.PutOptFn) error {
	if f.putErr != nil {
		return f.putErr
	}
	f.putKey = objectKey
	f.objects[objectKey] = &infraStorage.FileInfo{Key: objectKey, Size: 1}
	return nil
}

func (f *fakeStore) PutObjectWithReader(context.Context, string, io.Reader, ...infraStorage.PutOptFn) error {
	return nil
}

func (f *fakeStore) GetObject(context.Context, string) ([]byte, error) { return nil, nil }

func (f *fakeStore) DeleteObject(_ context.Context, objectKey string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deletedKey = objectKey
	delete(f.objects, objectKey)
	return nil
}

func (f *fakeStore) GetObjectUrl(_ context.Context, objectKey string, _ ...infraStorage.GetOptFn) (string, error) {
	if f.urlErr != nil {
		return "", f.urlErr
	}
	return "http://fake.local/" + objectKey, nil
}

func (f *fakeStore) HeadObject(context.Context, string, ...infraStorage.GetOptFn) (*infraStorage.FileInfo, error) {
	return nil, nil
}

func (f *fakeStore) ListAllObjects(context.Context, string, ...infraStorage.GetOptFn) ([]*infraStorage.FileInfo, error) {
	return nil, nil
}

// ListObjectsPaginated 按 order 顺序返回，支持 Cursor 分页与 Prefix 过滤。
func (f *fakeStore) ListObjectsPaginated(_ context.Context, input *infraStorage.ListObjectsPaginatedInput, _ ...infraStorage.GetOptFn) (*infraStorage.ListObjectsPaginatedOutput, error) {
	f.listCalls++
	if f.listErr != nil {
		return nil, f.listErr
	}
	var matched []*infraStorage.FileInfo
	for _, k := range f.order {
		if obj, ok := f.objects[k]; ok && strings.HasPrefix(k, input.Prefix) {
			matched = append(matched, obj)
		}
	}
	start := 0
	if input.Cursor != "" {
		for i, obj := range matched {
			if obj.Key == input.Cursor {
				start = i + 1
				break
			}
		}
	}
	end := start + input.PageSize
	if end > len(matched) {
		end = len(matched)
	}
	page := matched[start:end]
	out := &infraStorage.ListObjectsPaginatedOutput{Files: page}
	if end < len(matched) && len(page) > 0 {
		out.IsTruncated = true
		out.Cursor = page[len(page)-1].Key
	}
	return out, nil
}

// errCode 提取 errorx 业务错误码（照 menu/service_test 惯例）。
func errCode(err error) int32 {
	var se errorx.StatusError
	if errors.As(err, &se) {
		return se.Code()
	}
	return 0
}

// chdirRoot 让 conf.GetConf() 能读到仓库根 conf/conf.yaml（Info 需要配置）。
func chdirRoot(t *testing.T) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir("../../../.."); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

func TestInfoAggregatesConnected(t *testing.T) {
	chdirRoot(t)
	svc := NewService(newFakeStore("user_avatar/1.png", "uploads/a.png"))
	info, err := svc.Info(context.Background())
	if err != nil {
		t.Fatalf("Info err: %v", err)
	}
	if !info.Connected {
		t.Fatalf("connected = false, message = %q", info.Message)
	}
	if info.ObjectCount != 2 {
		t.Fatalf("objectCount = %d, want 2", info.ObjectCount)
	}
	if info.TotalSize != 100+101 {
		t.Fatalf("totalSize = %d, want 201", info.TotalSize)
	}
}

func TestInfoDisconnectedReturnsBusinessEnvelope(t *testing.T) {
	chdirRoot(t)
	store := newFakeStore()
	store.listErr = errors.New("connection refused")
	info, err := NewService(store).Info(context.Background())
	if err != nil {
		t.Fatalf("断连时应返回 nil error（HTTP 200 业务信封），got %v", err)
	}
	if info.Connected {
		t.Fatal("connected 应为 false")
	}
	if !strings.Contains(info.Message, "connection refused") {
		t.Fatalf("message = %q, want 含 connection refused", info.Message)
	}
}

func TestListFiltersAndPaginates(t *testing.T) {
	svc := NewService(newFakeStore("user_avatar/1.png", "user_avatar/2.png", "uploads/3.png"))

	list, total, err := svc.List(context.Background(), 1, 1, "user_avatar/", "")
	if err != nil {
		t.Fatalf("List err: %v", err)
	}
	if total != 2 || len(list) != 1 || list[0].Key != "user_avatar/1.png" {
		t.Fatalf("page1 = %v total=%d, want first of 2", list, total)
	}

	list, total, err = svc.List(context.Background(), 2, 1, "user_avatar/", "")
	if err != nil {
		t.Fatalf("List page2 err: %v", err)
	}
	if total != 2 || len(list) != 1 || list[0].Key != "user_avatar/2.png" {
		t.Fatalf("page2 = %v total=%d", list, total)
	}

	// 越界页返回空列表而非报错
	list, total, err = svc.List(context.Background(), 9, 10, "", "")
	if err != nil {
		t.Fatalf("越界页 err: %v", err)
	}
	if total != 3 || len(list) != 0 {
		t.Fatalf("越界页 list=%v total=%d, want empty/3", list, total)
	}

	// keyword 包含过滤
	list, total, err = svc.List(context.Background(), 1, 10, "", "uploads")
	if err != nil {
		t.Fatalf("keyword err: %v", err)
	}
	if total != 1 || list[0].Key != "uploads/3.png" {
		t.Fatalf("keyword list=%v total=%d", list, total)
	}
}

func TestListNormalizesPaging(t *testing.T) {
	svc := NewService(newFakeStore("a", "b", "c"))
	// page<=0 归一为 1；pageSize 超上限归一（fake 每次至多返回 PageSize 条）
	list, total, err := svc.List(context.Background(), 0, 9999, "", "")
	if err != nil {
		t.Fatalf("List err: %v", err)
	}
	if total != 3 || len(list) == 0 {
		t.Fatalf("归一化后 list len=%d total=%d", len(list), total)
	}
}

func TestListStoreError(t *testing.T) {
	store := newFakeStore()
	store.listErr = errors.New("boom")
	_, _, err := NewService(store).List(context.Background(), 1, 10, "", "")
	if err == nil {
		t.Fatal("期望返回 error")
	}
	if errCode(err) != errno.ErrStorageListFailed {
		t.Fatalf("err = %v, want errno %d", err, errno.ErrStorageListFailed)
	}
}

func TestDeleteValidationAndError(t *testing.T) {
	if err := NewService(newFakeStore()).Delete(context.Background(), "  "); err == nil {
		t.Fatal("空 key 应报错")
	}

	store := newFakeStore("x.png")
	if err := NewService(store).Delete(context.Background(), "x.png"); err != nil {
		t.Fatalf("正常删除 err: %v", err)
	}
	if store.deletedKey != "x.png" {
		t.Fatalf("deletedKey = %q", store.deletedKey)
	}

	store = newFakeStore("x.png")
	store.deleteErr = errors.New("denied")
	err := NewService(store).Delete(context.Background(), "x.png")
	if errCode(err) != errno.ErrStorageDeleteFailed {
		t.Fatalf("err = %v, want errno %d", err, errno.ErrStorageDeleteFailed)
	}
}

func TestUpload(t *testing.T) {
	store := newFakeStore()
	vo, err := NewService(store).Upload(context.Background(), "Avatar.PNG", "image/png", []byte("data"))
	if err != nil {
		t.Fatalf("Upload err: %v", err)
	}
	if !strings.HasPrefix(vo.Key, "uploads/") || !strings.HasSuffix(vo.Key, ".png") {
		t.Fatalf("key = %q, want uploads/<date>/<uuid>.png", vo.Key)
	}
	if vo.WebURI == "" || store.putKey != vo.Key {
		t.Fatalf("webURI=%q putKey=%q", vo.WebURI, store.putKey)
	}

	if _, err := NewService(store).Upload(context.Background(), "a.png", "image/png", nil); errCode(err) != errno.ErrUploadInvalidParamCode {
		t.Fatalf("空文件 err = %v", err)
	}
	big := make([]byte, MaxUploadBytes+1)
	if _, err := NewService(store).Upload(context.Background(), "a.png", "image/png", big); errCode(err) != errno.ErrUploadInvalidFileSizeCode {
		t.Fatalf("超限 err = %v", err)
	}
}

func TestBuildUploadKey(t *testing.T) {
	before := time.Now().Format("20060102")
	key := buildUploadKey("photo.PNG")
	if !strings.HasPrefix(key, "uploads/"+before+"/") {
		t.Fatalf("key = %q", key)
	}
	if !strings.HasSuffix(key, ".png") {
		t.Fatalf("ext 应小写保留: %q", key)
	}
	// 扩展名只保留字母数字
	if k := buildUploadKey("a.b.tar.gz"); !strings.HasSuffix(k, ".gztar") && !strings.HasSuffix(k, ".gz") {
		// path.Ext 取 .gz；确保无点号残留即可
		if strings.Contains(strings.TrimPrefix(k, "uploads/"+before+"/"), "..") {
			t.Fatalf("key 含非法扩展: %q", k)
		}
	}
}
