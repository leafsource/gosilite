package scheduler

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type testJob struct {
	key string
	run func(context.Context, string) (string, error)
}

func (j testJob) Key() string                                     { return j.key }
func (j testJob) Run(c context.Context, p string) (string, error) { return j.run(c, p) }

type testLease struct {
	mu        sync.Mutex
	held      bool
	fail      bool
	renewFail bool
}

func (l *testLease) Acquire(_ context.Context, _, _ string, _ time.Duration) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.fail {
		return false, errors.New("down")
	}
	if l.held {
		return false, nil
	}
	l.held = true
	return true, nil
}
func (l *testLease) Renew(context.Context, string, string, time.Duration) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.renewFail {
		l.held = false
		return false, nil
	}
	return true, nil
}
func (l *testLease) Release(context.Context, string, string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.held = false
	return true, nil
}

type testRuns struct {
	mu              sync.Mutex
	runs            []*Run
	finished        []int64
	activeRunTokens map[int64]string
}

func (r *testRuns) ClaimActiveRun(_ context.Context, jobID int64, token string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.activeRunTokens == nil {
		r.activeRunTokens = map[int64]string{}
	}
	if r.activeRunTokens[jobID] != "" {
		return false, nil
	}
	r.activeRunTokens[jobID] = token
	return true, nil
}

func (r *testRuns) ClearActiveRun(_ context.Context, jobID int64, token string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.activeRunTokens[jobID] != token {
		return false, nil
	}
	r.activeRunTokens[jobID] = ""
	return true, nil
}

func (r *testRuns) CreateRun(_ context.Context, x *Run) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	x.ID = int64(len(r.runs) + 100)
	r.runs = append(r.runs, x)
	return x.ID, nil
}

func (r *testRuns) FinishRun(_ context.Context, x *Run) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, run := range r.runs {
		if run.ID == x.ID {
			r.finished = append(r.finished, x.ID)
			return nil
		}
	}
	return errors.New("run not found")
}
func TestRegistryRejectsDuplicateKey(t *testing.T) {
	r := NewRegistry()
	j := testJob{"system.scheduler.test", func(context.Context, string) (string, error) { return "", nil }}
	if err := r.Register(j); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(j); err == nil {
		t.Fatal("expected duplicate rejection")
	}
}
func TestExecutorRetriesFailure(t *testing.T) {
	r := NewRegistry()
	n := 0
	_ = r.Register(testJob{"system.scheduler.test", func(context.Context, string) (string, error) {
		n++
		if n == 1 {
			return "", errors.New("fail")
		}
		return "ok", nil
	}})
	runs := &testRuns{}
	e := NewExecutor(r, &testLease{}, runs, time.Second)
	if err := e.Execute(context.Background(), Definition{ID: 1, JobKey: "system.scheduler.test", Timeout: time.Second, RetryCount: 1}, "manual"); err != nil {
		t.Fatal(err)
	}
	if len(runs.runs) != 2 || runs.runs[0].Status != RunStatusFailed || runs.runs[1].Status != RunStatusSucceeded {
		t.Fatalf("runs=%+v", runs.runs)
	}
	if runs.runs[0].ID == 0 || runs.runs[1].ID == 0 || runs.runs[0].ID == runs.runs[1].ID {
		t.Fatalf("run IDs were not assigned uniquely: %+v", runs.runs)
	}
	if len(runs.finished) != 2 || runs.finished[0] != runs.runs[0].ID || runs.finished[1] != runs.runs[1].ID {
		t.Fatalf("finish must target created IDs: finished=%v runs=%+v", runs.finished, runs.runs)
	}
}
func TestExecutorRecordsLockSkip(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(testJob{"system.scheduler.test", func(context.Context, string) (string, error) { return "", nil }})
	runs := &testRuns{}
	e := NewExecutor(r, &testLease{held: true}, runs, time.Second)
	if err := e.Execute(context.Background(), Definition{JobKey: "system.scheduler.test", Timeout: time.Second}, "cron"); err != nil {
		t.Fatal(err)
	}
	if runs.runs[0].Status != RunStatusSkippedLocked {
		t.Fatal(runs.runs[0].Status)
	}
}
func TestExecutorTimesOut(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(testJob{"system.scheduler.test", func(ctx context.Context, _ string) (string, error) { <-ctx.Done(); return "", ctx.Err() }})
	runs := &testRuns{}
	e := NewExecutor(r, &testLease{}, runs, time.Second)
	err := e.Execute(context.Background(), Definition{JobKey: "system.scheduler.test", Timeout: time.Millisecond}, "manual")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v", err)
	}
	deadline := time.After(time.Second)
	for {
		runs.mu.Lock()
		status := runs.runs[0].Status
		finished := len(runs.finished)
		runs.mu.Unlock()
		if status == RunStatusTimedOut && finished == 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("timed-out run was not finalized: status=%s finished=%d", status, finished)
		case <-time.After(time.Millisecond):
		}
	}
}
func TestExecutorLeaseRenewalLossRetainsDurableFenceUntilWorkerExits(t *testing.T) {
	r := NewRegistry()
	block := make(chan struct{})
	started := make(chan struct{})
	_ = r.Register(testJob{"system.scheduler.test", func(context.Context, string) (string, error) {
		close(started)
		<-block
		return "finished", nil
	}})
	runs := &testRuns{}
	leases := &testLease{renewFail: true}
	e := NewExecutor(r, leases, runs, 10*time.Millisecond)

	err := e.Execute(context.Background(), Definition{ID: 1, JobKey: "system.scheduler.test", Timeout: time.Second}, "manual")
	if err == nil || err.Error() != "scheduler lease lost" {
		t.Fatalf("err=%v", err)
	}
	<-started
	if err := e.Execute(context.Background(), Definition{ID: 1, JobKey: "system.scheduler.test", Timeout: time.Second}, "manual"); err != nil {
		t.Fatal(err)
	}
	runs.mu.Lock()
	second := runs.runs[1]
	activeToken := runs.activeRunTokens[1]
	runs.mu.Unlock()
	if second.Status != RunStatusSkippedLocked || activeToken == "" {
		t.Fatalf("competing execution ran despite lost Redis lease: second=%+v activeToken=%q", second, activeToken)
	}

	close(block)
	deadline := time.After(time.Second)
	for {
		runs.mu.Lock()
		finished := len(runs.finished)
		activeToken := runs.activeRunTokens[1]
		runs.mu.Unlock()
		if finished == 2 && activeToken == "" {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("active fence was not cleared after worker exit: finished=%d activeToken=%q", finished, activeToken)
		case <-time.After(time.Millisecond):
		}
	}
}

func TestExecutorTimeoutRetainsLeaseAndRunUntilJobExits(t *testing.T) {
	r := NewRegistry()
	block := make(chan struct{})
	started := make(chan struct{})
	_ = r.Register(testJob{"system.scheduler.test", func(context.Context, string) (string, error) {
		close(started)
		<-block
		return "finished", nil
	}})
	runs := &testRuns{}
	leases := &testLease{}
	e := NewExecutor(r, leases, runs, 10*time.Millisecond)
	start := time.Now()
	err := e.Execute(context.Background(), Definition{JobKey: "system.scheduler.test", Timeout: time.Millisecond, RetryCount: 1}, "manual")
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > 100*time.Millisecond {
		t.Fatalf("err=%v elapsed=%v", err, time.Since(start))
	}
	<-started
	if len(runs.finished) != 0 || len(runs.runs) != 1 {
		t.Fatalf("timed-out job must remain active: finished=%v runs=%+v", runs.finished, runs.runs)
	}
	if err := e.Execute(context.Background(), Definition{JobKey: "system.scheduler.test", Timeout: time.Second}, "manual"); err != nil {
		t.Fatal(err)
	}
	if len(runs.runs) != 2 || runs.runs[1].Status != RunStatusSkippedLocked {
		t.Fatalf("second execution must be skipped while worker is alive: %+v", runs.runs)
	}
	close(block)
	deadline := time.After(time.Second)
	for {
		runs.mu.Lock()
		finished := len(runs.finished)
		status := runs.runs[0].Status
		runs.mu.Unlock()
		if finished == 2 && status == RunStatusTimedOut {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("run was not finalized after worker exit: finished=%d status=%s", finished, status)
		case <-time.After(time.Millisecond):
		}
	}
}
