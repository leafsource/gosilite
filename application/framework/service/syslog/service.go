package syslog

import (
	"context"
	"strconv"
	"strings"
	"time"

	"gosilite/application/base/logq"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

// Service 日志管理查询/清理/处理（采集写入走 sink.go）。
type Service struct {
	repo repository.SysLogRepository
}

// SysLogSVC 包级门面（InitService 安装）。
var SysLogSVC *Service

// sinkOnce 保证异步 worker 只启动一次。
var sinkOnce struct {
	done bool
	s    *sink
}

// InitService 安装门面并启动有界异步批量写。
func InitService(repo repository.SysLogRepository) {
	SysLogSVC = &Service{repo: repo}
	if !sinkOnce.done {
		sinkOnce.s = newSink(repo)
		sinkOnce.done = true
		logq.Set(sinkOnce.s)
		logq.SetFlusher(sinkOnce.s)
	}
}

// newService 测试用（不起 worker）。
func newService(repo repository.SysLogRepository) *Service {
	return &Service{repo: repo}
}

const (
	timeLayout  = "2006-01-02 15:04:05"
	maxPageSize = 500
)

func normalize(page, pageSize int32) (int, int) {
	p, ps := int(page), int(pageSize)
	if p < 1 {
		p = 1
	}
	if ps < 1 {
		ps = 10
	}
	if ps > maxPageSize {
		ps = maxPageSize
	}
	return p, ps
}

// pageRange 解析时间范围为可空指针；日期-only 时起始补 00:00:00、结束补 23:59:59。
func pageRange(start, end *string) (*time.Time, *time.Time, error) {
	var st, et *time.Time
	if start != nil && strings.TrimSpace(*start) != "" {
		v := strings.TrimSpace(*start)
		t, err := time.ParseInLocation(timeLayout, v, time.Local)
		if err != nil {
			t, err = time.ParseInLocation("2006-01-02", v, time.Local)
		}
		if err != nil {
			return nil, nil, errorx.New(errno.ErrSystemInvalidParam, errorx.KV("field", "startTime"))
		}
		st = &t
	}
	if end != nil && strings.TrimSpace(*end) != "" {
		v := strings.TrimSpace(*end)
		t, err := time.ParseInLocation(timeLayout, v, time.Local)
		if err != nil {
			t, err = time.ParseInLocation("2006-01-02", v, time.Local)
			if err == nil {
				t = t.Add(24*time.Hour - time.Second)
			}
		}
		if err != nil {
			return nil, nil, errorx.New(errno.ErrSystemInvalidParam, errorx.KV("field", "endTime"))
		}
		et = &t
	}
	return st, et, nil
}

// ===== 操作日志 =====

func (s *Service) OperPage(ctx context.Context, page, pageSize int32, traceID, userName, method, httpMethod string,
	operType, execStatus *int32, start, end *string) ([]*model.SystemOperLog, int64, error) {
	p, ps := normalize(page, pageSize)
	st, et, err := pageRange(start, end)
	if err != nil {
		return nil, 0, err
	}
	f := repository.OperPageFilter{
		TraceID: strings.TrimSpace(traceID), UserName: strings.TrimSpace(userName),
		Method: strings.TrimSpace(method), HTTPMethod: strings.TrimSpace(httpMethod),
		OperType: operType, ExecStatus: execStatus,
		StartTime: st, EndTime: et,
	}
	return s.repo.OperPage(ctx, p, ps, f)
}

func (s *Service) OperDetail(ctx context.Context, id int64) (*model.SystemOperLog, error) {
	row, err := s.repo.OperByID(ctx, id)
	if err != nil {
		return nil, errorx.New(errno.ErrOperLogNotFound, errorx.KV("id", strconv.FormatInt(id, 10)))
	}
	return row, nil
}

func (s *Service) CleanOper(ctx context.Context, ids []int64, all bool) (int64, error) {
	return s.repo.CleanOper(ctx, ids, all)
}

// ===== 系统异常日志 =====

func (s *Service) ErrorPage(ctx context.Context, page, pageSize int32, traceID, userName, keyword, httpMethod string,
	handleStatus, errLevel *int32, start, end *string) ([]*model.SystemErrorLog, int64, error) {
	p, ps := normalize(page, pageSize)
	st, et, err := pageRange(start, end)
	if err != nil {
		return nil, 0, err
	}
	f := repository.ErrorPageFilter{
		TraceID: strings.TrimSpace(traceID), UserName: strings.TrimSpace(userName),
		Keyword: strings.TrimSpace(keyword), HTTPMethod: strings.TrimSpace(httpMethod),
		HandleStatus: handleStatus, ErrLevel: errLevel,
		StartTime: st, EndTime: et,
	}
	return s.repo.ErrorPage(ctx, p, ps, f)
}

func (s *Service) ErrorDetail(ctx context.Context, id int64) (*model.SystemErrorLog, error) {
	row, err := s.repo.ErrorByID(ctx, id)
	if err != nil {
		return nil, errorx.New(errno.ErrErrorLogNotFound, errorx.KV("id", strconv.FormatInt(id, 10)))
	}
	return row, nil
}

func (s *Service) HandleError(ctx context.Context, id int64, remark string) (int64, error) {
	n, err := s.repo.MarkErrorHandled(ctx, id, strings.TrimSpace(remark))
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, errorx.New(errno.ErrErrorLogNotFound, errorx.KV("id", strconv.FormatInt(id, 10)))
	}
	return n, nil
}

func (s *Service) CleanError(ctx context.Context, ids []int64, all bool) (int64, error) {
	return s.repo.CleanError(ctx, ids, all)
}

// ===== 登录日志 =====

func (s *Service) LoginPage(ctx context.Context, page, pageSize int32, userName, clientIP string,
	status *int32, start, end *string) ([]*model.SystemLoginLog, int64, error) {
	p, ps := normalize(page, pageSize)
	st, et, err := pageRange(start, end)
	if err != nil {
		return nil, 0, err
	}
	f := repository.LoginPageFilter{
		UserName: strings.TrimSpace(userName), ClientIP: strings.TrimSpace(clientIP),
		Status: status, StartTime: st, EndTime: et,
	}
	return s.repo.LoginPage(ctx, p, ps, f)
}
