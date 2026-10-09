package user

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"gosilite/application/infra/cache"
)

var errEmpty = &emptyErr{}

type emptyErr struct{}

func (*emptyErr) Error() string { return "cache: empty" }

// memCache is a minimal in-memory CacheClient for SessionStore tests.
type memCache struct {
	mu     sync.Mutex
	data   map[string]string
	hashes map[string]map[string]string
}

func newMemCache() *memCache {
	cache.SetEmptyValueError(errEmpty)
	return &memCache{
		data:   make(map[string]string),
		hashes: make(map[string]map[string]string),
	}
}

type memStringCmd struct {
	val string
	err error
}

func (c *memStringCmd) Err() error               { return c.err }
func (c *memStringCmd) Result() (string, error)  { return c.val, c.err }
func (c *memStringCmd) Val() string              { return c.val }
func (c *memStringCmd) Int64() (int64, error)    { return strconv.ParseInt(c.val, 10, 64) }
func (c *memStringCmd) Bytes() ([]byte, error)   { return []byte(c.val), c.err }

type memIntCmd struct {
	val int64
	err error
}

func (c *memIntCmd) Err() error             { return c.err }
func (c *memIntCmd) Result() (int64, error) { return c.val, c.err }
func (c *memIntCmd) Val() int64             { return c.val }

type memBoolCmd struct{ val bool }

func (c *memBoolCmd) Err() error             { return nil }
func (c *memBoolCmd) Result() (bool, error)  { return c.val, nil }
func (c *memBoolCmd) Val() bool              { return c.val }

type memStatusCmd struct{ val string }

func (c *memStatusCmd) Err() error              { return nil }
func (c *memStatusCmd) Result() (string, error) { return c.val, nil }
func (c *memStatusCmd) Val() string             { return c.val }

type memMapCmd struct{ val map[string]string }

func (c *memMapCmd) Err() error                         { return nil }
func (c *memMapCmd) Result() (map[string]string, error) { return c.val, nil }
func (c *memMapCmd) Val() map[string]string             { return c.val }

type memSliceCmd struct{ val []string }

func (c *memSliceCmd) Err() error                { return nil }
func (c *memSliceCmd) Result() ([]string, error) { return c.val, nil }
func (c *memSliceCmd) Val() []string             { return c.val }

func (m *memCache) CreatePipeline() cache.Pipeline { return nil }

func (m *memCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) cache.StatusCmd {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = toString(value)
	return &memStatusCmd{val: "OK"}
}

func (m *memCache) Get(ctx context.Context, key string) cache.StringCmd {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.data[key]
	if !ok {
		return &memStringCmd{err: errEmpty}
	}
	return &memStringCmd{val: v}
}

func (m *memCache) IncrBy(ctx context.Context, key string, value int64) cache.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, _ := strconv.ParseInt(m.data[key], 10, 64)
	n += value
	m.data[key] = strconv.FormatInt(n, 10)
	return &memIntCmd{val: n}
}

func (m *memCache) Incr(ctx context.Context, key string) cache.IntCmd {
	return m.IncrBy(ctx, key, 1)
}

func (m *memCache) HSet(ctx context.Context, key string, values ...interface{}) cache.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.hashes[key]
	if !ok {
		h = make(map[string]string)
		m.hashes[key] = h
	}
	var n int64
	for i := 0; i+1 < len(values); i += 2 {
		f := toString(values[i])
		v := toString(values[i+1])
		if _, exists := h[f]; !exists {
			n++
		}
		h[f] = v
	}
	return &memIntCmd{val: n}
}

func (m *memCache) HGetAll(ctx context.Context, key string) cache.MapStringStringCmd {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]string, len(m.hashes[key]))
	for f, v := range m.hashes[key] {
		out[f] = v
	}
	return &memMapCmd{val: out}
}

func (m *memCache) HDel(ctx context.Context, key string, fields ...string) cache.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for _, f := range fields {
		if _, ok := m.hashes[key][f]; ok {
			delete(m.hashes[key], f)
			n++
		}
	}
	return &memIntCmd{val: n}
}

func (m *memCache) Delete(ctx context.Context, keys ...string) cache.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for _, k := range keys {
		if _, ok := m.data[k]; ok {
			delete(m.data, k)
			n++
		}
	}
	return &memIntCmd{val: n}
}

func (m *memCache) Exists(ctx context.Context, keys ...string) cache.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for _, k := range keys {
		if _, ok := m.data[k]; ok {
			n++
		}
	}
	return &memIntCmd{val: n}
}

func (m *memCache) Expire(ctx context.Context, key string, expiration time.Duration) cache.BoolCmd {
	return &memBoolCmd{val: true}
}

func (m *memCache) LIndex(ctx context.Context, key string, index int64) cache.StringCmd {
	return &memStringCmd{err: errEmpty}
}
func (m *memCache) LPush(ctx context.Context, key string, values ...interface{}) cache.IntCmd {
	return &memIntCmd{}
}
func (m *memCache) RPush(ctx context.Context, key string, values ...interface{}) cache.IntCmd {
	return &memIntCmd{}
}
func (m *memCache) LSet(ctx context.Context, key string, index int64, value interface{}) cache.StatusCmd {
	return &memStatusCmd{val: "OK"}
}
func (m *memCache) LPop(ctx context.Context, key string) cache.StringCmd {
	return &memStringCmd{err: errEmpty}
}
func (m *memCache) LRange(ctx context.Context, key string, start, stop int64) cache.StringSliceCmd {
	return &memSliceCmd{}
}

func (m *memCache) ScanAll(ctx context.Context, match string, limit int) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	prefix := match
	if len(prefix) > 0 && prefix[len(prefix)-1] == '*' {
		prefix = prefix[:len(prefix)-1]
	}
	var out []string
	for k := range m.data {
		if len(prefix) == 0 || len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			out = append(out, k)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	for k := range m.hashes {
		if len(prefix) == 0 || len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			out = append(out, k)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (m *memCache) TypeOf(ctx context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.hashes[key]; ok {
		return "hash", nil
	}
	if _, ok := m.data[key]; ok {
		return "string", nil
	}
	return "none", nil
}

func toString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	default:
		return ""
	}
}

func newSessionStore() *SessionStore {
	return NewSessionStore(newMemCache())
}

func TestIssueAndValidateRefreshToken(t *testing.T) {
	ctx := context.Background()
	s := newSessionStore()

	token, err := s.IssueRefreshToken(ctx, 42)
	if err != nil {
		t.Fatalf("IssueRefreshToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty refresh token")
	}

	userID, err := s.ValidateRefreshToken(ctx, token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken: %v", err)
	}
	if userID != 42 {
		t.Fatalf("userID = %d, want 42", userID)
	}
}

func TestValidateRefreshTokenRejectsEmptyAndUnknown(t *testing.T) {
	ctx := context.Background()
	s := newSessionStore()

	if _, err := s.ValidateRefreshToken(ctx, ""); err == nil {
		t.Fatal("expected error for empty refresh token")
	}
	if _, err := s.ValidateRefreshToken(ctx, "not-a-real-token"); err == nil {
		t.Fatal("expected error for unknown refresh token")
	}
}

func TestRefreshRotationInvalidatesOldToken(t *testing.T) {
	ctx := context.Background()
	s := newSessionStore()

	oldToken, err := s.IssueRefreshToken(ctx, 7)
	if err != nil {
		t.Fatalf("IssueRefreshToken: %v", err)
	}

	if err := s.RevokeUserRefreshTokens(ctx, 7); err != nil {
		t.Fatalf("RevokeUserRefreshTokens: %v", err)
	}
	newToken, err := s.IssueRefreshToken(ctx, 7)
	if err != nil {
		t.Fatalf("IssueRefreshToken after revoke: %v", err)
	}

	if _, err := s.ValidateRefreshToken(ctx, oldToken); err == nil {
		t.Fatal("old refresh token must be invalid after rotation")
	}
	if _, err := s.ValidateRefreshToken(ctx, newToken); err != nil {
		t.Fatalf("new refresh token should be valid: %v", err)
	}
}

func TestRevokeRefreshToken(t *testing.T) {
	ctx := context.Background()
	s := newSessionStore()

	token, err := s.IssueRefreshToken(ctx, 9)
	if err != nil {
		t.Fatalf("IssueRefreshToken: %v", err)
	}
	if err := s.RevokeRefreshToken(ctx, token); err != nil {
		t.Fatalf("RevokeRefreshToken: %v", err)
	}
	if _, err := s.ValidateRefreshToken(ctx, token); err == nil {
		t.Fatal("revoked token must not validate")
	}
}

func TestDenyAccessToken(t *testing.T) {
	ctx := context.Background()
	s := newSessionStore()

	token := "access-token-abc"
	denied, err := s.IsAccessTokenDenied(ctx, token)
	if err != nil {
		t.Fatalf("IsAccessTokenDenied: %v", err)
	}
	if denied {
		t.Fatal("fresh token must not be denied")
	}

	if err := s.DenyAccessToken(ctx, token, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("DenyAccessToken: %v", err)
	}
	denied, err = s.IsAccessTokenDenied(ctx, token)
	if err != nil {
		t.Fatalf("IsAccessTokenDenied after deny: %v", err)
	}
	if !denied {
		t.Fatal("denied token must report denied")
	}

	if denied, err := s.IsAccessTokenDenied(ctx, ""); err != nil || denied {
		t.Fatalf("empty token: denied=%v err=%v", denied, err)
	}
}
