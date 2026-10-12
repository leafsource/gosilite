package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	core "gosilite/application/framework/scheduler"
	"gosilite/types/jobkeys"
)

type RunLogCleaner interface {
	DeleteRunLogsBefore(context.Context, time.Time) (int64, error)
}

type SchedulerTest struct{}

func (SchedulerTest) Key() string { return jobkeys.SchedulerTest }

func (SchedulerTest) Run(ctx context.Context, _ string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	core.Info(ctx, "execute", "built-in scheduler test executed")
	return "scheduler test executed", nil
}

type SchedulerLogCleanup struct {
	Repo          RunLogCleaner
	RetentionDays int
}

func (SchedulerLogCleanup) Key() string { return jobkeys.SchedulerLogCleanup }

func (j SchedulerLogCleanup) Run(ctx context.Context, _ string) (string, error) {
	cutoff := time.Now().AddDate(0, 0, -j.RetentionDays)
	deleted, err := j.Repo.DeleteRunLogsBefore(ctx, cutoff)
	if err != nil {
		return "", err
	}
	message := fmt.Sprintf("deleted %d execution logs older than %d days", deleted, j.RetentionDays)
	hlog.Infof("[scheduler] %s", message)
	core.Info(ctx, "cleanup", "%s", message)
	return message, nil
}
