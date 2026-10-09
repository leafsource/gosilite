package user

import (
	"context"
	"sort"
	"strings"
	"time"

	"gosilite/application/base/useragent"
	"gosilite/application/framework/service/dept"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// OnlineRow 在线会话行（含部门名，列表直接可渲染）。
type OnlineRow struct {
	AccessHash string
	UserID     int64
	Account    string
	DeptID     int64
	DeptName   string
	IP         string
	Browser    string
	OS         string
	LoginAt    time.Time
}

// RegisterOnline 登录成功后登记在线会话。
// 仅告警降级：Redis/查询失败绝不阻塞登录主流程。
func (s *Service) RegisterOnline(ctx context.Context, userID int64, accessRaw, refreshRaw, ip, ua string, expire time.Time) {
	info, err := s.onlineInfo(ctx, userID, ip, ua)
	if err != nil {
		hlog.CtxWarnf(ctx, "[online] register info failed: user=%d err=%v", userID, err)
		return
	}
	if err := s.sessions.PutOnline(ctx, accessRaw, refreshRaw, info, expire); err != nil {
		hlog.CtxWarnf(ctx, "[online] put failed: user=%d err=%v", userID, err)
	}
}

// MoveOnline 刷新换发后迁移登记（旧 access 摘除、新 access 写入）。
func (s *Service) MoveOnline(ctx context.Context, userID int64, oldAccess, newAccess, newRefresh, ip, ua string, expire time.Time) {
	info, err := s.onlineInfo(ctx, userID, ip, ua)
	if err != nil {
		hlog.CtxWarnf(ctx, "[online] move info failed: user=%d err=%v", userID, err)
		return
	}
	if err := s.sessions.MoveOnline(ctx, oldAccess, newAccess, newRefresh, info, expire); err != nil {
		hlog.CtxWarnf(ctx, "[online] move failed: user=%d err=%v", userID, err)
	}
}

// DeregisterOnline 登出摘除登记（容错降级）。
func (s *Service) DeregisterOnline(ctx context.Context, accessRaw string) {
	if err := s.sessions.RemoveOnline(ctx, accessRaw); err != nil {
		hlog.CtxWarnf(ctx, "[online] remove failed: %v", err)
	}
}

func (s *Service) onlineInfo(ctx context.Context, userID int64, ip, ua string) (OnlineSession, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return OnlineSession{}, err
	}
	browser, osName := useragent.Parse(ua)
	return OnlineSession{
		UserID:  user.ID,
		Account: user.Account,
		DeptID:  user.DeptID,
		IP:      ip,
		Browser: browser,
		OS:      osName,
		LoginAt: time.Now().Unix(),
	}, nil
}

// OnlinePage 在线会话分页：登录时间倒序，内存分页（在线量小）。
func (s *Service) OnlinePage(ctx context.Context, keyword string, page, pageSize int) ([]OnlineRow, int64, error) {
	entries, err := s.sessions.ListOnline(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows := make([]OnlineRow, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, OnlineRow{
			AccessHash: e.AccessHash,
			UserID:     e.Session.UserID,
			Account:    e.Session.Account,
			DeptID:     e.Session.DeptID,
			IP:         e.Session.IP,
			Browser:    e.Session.Browser,
			OS:         e.Session.OS,
			LoginAt:    time.Unix(e.Session.LoginAt, 0),
		})
	}
	rows2, total := pageOnline(rows, keyword, page, pageSize, s.deptNameMap(ctx))
	return rows2, total, nil
}

// deptNameMap 部门 id→名称（改名实时反映；失败降级为空名）。
func (s *Service) deptNameMap(ctx context.Context) map[int64]string {
	names := make(map[int64]string)
	list, err := dept.SystemDeptSVC.List(ctx)
	if err != nil {
		hlog.CtxWarnf(ctx, "[online] dept list failed: %v", err)
		return names
	}
	for _, d := range list {
		names[d.ID] = d.DeptName
	}
	return names
}

// pageOnline 纯函数：按登录名称过滤 → 登录时间倒序 → 内存分页 → 附部门名。
func pageOnline(rows []OnlineRow, keyword string, page, pageSize int, deptNames map[int64]string) ([]OnlineRow, int64) {
	kw := strings.ToLower(strings.TrimSpace(keyword))
	filtered := make([]OnlineRow, 0, len(rows))
	for _, r := range rows {
		if kw != "" && !strings.Contains(strings.ToLower(r.Account), kw) {
			continue
		}
		r.DeptName = deptNames[r.DeptID]
		filtered = append(filtered, r)
	}
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].LoginAt.After(filtered[j].LoginAt) })

	total := int64(len(filtered))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	start := (page - 1) * pageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total
}

// ForceOffline 会话级强退。
func (s *Service) ForceOffline(ctx context.Context, accessHash string) error {
	if strings.TrimSpace(accessHash) == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("field", "id"))
	}
	if err := s.sessions.ForceOnline(ctx, accessHash); err != nil {
		if err == ErrOnlineNotFound {
			return errorx.New(errno.ErrOnlineSessionNotFound, errorx.KV("id", accessHash))
		}
		return err
	}
	return nil
}
