package model

import "time"

type SystemJob struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement:true"`
	JobKey         string    `gorm:"column:job_key"`
	JobName        string    `gorm:"column:job_name"`
	Cron           string    `gorm:"column:cron"`
	TimeoutSec     int32     `gorm:"column:timeout_sec"`
	ParamsJSON     string    `gorm:"column:params_json"`
	RetryCount     int32     `gorm:"column:retry_count"`
	Status         int16     `gorm:"column:status"`
	ActiveRunToken string    `gorm:"column:active_run_token"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (*SystemJob) TableName() string { return "system_job" }

type SystemJobRun struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement:true"`
	JobID       int64      `gorm:"column:job_id"`
	JobKey      string     `gorm:"column:job_key"`
	Trigger     string     `gorm:"column:trigger_type"`
	Attempt     int32      `gorm:"column:attempt"`
	Status      string     `gorm:"column:status"`
	TriggeredAt time.Time  `gorm:"column:triggered_at"`
	FinishedAt  *time.Time `gorm:"column:finished_at"`
	Output      string     `gorm:"column:output"`
	Error       string     `gorm:"column:error_message"`
}

func (*SystemJobRun) TableName() string { return "system_job_run" }
