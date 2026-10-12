package scheduler

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Job is executable scheduler behavior registered by a stable business key.
type Job interface {
	Key() string
	Run(context.Context, string) (string, error)
}

// Registry holds all code-defined scheduler jobs.
type Registry struct {
	mu   sync.RWMutex
	jobs map[string]Job
}

func NewRegistry() *Registry {
	return &Registry{jobs: make(map[string]Job)}
}

func (r *Registry) Register(job Job) error {
	if job == nil {
		return fmt.Errorf("scheduler job is required")
	}
	key := strings.TrimSpace(job.Key())
	if key == "" || key != strings.ToLower(key) || strings.ContainsAny(key, " \t\n") {
		return fmt.Errorf("scheduler job key must be lowercase and non-empty: %q", key)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.jobs[key]; exists {
		return fmt.Errorf("scheduler job already registered: %s", key)
	}
	r.jobs[key] = job
	return nil
}

func (r *Registry) Get(key string) (Job, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	job, ok := r.jobs[key]
	return job, ok
}

func (r *Registry) Keys() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := make([]string, 0, len(r.jobs))
	for key := range r.jobs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
