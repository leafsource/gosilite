package syslog

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"gosilite/application/base/logq"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

func errCode(err error) int32 {
	var se errorx.StatusError
	if errors.As(err, &se) {
		return se.Code()
	}
	return 0
}

// fakeRepo 记录调用的仓储替身。
type fakeRepo struct {
	mu sync.Mutex

	operPageCalled   bool
	operFilter       repository.OperPageFilter
	operPage, ps     int
	operRows         []*model.SystemOperLog
	operTotal        int64

	operByIDErr      error
	deletedOperIDs   []int64
	deletedOperAll   bool
	cleanOperN       int64

	// sink 写入
	gotOper   []*model.SystemOperLog
	gotError  []*model.SystemErrorLog
	gotLogin  []*model.SystemLoginLog
	insertErr error
	blockCh   chan struct{} // 非 nil 时 BatchInsertOper 阻塞（模拟慢库）
}

func (f *fakeRepo) OperPage(_ context.Context, page, pageSize int, fo repository.OperPageFilter) ([]*model.SystemOperLog, int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.operPageCalled, f.operPage, f.ps, f.operFilter = true, page, pageSize, fo
	return f.operRows, f.operTotal, nil
}

func (f *fakeRepo) OperByID(_ context.Context, _ int64) (*model.SystemOperLog, error) {
	if f.operByIDErr != nil {
		return nil, f.operByIDErr
	}
	return &model.SystemOperLog{ID: 1, TraceID: "t"}, nil
}

func (f *fakeRepo) CleanOper(_ context.Context, ids []int64, all bool) (int64, error) {
	f.deletedOperIDs, f.deletedOperAll = ids, all
	return f.cleanOperN, nil
}

func (f *fakeRepo) ErrorPage(context.Context, int, int, repository.ErrorPageFilter) ([]*model.SystemErrorLog, int64, error) {
	return nil, 0, nil
}

func (f *fakeRepo) ErrorByID(context.Context, int64) (*model.SystemErrorLog, error) {
	return nil, errors.New("not found")
}

func (f *fakeRepo) MarkErrorHandled(_ context.Context, id int64, _ string) (int64, error) {
	if id == 404 {
		return 0, nil
	}
	return 1, nil
}

func (f *fakeRepo) CleanError(context.Context, []int64, bool) (int64, error) { return 0, nil }

func (f *fakeRepo) LoginPage(context.Context, int, int, repository.LoginPageFilter) ([]*model.SystemLoginLog, int64, error) {
	return nil, 0, nil
}

func (f *fakeRepo) BatchInsertOper(_ context.Context, rows []*model.SystemOperLog) error {
	if f.blockCh != nil {
		<-f.blockCh
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.insertErr != nil {
		return f.insertErr
	}
	f.gotOper = append(f.gotOper, rows...)
	return nil
}

func (f *fakeRepo) BatchInsertError(_ context.Context, rows []*model.SystemErrorLog) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gotError = append(f.gotError, rows...)
	return nil
}

func (f *fakeRepo) BatchInsertLogin(_ context.Context, rows []*model.SystemLoginLog) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gotLogin = append(f.gotLogin, rows...)
	return nil
}

func strPtr(s string) *string { return &s }

func TestOperPageNormalizesPaging(t *testing.T) {
	repo := &fakeRepo{}
	svc := newService(repo)
	_, _, err := svc.OperPage(context.Background(), 0, 9999, "  tid ", "alice", "/api", "GET",
		nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if repo.operPage != 1 {
		t.Fatalf("page want 1, got %d", repo.operPage)
	}
	if repo.ps != 500 {
		t.Fatalf("pageSize want clamp 500, got %d", repo.ps)
	}
	if repo.operFilter.TraceID != "tid" || repo.operFilter.UserName != "alice" {
		t.Fatalf("filter not trimmed: %+v", repo.operFilter)
	}
}

func TestPageRangeDateOnly(t *testing.T) {
	st, et, err := pageRange(strPtr("2026-02-01"), strPtr("2026-02-03"))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if st == nil || st.Format(timeLayout) != "2026-02-01 00:00:00" {
		t.Fatalf("start wrong: %v", st)
	}
	if et == nil || et.Format(timeLayout) != "2026-02-03 23:59:59" {
		t.Fatalf("end wrong: %v", et)
	}
}

func TestPageRangeInvalid(t *testing.T) {
	_, _, err := pageRange(strPtr("not-a-time"), nil)
	if errCode(err) != errno.ErrSystemInvalidParam {
		t.Fatalf("want ErrSystemInvalidParam, got %v", err)
	}
}

func TestOperDetailNotFound(t *testing.T) {
	repo := &fakeRepo{operByIDErr: errors.New("record not found")}
	svc := newService(repo)
	_, err := svc.OperDetail(context.Background(), 9)
	if errCode(err) != errno.ErrOperLogNotFound {
		t.Fatalf("want ErrOperLogNotFound, got %v", err)
	}
}

func TestHandleErrorNotFound(t *testing.T) {
	repo := &fakeRepo{}
	svc := newService(repo)
	if _, err := svc.HandleError(context.Background(), 404, "x"); errCode(err) != errno.ErrErrorLogNotFound {
		t.Fatalf("want ErrErrorLogNotFound, got %v", err)
	}
	if _, err := svc.HandleError(context.Background(), 1, "已修复"); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestCleanOperPassThrough(t *testing.T) {
	repo := &fakeRepo{cleanOperN: 3}
	svc := newService(repo)
	n, err := svc.CleanOper(context.Background(), []int64{1, 2, 3}, false)
	if err != nil || n != 3 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if repo.deletedOperAll || len(repo.deletedOperIDs) != 3 {
		t.Fatalf("ids/all not passed: all=%v ids=%v", repo.deletedOperAll, repo.deletedOperIDs)
	}
	if _, err := svc.CleanOper(context.Background(), nil, true); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !repo.deletedOperAll {
		t.Fatalf("all not passed")
	}
}

// ===== sink：异步批量写 + 有界丢弃 =====

func TestSinkFlushOnInterval(t *testing.T) {
	repo := &fakeRepo{}
	s := newSink(repo)
	defer s.Flush(context.Background())

	now := time.Now()
	s.EnqueueOper(&logq.OperEntry{TraceID: "a", OperType: 4, CreatedAt: now})
	s.EnqueueError(&logq.ErrorEntry{TraceID: "b", ErrType: 1, CreatedAt: now})
	s.EnqueueLogin(&logq.LoginEntry{TraceID: "c", Status: 0, CreatedAt: now})

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		repo.mu.Lock()
		done := len(repo.gotOper) == 1 && len(repo.gotError) == 1 && len(repo.gotLogin) == 1
		repo.mu.Unlock()
		if done {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if len(repo.gotOper) != 1 || repo.gotOper[0].TraceID != "a" {
		t.Fatalf("oper rows: %+v", repo.gotOper)
	}
	if len(repo.gotError) != 1 || repo.gotError[0].TraceID != "b" {
		t.Fatalf("error rows: %+v", repo.gotError)
	}
	if len(repo.gotLogin) != 1 || repo.gotLogin[0].TraceID != "c" {
		t.Fatalf("login rows: %+v", repo.gotLogin)
	}
	// 字段转换保真
	if repo.gotOper[0].OperType != 4 {
		t.Fatalf("oper_type conversion: %+v", repo.gotOper[0])
	}
}

func TestSinkQueueBoundDropsInsteadOfBlocking(t *testing.T) {
	block := make(chan struct{})
	repo := &fakeRepo{insertErr: errors.New("db down"), blockCh: block}
	s := newSink(repo)
	defer s.Flush(context.Background())
	// 后注册的 defer 先执行：先解除阻塞再排空
	defer close(block)

	// 洪峰前先让 worker 卡在 flush（慢库），队列灌满后必须丢弃而非阻塞
	for i := 0; i < queueOperCap+flushBatch+100; i++ {
		s.EnqueueOper(&logq.OperEntry{TraceID: strconv.Itoa(i)})
	}
	if s.dropped.Load() == 0 {
		t.Fatal("expected drops when queue overflows")
	}
	// 满仓后再投递必须立即返回
	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			s.EnqueueOper(&logq.OperEntry{TraceID: strconv.Itoa(i)})
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("enqueue blocked: queue must be non-blocking")
	}
}
