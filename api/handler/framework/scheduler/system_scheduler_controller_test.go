package scheduler

import (
	"testing"
	"time"

	dalModel "gosilite/application/framework/dal/model"
)

func TestRunDTOUsesRetryCountAndTriggeredAt(t *testing.T) {
	triggeredAt := time.Date(2026, 10, 12, 13, 50, 0, 0, time.UTC)
	finishedAt := triggeredAt.Add(time.Second)
	run := &dalModel.SystemJobRun{
		ID:          1,
		JobID:       2,
		JobKey:      "system.scheduler.test",
		Trigger:     "manual",
		Attempt:     1,
		Status:      "succeeded",
		TriggeredAt: triggeredAt,
		FinishedAt:  &finishedAt,
		Output:      "done",
	}

	actual := runDTO(run)
	if actual.RetryCount != 0 {
		t.Fatalf("retry count = %d, want 0", actual.RetryCount)
	}
	if actual.TriggeredAt != triggeredAt.Format(time.RFC3339) {
		t.Fatalf("triggered at = %q, want %q", actual.TriggeredAt, triggeredAt.Format(time.RFC3339))
	}
	if actual.FinishedAt != finishedAt.Format(time.RFC3339) {
		t.Fatalf("finished at = %q, want %q", actual.FinishedAt, finishedAt.Format(time.RFC3339))
	}
}
