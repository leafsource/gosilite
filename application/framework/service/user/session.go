package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
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

// ===== 在线会话登记（Redis Hash：field=access 摘要，value=JSON；惰性过期清理） =====

// OnlineSession 活跃会话快照（登录时物化，列表直接可读）。
type OnlineSession struct {
	UserID      int64  `json:"userId"`
	Account     string `json:"account"`
	DeptID      int64  `json:"deptId"`
	IP          string `json:"ip"`
	Browser     string `json:"browser"`
	OS          string `json:"os"`
	LoginAt     int64  `json:"loginAt"` // unix 秒
	Expire      int64  `json:"expire"`  // access 过期 unix 秒
	RefreshHash string `json:"refreshHash"`
}

// OnlineEntry 在线会话 + 行标识（access 摘要）。
type OnlineEntry struct {
	AccessHash string
	Session    OnlineSession
}

// PutOnline 登记/更新在线会话（refresh 摘要用于会话级强退吊销）。
func (s *SessionStore) PutOnline(ctx context.Context, accessRaw, refreshRaw string, info OnlineSession, expire time.Time) error {
	info.Expire = expire.Unix()
	info.RefreshHash = hashToken(refreshRaw)
	b, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return s.cache.HSet(ctx, consts.AuthOnlineSessionsKey, hashToken(accessRaw), string(b)).Err()
}

// RemoveOnline 摘除在线登记（登出）。
func (s *SessionStore) RemoveOnline(ctx context.Context, accessRaw string) error {
	return s.cache.HDel(ctx, consts.AuthOnlineSessionsKey, hashToken(accessRaw)).Err()
}

// MoveOnline 刷新换发：先写新会话再摘旧（瞬间并存无害），保证 rotation 后列表不丢。
// 登录时间沿用原会话（refresh 不重置会话起始时刻）。
func (s *SessionStore) MoveOnline(ctx context.Context, oldAccessRaw, newAccessRaw, newRefreshRaw string, info OnlineSession, expire time.Time) error {
	if all, err := s.cache.HGetAll(ctx, consts.AuthOnlineSessionsKey).Result(); err == nil {
		if raw, ok := all[hashToken(oldAccessRaw)]; ok {
			var old OnlineSession
			if json.Unmarshal([]byte(raw), &old) == nil && old.LoginAt > 0 {
				info.LoginAt = old.LoginAt
			}
		}
	}
	if err := s.PutOnline(ctx, newAccessRaw, newRefreshRaw, info, expire); err != nil {
		return err
	}
	return s.RemoveOnline(ctx, oldAccessRaw)
}

// ListOnline 列出未过期会话；顺带惰性清理过期 field（Hash 无 field TTL）。
func (s *SessionStore) ListOnline(ctx context.Context) ([]OnlineEntry, error) {
	all, err := s.cache.HGetAll(ctx, consts.AuthOnlineSessionsKey).Result()
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	out := make([]OnlineEntry, 0, len(all))
	var stale []string
	for field, raw := range all {
		var sess OnlineSession
		if err := json.Unmarshal([]byte(raw), &sess); err != nil || sess.Expire < now {
			stale = append(stale, field)
			continue
		}
		out = append(out, OnlineEntry{AccessHash: field, Session: sess})
	}
	if len(stale) > 0 {
		_ = s.cache.HDel(ctx, consts.AuthOnlineSessionsKey, stale...).Err()
	}
	return out, nil
}

// ForceOnline 会话级强退：access 进 denylist + 吊销该会话 refresh + 摘登记。
// 不存在该会话时返回 ErrOnlineNotFound。
func (s *SessionStore) ForceOnline(ctx context.Context, accessHash string) error {
	all, err := s.cache.HGetAll(ctx, consts.AuthOnlineSessionsKey).Result()
	if err != nil {
		return err
	}
	raw, ok := all[accessHash]
	if !ok {
		return ErrOnlineNotFound
	}
	var sess OnlineSession
	if err := json.Unmarshal([]byte(raw), &sess); err != nil {
		return ErrOnlineNotFound
	}
	ttl := time.Until(time.Unix(sess.Expire, 0))
	if ttl <= 0 {
		ttl = accessDenyTTL
	}
	if ttl > accessDenyTTL {
		ttl = accessDenyTTL
	}
	// denylist key = 前缀 + access 摘要（与 accessDenyKey 同构，field 已是摘要）
	if err := s.cache.Set(ctx, consts.AuthAccessTokenDenyPrefix+accessHash, "1", ttl).Err(); err != nil {
		return err
	}
	// 该会话 refresh（key = 前缀 + 摘要）
	if sess.RefreshHash != "" {
		_ = s.cache.Delete(ctx, consts.AuthRefreshTokenPrefix+sess.RefreshHash).Err()
	}
	return s.cache.HDel(ctx, consts.AuthOnlineSessionsKey, accessHash).Err()
}

// ErrOnlineNotFound 在线会话不存在（已被踢/过期）。
var ErrOnlineNotFound = errors.New("online session not found")
