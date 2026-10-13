package user

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	configService "gosilite/application/framework/service/config"
	"gosilite/conf"
	"gosilite/pkg/bcrypt"
)

type loginSecurityRepo struct {
	repository.ISystemUserRepository
	user *model.SystemUser
}

func (r *loginSecurityRepo) GetByAccount(_ context.Context, account string) (*model.SystemUser, error) {
	if r.user == nil || r.user.Account != account {
		return nil, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *loginSecurityRepo) GetByID(_ context.Context, id int64) (*model.SystemUser, error) {
	if r.user == nil || r.user.ID != id {
		return nil, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *loginSecurityRepo) RecordLoginFailure(_ context.Context, _ int64, threshold int32) (bool, error) {
	r.user.LoginFailureCount++
	if r.user.LoginFailureCount >= threshold {
		r.user.LoginPermanentlyLocked = true
	}
	return r.user.LoginPermanentlyLocked, nil
}

func (r *loginSecurityRepo) ResetLoginSecurity(_ context.Context, _ int64) error {
	r.user.LoginFailureCount = 0
	r.user.LoginPermanentlyLocked = false
	return nil
}

func (r *loginSecurityRepo) ExistsAccount(context.Context, string, int64) (bool, error) {
	return false, nil
}
func (r *loginSecurityRepo) Update(_ context.Context, user *model.SystemUser) error {
	r.user = user
	return nil
}
func (r *loginSecurityRepo) ReplaceRoles(context.Context, int64, []int64) error { return nil }

type loginSecurityConfigRepo struct {
	repository.IConfigRepository
	values map[string]string
}

func (r *loginSecurityConfigRepo) GetByKey(_ context.Context, key string) (*model.SystemConfig, error) {
	value, ok := r.values[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return &model.SystemConfig{ConfigKey: key, ConfigValue: value, Status: 0}, nil
}

func newLoginSecurityService(t *testing.T) (*Service, *loginSecurityRepo, *memCache) {
	t.Helper()
	hash, err := bcrypt.HashPassword("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	userRepo := &loginSecurityRepo{user: &model.SystemUser{ID: 1, Account: "alice", State: "enabled", Password: hash}}
	cache := newMemCache()
	previous := configService.SystemConfigSVC
	configService.SystemConfigSVC = configService.NewService(&loginSecurityConfigRepo{values: map[string]string{
		"sys.login.max_failures":           "5",
		"sys.login.failure_window_minutes": "15",
		"sys.login.lock_minutes":           "30",
	}})
	t.Cleanup(func() { configService.SystemConfigSVC = previous })
	return &Service{
		repo: userRepo, cache: cache, sessions: NewSessionStore(cache),
		loginSecurityDefaults: conf.LoginSecurity{MaxFailures: 5, FailureWindowMinutes: 15, LockMinutes: 30},
	}, userRepo, cache
}

func TestLoginTemporaryLockAndSuccessfulReset(t *testing.T) {
	ctx := context.Background()
	svc, repo, cache := newLoginSecurityService(t)

	for i := 0; i < 5; i++ {
		if _, err := svc.login(ctx, "Alice", "wrong-password"); !errors.Is(err, errInvalidCredentials) {
			t.Fatalf("attempt %d: want unified credential error, got %v", i+1, err)
		}
	}
	if exists, _ := cache.Exists(ctx, svc.loginLockKey("alice")).Result(); exists != 1 {
		t.Fatal("fifth failure must create temporary lock")
	}
	if repo.user.LoginFailureCount != 5 {
		t.Fatalf("failure count = %d, want 5", repo.user.LoginFailureCount)
	}
	if _, err := svc.login(ctx, "alice", "correct-password"); !errors.Is(err, errTemporaryLocked) {
		t.Fatalf("temporary lock must reject correct password, got %v", err)
	}
	if repo.user.LoginFailureCount != 5 {
		t.Fatalf("correct password during temporary lock must not count as a failure, got %d", repo.user.LoginFailureCount)
	}
	if _, err := cache.Delete(ctx, svc.loginLockKey("alice")).Result(); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.login(ctx, "alice", "correct-password"); err != nil {
		t.Fatalf("successful login after temporary lock expires: %v", err)
	}
	if repo.user.LoginFailureCount != 0 || repo.user.LoginPermanentlyLocked {
		t.Fatalf("successful login must reset security state: %+v", repo.user)
	}
}

func TestLoginPermanentlyLocksAfterTenFailures(t *testing.T) {
	ctx := context.Background()
	svc, repo, cache := newLoginSecurityService(t)

	for i := 0; i < 10; i++ {
		_, _ = cache.Delete(ctx, svc.loginLockKey("alice")).Result()
		if _, err := svc.login(ctx, "alice", "wrong-password"); !errors.Is(err, errInvalidCredentials) {
			t.Fatalf("attempt %d: %v", i+1, err)
		}
	}
	if !repo.user.LoginPermanentlyLocked || repo.user.LoginFailureCount != 10 {
		t.Fatalf("expected permanent lock after ten failures: %+v", repo.user)
	}
	if _, err := svc.login(ctx, "alice", "correct-password"); !errors.Is(err, errInvalidCredentials) {
		t.Fatalf("permanent lock must retain unified error, got %v", err)
	}
}

func TestAdminPasswordResetClearsPermanentLock(t *testing.T) {
	ctx := context.Background()
	svc, repo, cache := newLoginSecurityService(t)
	repo.user.LoginFailureCount = 10
	repo.user.LoginPermanentlyLocked = true
	_, _ = cache.Set(ctx, svc.loginLockKey("alice"), "1", 0).Result()

	update := *repo.user
	if err := svc.AdminUpdate(ctx, &update, nil, nil, "new-password"); err != nil {
		t.Fatalf("admin password reset: %v", err)
	}
	if repo.user.LoginFailureCount != 0 || repo.user.LoginPermanentlyLocked {
		t.Fatalf("reset password must unlock: %+v", repo.user)
	}
	if exists, _ := cache.Exists(ctx, svc.loginLockKey("alice")).Result(); exists != 0 {
		t.Fatal("password reset must clear temporary lock")
	}
}
