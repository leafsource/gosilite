package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	"gosilite/application/framework/jobs"
	core "gosilite/application/framework/scheduler"
	"gosilite/conf"
	"gosilite/types/jobkeys"
)

var SystemSchedulerSVC *Service

const (
	TestJobKey       = jobkeys.SchedulerTest
	LogCleanupJobKey = jobkeys.SchedulerLogCleanup
)

type Service struct {
	repo     repository.SchedulerRepository
	registry *core.Registry
	executor *core.Executor
	logs     *core.LogSink
	cron     *cron.Cron
	mu       sync.Mutex
	entries  map[int64]cron.EntryID
}

func NewService(repo repository.SchedulerRepository, cacheLease core.LeaseStore) *Service {
	return NewServiceWithRetention(repo, cacheLease, conf.GetConf().Scheduler.LogRetentionDays)
}

func NewServiceWithRetention(repo repository.SchedulerRepository, cacheLease core.LeaseStore, retentionDays int) *Service {
	if retentionDays < 1 {
		retentionDays = 30
	}
	r := core.NewRegistry()
	_ = r.Register(jobs.SchedulerTest{})
	_ = r.Register(jobs.SchedulerLogCleanup{Repo: repo, RetentionDays: retentionDays})
	logs := core.NewLogSink(repo)
	executor := core.NewExecutor(r, cacheLease, repo, time.Minute)
	executor.SetLogSink(logs)
	return &Service{repo: repo, registry: r, executor: executor, logs: logs, cron: cron.New(cron.WithSeconds(), cron.WithLogger(core.CronLogger{}), cron.WithChain(cron.Recover(core.CronLogger{}))), entries: map[int64]cron.EntryID{}}
}
func InitService(repo repository.SchedulerRepository, lease core.LeaseStore) *Service {
	SystemSchedulerSVC = NewService(repo, lease)
	SystemSchedulerSVC.Load(context.Background())
	SystemSchedulerSVC.cron.Start()
	return SystemSchedulerSVC
}
func (s *Service) Keys() []string { return s.registry.Keys() }
func (s *Service) List(ctx context.Context, p, n int) ([]*model.SystemJob, int64, error) {
	return s.repo.ListJobs(ctx, p, n)
}
func (s *Service) Get(ctx context.Context, id int64) (*model.SystemJob, error) {
	return s.repo.GetJob(ctx, id)
}
func (s *Service) Save(ctx context.Context, m *model.SystemJob) error {
	if m.MaxInstances == 0 {
		m.MaxInstances = 1
	}
	if err := s.validate(m); err != nil {
		return err
	}
	previous, err := s.previous(ctx, m.ID)
	if err != nil {
		return err
	}
	if err := s.repo.SaveJob(ctx, m); err != nil {
		return err
	}
	if err := s.refresh(m); err != nil {
		if previous != nil {
			_ = s.repo.SaveJob(ctx, previous)
			_ = s.refresh(previous)
		} else {
			_ = s.repo.DeleteJob(ctx, m.ID)
		}
		return err
	}
	return nil
}
func (s *Service) SetStatus(ctx context.Context, id int64, status int16) error {
	if status != 0 && status != 1 {
		return fmt.Errorf("invalid status")
	}
	m, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	previous := *m
	m.Status = status
	if err := s.refresh(m); err != nil {
		return err
	}
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		_ = s.refresh(&previous)
		return err
	}
	return nil
}
func (s *Service) Delete(ctx context.Context, id int64) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}
	if err := s.repo.DeleteJob(ctx, id); err != nil {
		return err
	}
	s.mu.Lock()
	if entry, ok := s.entries[id]; ok {
		s.cron.Remove(entry)
		delete(s.entries, id)
	}
	s.mu.Unlock()
	return nil
}
func (s *Service) Run(ctx context.Context, id int64) error {
	m, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	s.executeAsync(toDef(m))
	return nil
}

func (s *Service) Retry(ctx context.Context, runID int64) error {
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if run.Status == core.RunStatusRunning || run.Status == core.RunStatusStopping {
		return fmt.Errorf("scheduler run is still active")
	}
	s.executeAsync(core.Definition{ID: run.JobID, JobKey: run.JobKey, JobName: run.JobName, ParamsJSON: run.ParamsJSON, Timeout: time.Duration(run.TimeoutSec) * time.Second, ConfigVersion: run.ConfigVersion})
	return nil
}

func (s *Service) executeAsync(def core.Definition) {
	go func() {
		_ = s.executor.Execute(context.Background(), def, "manual")
	}()
}

func (s *Service) Stop(ctx context.Context, runID int64) error {
	stopped, err := s.executor.Stop(ctx, runID)
	if err != nil {
		return err
	}
	if !stopped {
		return fmt.Errorf("scheduler run is not running")
	}
	return nil
}

func (s *Service) Runs(ctx context.Context, jobID *int64, p, n int) ([]*model.SystemJobRun, int64, error) {
	return s.repo.ListRuns(ctx, jobID, p, n)
}
func (s *Service) RunLogs(ctx context.Context, runID, afterID int64, p, n int) ([]*model.SystemJobRunLog, int64, error) {
	return s.repo.ListRunLogs(ctx, runID, afterID, p, n)
}
func (s *Service) SubscribeRunLogs(runID int64) (<-chan *core.LogEntry, func()) {
	return s.logs.Subscribe(runID)
}
func (s *Service) Load(ctx context.Context) {
	const pageSize = 200
	for page := 1; ; page++ {
		jobs, total, err := s.repo.ListJobs(ctx, page, pageSize)
		if err != nil {
			return
		}
		for _, m := range jobs {
			_ = s.refresh(m)
		}
		if int64(page*pageSize) >= total || len(jobs) == 0 {
			return
		}
	}
}

func (s *Service) previous(ctx context.Context, id int64) (*model.SystemJob, error) {
	if id == 0 {
		return nil, nil
	}
	m, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	copy := *m
	return &copy, nil
}
func (s *Service) refresh(m *model.SystemJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.entries[m.ID]; ok {
		s.cron.Remove(id)
		delete(s.entries, m.ID)
	}
	if m.Status == 0 {
		return nil
	}
	id, err := s.cron.AddFunc(m.Cron, func() { _ = s.executor.Execute(context.Background(), toDef(m), "cron") })
	if err == nil {
		s.entries[m.ID] = id
	}
	return err
}
func (s *Service) validate(m *model.SystemJob) error {
	m.JobKey = strings.TrimSpace(m.JobKey)
	if _, ok := s.registry.Get(m.JobKey); !ok {
		return fmt.Errorf("scheduler job not registered: %s", m.JobKey)
	}
	if strings.TrimSpace(m.JobName) == "" || strings.TrimSpace(m.Cron) == "" || m.TimeoutSec <= 0 || m.RetryCount < 0 || m.MaxInstances < 1 {
		return fmt.Errorf("invalid scheduler job configuration")
	}
	if !json.Valid([]byte(m.ParamsJSON)) {
		return fmt.Errorf("paramsJson must be valid JSON")
	}
	_, err := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(m.Cron)
	return err
}
func toDef(m *model.SystemJob) core.Definition {
	return core.Definition{ID: m.ID, JobKey: m.JobKey, JobName: m.JobName, Timeout: time.Duration(m.TimeoutSec) * time.Second, RetryCount: int(m.RetryCount), ParamsJSON: m.ParamsJSON, ConfigVersion: m.ConfigVersion}
}
