package scheduler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/robfig/cron/v3"
)

const (
	RunStatusRunning       = "running"
	RunStatusStopping      = "stopping"
	RunStatusStopped       = "stopped"
	RunStatusSucceeded     = "succeeded"
	RunStatusFailed        = "failed"
	RunStatusTimedOut      = "timed_out"
	RunStatusSkippedLocked = "skipped_locked"
)

type Definition struct {
	ID            int64
	JobKey        string
	Timeout       time.Duration
	RetryCount    int
	ParamsJSON    string
	JobName       string
	ConfigVersion int32
}

type Run struct {
	ID            int64
	JobID         int64
	JobKey        string
	Trigger       string
	Attempt       int
	Status        string
	TriggeredAt   time.Time
	FinishedAt    time.Time
	Output        string
	Error         string
	JobName       string
	ParamsJSON    string
	TimeoutSec    int32
	ConfigVersion int32
}

type RunRepository interface {
	AcquireRunSlot(context.Context, int64) (bool, error)
	ReleaseRunSlot(context.Context, int64) (bool, error)
	CreateRun(context.Context, *Run) (int64, error)
	FinishRun(context.Context, *Run) error
	RequestStopRun(context.Context, int64) (bool, error)
	IsRunStopping(context.Context, int64) (bool, error)
}

const durableStopPollInterval = time.Second

type Executor struct {
	registry         *Registry
	leases           LeaseStore
	runs             RunRepository
	logs             *LogSink
	leaseTTL         time.Duration
	stopPollInterval time.Duration
	now              func() time.Time
	mu               sync.Mutex
	cancels          map[int64]context.CancelFunc
}

type jobResult struct {
	output string
	err    error
}

func NewExecutor(registry *Registry, leases LeaseStore, runs RunRepository, leaseTTL time.Duration) *Executor {
	if leaseTTL <= 0 {
		leaseTTL = time.Minute
	}
	return &Executor{registry: registry, leases: leases, runs: runs, leaseTTL: leaseTTL, stopPollInterval: durableStopPollInterval, now: time.Now, cancels: map[int64]context.CancelFunc{}}
}

func (e *Executor) SetLogSink(logs *LogSink) { e.logs = logs }

func (e *Executor) Stop(ctx context.Context, runID int64) (bool, error) {
	stopped, err := e.runs.RequestStopRun(ctx, runID)
	if err != nil || !stopped {
		return stopped, err
	}
	e.mu.Lock()
	cancel := e.cancels[runID]
	e.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	return true, nil
}

func (e *Executor) registerCancel(runID int64, cancel context.CancelFunc) {
	e.mu.Lock()
	e.cancels[runID] = cancel
	e.mu.Unlock()
}

func (e *Executor) clearCancel(runID int64) {
	e.mu.Lock()
	delete(e.cancels, runID)
	e.mu.Unlock()
}

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
		run := &Run{JobID: def.ID, JobKey: def.JobKey, Trigger: trigger, Attempt: attempt, Status: RunStatusRunning, TriggeredAt: e.now(), JobName: def.JobName, ParamsJSON: def.ParamsJSON, TimeoutSec: int32(def.Timeout / time.Second), ConfigVersion: def.ConfigVersion}
		runID, err := e.runs.CreateRun(ctx, run)
		if err != nil {
			return err
		}
		run.ID = runID
		e.log(run, "info", "trigger", "run created: trigger=%s attempt=%d", trigger, attempt)
		acquired, acquireErr := e.runs.AcquireRunSlot(ctx, def.ID)
		if acquireErr != nil {
			e.log(run, "error", "capacity", "execution capacity unavailable: %v", acquireErr)
			run.Status, run.Error, run.FinishedAt = RunStatusFailed, "execution capacity unavailable: "+acquireErr.Error(), e.now()
			if finishErr := e.runs.FinishRun(ctx, run); finishErr != nil {
				return finishErr
			}
			return acquireErr
		}
		if !acquired {
			e.log(run, "warn", "capacity", "job execution capacity exhausted")
			run.Status, run.Error, run.FinishedAt = RunStatusSkippedLocked, "job execution capacity exhausted", e.now()
			return e.runs.FinishRun(ctx, run)
		}

		lease, locked, err := AcquireLease(ctx, e.leases, def.JobKey+":"+randomToken(), randomToken(), e.leaseTTL)
		if err != nil {
			_, _ = e.runs.ReleaseRunSlot(context.Background(), def.ID)
			e.log(run, "error", "lock", "redis lease unavailable: %v", err)
			run.Status, run.Error, run.FinishedAt = RunStatusFailed, "redis lease unavailable: "+err.Error(), e.now()
			if finishErr := e.runs.FinishRun(ctx, run); finishErr != nil {
				return finishErr
			}
			return err
		}
		if !locked {
			_, _ = e.runs.ReleaseRunSlot(context.Background(), def.ID)
			e.log(run, "warn", "lock", "redis lease held by another instance")
			run.Status, run.Error, run.FinishedAt = RunStatusSkippedLocked, "job lease held by another instance", e.now()
			return e.runs.FinishRun(ctx, run)
		}

		e.log(run, "info", "lock", "redis lease acquired")
		output, runErr, pending := e.runWithRenewal(ctx, job, def, lease, run)
		if pending {
			return runErr
		}
		if errors.Is(runErr, context.Canceled) {
			stopping, err := e.runs.IsRunStopping(context.Background(), run.ID)
			if err != nil {
				return err
			}
			if stopping {
				run.Status = RunStatusStopped
				run.Error = ""
				run.Output = output
				run.FinishedAt = e.now()
				e.log(run, "info", "finish", "run stopped cooperatively")
				if err := e.runs.FinishRun(ctx, run); err != nil {
					return err
				}
				_, _ = lease.Release(context.Background())
				_, err := e.runs.ReleaseRunSlot(context.Background(), def.ID)
				return err
			}
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
		_, releaseSlotErr := e.runs.ReleaseRunSlot(context.Background(), def.ID)
		if releaseSlotErr != nil {
			return releaseSlotErr
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

func (e *Executor) runWithRenewal(ctx context.Context, job Job, def Definition, lease *Lease, run *Run) (string, error, bool) {
	jobCtx, cancel := context.WithTimeout(ctx, def.Timeout)
	stopping, err := e.runs.IsRunStopping(ctx, run.ID)
	if err != nil {
		cancel()
		return "", err, false
	}
	if stopping {
		cancel()
	}
	e.registerCancel(run.ID, cancel)
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
	renewTicker := time.NewTicker(e.leaseTTL / 2)
	defer renewTicker.Stop()
	stopPollTicker := time.NewTicker(e.stopPollInterval)
	defer stopPollTicker.Stop()
	for {
		select {
		case result := <-finished:
			cancel()
			e.clearCancel(run.ID)
			return result.output, result.err, false
		case <-jobCtx.Done():
			cancel()
			e.clearCancel(run.ID)
			if errors.Is(jobCtx.Err(), context.Canceled) {
				e.log(run, "info", "execute", "stop requested; awaiting worker exit")
				e.finishPendingRun(run, def.ID, lease, finished, context.Canceled)
				return "", context.Canceled, true
			}
			e.log(run, "error", "execute", "job context ended: %v", jobCtx.Err())
			e.finishPendingRun(run, def.ID, lease, finished, jobCtx.Err())
			return "", jobCtx.Err(), true
		case <-stopPollTicker.C:
			stopping, err := e.runs.IsRunStopping(context.Background(), run.ID)
			if err != nil {
				e.log(run, "warn", "execute", "durable stop status poll failed: %v", err)
			} else if stopping {
				cancel()
			}
		case <-renewTicker.C:
			ok, err := lease.Renew(ctx)
			if err != nil || !ok {
				cancel()
				leaseErr := errors.New("scheduler lease lost")
				if err != nil {
					leaseErr = err
				}
				e.log(run, "error", "lock", "redis lease renewal lost: %v", leaseErr)
				e.finishPendingRun(run, def.ID, lease, finished, leaseErr)
				return "", leaseErr, true
			}
		}
	}
}

func (e *Executor) finishPendingRun(run *Run, jobID int64, lease *Lease, finished <-chan jobResult, terminalErr error) {
	go func() {
		ticker := time.NewTicker(e.leaseTTL / 2)
		defer ticker.Stop()
		for {
			select {
			case result := <-finished:
				_, _ = lease.Release(context.Background())
				stopping, _ := e.runs.IsRunStopping(context.Background(), run.ID)
				if errors.Is(terminalErr, context.Canceled) && stopping {
					e.log(run, "info", "finish", "worker stopped cooperatively")
					run.Output, run.Status, run.Error, run.FinishedAt = result.output, RunStatusStopped, "", e.now()
				} else {
					e.log(run, "error", "finish", "worker exited after terminal state: %v", terminalErr)
					run.Output, run.Status, run.Error, run.FinishedAt = result.output, statusForError(terminalErr), terminalErr.Error(), e.now()
				}
				_ = e.runs.FinishRun(context.Background(), run)
				_, _ = e.runs.ReleaseRunSlot(context.Background(), jobID)
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
