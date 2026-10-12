package scheduler

import (
	"bytes"
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type logStore struct {
	mu      sync.Mutex
	entries []*LogEntry
}

func (s *logStore) CreateRunLog(_ context.Context, entry *LogEntry) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := *entry
	copy.ID = int64(len(s.entries) + 1)
	s.entries = append(s.entries, &copy)
	return copy.ID, nil
}

func TestLogSinkPersistsThenBroadcasts(t *testing.T) {
	store := &logStore{}
	sink := NewLogSink(store)
	live, unsubscribe := sink.Subscribe(7)
	defer unsubscribe()
	sink.Write(&LogEntry{RunID: 7, JobKey: "system.scheduler.test", Level: "info", Stage: "execute", Message: "started"})
	select {
	case entry := <-live:
		if entry.ID != 1 || entry.Message != "started" {
			t.Fatalf("entry=%+v", entry)
		}
	case <-time.After(time.Second):
		t.Fatal("log was not broadcast after persistence")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if len(store.entries) != 1 || store.entries[0].ID != 1 {
		t.Fatalf("entries=%+v", store.entries)
	}
}

func TestRunLoggerUsesContext(t *testing.T) {
	store := &logStore{}
	sink := NewLogSink(store)
	live, unsubscribe := sink.Subscribe(9)
	defer unsubscribe()
	Info(WithRunLogger(context.Background(), sink, 9, "system.scheduler.test"), "execute", "step %d", 1)
	select {
	case entry := <-live:
		if entry.RunID != 9 || entry.Stage != "execute" || entry.Message != "step 1" {
			t.Fatalf("entry=%+v", entry)
		}
	case <-time.After(time.Second):
		t.Fatal("context log was not persisted")
	}
}

func TestRunLoggerEmitsContextToHertzLog(t *testing.T) {
	var output bytes.Buffer
	logger := hlog.DefaultLogger()
	logger.SetOutput(&output)
	t.Cleanup(func() { logger.SetOutput(os.Stderr) })

	store := &logStore{}
	sink := NewLogSink(store)
	live, unsubscribe := sink.Subscribe(9)
	defer unsubscribe()
	ctx := WithRunLogger(context.Background(), sink, 9, "system.scheduler.test")
	Info(ctx, "execute", "started")
	Warn(ctx, "retry", "attempt %d", 2)
	Error(ctx, "finish", "failed: %s", "boom")

	for range 3 {
		select {
		case <-live:
		case <-time.After(time.Second):
			t.Fatal("context log was not persisted")
		}
	}
	for _, want := range []string{
		"[scheduler] run=9 job=system.scheduler.test stage=execute started",
		"[scheduler] run=9 job=system.scheduler.test stage=retry attempt 2",
		"[scheduler] run=9 job=system.scheduler.test stage=finish failed: boom",
	} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("Hertz log = %q, missing %q", output.String(), want)
		}
	}
}
