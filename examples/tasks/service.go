package tasks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/leafsource/gosilite/httpx"
)

// Service holds task business rules. HTTP and MCP both call this layer.
type Service struct {
	repo Repository
	now  func() time.Time
}

// NewService injects a repository and clock.
func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// WithClock overrides the clock (tests).
func (s *Service) WithClock(now func() time.Time) *Service {
	s.now = now
	return s
}

// Create validates input and inserts a task.
func (s *Service) Create(ctx context.Context, in CreateTaskRequest) (Task, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return Task{}, httpx.Invalid("title is required")
	}
	status := StatusTodo
	if strings.TrimSpace(in.Status) != "" {
		parsed, ok := ParseStatus(in.Status)
		if !ok {
			return Task{}, httpx.WithDetail(httpx.Invalid("status must be todo|doing|done"), "status", in.Status)
		}
		status = parsed
	}
	now := s.now().UTC()
	task := Task{
		ID:          newID(),
		Title:       title,
		Description: strings.TrimSpace(in.Description),
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, &task); err != nil {
		return Task{}, err
	}
	return task, nil
}

// Get returns one task or not_found.
func (s *Service) Get(ctx context.Context, id string) (Task, error) {
	if strings.TrimSpace(id) == "" {
		return Task{}, httpx.Invalid("id is required")
	}
	t, err := s.repo.Get(ctx, id)
	if err != nil {
		return Task{}, err
	}
	return *t, nil
}

// List returns tasks, optionally filtered by status.
func (s *Service) List(ctx context.Context, q ListTasksQuery) ([]Task, error) {
	filter := ListFilter{}
	if strings.TrimSpace(q.Status) != "" {
		st, ok := ParseStatus(q.Status)
		if !ok {
			return nil, httpx.WithDetail(httpx.Invalid("status must be todo|doing|done"), "status", q.Status)
		}
		filter.Status = st
	}
	return s.repo.List(ctx, filter)
}

// Update applies a partial update.
func (s *Service) Update(ctx context.Context, id string, in UpdateTaskRequest) (Task, error) {
	if strings.TrimSpace(id) == "" {
		return Task{}, httpx.Invalid("id is required")
	}
	current, err := s.repo.Get(ctx, id)
	if err != nil {
		return Task{}, err
	}
	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return Task{}, httpx.Invalid("title cannot be empty")
		}
		current.Title = title
	}
	if in.Description != nil {
		current.Description = strings.TrimSpace(*in.Description)
	}
	if in.Status != nil {
		st, ok := ParseStatus(*in.Status)
		if !ok {
			return Task{}, httpx.WithDetail(httpx.Invalid("status must be todo|doing|done"), "status", *in.Status)
		}
		current.Status = st
	}
	current.UpdatedAt = s.now().UTC()
	if err := s.repo.Update(ctx, current); err != nil {
		return Task{}, err
	}
	return *current, nil
}

// Delete removes a task.
func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return httpx.Invalid("id is required")
	}
	return s.repo.Delete(ctx, id)
}

func newID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
