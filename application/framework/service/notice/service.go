package notice

import (
	"context"
	"strconv"
	"strings"

	"gosilite/application/base/ctxutil"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

// SystemNoticeSVC 门面
var SystemNoticeSVC *Service

// Service 通知公告
type Service struct {
	repo repository.INoticeRepository
}

// NewService 构造
func NewService(repo repository.INoticeRepository) *Service {
	return &Service{repo: repo}
}

// InitService 安装门面
func InitService(repo repository.INoticeRepository) *Service {
	SystemNoticeSVC = NewService(repo)
	return SystemNoticeSVC
}

// Page 分页（noticeType/status 为 nil 时不过滤）
func (s *Service) Page(ctx context.Context, page, pageSize int, keyword string, noticeType, status *int16) ([]*model.SystemNotice, int64, error) {
	return s.repo.Page(ctx, page, pageSize, strings.TrimSpace(keyword), noticeType, status)
}

// Get 详情
func (s *Service) Get(ctx context.Context, id int64) (*model.SystemNotice, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errorx.New(errno.ErrSystemNoticeNotFound, errorx.KV("id", formatID(id)))
	}
	return m, nil
}

// Create 创建（publisher 取登录态）
func (s *Service) Create(ctx context.Context, m *model.SystemNotice) error {
	if strings.TrimSpace(m.Title) == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "标题不能为空"))
	}
	m.Title = strings.TrimSpace(m.Title)
	if login := ctxutil.GetLoginUser(ctx); login != nil {
		if login.Username != "" {
			m.Publisher = login.Username
		} else {
			m.Publisher = login.Account
		}
	}
	m.DelFlag = 0
	return s.repo.Create(ctx, m)
}

// Update 更新（不改 publisher）
func (s *Service) Update(ctx context.Context, m *model.SystemNotice) error {
	if strings.TrimSpace(m.Title) == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "标题不能为空"))
	}
	m.Title = strings.TrimSpace(m.Title)
	if _, err := s.repo.GetByID(ctx, m.ID); err != nil {
		return errorx.New(errno.ErrSystemNoticeNotFound, errorx.KV("id", formatID(m.ID)))
	}
	return s.repo.Save(ctx, m)
}

// Delete 软删
func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.SoftDelete(ctx, id)
}

// ========== 消费侧（通知中心） ==========

// InboxItem 收件箱条目（仓储行别名：通知 + 是否已读）
type InboxItem = repository.NoticeInboxRow

// Inbox 当前用户收件箱（已发布、按 id DESC；keyword 标题过滤；含 total 与 unread）
func (s *Service) Inbox(ctx context.Context, page, pageSize int, keyword string) ([]InboxItem, int64, int64, error) {
	uid, err := loginUserID(ctx)
	if err != nil {
		return nil, 0, 0, err
	}
	return s.repo.InboxPage(ctx, uid, page, pageSize, keyword)
}

// Detail 通知详情：仅已发布可见；成功后幂等标记本人已读（打开详情自动已读）
func (s *Service) Detail(ctx context.Context, id int64) (*model.SystemNotice, error) {
	uid, err := loginUserID(ctx)
	if err != nil {
		return nil, err
	}
	m, err := s.repo.GetPublishedByID(ctx, id)
	if err != nil {
		return nil, errorx.New(errno.ErrSystemNoticeNotFound, errorx.KV("id", formatID(id)))
	}
	if err := s.repo.MarkRead(ctx, uid, m.ID); err != nil {
		return nil, err
	}
	return m, nil
}

// UnreadCount 未读数 = 已发布 − 本人已读
func (s *Service) UnreadCount(ctx context.Context) (int64, error) {
	uid, err := loginUserID(ctx)
	if err != nil {
		return 0, err
	}
	return s.repo.UnreadCount(ctx, uid)
}

// Read 标记已读：id 缺省=全部已读；返回影响行数
func (s *Service) Read(ctx context.Context, id *int64) (int64, error) {
	uid, err := loginUserID(ctx)
	if err != nil {
		return 0, err
	}
	if id == nil {
		return s.repo.MarkAllRead(ctx, uid)
	}
	if _, err := s.repo.GetPublishedByID(ctx, *id); err != nil {
		return 0, errorx.New(errno.ErrSystemNoticeNotFound, errorx.KV("id", formatID(*id)))
	}
	if err := s.repo.MarkRead(ctx, uid, *id); err != nil {
		return 0, err
	}
	return 1, nil
}

func loginUserID(ctx context.Context) (int64, error) {
	if login := ctxutil.GetLoginUser(ctx); login != nil {
		return login.ID, nil
	}
	return 0, errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("reason", "not login"))
}

func formatID(v int64) string {
	return strconv.FormatInt(v, 10)
}
