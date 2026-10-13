package user

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"

	"gosilite/application/framework/dal/model"
	configService "gosilite/application/framework/service/config"
	"gosilite/application/infra/cache"
	"gosilite/pkg/bcrypt"
	"gosilite/types/consts"
)

const permanentLoginFailureThreshold int32 = 10

var (
	errInvalidCredentials = errors.New("invalid credentials")
	errTemporaryLocked    = errors.New("temporarily locked")
)

type loginSecurityRule struct {
	maxFailures   int
	failureWindow time.Duration
	lockDuration  time.Duration
}

func (s *Service) login(ctx context.Context, account, password string) (*int64, error) {
	account = normalizeLoginAccount(account)
	rule := s.loginSecurityRule(ctx)

	user, err := s.repo.GetByAccount(ctx, account)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := s.ensureLoginAllowed(ctx, account); err != nil {
			return nil, err
		}
		if err := s.recordLoginFailure(ctx, account, nil, rule); err != nil {
			return nil, err
		}
		return nil, errInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if user.LoginPermanentlyLocked || user.State != "enabled" {
		return nil, errInvalidCredentials
	}
	passwordValid := bcrypt.CheckPassword(user.Password, password)
	if err := s.ensureLoginAllowed(ctx, account); err != nil {
		if errors.Is(err, errTemporaryLocked) && !passwordValid {
			if err := s.recordLoginFailure(ctx, account, user, rule); err != nil {
				return nil, err
			}
		}
		return nil, err
	}
	if !passwordValid {
		if err := s.recordLoginFailure(ctx, account, user, rule); err != nil {
			return nil, err
		}
		return nil, errInvalidCredentials
	}
	if err := s.resetLoginFailure(ctx, account, user.ID); err != nil {
		return nil, err
	}
	return &user.ID, nil
}

func normalizeLoginAccount(account string) string {
	return strings.ToLower(strings.TrimSpace(account))
}

func (s *Service) ensureLoginAllowed(ctx context.Context, account string) error {
	locked, err := s.cache.Exists(ctx, s.loginLockKey(account)).Result()
	if err != nil {
		return err
	}
	if locked > 0 {
		return errTemporaryLocked
	}
	return nil
}

func (s *Service) recordLoginFailure(ctx context.Context, account string, user *model.SystemUser, rule loginSecurityRule) error {
	counter, ok := s.cache.(cache.AtomicCounter)
	if !ok {
		return errors.New("cache does not support atomic login failure counting")
	}
	failures, err := counter.IncrWithTTL(ctx, s.loginFailureKey(account), rule.failureWindow)
	if err != nil {
		return err
	}
	if user != nil {
		permanentlyLocked, err := s.repo.RecordLoginFailure(ctx, user.ID, permanentLoginFailureThreshold)
		if err != nil {
			return err
		}
		if permanentlyLocked {
			return nil
		}
	}
	if failures >= int64(rule.maxFailures) {
		if _, err := s.cache.Set(ctx, s.loginLockKey(account), "1", rule.lockDuration).Result(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) resetLoginFailure(ctx context.Context, account string, userID int64) error {
	if _, err := s.cache.Delete(ctx, s.loginFailureKey(account), s.loginLockKey(account)).Result(); err != nil {
		return err
	}
	return s.repo.ResetLoginSecurity(ctx, userID)
}

func (s *Service) clearLoginLock(ctx context.Context, account string) error {
	_, err := s.cache.Delete(ctx, s.loginFailureKey(normalizeLoginAccount(account)), s.loginLockKey(normalizeLoginAccount(account))).Result()
	return err
}

func (s *Service) loginFailureKey(account string) string {
	return consts.AuthLoginFailurePrefix + account
}

func (s *Service) loginLockKey(account string) string {
	return consts.AuthLoginTemporaryLockPrefix + account
}

func (s *Service) loginSecurityRule(ctx context.Context) loginSecurityRule {
	defaults := s.loginSecurityDefaults
	maxFailures := s.loginSecurityValue(ctx, "sys.login.max_failures", defaults.MaxFailures)
	if maxFailures >= int(permanentLoginFailureThreshold) {
		hlog.CtxWarnf(ctx, "[login-security] max failures %d must be below permanent threshold; using YAML default", maxFailures)
		maxFailures = defaults.MaxFailures
	}
	return loginSecurityRule{
		maxFailures:   maxFailures,
		failureWindow: time.Duration(s.loginSecurityValue(ctx, "sys.login.failure_window_minutes", defaults.FailureWindowMinutes)) * time.Minute,
		lockDuration:  time.Duration(s.loginSecurityValue(ctx, "sys.login.lock_minutes", defaults.LockMinutes)) * time.Minute,
	}
}

func (s *Service) loginSecurityValue(ctx context.Context, key string, fallback int) int {
	if configService.SystemConfigSVC == nil {
		return fallback
	}
	config, err := configService.SystemConfigSVC.GetByKey(ctx, key)
	if err != nil || config.Status != 0 {
		return fallback
	}
	value, err := strconv.Atoi(strings.TrimSpace(config.ConfigValue))
	if err != nil || value < 1 {
		hlog.CtxWarnf(ctx, "[login-security] invalid %s=%q; using YAML default", key, config.ConfigValue)
		return fallback
	}
	return value
}
