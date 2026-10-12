package scheduler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/robfig/cron/v3"
)

const (
	RunStatusSucceeded     = "succeeded"
	RunStatusFailed        = "failed"
	RunStatusTimedOut      = "timed_out"
	RunStatusSkippedLocked = "skipped_locked"
)

type Definition struct {
	ID         int64
	JobKey     string
	Timeout    time.Duration
	RetryCount int
	ParamsJSON string
}

type Run struct {
	ID          int64
	JobID       int64
	JobKey      string
	Trigger     string
	Attempt     int
	Status      string
	TriggeredAt time.Time
	FinishedAt  time.Time
	Output      string
	Error       string
}

type RunRepository interface {
	ClaimActiveRun(context.Context, int64, string) (bool, error)
	ClearActiveRun(context.Context, int64, string) (bool, error)
	CreateRun(context.Context, *Run) (int64, error)
	FinishRun(context.Context, *Run) error
}

type Executor struct {
	registry *Registry
	leases   LeaseStore
	runs     RunRepository
	logs     *LogSink
	leaseTTL time.Duration
	now      func() time.Time
}

type jobResult struct {
	output string
	err    error
}

func NewExecutor(registry *Registry, leases LeaseStore, runs RunRepository, leaseTTL time.Duration) *Executor {
	if leaseTTL <= 0 {
		leaseTTL = time.Minute
	}
	return &Executor{registry: registry, leases: leases, runs: runs, leaseTTL: leaseTTL, now: time.Now}
}

func (e *Executor) SetLogSink(logs *LogSink) { e.logs = logs }

func (e *Executor) log(run *Run, level, stage, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	switch level {
	case "error":
		hlog.Errorf("[scheduler] run=%d job=%s stage=%s %s", run.ID, run.JobKey, stage, message)
	case "warn":
		hlog.Warnf("[scheduler] run=%d job=%s stage=%s %s", run.ID, run.JobKey, stage, message)
	default:
		hlog.Infof("[scheduler] run=%d job=%s stage=%s %s", run.ID, run.JobKey, stage, message)
	}
	if e.logs != nil {
		e.logs.Write(&LogEntry{RunID: run.ID, JobKey: run.JobKey, Level: level, Stage: stage, Message: message})
	}
}

func (e *Executor) Execute(ctx context.Context, def Definition, trigger string) error {
	job, ok := e.registry.Get(def.JobKey)
	if !ok {
		return fmt.Errorf("scheduler job not registered: %s", def.JobKey)
	}
	attempts := def.RetryCount + 1
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		run := &Run{JobID: def.ID, JobKey: def.JobKey, Trigger: trigger, Attempt: attempt, TriggeredAt: e.now()}
		runID, err := e.runs.CreateRun(ctx, run)
		if err != nil {
			return err
		}
		run.ID = runID
		e.log(run, "info", "trigger", "run created: trigger=%s attempt=%d", trigger, attempt)
		activeToken := randomToken()
		claimed, claimErr := e.runs.ClaimActiveRun(ctx, def.ID, activeToken)
		if claimErr != nil {
			e.log(run, "error", "lock", "active execution fence unavailable: %v", claimErr)
			run.Status, run.Error, run.FinishedAt = RunStatusFailed, "active execution fence unavailable: "+claimErr.Error(), e.now()
			if finishErr := e.runs.FinishRun(ctx, run); finishErr != nil {
				return finishErr
			}
			return claimErr
		}
		if !claimed {
			e.log(run, "warn", "lock", "active execution fence held")
			run.Status, run.Error, run.FinishedAt = RunStatusSkippedLocked, "job active execution fence held", e.now()
			return e.runs.FinishRun(ctx, run)
		}

		lease, locked, err := AcquireLease(ctx, e.leases, def.JobKey, randomToken(), e.leaseTTL)
		if err != nil {
			_, _ = e.runs.ClearActiveRun(context.Background(), def.ID, activeToken)
			e.log(run, "error", "lock", "redis lease unavailable: %v", err)
			run.Status, run.Error, run.FinishedAt = RunStatusFailed, "redis lease unavailable: "+err.Error(), e.now()
			if finishErr := e.runs.FinishRun(ctx, run); finishErr != nil {
				return finishErr
			}
			return err
		}
		if !locked {
			_, _ = e.runs.ClearActiveRun(context.Background(), def.ID, activeToken)
			e.log(run, "warn", "lock", "redis lease held by another instance")
			run.Status, run.Error, run.FinishedAt = RunStatusSkippedLocked, "job lease held by another instance", e.now()
			return e.runs.FinishRun(ctx, run)
		}

		e.log(run, "info", "lock", "redis lease acquired")
		output, runErr, pending := e.runWithRenewal(ctx, job, def, lease, run, activeToken)
		if pending {
			return runErr
		}
		_, releaseErr := lease.Release(context.Background())
		run.Output, run.FinishedAt = output, e.now()
		if runErr == nil {
			run.Status = RunStatusSucceeded
		} else {
			run.Status, run.Error = statusForError(runErr), runErr.Error()
		}
		if runErr == nil {
			e.log(run, "info", "finish", "run succeeded")
		} else {
			e.log(run, "error", "finish", "run finished: %v", runErr)
		}
		finishErr := e.runs.FinishRun(ctx, run)
		if finishErr != nil {
			return finishErr
		}
		_, clearErr := e.runs.ClearActiveRun(context.Background(), def.ID, activeToken)
		if clearErr != nil {
			return clearErr
		}
		if releaseErr != nil {
			return releaseErr
		}
		if runErr == nil {
			return nil
		}
		lastErr = runErr
	}
	return lastErr
}

func (e *Executor) runWithRenewal(ctx context.Context, job Job, def Definition, lease *Lease, run *Run, activeToken string) (string, error, bool) {
	jobCtx, cancel := context.WithTimeout(ctx, def.Timeout)
	jobCtx = WithRunLogger(jobCtx, e.logs, run.ID, run.JobKey)
	e.log(run, "info", "execute", "job execution started")
	finished := make(chan jobResult, 1)
	go func() {
		var result jobResult
		logger := newRunCronLogger(e.logs, run.ID, run.JobKey, func(err error) {
			result.err = fmt.Errorf("scheduler job panic: %w", err)
		})
		cron.NewChain(cron.Recover(logger)).Then(cron.FuncJob(func() {
			result.output, result.err = job.Run(jobCtx, def.ParamsJSON)
		})).Run()
		finished <- result
	}()
	ticker := time.NewTicker(e.leaseTTL / 2)
	defer ticker.Stop()
	for {
		select {
		case result := <-finished:
			cancel()
			return result.output, result.err, false
		case <-jobCtx.Done():
			cancel()
			e.log(run, "error", "execute", "job context ended: %v", jobCtx.Err())
			e.finishPendingRun(run, def.ID, activeToken, lease, finished, jobCtx.Err())
			return "", jobCtx.Err(), true
		case <-ticker.C:
			ok, err := lease.Renew(ctx)
			if err != nil || !ok {
				cancel()
				leaseErr := errors.New("scheduler lease lost")
				if err != nil {
					leaseErr = err
				}
				e.log(run, "error", "lock", "redis lease renewal lost: %v", leaseErr)
				e.finishPendingRun(run, def.ID, activeToken, lease, finished, leaseErr)
				return "", leaseErr, true
			}
		}
	}
}

func (e *Executor) finishPendingRun(run *Run, jobID int64, activeToken string, lease *Lease, finished <-chan jobResult, terminalErr error) {
	go func() {
		ticker := time.NewTicker(e.leaseTTL / 2)
		defer ticker.Stop()
		for {
			select {
			case result := <-finished:
				_, _ = lease.Release(context.Background())
				e.log(run, "error", "finish", "worker exited after terminal state: %v", terminalErr)
				run.Output, run.Status, run.Error, run.FinishedAt = result.output, statusForError(terminalErr), terminalErr.Error(), e.now()
				_ = e.runs.FinishRun(context.Background(), run)
				_, _ = e.runs.ClearActiveRun(context.Background(), jobID, activeToken)
				return
			case <-ticker.C:
				_, _ = lease.Renew(context.Background())
			}
		}
	}()
}

func statusForError(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return RunStatusTimedOut
	}
	return RunStatusFailed
}

func randomToken() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}
