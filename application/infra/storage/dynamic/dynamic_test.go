package dynamic

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"gosilite/application/infra/storage"
)

// fakeStore 最小 storage.Storage 桩：只实现接口方法签名，不做真实 IO。
type fakeStore struct{ tag string }

func (f *fakeStore) PutObject(ctx context.Context, objectKey string, content []byte, opts ...storage.PutOptFn) error {
	return nil
}
func (f *fakeStore) PutObjectWithReader(ctx context.Context, objectKey string, content io.Reader, opts ...storage.PutOptFn) error {
	return nil
}
func (f *fakeStore) GetObject(ctx context.Context, objectKey string) ([]byte, error) { return nil, nil }
func (f *fakeStore) DeleteObject(ctx context.Context, objectKey string) error        { return nil }
func (f *fakeStore) GetObjectUrl(ctx context.Context, objectKey string, opts ...storage.GetOptFn) (string, error) {
	return f.tag, nil
}
func (f *fakeStore) HeadObject(ctx context.Context, objectKey string, opts ...storage.GetOptFn) (*storage.FileInfo, error) {
	return nil, nil
}
func (f *fakeStore) ListAllObjects(ctx context.Context, prefix string, opts ...storage.GetOptFn) ([]*storage.FileInfo, error) {
	return nil, nil
}
func (f *fakeStore) ListObjectsPaginated(ctx context.Context, input *storage.ListObjectsPaginatedInput, opts ...storage.GetOptFn) (*storage.ListObjectsPaginatedOutput, error) {
	return nil, nil
}

var _ storage.Storage = (*fakeStore)(nil)


// chdirRoot 让 conf.GetConf() 能读到仓库根 conf/conf.yaml（fallbackMeta 需要）。
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

func TestNoDefaultConfigFallsBack(t *testing.T) {
	chdirRoot(t)
	fallback := &fakeStore{tag: "conf"}
	c := NewWithBuilder(
		func(ctx context.Context) (*Config, error) { return nil, nil },
		fallback,
		func(ctx context.Context, cfg *Config) (storage.Storage, error) {
			t.Fatal("builder should not be called when resolver returns nil")
			return nil, nil
		},
	)
	s, meta, err := c.Resolve(context.Background())
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if s != fallback {
		t.Fatalf("expect fallback store")
	}
	if meta == nil || meta.ConfigKey != "conf" {
		t.Fatalf("expect conf meta, got %+v", meta)
	}
}

func TestResolveErrorFallsBack(t *testing.T) {
	fallback := &fakeStore{tag: "conf"}
	c := NewWithBuilder(
		func(ctx context.Context) (*Config, error) { return nil, errors.New("db down") },
		fallback,
		func(ctx context.Context, cfg *Config) (storage.Storage, error) { return nil, nil },
	)
	s, _, err := c.Resolve(context.Background())
	if err != nil {
		t.Fatalf("resolve should swallow resolver error: %v", err)
	}
	if s != fallback {
		t.Fatalf("expect fallback store")
	}
}

func TestBuildFailureFallsBack(t *testing.T) {
	fallback := &fakeStore{tag: "conf"}
	c := NewWithBuilder(
		func(ctx context.Context) (*Config, error) { return &Config{ConfigKey: "bad"}, nil },
		fallback,
		func(ctx context.Context, cfg *Config) (storage.Storage, error) { return nil, errors.New("build failed") },
	)
	s, _, err := c.Resolve(context.Background())
	if err != nil || s != fallback {
		t.Fatalf("expect fallback, got %v err=%v", s, err)
	}
}

func TestInvalidateRebuildsWithNewConfig(t *testing.T) {
	fallback := &fakeStore{tag: "conf"}
	key := "minio"
	c := NewWithBuilder(
		func(ctx context.Context) (*Config, error) {
			return &Config{ConfigKey: key, Domain: "example.com", IsHTTPS: "Y", AccessPolicy: "0"}, nil
		},
		fallback,
		func(ctx context.Context, cfg *Config) (storage.Storage, error) {
			return &fakeStore{tag: cfg.ConfigKey}, nil
		},
	)

	s, meta, err := c.Resolve(context.Background())
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if s.(*fakeStore).tag != "minio" || meta.ConfigKey != "minio" || !meta.IsHTTPS || meta.AccessPolicy != "0" {
		t.Fatalf("unexpected first build: %q %+v", s.(*fakeStore).tag, meta)
	}

	// 模拟设默认切换配置后 Invalidate
	key = "aliyun-s3"
	c.Invalidate()
	s2, meta2, err := c.Resolve(context.Background())
	if err != nil {
		t.Fatalf("resolve2: %v", err)
	}
	if s2.(*fakeStore).tag != "aliyun-s3" || meta2.ConfigKey != "aliyun-s3" {
		t.Fatalf("expect rebuilt client, got %q %+v", s2.(*fakeStore).tag, meta2)
	}
}

func TestResolveCachedWithoutRebuild(t *testing.T) {
	fallback := &fakeStore{tag: "conf"}
	builds := 0
	c := NewWithBuilder(
		func(ctx context.Context) (*Config, error) { return &Config{ConfigKey: "minio"}, nil },
		fallback,
		func(ctx context.Context, cfg *Config) (storage.Storage, error) {
			builds++
			return &fakeStore{tag: cfg.ConfigKey}, nil
		},
	)
	_, _, _ = c.Resolve(context.Background())
	_, _, _ = c.Resolve(context.Background())
	if builds != 1 {
		t.Fatalf("expect 1 build, got %d", builds)
	}
}
