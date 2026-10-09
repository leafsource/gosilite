package repository

import (
	"context"
	"time"

	"gosilite/application/framework/dal/model"

	"gorm.io/gorm"
)

// ===== 筛选条件（服务层组装） =====

type OperPageFilter struct {
	TraceID    string
	UserName   string
	Method     string // request_uri 模糊
	HTTPMethod string
	OperType   *int32
	ExecStatus *int32
	StartTime  *time.Time
	EndTime    *time.Time
}

type ErrorPageFilter struct {
	TraceID      string
	UserName     string
	Keyword      string // err_msg 模糊
	HTTPMethod   string
	HandleStatus *int32
	ErrLevel     *int32
	StartTime    *time.Time
	EndTime      *time.Time
}

type LoginPageFilter struct {
	UserName  string
	ClientIP  string
	Status    *int32
	StartTime *time.Time
	EndTime   *time.Time
}

// SysLogRepository 三张日志表的读写入口。
type SysLogRepository interface {
	// 批量写（异步 worker 调用）
	BatchInsertOper(ctx context.Context, rows []*model.SystemOperLog) error
	BatchInsertError(ctx context.Context, rows []*model.SystemErrorLog) error
	BatchInsertLogin(ctx context.Context, rows []*model.SystemLoginLog) error

	// 操作日志：列表不含大字段（headers/body），详情单查
	OperPage(ctx context.Context, page, pageSize int, f OperPageFilter) ([]*model.SystemOperLog, int64, error)
	OperByID(ctx context.Context, id int64) (*model.SystemOperLog, error)
	CleanOper(ctx context.Context, ids []int64, all bool) (int64, error)

	ErrorPage(ctx context.Context, page, pageSize int, f ErrorPageFilter) ([]*model.SystemErrorLog, int64, error)
	ErrorByID(ctx context.Context, id int64) (*model.SystemErrorLog, error)
	MarkErrorHandled(ctx context.Context, id int64, remark string) (int64, error)
	CleanError(ctx context.Context, ids []int64, all bool) (int64, error)

	LoginPage(ctx context.Context, page, pageSize int, f LoginPageFilter) ([]*model.SystemLoginLog, int64, error)
}

type syslogRepository struct{ db *gorm.DB }

// NewSysLogRepository 构造日志仓储。
func NewSysLogRepository(db *gorm.DB) SysLogRepository {
	return &syslogRepository{db: db}
}

// 列表查询排除的大字段保持不 SELECT，读写路径分离（详情才拉全量）。
const operListColumns = "id, trace_id, user_id, user_name, oper_type, request_uri, http_method, module, oper_status, error_msg, remark, cost_ms, client_ip, created_at"

const errorListColumns = "id, trace_id, user_id, user_name, err_type, err_level, handle_status, err_code, err_msg, request_uri, http_method, module, cost_ms, client_ip, handled_at, created_at"

func (r *syslogRepository) BatchInsertOper(ctx context.Context, rows []*model.SystemOperLog) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(rows, 100).Error
}

func (r *syslogRepository) BatchInsertError(ctx context.Context, rows []*model.SystemErrorLog) error {
	if len(rows) == 0 {
		return nil
	}
	// handled_at/handle_remark 由「标记处理」更新写入，批量插入保持 NULL（零值时间会落成 0001-01-01）
	return r.db.WithContext(ctx).Omit("handled_at", "handle_remark").CreateInBatches(rows, 100).Error
}

func (r *syslogRepository) BatchInsertLogin(ctx context.Context, rows []*model.SystemLoginLog) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(rows, 100).Error
}

// ===== 操作日志 =====

func (r *syslogRepository) OperPage(ctx context.Context, page, pageSize int, f OperPageFilter) ([]*model.SystemOperLog, int64, error) {
	q := r.operCond(f)
	var total int64
	if err := q.WithContext(ctx).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}
	var rows []*model.SystemOperLog
	err := q.WithContext(ctx).
		Select(operListColumns).
		Order("created_at DESC, id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&rows).Error
	return rows, total, err
}

func (r *syslogRepository) operCond(f OperPageFilter) *gorm.DB {
	q := r.db.Model(&model.SystemOperLog{})
	if f.TraceID != "" {
		q = q.Where("trace_id = ?", f.TraceID)
	}
	if f.UserName != "" {
		q = q.Where("user_name LIKE ?", "%"+f.UserName+"%")
	}
	if f.Method != "" {
		q = q.Where("request_uri LIKE ?", "%"+f.Method+"%")
	}
	if f.HTTPMethod != "" {
		q = q.Where("http_method = ?", f.HTTPMethod)
	}
	if f.OperType != nil {
		q = q.Where("oper_type = ?", *f.OperType)
	}
	if f.ExecStatus != nil {
		q = q.Where("oper_status = ?", *f.ExecStatus)
	}
	if f.StartTime != nil {
		q = q.Where("created_at >= ?", *f.StartTime)
	}
	if f.EndTime != nil {
		q = q.Where("created_at <= ?", *f.EndTime)
	}
	return q
}

func (r *syslogRepository) OperByID(ctx context.Context, id int64) (*model.SystemOperLog, error) {
	var row model.SystemOperLog
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *syslogRepository) CleanOper(ctx context.Context, ids []int64, all bool) (int64, error) {
	q := r.db.WithContext(ctx).Where("1 = 1")
	if !all {
		if len(ids) == 0 {
			return 0, nil
		}
		q = q.Where("id IN ?", ids)
	}
	res := q.Delete(&model.SystemOperLog{})
	return res.RowsAffected, res.Error
}

// ===== 系统异常日志 =====

func (r *syslogRepository) ErrorPage(ctx context.Context, page, pageSize int, f ErrorPageFilter) ([]*model.SystemErrorLog, int64, error) {
	q := r.errorCond(f)
	var total int64
	if err := q.WithContext(ctx).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}
	var rows []*model.SystemErrorLog
	err := q.WithContext(ctx).
		Select(errorListColumns).
		Order("created_at DESC, id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&rows).Error
	return rows, total, err
}

func (r *syslogRepository) errorCond(f ErrorPageFilter) *gorm.DB {
	q := r.db.Model(&model.SystemErrorLog{})
	if f.TraceID != "" {
		q = q.Where("trace_id = ?", f.TraceID)
	}
	if f.UserName != "" {
		q = q.Where("user_name LIKE ?", "%"+f.UserName+"%")
	}
	if f.Keyword != "" {
		q = q.Where("err_msg LIKE ?", "%"+f.Keyword+"%")
	}
	if f.HTTPMethod != "" {
		q = q.Where("http_method = ?", f.HTTPMethod)
	}
	if f.HandleStatus != nil {
		q = q.Where("handle_status = ?", *f.HandleStatus)
	}
	if f.ErrLevel != nil {
		q = q.Where("err_level = ?", *f.ErrLevel)
	}
	if f.StartTime != nil {
		q = q.Where("created_at >= ?", *f.StartTime)
	}
	if f.EndTime != nil {
		q = q.Where("created_at <= ?", *f.EndTime)
	}
	return q
}

func (r *syslogRepository) ErrorByID(ctx context.Context, id int64) (*model.SystemErrorLog, error) {
	var row model.SystemErrorLog
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *syslogRepository) MarkErrorHandled(ctx context.Context, id int64, remark string) (int64, error) {
	res := r.db.WithContext(ctx).
		Model(&model.SystemErrorLog{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"handle_status": 1,
			"handled_at":    time.Now(),
			"handle_remark": remark,
		})
	return res.RowsAffected, res.Error
}

func (r *syslogRepository) CleanError(ctx context.Context, ids []int64, all bool) (int64, error) {
	q := r.db.WithContext(ctx).Where("1 = 1")
	if !all {
		if len(ids) == 0 {
			return 0, nil
		}
		q = q.Where("id IN ?", ids)
	}
	res := q.Delete(&model.SystemErrorLog{})
	return res.RowsAffected, res.Error
}

// ===== 登录日志 =====

func (r *syslogRepository) LoginPage(ctx context.Context, page, pageSize int, f LoginPageFilter) ([]*model.SystemLoginLog, int64, error) {
	q := r.db.Model(&model.SystemLoginLog{})
	if f.UserName != "" {
		q = q.Where("user_name LIKE ?", "%"+f.UserName+"%")
	}
	if f.ClientIP != "" {
		q = q.Where("client_ip LIKE ?", "%"+f.ClientIP+"%")
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	if f.StartTime != nil {
		q = q.Where("created_at >= ?", *f.StartTime)
	}
	if f.EndTime != nil {
		q = q.Where("created_at <= ?", *f.EndTime)
	}
	var total int64
	if err := q.WithContext(ctx).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}
	var rows []*model.SystemLoginLog
	err := q.WithContext(ctx).
		Order("created_at DESC, id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&rows).Error
	return rows, total, err
}
