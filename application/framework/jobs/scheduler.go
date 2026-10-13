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
	const totalSteps = 10

	core.Info(ctx, "execute", "built-in scheduler test started: %d seconds", totalSteps)
	for step := 1; step <= totalSteps; step++ {
		select {
		case <-ctx.Done():
			core.Warn(ctx, "execute", "built-in scheduler test canceled at step %d/%d", step, totalSteps)
			return "", ctx.Err()
		case <-time.After(time.Second):
			core.Info(ctx, "execute", "built-in scheduler test progress: %d/%d", step, totalSteps)
		}
	}
	core.Info(ctx, "execute", "built-in scheduler test completed")
	return "scheduler test completed after 10 seconds", nil
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
