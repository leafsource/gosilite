package syslog

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"gosilite/application/base/logq"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// 有界异步批量写：请求路径只做非阻塞入队；worker 按「批满 500 或 200ms」
// 先到者 flush，GORM CreateInBatches 多值插入，失败重试 1 次后整批丢弃。
// 队列满即丢弃 + 计数 + 限频告警，日志量洪峰绝不反压业务请求。
const (
	queueOperCap  = 8192
	queueErrCap   = 2048
	queueLoginCap = 1024
	flushInterval = 200 * time.Millisecond
	flushBatch    = 500
	flushTimeout  = 5 * time.Second
	warnInterval  = 60 * time.Second
)

type sink struct {
	repo repository.SysLogRepository

	operCh  chan *logq.OperEntry
	errCh   chan *logq.ErrorEntry
	loginCh chan *logq.LoginEntry

	quit     chan struct{}
	closeOne sync.Once
	wg       sync.WaitGroup

	dropped  atomic.Int64
	lastWarn atomic.Int64
}

func newSink(repo repository.SysLogRepository) *sink {
	s := &sink{
		repo:    repo,
		operCh:  make(chan *logq.OperEntry, queueOperCap),
		errCh:   make(chan *logq.ErrorEntry, queueErrCap),
		loginCh: make(chan *logq.LoginEntry, queueLoginCap),
		quit:    make(chan struct{}),
	}
	s.wg.Add(3)
	go runQueue(s, s.operCh, s.flushOper)
	go runQueue(s, s.errCh, s.flushError)
	go runQueue(s, s.loginCh, s.flushLogin)
	return s
}

// ===== 有界非阻塞入队（Sink 实现） =====

func (s *sink) EnqueueOper(e *logq.OperEntry) {
	select {
	case s.operCh <- e:
	default:
		s.noteDrop()
	}
}

func (s *sink) EnqueueError(e *logq.ErrorEntry) {
	select {
	case s.errCh <- e:
	default:
		s.noteDrop()
	}
}

func (s *sink) EnqueueLogin(e *logq.LoginEntry) {
	select {
	case s.loginCh <- e:
	default:
		s.noteDrop()
	}
}

// noteDrop 丢弃计数 + 每分钟至多一条告警（避免告警风暴自身造成开销）。
func (s *sink) noteDrop() {
	s.dropped.Add(1)
	now := time.Now().Unix()
	last := s.lastWarn.Load()
	if now-last >= int64(warnInterval/time.Second) && s.lastWarn.CompareAndSwap(last, now) {
		hlog.Errorf("[syslog] queue full, total dropped=%d", s.dropped.Load())
	}
}

// ===== worker =====

// runQueue 通用收集循环：批满 / 时间窗 / 退出排空 三种触发。
func runQueue[T any](s *sink, ch <-chan T, flush func(batch []T)) {
	defer s.wg.Done()
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()
	batch := make([]T, 0, flushBatch)
	doFlush := func() {
		if len(batch) > 0 {
			flush(batch)
			batch = batch[:0]
		}
	}
	for {
		select {
		case e := <-ch:
			batch = append(batch, e)
			if len(batch) >= flushBatch {
				flush(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			doFlush()
		case <-s.quit:
			// 退出：排空残留（容量有界，排空是微秒~毫秒级）
			for {
				select {
				case e := <-ch:
					batch = append(batch, e)
					if len(batch) >= flushBatch {
						flush(batch)
						batch = batch[:0]
					}
					continue
				default:
				}
				break
			}
			doFlush()
			return
		}
	}
}

// doInsert 带超时的一次写入 + 重试一次；仍失败整批丢弃（只告警，不反压）。
func doInsert(tag string, fn func(ctx context.Context) error) {
	ctx, cancel := context.WithTimeout(context.Background(), flushTimeout)
	err := fn(ctx)
	cancel()
	if err == nil {
		return
	}
	hlog.Errorf("[syslog] flush %s failed: %v, retrying once", tag, err)
	ctx2, cancel2 := context.WithTimeout(context.Background(), flushTimeout)
	err2 := fn(ctx2)
	cancel2()
	if err2 != nil {
		hlog.Errorf("[syslog] flush %s dropped after retry: %v", tag, err2)
	}
}

func (s *sink) flushOper(batch []*logq.OperEntry) {
	rows := make([]*model.SystemOperLog, 0, len(batch))
	for _, e := range batch {
		rows = append(rows, &model.SystemOperLog{
			TraceID: e.TraceID, UserID: e.UserID, UserName: e.UserName,
			OperType: int16(e.OperType), RequestURI: e.RequestURI,
			HTTPMethod: e.HTTPMethod, Module: e.Module, OperStatus: int16(e.OperStatus),
			RequestHeaders: e.RequestHeaders, RequestBody: e.RequestBody,
			ResponseBody: e.ResponseBody, ErrorMsg: e.ErrorMsg, Remark: e.Remark,
			CostMs: e.CostMs, ClientIP: e.ClientIP, CreatedAt: e.CreatedAt,
		})
	}
	doInsert("oper", func(ctx context.Context) error { return s.repo.BatchInsertOper(ctx, rows) })
}

func (s *sink) flushError(batch []*logq.ErrorEntry) {
	rows := make([]*model.SystemErrorLog, 0, len(batch))
	for _, e := range batch {
		rows = append(rows, &model.SystemErrorLog{
			TraceID: e.TraceID, UserID: e.UserID, UserName: e.UserName,
			ErrType: int16(e.ErrType), ErrLevel: int16(e.ErrLevel), HandleStatus: 0,
			ErrCode: e.ErrCode, ErrMsg: e.ErrMsg, Stack: e.Stack,
			RequestURI: e.RequestURI, HTTPMethod: e.HTTPMethod, Module: e.Module,
			RequestHeaders: e.RequestHeaders, RequestBody: e.RequestBody,
			CostMs: e.CostMs, ClientIP: e.ClientIP, CreatedAt: e.CreatedAt,
		})
	}
	doInsert("error", func(ctx context.Context) error { return s.repo.BatchInsertError(ctx, rows) })
}

func (s *sink) flushLogin(batch []*logq.LoginEntry) {
	rows := make([]*model.SystemLoginLog, 0, len(batch))
	for _, e := range batch {
		rows = append(rows, &model.SystemLoginLog{
			TraceID: e.TraceID, UserID: e.UserID, UserName: e.UserName,
			Status: int16(e.Status), FailReason: e.FailReason,
			ClientIP: e.ClientIP, UserAgent: e.UserAgent,
			Browser: e.Browser, Os: e.OS, CreatedAt: e.CreatedAt,
		})
	}
	doInsert("login", func(ctx context.Context) error { return s.repo.BatchInsertLogin(ctx, rows) })
}

// ===== Flush（进程退出排空，logq.Flusher 实现） =====

func (s *sink) Flush(ctx context.Context) error {
	s.closeOne.Do(func() { close(s.quit) })
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
