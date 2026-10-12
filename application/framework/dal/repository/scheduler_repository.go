package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/scheduler"
)

type SchedulerRepository interface {
	ListJobs(context.Context, int, int) ([]*model.SystemJob, int64, error)
	GetJob(context.Context, int64) (*model.SystemJob, error)
	SaveJob(context.Context, *model.SystemJob) error
	DeleteJob(context.Context, int64) error
	UpdateStatus(context.Context, int64, int16) error
	ClaimActiveRun(context.Context, int64, string) (bool, error)
	ClearActiveRun(context.Context, int64, string) (bool, error)
	CreateRun(context.Context, *scheduler.Run) (int64, error)
	FinishRun(context.Context, *scheduler.Run) error
	ListRuns(context.Context, int64, int, int) ([]*model.SystemJobRun, int64, error)
}

type schedulerRepository struct{ db *gorm.DB }

func NewSchedulerRepository(db *gorm.DB) SchedulerRepository { return &schedulerRepository{db} }

func (r *schedulerRepository) ListJobs(c context.Context, p, s int) ([]*model.SystemJob, int64, error) {
	if p < 1 {
		p = 1
	}
	if s < 1 || s > 200 {
		s = 10
	}
	var out []*model.SystemJob
	var n int64
	q := r.db.WithContext(c).Model(&model.SystemJob{})
	if e := q.Count(&n).Error; e != nil {
		return nil, 0, e
	}
	e := q.Order("id DESC").Offset((p - 1) * s).Limit(s).Find(&out).Error
	return out, n, e
}

func (r *schedulerRepository) GetJob(c context.Context, id int64) (*model.SystemJob, error) {
	var m model.SystemJob
	e := r.db.WithContext(c).First(&m, id).Error
	return &m, e
}

func (r *schedulerRepository) SaveJob(c context.Context, m *model.SystemJob) error {
	if m.ID == 0 {
		return r.db.WithContext(c).Create(m).Error
	}
	now := time.Now()
	result := r.db.WithContext(c).Model(&model.SystemJob{}).Where("id = ?", m.ID).Updates(map[string]any{
		"job_key":     m.JobKey,
		"job_name":    m.JobName,
		"cron":        m.Cron,
		"timeout_sec": m.TimeoutSec,
		"params_json": m.ParamsJSON,
		"retry_count": m.RetryCount,
		"status":      m.Status,
		"updated_at":  now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	m.UpdatedAt = now
	return nil
}

func (r *schedulerRepository) DeleteJob(c context.Context, id int64) error {
	result := r.db.WithContext(c).Delete(&model.SystemJob{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *schedulerRepository) UpdateStatus(c context.Context, id int64, status int16) error {
	result := r.db.WithContext(c).Model(&model.SystemJob{}).Where("id = ?", id).Updates(map[string]any{"status": status, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *schedulerRepository) ClaimActiveRun(c context.Context, id int64, token string) (bool, error) {
	result := r.db.WithContext(c).Model(&model.SystemJob{}).Where("id = ? AND active_run_token = ?", id, "").Updates(map[string]any{"active_run_token": token, "updated_at": time.Now()})
	return result.RowsAffected == 1, result.Error
}

func (r *schedulerRepository) ClearActiveRun(c context.Context, id int64, token string) (bool, error) {
	result := r.db.WithContext(c).Model(&model.SystemJob{}).Where("id = ? AND active_run_token = ?", id, token).Updates(map[string]any{"active_run_token": "", "updated_at": time.Now()})
	return result.RowsAffected == 1, result.Error
}

func (r *schedulerRepository) CreateRun(c context.Context, x *scheduler.Run) (int64, error) {
	m := &model.SystemJobRun{JobID: x.JobID, JobKey: x.JobKey, Trigger: x.Trigger, Attempt: int32(x.Attempt), Status: x.Status, TriggeredAt: x.TriggeredAt}
	if err := r.db.WithContext(c).Create(m).Error; err != nil {
		return 0, err
	}
	return m.ID, nil
}

func (r *schedulerRepository) FinishRun(c context.Context, x *scheduler.Run) error {
	result := r.db.WithContext(c).Model(&model.SystemJobRun{}).Where("id = ?", x.ID).Updates(map[string]any{
		"status":        x.Status,
		"finished_at":   x.FinishedAt,
		"output":        x.Output,
		"error_message": x.Error,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *schedulerRepository) ListRuns(c context.Context, id int64, p, s int) ([]*model.SystemJobRun, int64, error) {
	if p < 1 {
		p = 1
	}
	if s < 1 || s > 200 {
		s = 10
	}
	var out []*model.SystemJobRun
	var n int64
	q := r.db.WithContext(c).Model(&model.SystemJobRun{}).Where("job_id = ?", id)
	if e := q.Count(&n).Error; e != nil {
		return nil, 0, e
	}
	e := q.Order("id DESC").Offset((p - 1) * s).Limit(s).Find(&out).Error
	return out, n, e
}
