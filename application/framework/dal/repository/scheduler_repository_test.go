package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gosilite/application/framework/dal/model"
)

func TestSaveJobUpdatePreservesActiveRunToken(t *testing.T) {
	matcher := sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		for _, column := range []string{"job_key", "job_name", "cron", "timezone", "timeout_sec", "params_json", "retry_count", "status", "updated_at"} {
			if !strings.Contains(actualSQL, `"`+column+`"`) {
				return fmt.Errorf("update does not include %s: %s", column, actualSQL)
			}
		}
		for _, column := range []string{"active_run_token", "created_at"} {
			if strings.Contains(actualSQL, `"`+column+`"`) {
				return fmt.Errorf("update must not modify %s: %s", column, actualSQL)
			}
		}
		return nil
	})
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	repo := NewSchedulerRepository(gormDB)
	mock.ExpectBegin()
	mock.ExpectExec("scheduler job update").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	job := &model.SystemJob{
		ID:             1,
		JobKey:         "system.scheduler.test",
		JobName:        "renamed",
		Cron:           "0 0 * * * *",
		Timezone:       "Asia/Shanghai",
		TimeoutSec:     30,
		ParamsJSON:     `{"key":"value"}`,
		RetryCount:     2,
		Status:         1,
		ActiveRunToken: "must-not-be-written",
		CreatedAt:      time.Now().Add(-time.Hour),
	}
	if err := repo.SaveJob(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
