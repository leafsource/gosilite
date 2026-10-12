package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	"gosilite/application/framework/jobs"
	core "gosilite/application/framework/scheduler"
)

func TestTestJobKeyIsStable(t *testing.T) {
	if TestJobKey != "system.scheduler.test" {
		t.Fatalf("test job key = %q, want %q", TestJobKey, "system.scheduler.test")
	}
}

func TestValidateRequiresSixFieldCron(t *testing.T) {
	registry := core.NewRegistry()
	if err := registry.Register(jobs.SchedulerTest{}); err != nil {
		t.Fatal(err)
	}
	service := &Service{registry: registry}

	job := &model.SystemJob{
		JobKey:     TestJobKey,
		JobName:    "test",
		TimeoutSec: 1,
		ParamsJSON: "{}",
		Cron:       "0 0 * * * *",
	}
	if err := service.validate(job); err != nil {
		t.Fatalf("six-field cron must validate: %v", err)
	}

	job.Cron = "0 * * * *"
	if err := service.validate(job); err == nil {
		t.Fatal("five-field cron must be rejected")
	}
}

func TestServiceCronUsesSecondsParser(t *testing.T) {
	service := &Service{cron: cron.New(cron.WithSeconds())}
	if _, err := service.cron.AddFunc("0 0 * * * *", func() {}); err != nil {
		t.Fatalf("six-field cron must schedule: %v", err)
	}
}

type retryRepository struct {
	repository.SchedulerRepository
	run     *model.SystemJobRun
	created *core.Run
}

type retryJob struct {
	key string
	run func(context.Context, string) (string, error)
}

func (j retryJob) Key() string { return j.key }

func (j retryJob) Run(ctx context.Context, params string) (string, error) {
	return j.run(ctx, params)
}

type retryLease struct{}

func (retryLease) Acquire(context.Context, string, string, time.Duration) (bool, error) {
	return true, nil
}

func (retryLease) Renew(context.Context, string, string, time.Duration) (bool, error) {
	return true, nil
}

func (retryLease) Release(context.Context, string, string) (bool, error) {
	return true, nil
}

func (r *retryRepository) GetRun(context.Context, int64) (*model.SystemJobRun, error) {
	return r.run, nil
}

func (r *retryRepository) ClaimActiveRun(context.Context, int64, string) (bool, error) {
	return true, nil
}

func (r *retryRepository) ClearActiveRun(context.Context, int64, string) (bool, error) {
	return true, nil
}

func (r *retryRepository) CreateRun(_ context.Context, run *core.Run) (int64, error) {
	r.created = run
	return 99, nil
}

func (r *retryRepository) FinishRun(context.Context, *core.Run) error { return nil }

func (r *retryRepository) IsRunStopping(context.Context, int64) (bool, error) { return false, nil }

func TestRetryStoppedRunUsesImmutableSnapshot(t *testing.T) {
	repo := &retryRepository{run: &model.SystemJobRun{
		ID:            12,
		JobID:         34,
		JobKey:        TestJobKey,
		JobName:       "snapshot name",
		ParamsJSON:    `{"snapshot":true}`,
		TimeoutSec:    7,
		ConfigVersion: 9,
		Status:        core.RunStatusStopped,
	}}
	registry := core.NewRegistry()
	var receivedParams string
	if err := registry.Register(retryJob{key: TestJobKey, run: func(_ context.Context, params string) (string, error) {
		receivedParams = params
		return "ok", nil
	}}); err != nil {
		t.Fatal(err)
	}
	service := &Service{
		repo:     repo,
		registry: registry,
		executor: core.NewExecutor(registry, retryLease{}, repo, time.Minute),
	}

	if err := service.Retry(context.Background(), repo.run.ID); err != nil {
		t.Fatalf("retry stopped run: %v", err)
	}
	if receivedParams != repo.run.ParamsJSON {
		t.Fatalf("job params = %q, want snapshot %q", receivedParams, repo.run.ParamsJSON)
	}
	if repo.created == nil {
		t.Fatal("retry did not create a run")
	}
	if repo.created.JobID != repo.run.JobID || repo.created.JobName != repo.run.JobName || repo.created.ParamsJSON != repo.run.ParamsJSON || repo.created.TimeoutSec != repo.run.TimeoutSec || repo.created.ConfigVersion != repo.run.ConfigVersion {
		t.Fatalf("retry definition = %+v, want immutable snapshot %+v", repo.created, repo.run)
	}
}
