package operlog

import (
	"context"
	"strings"
	"testing"

	"gosilite/application/base/logq"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"
)

// memSink 同步记录入队内容的测试替身。
type memSink struct {
	oper  []*logq.OperEntry
	errs  []*logq.ErrorEntry
	login []*logq.LoginEntry
}

func (m *memSink) EnqueueOper(e *logq.OperEntry)   { m.oper = append(m.oper, e) }
func (m *memSink) EnqueueError(e *logq.ErrorEntry) { m.errs = append(m.errs, e) }
func (m *memSink) EnqueueLogin(e *logq.LoginEntry) { m.login = append(m.login, e) }

// TestPanicCapturedIntoErrorLog 端到端验证 panic 链路：
// handler panic → recovery 回填 PanicKey → MW 收尾 → 异常日志含 ErrType=2 + 堆栈。
func TestPanicCapturedIntoErrorLog(t *testing.T) {
	prev := logq.Current()
	sink := &memSink{}
	logq.Set(sink)
	if prev != nil {
		t.Cleanup(func() { logq.Set(prev) })
	}

	// 与 main.registerMiddleware 相同的顺序：MW 在 recovery 之前（外层）
	h := server.New()
	h.Use(MW())
	h.Use(recovery.Recovery(
		recovery.WithRecoveryHandler(func(_ context.Context, c *app.RequestContext, err interface{}, stack []byte) {
			c.Set(PanicKey, &PanicInfo{Err: err, Stack: stack})
			c.SetStatusCode(500)
		}),
	))
	h.GET("/boom", func(_ context.Context, _ *app.RequestContext) {
		panic("boom-test")
	})

	w := ut.PerformRequest(h.Engine, "GET", "/boom", nil)
	if w.Code != 500 {
		t.Fatalf("http status want 500, got %d", w.Code)
	}

	if len(sink.errs) != 1 {
		t.Fatalf("error entries want 1, got %d", len(sink.errs))
	}
	e := sink.errs[0]
	if e.ErrType != 2 || e.ErrLevel != 2 {
		t.Fatalf("want panic type/level 2, got type=%d level=%d", e.ErrType, e.ErrLevel)
	}
	if !strings.Contains(e.ErrMsg, "boom-test") {
		t.Fatalf("errMsg should carry panic value, got %q", e.ErrMsg)
	}
	if len(e.Stack) == 0 || !strings.Contains(string(e.Stack), "panic(") {
		t.Fatalf("stack must contain panic frame dump, got %q", e.Stack)
	}
	if e.RequestURI != "/boom" {
		t.Fatalf("request uri: %q", e.RequestURI)
	}

	// 操作日志同时留痕且为失败态
	if len(sink.oper) != 1 || sink.oper[0].OperStatus != 1 {
		t.Fatalf("oper entries: %+v", sink.oper)
	}
}

// TestBusinessErrorEntryNotPanic 业务失败走 errType=1 且不带 panic 堆栈。
func TestBusinessErrorEntryNotPanic(t *testing.T) {
	_, toErr := classify(200, 701000028, true, false)
	if !toErr {
		t.Fatal("business failure must go to error log")
	}
	st, toErr := classify(200, 200, true, false)
	if st != 0 || toErr {
		t.Fatalf("success classify: st=%d toErr=%v", st, toErr)
	}
}
