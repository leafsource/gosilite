package scheduler

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/sse"
	dalModel "gosilite/application/framework/dal/model"
	core "gosilite/application/framework/scheduler"
)

func TestStreamRunLogsReplaysAndFollowsLiveLogs(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	h := server.New(server.WithListener(ln))
	live := make(chan *core.LogEntry, 1)
	h.GET("/stream", func(ctx context.Context, c *app.RequestContext) {
		writer := sse.NewWriter(c)
		_ = streamRunLogs(ctx, writer, []*dalModel.SystemJobRunLog{{
			ID: 1, RunID: 7, JobKey: "system.scheduler.test", Level: "info", Stage: "start", Message: "replayed", CreatedAt: time.Unix(0, 0).UTC(),
		}}, live)
	})
	go h.Spin()
	t.Cleanup(func() { _ = h.Close() })

	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+ln.Addr().String()+"/stream", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := (&http.Client{}).Do(request)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = response.Body.Close() })
	if contentType := response.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "text/event-stream") {
		t.Fatalf("Content-Type = %q", contentType)
	}

	lines := make(chan string, 8)
	go func() {
		scanner := bufio.NewScanner(response.Body)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()
	assertSSELines(t, lines, []string{"id: 1", "event: log", "data: {\"id\":\"1\""})

	live <- &core.LogEntry{ID: 2, RunID: 7, JobKey: "system.scheduler.test", Level: "info", Stage: "finish", Message: "live", CreatedAt: time.Unix(1, 0).UTC()}
	assertSSELines(t, lines, []string{"id: 2", "event: log", "data: {\"id\":\"2\""})
}

func assertSSELines(t *testing.T, lines <-chan string, prefixes []string) {
	t.Helper()
	for _, prefix := range prefixes {
		select {
		case line, ok := <-lines:
			if !ok || !strings.HasPrefix(line, prefix) {
				t.Fatalf("SSE line = %q, want prefix %q", line, prefix)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for SSE line %q", prefix)
		}
	}
	select {
	case line := <-lines:
		if line != "" {
			t.Fatalf("SSE event terminator = %q, want empty line", line)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for SSE event terminator")
	}
}

func TestStreamRunLogsSuppressesBufferedReplayDuplicate(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	h := server.New(server.WithListener(ln))
	live := make(chan *core.LogEntry, 2)
	live <- &core.LogEntry{ID: 2, RunID: 7, JobKey: "system.scheduler.test", Level: "info", Stage: "replayed", Message: "duplicate", CreatedAt: time.Unix(1, 0).UTC()}
	live <- &core.LogEntry{ID: 3, RunID: 7, JobKey: "system.scheduler.test", Level: "info", Stage: "finish", Message: "live", CreatedAt: time.Unix(2, 0).UTC()}
	h.GET("/stream", func(ctx context.Context, c *app.RequestContext) {
		writer := sse.NewWriter(c)
		_ = streamRunLogs(ctx, writer, []*dalModel.SystemJobRunLog{
			{ID: 1, RunID: 7, JobKey: "system.scheduler.test", Level: "info", Stage: "start", Message: "replayed", CreatedAt: time.Unix(0, 0).UTC()},
			{ID: 2, RunID: 7, JobKey: "system.scheduler.test", Level: "info", Stage: "execute", Message: "replayed", CreatedAt: time.Unix(1, 0).UTC()},
		}, live)
	})
	go h.Spin()
	t.Cleanup(func() { _ = h.Close() })

	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+ln.Addr().String()+"/stream", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := (&http.Client{}).Do(request)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = response.Body.Close() })

	lines := make(chan string, 12)
	go func() {
		scanner := bufio.NewScanner(response.Body)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()
	assertSSELines(t, lines, []string{"id: 1", "event: log", "data: {\"id\":\"1\""})
	assertSSELines(t, lines, []string{"id: 2", "event: log", "data: {\"id\":\"2\""})
	assertSSELines(t, lines, []string{"id: 3", "event: log", "data: {\"id\":\"3\""})
}

func TestRunDTOUsesRetryCountAndTriggeredAt(t *testing.T) {
	triggeredAt := time.Date(2026, 10, 12, 13, 50, 0, 0, time.UTC)
	finishedAt := triggeredAt.Add(time.Second)
	run := &dalModel.SystemJobRun{
		ID:          1,
		JobID:       2,
		JobKey:      "system.scheduler.test",
		Trigger:     "manual",
		Attempt:     1,
		Status:      "succeeded",
		TriggeredAt: triggeredAt,
		FinishedAt:  &finishedAt,
		Output:      "done",
	}

	actual := runDTO(run)
	if actual.RetryCount != 0 {
		t.Fatalf("retry count = %d, want 0", actual.RetryCount)
	}
	if actual.TriggeredAt != triggeredAt.Format(time.RFC3339) {
		t.Fatalf("triggered at = %q, want %q", actual.TriggeredAt, triggeredAt.Format(time.RFC3339))
	}
	if actual.FinishedAt != finishedAt.Format(time.RFC3339) {
		t.Fatalf("finished at = %q, want %q", actual.FinishedAt, finishedAt.Format(time.RFC3339))
	}
}
