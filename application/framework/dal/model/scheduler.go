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
	MaxInstances   int32     `gorm:"column:max_instances"`
	ActiveRunCount int32     `gorm:"column:active_run_count"`
	ActiveRunToken string    `gorm:"column:active_run_token"`
	ConfigVersion  int32     `gorm:"column:config_version"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (*SystemJob) TableName() string { return "system_job" }

type SystemJobRun struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement:true"`
	JobID         int64      `gorm:"column:job_id"`
	JobKey        string     `gorm:"column:job_key"`
	Trigger       string     `gorm:"column:trigger_type"`
	Attempt       int32      `gorm:"column:attempt"`
	Status        string     `gorm:"column:status"`
	TriggeredAt   time.Time  `gorm:"column:triggered_at"`
	FinishedAt    *time.Time `gorm:"column:finished_at"`
	Output        string     `gorm:"column:output"`
	Error         string     `gorm:"column:error_message"`
	JobName       string     `gorm:"column:job_name"`
	ParamsJSON    string     `gorm:"column:params_json"`
	TimeoutSec    int32      `gorm:"column:timeout_sec"`
	ConfigVersion int32      `gorm:"column:config_version"`
}

func (*SystemJobRun) TableName() string { return "system_job_run" }

type SystemJobRunLog struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement:true"`
	RunID     int64     `gorm:"column:run_id"`
	JobKey    string    `gorm:"column:job_key"`
	Level     string    `gorm:"column:level"`
	Stage     string    `gorm:"column:stage"`
	Message   string    `gorm:"column:message"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (*SystemJobRunLog) TableName() string { return "system_job_run_log" }
