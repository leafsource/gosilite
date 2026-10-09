// Package logq 是日志采集的中枢：中间件（生产者）与异步批量写（消费者）
// 之间的解耦点。entries 为纯数据结构，两个方向都不引入对方依赖。
package logq

import (
	"context"
	"sync/atomic"
	"time"
)

// OperEntry 操作日志条目（每请求一条）。
type OperEntry struct {
	TraceID        string
	UserID         int64
	UserName       string
	OperType       int8 // sys_oper_type
	RequestURI     string
	HTTPMethod     string
	Module         string
	OperStatus     int8 // 0 成功 1 失败
	RequestHeaders string
	RequestBody    string
	ResponseBody   string
	ErrorMsg       string
	Remark         string
	CostMs         int32
	ClientIP       string
	CreatedAt      time.Time
}

// ErrorEntry 系统异常日志条目。
type ErrorEntry struct {
	TraceID        string
	UserID         int64
	UserName       string
	ErrType        int8 // 1 业务异常 2 panic
	ErrLevel       int8 // sys_error_level: 1 普通 2 严重
	HandleStatus   int8 // 恒 0 待处理，由 worker 落库
	ErrCode        int32
	ErrMsg         string
	Stack          string
	RequestURI     string
	HTTPMethod     string
	Module         string
	RequestHeaders string
	RequestBody    string
	CostMs         int32
	ClientIP       string
	CreatedAt      time.Time
}

// LoginEntry 登录日志条目。
type LoginEntry struct {
	TraceID    string
	UserID     int64
	UserName   string
	Status     int8 // sys_login_status: 0 成功 1 失败
	FailReason string
	ClientIP   string
	UserAgent  string
	Browser    string
	OS         string
	CreatedAt  time.Time
}

// Sink 是有界非阻塞写入接口：入队满时丢弃并自计数，绝不阻塞业务请求。
type Sink interface {
	EnqueueOper(e *OperEntry)
	EnqueueError(e *ErrorEntry)
	EnqueueLogin(e *LoginEntry)
}

// Flusher 进程退出前排空队列（可选能力，由 sink 实现）。
type Flusher interface {
	Flush(ctx context.Context) error
}

var (
	sinkVal    atomic.Value // Sink
	flusherVal atomic.Value // Flusher
)

// Set 安装全局 sink（bootstrap 时调用一次）。
func Set(s Sink) { sinkVal.Store(s) }

// Current 返回当前 sink，可能为 nil（未初始化 / 单测环境）。
func Current() Sink {
	s, _ := sinkVal.Load().(Sink)
	return s
}

// SetFlusher 安装退出排空钩子。
func SetFlusher(f Flusher) { flusherVal.Store(f) }

// Flush 尝试排空全部队列；无 hook 时为空操作。
func Flush(ctx context.Context) error {
	f, _ := flusherVal.Load().(Flusher)
	if f == nil {
		return nil
	}
	return f.Flush(ctx)
}
