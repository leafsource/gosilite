package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"gosilite/application/infra/cache"
	"gosilite/types/consts"
)

const (
	refreshTokenTTL = 7 * 24 * time.Hour
	accessDenyTTL   = 24 * time.Hour
)

type SessionTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	Expire       time.Time
}

type SessionStore struct {
	cache cache.CacheClient
}

func NewSessionStore(c cache.CacheClient) *SessionStore {
	return &SessionStore{cache: c}
}

func (s *SessionStore) IssueRefreshToken(ctx context.Context, userID int64) (string, error) {
	raw, err := randomToken(32)
	if err != nil {
		return "", err
	}
	gen, err := s.refreshGeneration(ctx, userID)
	if err != nil {
		return "", err
	}
	value := fmt.Sprintf("%d:%d", userID, gen)
	if err := s.cache.Set(ctx, refreshTokenKey(raw), value, refreshTokenTTL).Err(); err != nil {
		return "", err
	}
	return raw, nil
}

func (s *SessionStore) ValidateRefreshToken(ctx context.Context, refreshToken string) (int64, error) {
	if refreshToken == "" {
		return 0, fmt.Errorf("refresh token is required")
	}
	value, err := s.cache.Get(ctx, refreshTokenKey(refreshToken)).Result()
	if err != nil {
		return 0, fmt.Errorf("invalid refresh token")
	}
	userID, gen, err := parseRefreshValue(value)
	if err != nil {
		return 0, err
	}
	current, err := s.refreshGeneration(ctx, userID)
	if err != nil {
		return 0, err
	}
	if gen != current {
		return 0, fmt.Errorf("invalid refresh token")
	}
	return userID, nil
}

func (s *SessionStore) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return s.cache.Delete(ctx, refreshTokenKey(refreshToken)).Err()
}

func (s *SessionStore) RevokeUserRefreshTokens(ctx context.Context, userID int64) error {
	gen := time.Now().UnixNano()
	return s.cache.Set(ctx, userGenerationKey(userID), strconv.FormatInt(gen, 10), refreshTokenTTL).Err()
}

func (s *SessionStore) DenyAccessToken(ctx context.Context, accessToken string, expire time.Time) error {
	if accessToken == "" {
		return nil
	}
	ttl := time.Until(expire)
	if ttl <= 0 {
		ttl = accessDenyTTL
	}
	if ttl > accessDenyTTL {
		ttl = accessDenyTTL
	}
	return s.cache.Set(ctx, accessDenyKey(accessToken), "1", ttl).Err()
}

func (s *SessionStore) IsAccessTokenDenied(ctx context.Context, accessToken string) (bool, error) {
	if accessToken == "" {
		return false, nil
	}
	cmd := s.cache.Exists(ctx, accessDenyKey(accessToken))
	exists, err := cmd.Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (s *SessionStore) refreshGeneration(ctx context.Context, userID int64) (int64, error) {
	value, err := s.cache.Get(ctx, userGenerationKey(userID)).Result()
	if err != nil {
		if err == cache.EmptyValue {
			return 0, nil
		}
		return 0, err
	}
	if value == "" {
		return 0, nil
	}
	gen, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return gen, nil
}

func refreshTokenKey(token string) string {
	return fmt.Sprintf("%s%s", consts.AuthRefreshTokenPrefix, hashToken(token))
}

func userGenerationKey(userID int64) string {
	return fmt.Sprintf("%s%d", consts.AuthUserRefreshGenerationPrefix, userID)
}

func accessDenyKey(token string) string {
	return fmt.Sprintf("%s%s", consts.AuthAccessTokenDenyPrefix, hashToken(token))
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func parseRefreshValue(value string) (int64, int64, error) {
	var userID, gen int64
	if _, err := fmt.Sscanf(value, "%d:%d", &userID, &gen); err != nil {
		return 0, 0, fmt.Errorf("invalid refresh token payload")
	}
	return userID, gen, nil
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
