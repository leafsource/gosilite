package scheduler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
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
		activeToken := randomToken()
		claimed, claimErr := e.runs.ClaimActiveRun(ctx, def.ID, activeToken)
		if claimErr != nil {
			run.Status, run.Error, run.FinishedAt = RunStatusFailed, "active execution fence unavailable: "+claimErr.Error(), e.now()
			if finishErr := e.runs.FinishRun(ctx, run); finishErr != nil {
				return finishErr
			}
			return claimErr
		}
		if !claimed {
			run.Status, run.Error, run.FinishedAt = RunStatusSkippedLocked, "job active execution fence held", e.now()
			return e.runs.FinishRun(ctx, run)
		}

		lease, locked, err := AcquireLease(ctx, e.leases, def.JobKey, randomToken(), e.leaseTTL)
		if err != nil {
			_, _ = e.runs.ClearActiveRun(context.Background(), def.ID, activeToken)
			run.Status, run.Error, run.FinishedAt = RunStatusFailed, "redis lease unavailable: "+err.Error(), e.now()
			if finishErr := e.runs.FinishRun(ctx, run); finishErr != nil {
				return finishErr
			}
			return err
		}
		if !locked {
			_, _ = e.runs.ClearActiveRun(context.Background(), def.ID, activeToken)
			run.Status, run.Error, run.FinishedAt = RunStatusSkippedLocked, "job lease held by another instance", e.now()
			return e.runs.FinishRun(ctx, run)
		}

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
	finished := make(chan jobResult, 1)
	go func() {
		output, err := job.Run(jobCtx, def.ParamsJSON)
		finished <- jobResult{output: output, err: err}
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
