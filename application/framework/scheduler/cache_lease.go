package scheduler

import (
	"context"
	"time"

	"gosilite/application/infra/cache"
)

type CacheLeaseStore struct{ cache cache.LeaseOperations }

func NewCacheLeaseStore(client cache.CacheClient) *CacheLeaseStore {
	return &CacheLeaseStore{cache: client.(cache.LeaseOperations)}
}
func (s *CacheLeaseStore) Acquire(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	return s.cache.AcquireLease(ctx, key, token, ttl)
}
func (s *CacheLeaseStore) Renew(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	return s.cache.RenewLease(ctx, key, token, ttl)
}
func (s *CacheLeaseStore) Release(ctx context.Context, key, token string) (bool, error) {
	return s.cache.ReleaseLease(ctx, key, token)
}
