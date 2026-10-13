package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type LogEntry struct {
	ID        int64
	RunID     int64
	JobKey    string
	Level     string
	Stage     string
	Message   string
	CreatedAt time.Time
}

type LogStore interface {
	CreateRunLog(context.Context, *LogEntry) (int64, error)
}

type LogSink struct {
	store       LogStore
	queue       chan *LogEntry
	mu          sync.RWMutex
	subscribers map[int64]map[chan *LogEntry]struct{}
}

func NewLogSink(store LogStore) *LogSink {
	s := &LogSink{store: store, queue: make(chan *LogEntry, 1024), subscribers: make(map[int64]map[chan *LogEntry]struct{})}
	go s.persist()
	return s
}

func (s *LogSink) Write(entry *LogEntry) {
	if entry == nil || entry.RunID == 0 {
		return
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	copy := *entry
	select {
	case s.queue <- &copy:
	default:
		hlog.Errorf("[scheduler] execution log queue full: run=%d", entry.RunID)
	}
}

func (s *LogSink) Subscribe(runID int64) (<-chan *LogEntry, func()) {
	ch := make(chan *LogEntry, 64)
	s.mu.Lock()
	if s.subscribers[runID] == nil {
		s.subscribers[runID] = make(map[chan *LogEntry]struct{})
	}
	s.subscribers[runID][ch] = struct{}{}
	s.mu.Unlock()
	return ch, func() {
		s.mu.Lock()
		if subscribers := s.subscribers[runID]; subscribers != nil {
			delete(subscribers, ch)
			if len(subscribers) == 0 {
				delete(s.subscribers, runID)
			}
		}
		s.mu.Unlock()
	}
}

func (s *LogSink) persist() {
	for entry := range s.queue {
		id, err := s.store.CreateRunLog(context.Background(), entry)
		if err != nil {
			hlog.Errorf("[scheduler] persist execution log failed: run=%d err=%v", entry.RunID, err)
			continue
		}
		entry.ID = id
		s.broadcast(entry)
	}
}

func (s *LogSink) broadcast(entry *LogEntry) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for ch := range s.subscribers[entry.RunID] {
		copy := *entry
		select {
		case ch <- &copy:
		default:
			hlog.Errorf("[scheduler] execution log subscriber slow: run=%d", entry.RunID)
		}
	}
}

type runLoggerKey struct{}

type RunLogger struct {
	sink   *LogSink
	runID  int64
	jobKey string
}

func WithRunLogger(ctx context.Context, sink *LogSink, runID int64, jobKey string) context.Context {
	return context.WithValue(ctx, runLoggerKey{}, &RunLogger{sink: sink, runID: runID, jobKey: jobKey})
}

func LoggerFromContext(ctx context.Context) *RunLogger {
	logger, _ := ctx.Value(runLoggerKey{}).(*RunLogger)
	return logger
}

func Log(ctx context.Context, level, stage, format string, args ...any) {
	logger := LoggerFromContext(ctx)
	if logger == nil || logger.sink == nil {
		return
	}
	message := fmt.Sprintf(format, args...)
	switch level {
	case "error":
		hlog.Errorf("[scheduler] run=%d job=%s stage=%s %s", logger.runID, logger.jobKey, stage, message)
	case "warn":
		hlog.Warnf("[scheduler] run=%d job=%s stage=%s %s", logger.runID, logger.jobKey, stage, message)
	default:
		hlog.Infof("[scheduler] run=%d job=%s stage=%s %s", logger.runID, logger.jobKey, stage, message)
	}
	logger.sink.Write(&LogEntry{RunID: logger.runID, JobKey: logger.jobKey, Level: level, Stage: stage, Message: message})
}

func Info(ctx context.Context, stage, format string, args ...any) {
	Log(ctx, "info", stage, format, args...)
}
func Warn(ctx context.Context, stage, format string, args ...any) {
	Log(ctx, "warn", stage, format, args...)
}
func Error(ctx context.Context, stage, format string, args ...any) {
	Log(ctx, "error", stage, format, args...)
}
