package scheduler

import (
	"fmt"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/robfig/cron/v3"
)

type CronLogger struct {
	sink    *LogSink
	runID   int64
	jobKey  string
	onError func(error)
}

func newRunCronLogger(sink *LogSink, runID int64, jobKey string, onError func(error)) CronLogger {
	return CronLogger{sink: sink, runID: runID, jobKey: jobKey, onError: onError}
}

func (l CronLogger) Info(msg string, keysAndValues ...interface{}) {
	hlog.Infof("[scheduler cron] %s %v", msg, keysAndValues)
	l.write("info", msg, keysAndValues...)
}

func (l CronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	hlog.Errorf("[scheduler cron] %s err=%v %v", msg, err, keysAndValues)
	l.write("error", msg+": "+err.Error(), keysAndValues...)
	if l.onError != nil {
		l.onError(err)
	}
}

func (l CronLogger) write(level, msg string, keysAndValues ...interface{}) {
	if l.sink == nil || l.runID == 0 {
		return
	}
	l.sink.Write(&LogEntry{RunID: l.runID, JobKey: l.jobKey, Level: level, Stage: "cron", Message: fmt.Sprintf("%s %v", msg, keysAndValues)})
}

var _ cron.Logger = CronLogger{}
