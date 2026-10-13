package notice

import (
	"context"
	"errors"
	"testing"

	"gosilite/application/base/ctxutil"
	"gosilite/application/framework/dal/model"
	"gosilite/pkg/ctxcache"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

type fakeNoticeRepo struct {
	byID      map[int64]*model.SystemNotice
	createErr error
	created   *model.SystemNotice
	saved     *model.SystemNotice
	deletedID int64

	// 消费侧（通知中心）
	inboxRows     []InboxItem
	inboxTotal    int64
	unread        int64
	markedRead    [][2]int64 // {userID, noticeID}
	markAllCalls  int
	markAllN      int64
	getPublishedE error
}

func (f *fakeNoticeRepo) Page(ctx context.Context, page, pageSize int, keyword string, noticeType, status *int16) ([]*model.SystemNotice, int64, error) {
	return nil, 0, nil
}

func (f *fakeNoticeRepo) GetByID(ctx context.Context, id int64) (*model.SystemNotice, error) {
	if m, ok := f.byID[id]; ok {
		return m, nil
	}
	return nil, errors.New("record not found")
}

func (f *fakeNoticeRepo) Create(ctx context.Context, m *model.SystemNotice) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = m
	return nil
}

func (f *fakeNoticeRepo) Save(ctx context.Context, m *model.SystemNotice) error {
	f.saved = m
	return nil
}

func (f *fakeNoticeRepo) SoftDelete(ctx context.Context, id int64) error {
	f.deletedID = id
	return nil
}

func (f *fakeNoticeRepo) InboxPage(ctx context.Context, userID int64, page, pageSize int, keyword string) ([]InboxItem, int64, int64, error) {
	return f.inboxRows, f.inboxTotal, f.unread, nil
}

func (f *fakeNoticeRepo) UnreadCount(ctx context.Context, userID int64) (int64, error) {
	return f.unread, nil
}

func (f *fakeNoticeRepo) GetPublishedByID(ctx context.Context, id int64) (*model.SystemNotice, error) {
	if f.getPublishedE != nil {
		return nil, f.getPublishedE
	}
	if m, ok := f.byID[id]; ok && m.Status == 0 && m.DelFlag == 0 {
		return m, nil
	}
	return nil, errors.New("record not found")
}

func (f *fakeNoticeRepo) MarkRead(ctx context.Context, userID, noticeID int64) error {
	f.markedRead = append(f.markedRead, [2]int64{userID, noticeID})
	return nil
}

func (f *fakeNoticeRepo) MarkAllRead(ctx context.Context, userID int64) (int64, error) {
	f.markAllCalls++
	return f.markAllN, nil
}

func newTestService() (*Service, *fakeNoticeRepo) {
	repo := &fakeNoticeRepo{byID: map[int64]*model.SystemNotice{}}
	return NewService(repo), repo
}

func loginCtx(account, username string) context.Context {
	ctx := ctxcache.Init(context.Background())
	ctxutil.SetLoginUser(ctx, &ctxutil.LoginUser{ID: 1, Account: account, Username: username})
	return ctx
}

func errCode(err error) int32 {
	var se errorx.StatusError
	if errors.As(err, &se) {
		return se.Code()
	}
	return 0
}

func TestCreateRejectsBlankTitle(t *testing.T) {
	svc, repo := newTestService()
	err := svc.Create(context.Background(), &model.SystemNotice{Title: "   "})
	if errCode(err) != errno.ErrSystemInvalidParam {
		t.Fatalf("want ErrSystemInvalidParam, got %v", err)
	}
	if repo.created != nil {
		t.Fatalf("repo.Create should not be called")
	}
}

func TestCreateFillsPublisherFromLoginUser(t *testing.T) {
	svc, repo := newTestService()
	m := &model.SystemNotice{Title: "春节放假通知", NoticeType: 1, Status: 0}
	if err := svc.Create(loginCtx("skyx", "张三"), m); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if repo.created == nil || repo.created.Publisher != "张三" {
		t.Fatalf("publisher should be login username, got %+v", repo.created)
	}
	if repo.created.DelFlag != 0 {
		t.Fatalf("del_flag should be 0")
	}
}

func TestCreateFallsBackToAccount(t *testing.T) {
	svc, repo := newTestService()
	m := &model.SystemNotice{Title: "t"}
	if err := svc.Create(loginCtx("skyx", ""), m); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if repo.created.Publisher != "skyx" {
		t.Fatalf("publisher should fall back to account, got %q", repo.created.Publisher)
	}
}

func TestUpdateMissingReturnsNotFound(t *testing.T) {
	svc, _ := newTestService()
	err := svc.Update(context.Background(), &model.SystemNotice{ID: 99, Title: "x"})
	if errCode(err) != errno.ErrSystemNoticeNotFound {
		t.Fatalf("want ErrSystemNoticeNotFound, got %v", err)
	}
}

func TestUpdateRejectsBlankTitle(t *testing.T) {
	svc, _ := newTestService()
	err := svc.Update(context.Background(), &model.SystemNotice{ID: 1, Title: ""})
	if errCode(err) != errno.ErrSystemInvalidParam {
		t.Fatalf("want ErrSystemInvalidParam, got %v", err)
	}
}

func TestGetMissingReturnsNotFound(t *testing.T) {
	svc, _ := newTestService()
	if _, err := svc.Get(context.Background(), 7); errCode(err) != errno.ErrSystemNoticeNotFound {
		t.Fatalf("want ErrSystemNoticeNotFound, got %v", err)
	}
}

func TestDeleteDelegatesToRepo(t *testing.T) {
	svc, repo := newTestService()
	if err := svc.Delete(context.Background(), 5); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if repo.deletedID != 5 {
		t.Fatalf("want soft delete id 5, got %d", repo.deletedID)
	}
}

// ========== 消费侧（通知中心） ==========

func TestDetailAutoMarksRead(t *testing.T) {
	svc, repo := newTestService()
	repo.byID[1] = &model.SystemNotice{ID: 1, Title: "t", Status: 0}
	m, err := svc.Detail(loginCtx("u", "张三"), 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if m.ID != 1 {
		t.Fatalf("want id 1, got %d", m.ID)
	}
	if len(repo.markedRead) != 1 || repo.markedRead[0] != [2]int64{1, 1} {
		t.Fatalf("want auto MarkRead(user=1, notice=1), got %v", repo.markedRead)
	}
}

func TestDetailWithdrawnReturnsNotFound(t *testing.T) {
	svc, repo := newTestService()
	repo.byID[2] = &model.SystemNotice{ID: 2, Status: 1} // 已撤回
	if _, err := svc.Detail(loginCtx("u", ""), 2); errCode(err) != errno.ErrSystemNoticeNotFound {
		t.Fatalf("want ErrSystemNoticeNotFound, got %v", err)
	}
	if len(repo.markedRead) != 0 {
		t.Fatalf("withdrawn notice must not be marked read")
	}
}

func TestDetailRequiresLogin(t *testing.T) {
	svc, _ := newTestService()
	if _, err := svc.Detail(context.Background(), 1); errCode(err) != errno.ErrUserAuthenticationFailed {
		t.Fatalf("want ErrUserAuthenticationFailed, got %v", err)
	}
}

func TestReadNilIDMarksAll(t *testing.T) {
	svc, repo := newTestService()
	repo.markAllN = 3
	n, err := svc.Read(loginCtx("u", ""), nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if n != 3 || repo.markAllCalls != 1 {
		t.Fatalf("want MarkAllRead once returning 3, got n=%d calls=%d", n, repo.markAllCalls)
	}
}

func TestReadSinglePublished(t *testing.T) {
	svc, repo := newTestService()
	repo.byID[7] = &model.SystemNotice{ID: 7, Status: 0}
	id := int64(7)
	n, err := svc.Read(loginCtx("u", ""), &id)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if n != 1 || len(repo.markedRead) != 1 || repo.markedRead[0] != [2]int64{1, 7} {
		t.Fatalf("want single read of 7, got n=%d marked=%v", n, repo.markedRead)
	}
	if repo.markAllCalls != 0 {
		t.Fatalf("single read must not trigger MarkAllRead")
	}
}

func TestReadSingleWithdrawnReturnsNotFound(t *testing.T) {
	svc, repo := newTestService()
	repo.byID[8] = &model.SystemNotice{ID: 8, Status: 1}
	id := int64(8)
	if _, err := svc.Read(loginCtx("u", ""), &id); errCode(err) != errno.ErrSystemNoticeNotFound {
		t.Fatalf("want ErrSystemNoticeNotFound, got %v", err)
	}
	if len(repo.markedRead) != 0 {
		t.Fatalf("withdrawn notice must not be marked read")
	}
}
