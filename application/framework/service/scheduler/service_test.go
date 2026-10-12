package scheduler

import (
	"testing"

	"github.com/robfig/cron/v3"
	"gosilite/application/framework/dal/model"
	core "gosilite/application/framework/scheduler"
)

func TestValidateRequiresSixFieldCron(t *testing.T) {
	registry := core.NewRegistry()
	if err := registry.Register(testJob{}); err != nil {
		t.Fatal(err)
	}
	service := &Service{registry: registry}

	job := &model.SystemJob{
		JobKey:     TestJobKey,
		JobName:    "test",
		Timezone:   "Asia/Shanghai",
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
