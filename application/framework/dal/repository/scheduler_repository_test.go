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

func TestDeleteRunLogsBeforeDeletesOnlyExpiredLogs(t *testing.T) {
	matcher := sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		if !strings.Contains(actualSQL, `DELETE FROM "system_job_run_log"`) || !strings.Contains(actualSQL, `created_at < $1`) {
			return fmt.Errorf("unexpected cleanup query: %s", actualSQL)
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
	cutoff := time.Now().Add(-30 * 24 * time.Hour)
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "system_job_run_log" WHERE created_at < \$1`).WithArgs(cutoff).WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	deleted, err := NewSchedulerRepository(gormDB).DeleteRunLogsBefore(context.Background(), cutoff)
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 3 {
		t.Fatalf("deleted = %d, want 3", deleted)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSaveJobUpdatePreservesActiveRunToken(t *testing.T) {
	matcher := sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		for _, column := range []string{"job_key", "job_name", "cron", "timeout_sec", "params_json", "retry_count", "status", "updated_at"} {
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
