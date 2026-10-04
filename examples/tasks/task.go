package tasks

import (
	"strings"
	"time"
)

// Status is the lifecycle state of a task.
type Status string

const (
	StatusTodo  Status = "todo"
	StatusDoing Status = "doing"
	StatusDone  Status = "done"
)

// Task is the domain entity.
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ParseStatus validates and normalizes a status string.
func ParseStatus(s string) (Status, bool) {
	switch Status(strings.ToLower(strings.TrimSpace(s))) {
	case StatusTodo:
		return StatusTodo, true
	case StatusDoing:
		return StatusDoing, true
	case StatusDone:
		return StatusDone, true
	default:
		return "", false
	}
}

// ValidStatus reports whether s is a known status.
func ValidStatus(s Status) bool {
	switch s {
	case StatusTodo, StatusDoing, StatusDone:
		return true
	default:
		return false
	}
}

// ListFilter selects tasks in list queries.
type ListFilter struct {
	Status Status
}
