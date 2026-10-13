package user

import (
	"context"
	"testing"
	"time"

	"gosilite/types/consts"
)

func newOnlineStore(t *testing.T) (*SessionStore, *memCache) {
	t.Helper()
	mc := newMemCache()
	return NewSessionStore(mc), mc
}

func TestPutAndListOnline(t *testing.T) {
	ctx := context.Background()
	st, _ := newOnlineStore(t)
	expire := time.Now().Add(time.Hour)

	info := OnlineSession{UserID: 7, Account: "skyx", DeptID: 3, IP: "127.0.0.1", LoginAt: time.Now().Unix()}
	if err := st.PutOnline(ctx, "access-raw-1", "refresh-raw-1", info, expire); err != nil {
		t.Fatalf("put: %v", err)
	}

	entries, err := st.ListOnline(ctx)
	if err != nil || len(entries) != 1 {
		t.Fatalf("list: n=%d err=%v", len(entries), err)
	}
	e := entries[0]
	if e.AccessHash != hashToken("access-raw-1") {
		t.Fatalf("field should be access hash, got %s", e.AccessHash)
	}
	if e.Session.Account != "skyx" || e.Session.RefreshHash != hashToken("refresh-raw-1") {
		t.Fatalf("session snapshot wrong: %+v", e.Session)
	}
	if e.Session.Expire != expire.Unix() {
		t.Fatalf("expire not materialized: %d", e.Session.Expire)
	}
}

func TestListOnlineLazyExpires(t *testing.T) {
	ctx := context.Background()
	st, mc := newOnlineStore(t)

	// 过期会话
	_ = st.PutOnline(ctx, "old", "r-old", OnlineSession{UserID: 1, Expire: 0}, time.Now().Add(-time.Hour))
	// 活跃会话
	_ = st.PutOnline(ctx, "new", "r-new", OnlineSession{UserID: 2}, time.Now().Add(time.Hour))

	entries, err := st.ListOnline(ctx)
	if err != nil || len(entries) != 1 {
		t.Fatalf("want only active session, n=%d err=%v", len(entries), err)
	}
	if entries[0].AccessHash != hashToken("new") {
		t.Fatalf("wrong survivor: %+v", entries[0])
	}
	// 过期 field 已被惰性 HDel
	all, _ := mc.HGetAll(ctx, consts.AuthOnlineSessionsKey).Result()
	if _, ok := all[hashToken("old")]; ok {
		t.Fatal("expired field should be removed lazily")
	}
}

func TestMoveOnline(t *testing.T) {
	ctx := context.Background()
	st, _ := newOnlineStore(t)

	origLogin := int64(1700000000)
	_ = st.PutOnline(ctx, "acc-v1", "ref-v1", OnlineSession{UserID: 3, Account: "u", LoginAt: origLogin}, time.Now().Add(time.Hour))
	err := st.MoveOnline(ctx, "acc-v1", "acc-v2", "ref-v2", OnlineSession{UserID: 3, Account: "u", LoginAt: time.Now().Unix()}, time.Now().Add(2*time.Hour))
	if err != nil {
		t.Fatalf("move: %v", err)
	}
	entries, _ := st.ListOnline(ctx)
	if len(entries) != 1 || entries[0].AccessHash != hashToken("acc-v2") {
		t.Fatalf("want only new session: %+v", entries)
	}
	if entries[0].Session.RefreshHash != hashToken("ref-v2") {
		t.Fatalf("refresh hash not updated: %+v", entries[0].Session)
	}
	// refresh 不重置登录时间
	if entries[0].Session.LoginAt != origLogin {
		t.Fatalf("loginAt should be preserved, got %d want %d", entries[0].Session.LoginAt, origLogin)
	}
}

func TestForceOnline(t *testing.T) {
	ctx := context.Background()
	st, mc := newOnlineStore(t)

	// 不存在
	if err := st.ForceOnline(ctx, "no-such-hash"); err != ErrOnlineNotFound {
		t.Fatalf("want ErrOnlineNotFound, got %v", err)
	}

	expire := time.Now().Add(30 * time.Minute)
	_ = st.PutOnline(ctx, "acc-kick", "ref-kick", OnlineSession{UserID: 9, Account: "victim"}, expire)
	hash := hashToken("acc-kick")

	if err := st.ForceOnline(ctx, hash); err != nil {
		t.Fatalf("force: %v", err)
	}
	// 登记已摘
	if entries, _ := st.ListOnline(ctx); len(entries) != 0 {
		t.Fatalf("session should be gone: %+v", entries)
	}
	// access 已进 denylist
	if ok, _ := st.IsAccessTokenDenied(ctx, "acc-kick"); !ok {
		t.Fatal("access should be denied after force")
	}
	// 该会话 refresh 已吊销（key 消失）
	if _, err := mc.Get(ctx, consts.AuthRefreshTokenPrefix+hashToken("ref-kick")).Result(); err == nil {
		t.Fatal("session refresh should be revoked")
	}
	// 其他会话不受影响
	_ = st.PutOnline(ctx, "acc-other", "ref-other", OnlineSession{UserID: 9, Account: "victim"}, expire)
	if err := st.ForceOnline(ctx, hash); err != ErrOnlineNotFound {
		t.Fatalf("kicked session should not exist: %v", err)
	}
	if entries, _ := st.ListOnline(ctx); len(entries) != 1 {
		t.Fatalf("other session must survive: %+v", entries)
	}
}
