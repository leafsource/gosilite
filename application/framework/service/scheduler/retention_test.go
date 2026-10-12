package scheduler

import (
	"context"
	"testing"
	"time"
)

type cleanupStore struct {
	before  time.Time
	deleted int64
}

func (s *cleanupStore) DeleteRunLogsBefore(_ context.Context, before time.Time) (int64, error) {
	s.before = before
	return s.deleted, nil
}

func TestLogCleanupJobDeletesExpiredLogsAndReportsCount(t *testing.T) {
	store := &cleanupStore{deleted: 3}
	job := logCleanupJob{repo: store, retentionDays: 14}
	before := time.Now().Add(-14 * 24 * time.Hour)

	output, err := job.Run(context.Background(), "{}")
	if err != nil {
		t.Fatal(err)
	}
	if output != "deleted 3 execution logs older than 14 days" {
		t.Fatalf("output = %q", output)
	}
	if store.before.Before(before.Add(-2*time.Second)) || store.before.After(time.Now().Add(-14*24*time.Hour).Add(2*time.Second)) {
		t.Fatalf("cutoff = %v, want approximately %v", store.before, before)
	}
}
