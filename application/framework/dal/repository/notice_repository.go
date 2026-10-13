package repository

import (
	"context"
	"strings"
	"time"

	"gosilite/application/framework/dal/model"

	"gorm.io/gorm"
)

// INoticeRepository 通知公告仓储
type INoticeRepository interface {
	Page(ctx context.Context, page, pageSize int, keyword string, noticeType, status *int16) ([]*model.SystemNotice, int64, error)
	GetByID(ctx context.Context, id int64) (*model.SystemNotice, error)
	Create(ctx context.Context, m *model.SystemNotice) error
	// Save 仅更新编辑允许变更的列（title/notice_type/content/status），
	// 不触碰 publisher/created_at/del_flag，避免整行 Save 把登录态之外的列清零。
	Save(ctx context.Context, m *model.SystemNotice) error
	SoftDelete(ctx context.Context, id int64) error

	// ===== 消费侧（通知中心） =====
	// InboxPage 当前用户收件箱：已发布（status=0）按 id DESC；keyword 标题过滤；返回条目 + total + unread
	InboxPage(ctx context.Context, userID int64, page, pageSize int, keyword string) ([]NoticeInboxRow, int64, int64, error)
	UnreadCount(ctx context.Context, userID int64) (int64, error)
	// GetPublishedByID 仅返回已发布的（撤回/删除对用户侧等同不存在）
	GetPublishedByID(ctx context.Context, id int64) (*model.SystemNotice, error)
	// MarkRead 幂等：insert … on conflict do nothing
	MarkRead(ctx context.Context, userID, noticeID int64) error
	// MarkAllRead 对本人全部已发布未读通知插入已读；返回插入行数
	MarkAllRead(ctx context.Context, userID int64) (int64, error)
}

// NoticeInboxRow 收件箱条目：通知 + 本人是否已读
type NoticeInboxRow struct {
	Notice *model.SystemNotice
	Read   bool
}

type noticeRepository struct{ db *gorm.DB }

// NewNoticeRepository 创建通知公告仓储
func NewNoticeRepository(db *gorm.DB) INoticeRepository {
	return &noticeRepository{db: db}
}

func (r *noticeRepository) Page(ctx context.Context, page, pageSize int, keyword string, noticeType, status *int16) ([]*model.SystemNotice, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}
	q := r.db.WithContext(ctx).Model(&model.SystemNotice{}).Where("del_flag = 0")
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("title LIKE ? OR content LIKE ?", kw, kw)
	}
	if noticeType != nil {
		q = q.Where("notice_type = ?", *noticeType)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.SystemNotice
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *noticeRepository) GetByID(ctx context.Context, id int64) (*model.SystemNotice, error) {
	var m model.SystemNotice
	err := r.db.WithContext(ctx).Where("id = ? AND del_flag = 0", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *noticeRepository) Create(ctx context.Context, m *model.SystemNotice) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *noticeRepository) Save(ctx context.Context, m *model.SystemNotice) error {
	return r.db.WithContext(ctx).Model(&model.SystemNotice{}).
		Where("id = ? AND del_flag = 0", m.ID).
		Updates(map[string]any{
			"title":       m.Title,
			"notice_type": m.NoticeType,
			"content":     m.Content,
			"status":      m.Status,
			"updated_at":  time.Now(),
		}).Error
}

func (r *noticeRepository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.SystemNotice{}).
		Where("id = ?", id).Update("del_flag", 2).Error
}

// ========== 消费侧（通知中心） ==========

// publishedCond 用户可见集：未删除且已发布
const publishedCond = "del_flag = 0 AND status = 0"

func (r *noticeRepository) InboxPage(ctx context.Context, userID int64, page, pageSize int, keyword string) ([]NoticeInboxRow, int64, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}
	keyword = strings.TrimSpace(keyword)
	published := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.SystemNotice{}).Where(publishedCond)
		if keyword != "" {
			q = q.Where("title LIKE ?", "%"+keyword+"%")
		}
		return q
	}
	var total int64
	if err := published().Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}
	unread, err := r.unreadCount(ctx, userID)
	if err != nil {
		return nil, 0, 0, err
	}
	var list []*model.SystemNotice
	if err := published().
		Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, 0, err
	}
	readIDs := map[int64]bool{}
	if len(list) > 0 {
		ids := make([]int64, 0, len(list))
		for _, m := range list {
			ids = append(ids, m.ID)
		}
		var hit []int64
		if err := r.db.WithContext(ctx).Model(&model.SystemNoticeRead{}).
			Where("user_id = ? AND notice_id IN ?", userID, ids).
			Pluck("notice_id", &hit).Error; err != nil {
			return nil, 0, 0, err
		}
		for _, id := range hit {
			readIDs[id] = true
		}
	}
	rows := make([]NoticeInboxRow, len(list))
	for i, m := range list {
		rows[i] = NoticeInboxRow{Notice: m, Read: readIDs[m.ID]}
	}
	return rows, total, unread, nil
}

func (r *noticeRepository) unreadCount(ctx context.Context, userID int64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.SystemNotice{}).
		Where(publishedCond).
		Where("NOT EXISTS (SELECT 1 FROM public.system_notice_read rr WHERE rr.notice_id = system_notice.id AND rr.user_id = ?)", userID).
		Count(&n).Error
	return n, err
}

func (r *noticeRepository) UnreadCount(ctx context.Context, userID int64) (int64, error) {
	return r.unreadCount(ctx, userID)
}

func (r *noticeRepository) GetPublishedByID(ctx context.Context, id int64) (*model.SystemNotice, error) {
	var m model.SystemNotice
	if err := r.db.WithContext(ctx).Where("id = ? AND "+publishedCond, id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *noticeRepository) MarkRead(ctx context.Context, userID, noticeID int64) error {
	return r.db.WithContext(ctx).Exec(
		`INSERT INTO public.system_notice_read (user_id, notice_id) VALUES (?, ?)
		 ON CONFLICT (user_id, notice_id) DO NOTHING`,
		userID, noticeID,
	).Error
}

func (r *noticeRepository) MarkAllRead(ctx context.Context, userID int64) (int64, error) {
	res := r.db.WithContext(ctx).Exec(
		`INSERT INTO public.system_notice_read (user_id, notice_id)
		 SELECT ?, n.id FROM public.system_notice n
		 WHERE n.del_flag = 0 AND n.status = 0
		   AND NOT EXISTS (
		     SELECT 1 FROM public.system_notice_read r
		     WHERE r.user_id = ? AND r.notice_id = n.id
		   )
		 ON CONFLICT (user_id, notice_id) DO NOTHING`,
		userID, userID,
	)
	return res.RowsAffected, res.Error
}
