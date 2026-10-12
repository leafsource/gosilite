package scheduler

import (
	"context"
	"time"
)

// LeaseStore provides the atomic Redis operations required by job leases.
type LeaseStore interface {
	Acquire(ctx context.Context, key, token string, ttl time.Duration) (bool, error)
	Renew(ctx context.Context, key, token string, ttl time.Duration) (bool, error)
	Release(ctx context.Context, key, token string) (bool, error)
}

type Lease struct {
	store LeaseStore
	key   string
	token string
	ttl   time.Duration
}

func AcquireLease(ctx context.Context, store LeaseStore, jobKey, token string, ttl time.Duration) (*Lease, bool, error) {
	ok, err := store.Acquire(ctx, "scheduler:lease:"+jobKey, token, ttl)
	if err != nil || !ok {
		return nil, ok, err
	}
	return &Lease{store: store, key: "scheduler:lease:" + jobKey, token: token, ttl: ttl}, true, nil
}

func (l *Lease) Renew(ctx context.Context) (bool, error) {
	return l.store.Renew(ctx, l.key, l.token, l.ttl)
}

func (l *Lease) Release(ctx context.Context) (bool, error) {
	return l.store.Release(ctx, l.key, l.token)
}
