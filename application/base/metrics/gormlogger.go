package metrics

import (
	"context"
	"time"

	"gorm.io/gorm/logger"
)

// GORMLogger 包装默认 logger：Trace 时做廉价聚合（不额外查库）。
type GORMLogger struct {
	inner logger.Interface
}

// NewGORMLogger 为 GORM 安装监控 Trace。
func NewGORMLogger(inner logger.Interface) logger.Interface {
	if inner == nil {
		inner = logger.Default
	}
	return &GORMLogger{inner: inner}
}

func (l *GORMLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &GORMLogger{inner: l.inner.LogMode(level)}
}

func (l *GORMLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.inner.Info(ctx, msg, data...)
}

func (l *GORMLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.inner.Warn(ctx, msg, data...)
}

func (l *GORMLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.inner.Error(ctx, msg, data...)
}

func (l *GORMLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, _ := fc()
	if sql != "" {
		ObserveSQL(sql, time.Since(begin), err)
	}
	l.inner.Trace(ctx, begin, fc, err)
}

// ObserveSQL 记录一次 SQL 执行。
func ObserveSQL(sql string, cost time.Duration, err error) {
	sqlStore.Observe(sql, cost, err)
}
