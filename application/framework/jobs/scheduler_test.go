package jobs

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSchedulerTestStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := (SchedulerTest{}).Run(ctx, "")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestSchedulerTestCompletesAfterTenSeconds(t *testing.T) {
	if testing.Short() {
		t.Skip("scheduler demo takes ten seconds")
	}

	started := time.Now()
	result, err := (SchedulerTest{}).Run(context.Background(), "")
	if err != nil {
		t.Fatalf("run scheduler test: %v", err)
	}
	if result != "scheduler test completed after 10 seconds" {
		t.Fatalf("unexpected result: %q", result)
	}
	if elapsed := time.Since(started); elapsed < 10*time.Second {
		t.Fatalf("expected at least ten seconds, got %s", elapsed)
	}
}
