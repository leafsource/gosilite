package scheduler

import (
	"context"
	"errors"
	"strings"
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
	keys      map[string]bool
	fail      bool
	renewFail bool
}

func (l *testLease) Acquire(_ context.Context, key, _ string, _ time.Duration) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.fail {
		return false, errors.New("down")
	}
	if l.held && len(l.keys) == 0 {
		return false, nil
	}
	if l.keys == nil {
		l.keys = map[string]bool{}
	}
	if l.keys[key] {
		return false, nil
	}
	l.keys[key] = true
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
	activeRunCounts map[int64]int
	maxInstances    map[int64]int
	acquireCalls    int
	releaseCalls    int
}

func (r *testRuns) RequestStopRun(_ context.Context, runID int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, run := range r.runs {
		if run.ID == runID && run.Status == RunStatusRunning {
			run.Status = RunStatusStopping
			return true, nil
		}
	}
	return false, nil
}

func (r *testRuns) IsRunStopping(_ context.Context, runID int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, run := range r.runs {
		if run.ID == runID {
			return run.Status == RunStatusStopping, nil
		}
	}
	return false, nil
}

func (r *testRuns) AcquireRunSlot(_ context.Context, jobID int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.activeRunCounts == nil {
		r.activeRunCounts = map[int64]int{}
	}
	limit := r.maxInstances[jobID]
	if limit == 0 {
		limit = 1
	}
	if r.activeRunCounts[jobID] >= limit {
		return false, nil
	}
	r.activeRunCounts[jobID]++
	r.acquireCalls++
	return true, nil
}

func (r *testRuns) ReleaseRunSlot(_ context.Context, jobID int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.activeRunCounts[jobID] == 0 {
		return false, nil
	}
	r.activeRunCounts[jobID]--
	r.releaseCalls++
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
	if err := e.Execute(context.Background(), Definition{ID: 1, JobKey: "system.scheduler.test", JobName: "original", ParamsJSON: `{"value":1}`, Timeout: time.Second, RetryCount: 1, ConfigVersion: 7}, "manual"); err != nil {
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
	for _, run := range runs.runs {
		if run.JobName != "original" || run.ParamsJSON != `{"value":1}` || run.TimeoutSec != 1 || run.ConfigVersion != 7 {
			t.Fatalf("run snapshot = %+v", run)
		}
	}
	if runs.acquireCalls != 2 || runs.releaseCalls != 2 {
		t.Fatalf("each retry must reacquire and release a slot: acquire=%d release=%d", runs.acquireCalls, runs.releaseCalls)
	}
}

func TestExecutorAllowsConfiguredConcurrentInstances(t *testing.T) {
	r := NewRegistry()
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	_ = r.Register(testJob{"system.scheduler.test", func(context.Context, string) (string, error) {
		started <- struct{}{}
		<-release
		return "done", nil
	}})
	runs := &testRuns{maxInstances: map[int64]int{1: 2}}
	e := NewExecutor(r, &testLease{}, runs, time.Second)
	errs := make(chan error, 2)
	for range 2 {
		go func() {
			errs <- e.Execute(context.Background(), Definition{ID: 1, JobKey: "system.scheduler.test", Timeout: time.Second}, "manual")
		}()
	}
	<-started
	<-started
	runs.mu.Lock()
	active := runs.activeRunCounts[1]
	runs.mu.Unlock()
	if active != 2 {
		t.Fatalf("active run count = %d, want 2", active)
	}
	if err := e.Execute(context.Background(), Definition{ID: 1, JobKey: "system.scheduler.test", Timeout: time.Second}, "manual"); err != nil {
		t.Fatal(err)
	}
	runs.mu.Lock()
	third := runs.runs[2]
	runs.mu.Unlock()
	if third.Status != RunStatusSkippedLocked {
		t.Fatalf("third run status = %q, want capacity skip", third.Status)
	}
	close(release)
	if err := <-errs; err != nil {
		t.Fatal(err)
	}
	if err := <-errs; err != nil {
		t.Fatal(err)
	}
	runs.mu.Lock()
	active = runs.activeRunCounts[1]
	runs.mu.Unlock()
	if active != 0 {
		t.Fatalf("active run count = %d after exits, want 0", active)
	}
}
func TestExecutorPersistsCronWrapperPanicUnderRun(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(testJob{"system.scheduler.test", func(context.Context, string) (string, error) {
		panic("boom")
	}})
	runs := &testRuns{}
	store := &logStore{}
	e := NewExecutor(r, &testLease{}, runs, time.Second)
	e.SetLogSink(NewLogSink(store))

	err := e.Execute(context.Background(), Definition{ID: 1, JobKey: "system.scheduler.test", Timeout: time.Second}, "cron")
	if err == nil || err.Error() != "scheduler job panic: boom" {
		t.Fatalf("err=%v", err)
	}

	deadline := time.After(time.Second)
	for {
		store.mu.Lock()
		entries := append([]*LogEntry(nil), store.entries...)
		store.mu.Unlock()
		for _, entry := range entries {
			if entry.RunID == runs.runs[0].ID && entry.Stage == "cron" && entry.Level == "error" && strings.HasPrefix(entry.Message, "panic: boom [stack ") {
				return
			}
		}
		select {
		case <-deadline:
			t.Fatalf("panic log not persisted for run %d: %+v", runs.runs[0].ID, entries)
		case <-time.After(time.Millisecond):
		}
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
	activeToken := runs.activeRunCounts[1]
	runs.mu.Unlock()
	if second.Status != RunStatusSkippedLocked || activeToken != 1 {
		t.Fatalf("competing execution ran despite lost Redis lease: second=%+v activeRunCount=%d", second, activeToken)
	}

	close(block)
	deadline := time.After(time.Second)
	for {
		runs.mu.Lock()
		finished := len(runs.finished)
		activeToken := runs.activeRunCounts[1]
		runs.mu.Unlock()
		if finished == 2 && activeToken == 0 {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("active slot was not released after worker exit: finished=%d activeRunCount=%d", finished, activeToken)
		case <-time.After(time.Millisecond):
		}
	}
}

func TestExecutorStopCancelsWorkerAndPreservesFenceUntilExit(t *testing.T) {
	r := NewRegistry()
	started := make(chan struct{})
	exit := make(chan struct{})
	_ = r.Register(testJob{"system.scheduler.test", func(ctx context.Context, _ string) (string, error) {
		close(started)
		<-ctx.Done()
		<-exit
		return "stopped", ctx.Err()
	}})
	runs := &testRuns{}
	e := NewExecutor(r, &testLease{}, runs, time.Second)
	finished := make(chan error, 1)
	go func() {
		finished <- e.Execute(context.Background(), Definition{ID: 1, JobKey: "system.scheduler.test", Timeout: time.Minute}, "manual")
	}()
	<-started
	if stopped, err := e.Stop(context.Background(), 100); err != nil || !stopped {
		t.Fatalf("stop = %v, %v", stopped, err)
	}
	if err := e.Execute(context.Background(), Definition{ID: 1, JobKey: "system.scheduler.test", Timeout: time.Second}, "manual"); err != nil {
		t.Fatal(err)
	}
	runs.mu.Lock()
	status, token := runs.runs[1].Status, runs.activeRunCounts[1]
	runs.mu.Unlock()
	if status != RunStatusSkippedLocked || token != 1 {
		t.Fatalf("slot not retained: status=%s activeRunCount=%d", status, token)
	}
	close(exit)
	if err := <-finished; !errors.Is(err, context.Canceled) {
		t.Fatalf("execute = %v", err)
	}
	deadline := time.After(time.Second)
	for {
		runs.mu.Lock()
		status, token := runs.runs[0].Status, runs.activeRunCounts[1]
		runs.mu.Unlock()
		if status == RunStatusStopped && token == 0 {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("run was not stopped and slot released: status=%s activeRunCount=%d", status, token)
		case <-time.After(time.Millisecond):
		}
	}
}

func TestExecutorCrossReplicaStopPollsDurableStatusBeforeLeaseRenewal(t *testing.T) {
	r := NewRegistry()
	started := make(chan struct{})
	workerCanceled := make(chan struct{})
	exit := make(chan struct{})
	_ = r.Register(testJob{"system.scheduler.test", func(ctx context.Context, _ string) (string, error) {
		close(started)
		<-ctx.Done()
		close(workerCanceled)
		<-exit
		return "stopped", ctx.Err()
	}})
	runs := &testRuns{}
	leases := &testLease{}
	e := NewExecutor(r, leases, runs, time.Minute)
	e.stopPollInterval = time.Millisecond
	finished := make(chan error, 1)
	go func() {
		finished <- e.Execute(context.Background(), Definition{ID: 1, JobKey: "system.scheduler.test", Timeout: time.Minute, RetryCount: 1}, "manual")
	}()
	<-started

	// This executor represents another replica and has no local cancel registration for the run.
	remote := NewExecutor(r, &testLease{}, runs, 10*time.Millisecond)
	if stopped, err := remote.Stop(context.Background(), 100); err != nil || !stopped {
		t.Fatalf("cross-replica stop = %v, %v", stopped, err)
	}
	select {
	case <-workerCanceled:
	case <-time.After(time.Second):
		t.Fatal("worker context was not canceled after durable stop request")
	}
	if err := <-finished; !errors.Is(err, context.Canceled) {
		t.Fatalf("execute = %v", err)
	}
	runs.mu.Lock()
	runCount := len(runs.runs)
	activeToken := runs.activeRunCounts[1]
	runs.mu.Unlock()
	leases.mu.Lock()
	leaseHeld := leases.held
	leases.mu.Unlock()
	if runCount != 1 || activeToken != 1 || !leaseHeld {
		t.Fatalf("stopped worker must retain slot and lease without retry: runs=%d activeRunCount=%d leaseHeld=%v", runCount, activeToken, leaseHeld)
	}

	close(exit)
	deadline := time.After(time.Second)
	for {
		runs.mu.Lock()
		status := runs.runs[0].Status
		activeToken := runs.activeRunCounts[1]
		runs.mu.Unlock()
		if status == RunStatusStopped && activeToken == 0 {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("run was not finalized as stopped: status=%s activeRunCount=%d", status, activeToken)
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
