package tasks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/leafsource/gosilite/dbx"
	"github.com/leafsource/gosilite/httpx"
)

// Repository is the persistence port consumed by Service.
type Repository interface {
	Create(ctx context.Context, t *Task) error
	Get(ctx context.Context, id string) (*Task, error)
	List(ctx context.Context, filter ListFilter) ([]Task, error)
	Update(ctx context.Context, t *Task) error
	Delete(ctx context.Context, id string) error
}

// SQLRepository implements Repository with database/sql + dialect placeholders.
type SQLRepository struct {
	db      *dbx.DB
	dialect dbx.Dialect
}

// NewSQLRepository builds a repository on top of dbx.DB.
func NewSQLRepository(db *dbx.DB) *SQLRepository {
	return &SQLRepository{db: db, dialect: db.Dialect}
}

func (r *SQLRepository) Create(ctx context.Context, t *Task) error {
	const q = `INSERT INTO tasks (id, title, description, status, created_at, updated_at) VALUES (%s, %s, %s, %s, %s, %s)`
	query := bindPlaceholders(q, r.dialect, 6)
	_, err := r.db.SQL.ExecContext(ctx, query,
		t.ID, t.Title, t.Description, string(t.Status), t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return httpx.Internal("insert task", err)
	}
	return nil
}

func (r *SQLRepository) Get(ctx context.Context, id string) (*Task, error) {
	const q = `SELECT id, title, description, status, created_at, updated_at FROM tasks WHERE id = %s`
	query := bindPlaceholders(q, r.dialect, 1)
	row := r.db.SQL.QueryRowContext(ctx, query, id)
	var t Task
	var status string
	if err := row.Scan(&t.ID, &t.Title, &t.Description, &status, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, httpx.NotFound("task not found")
		}
		return nil, httpx.Internal("get task", err)
	}
	t.Status = Status(status)
	return &t, nil
}

func (r *SQLRepository) List(ctx context.Context, filter ListFilter) ([]Task, error) {
	base := `SELECT id, title, description, status, created_at, updated_at FROM tasks`
	var args []any
	if filter.Status != "" {
		base += ` WHERE status = %s`
		args = append(args, string(filter.Status))
	}
	base += ` ORDER BY created_at DESC, id`
	query := bindPlaceholders(base, r.dialect, len(args))
	rows, err := r.db.SQL.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, httpx.Internal("list tasks", err)
	}
	defer rows.Close()
	out := []Task{}
	for rows.Next() {
		var t Task
		var status string
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, httpx.Internal("scan task", err)
		}
		t.Status = Status(status)
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, httpx.Internal("list tasks", err)
	}
	return out, nil
}

func (r *SQLRepository) Update(ctx context.Context, t *Task) error {
	const q = `UPDATE tasks SET title = %s, description = %s, status = %s, updated_at = %s WHERE id = %s`
	query := bindPlaceholders(q, r.dialect, 5)
	res, err := r.db.SQL.ExecContext(ctx, query,
		t.Title, t.Description, string(t.Status), t.UpdatedAt, t.ID)
	if err != nil {
		return httpx.Internal("update task", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return httpx.Internal("update task", err)
	}
	if n == 0 {
		return httpx.NotFound("task not found")
	}
	return nil
}

func (r *SQLRepository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM tasks WHERE id = %s`
	query := bindPlaceholders(q, r.dialect, 1)
	res, err := r.db.SQL.ExecContext(ctx, query, id)
	if err != nil {
		return httpx.Internal("delete task", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return httpx.Internal("delete task", err)
	}
	if n == 0 {
		return httpx.NotFound("task not found")
	}
	return nil
}

// bindPlaceholders rewrites %s markers into dialect placeholders in order.
// The format string must contain exactly n placeholder markers.
func bindPlaceholders(format string, d dbx.Dialect, n int) string {
	parts := strings.Split(format, "%s")
	if len(parts)-1 != n {
		panic(fmt.Sprintf("tasks: placeholder count %d != %d", len(parts)-1, n))
	}
	var b strings.Builder
	for i, part := range parts {
		b.WriteString(part)
		if i < n {
			b.WriteString(d.Placeholder(i + 1))
		}
	}
	return b.String()
}
