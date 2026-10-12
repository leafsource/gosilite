package scheduler

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/sse"
	apiModel "gosilite/api/model/framework/scheduler"
	dalModel "gosilite/application/framework/dal/model"
	core "gosilite/application/framework/scheduler"
	svc "gosilite/application/framework/service/scheduler"
	R "gosilite/pkg/utils/response"
	"strconv"
	"time"
)

func dto(m *dalModel.SystemJob) *apiModel.SchedulerJobDTO {
	return &apiModel.SchedulerJobDTO{ID: m.ID, JobKey: m.JobKey, JobName: m.JobName, Cron: m.Cron, TimeoutSec: m.TimeoutSec, ParamsJson: m.ParamsJSON, RetryCount: m.RetryCount, Status: int32(m.Status), CreatedAt: m.CreatedAt.Format(time.RFC3339), UpdatedAt: m.UpdatedAt.Format(time.RFC3339)}
}
func save(r *apiModel.SchedulerJobSaveRequest) *dalModel.SystemJob {
	m := &dalModel.SystemJob{JobKey: r.JobKey, JobName: r.JobName, Cron: r.Cron, TimeoutSec: r.TimeoutSec, ParamsJSON: r.ParamsJson, RetryCount: r.RetryCount, Status: int16(r.Status)}
	if r.ID != nil {
		m.ID = *r.ID
	}
	return m
}

func runDTO(x *dalModel.SystemJobRun) *apiModel.SchedulerJobRunDTO {
	z := &apiModel.SchedulerJobRunDTO{ID: x.ID, JobId: x.JobID, JobKey: x.JobKey, Trigger: x.Trigger, RetryCount: x.Attempt - 1, Status: x.Status, TriggeredAt: x.TriggeredAt.Format(time.RFC3339), Output: x.Output, Error: x.Error, JobName: x.JobName, ParamsJson: x.ParamsJSON, TimeoutSec: x.TimeoutSec, ConfigVersion: x.ConfigVersion}
	if x.FinishedAt != nil {
		z.FinishedAt = x.FinishedAt.Format(time.RFC3339)
	}
	return z
}

func runLogDTO(x *dalModel.SystemJobRunLog) *apiModel.SchedulerJobRunLogDTO {
	return &apiModel.SchedulerJobRunLogDTO{ID: x.ID, RunId: x.RunID, JobKey: x.JobKey, Level: x.Level, Stage: x.Stage, Message: x.Message, CreatedAt: x.CreatedAt.Format(time.RFC3339)}
}

// JobList 定时任务列表
// @Summary 定时任务列表
// @Tags 监控
// @Produce json
// @Param data query apiModel.SchedulerJobListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/scheduler/job/list [GET]
func JobList(ctx context.Context, c *app.RequestContext) {
	var r apiModel.SchedulerJobListRequest
	if err := c.BindAndValidate(&r); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	xs, n, err := svc.SystemSchedulerSVC.List(ctx, int(r.Page), int(r.PageSize))
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	out := make([]*apiModel.SchedulerJobDTO, 0, len(xs))
	for _, x := range xs {
		out = append(out, dto(x))
	}
	R.SuccessWithData(ctx, c, &apiModel.SchedulerJobListResponse{List: out, Total: n})
}

// JobGet 定时任务详情
// @Summary 定时任务详情
// @Tags 监控
// @Produce json
// @Param data query apiModel.SchedulerJobIDRequest true "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/scheduler/job [GET]
func JobGet(ctx context.Context, c *app.RequestContext) {
	var r apiModel.SchedulerJobIDRequest
	if err := c.BindAndValidate(&r); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	x, err := svc.SystemSchedulerSVC.Get(ctx, r.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, dto(x))
}

// JobCreate 新增定时任务
// @Summary 新增定时任务
// @Tags 监控
// @Produce json
// @Param data body apiModel.SchedulerJobSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/scheduler/job [POST]
func JobCreate(ctx context.Context, c *app.RequestContext) {
	var r apiModel.SchedulerJobSaveRequest
	if err := c.BindAndValidate(&r); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	x := save(&r)
	if err := svc.SystemSchedulerSVC.Save(ctx, x); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, dto(x))
}

// JobUpdate 修改定时任务
// @Summary 修改定时任务
// @Tags 监控
// @Produce json
// @Param data body apiModel.SchedulerJobSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/scheduler/job [PUT]
func JobUpdate(ctx context.Context, c *app.RequestContext) {
	var r apiModel.SchedulerJobSaveRequest
	if err := c.BindAndValidate(&r); err != nil || r.ID == nil {
		R.InvalidParamRequestResponse(c, "id required")
		return
	}
	x := save(&r)
	if err := svc.SystemSchedulerSVC.Save(ctx, x); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, dto(x))
}

// JobRemove 删除定时任务
// @Summary 删除定时任务
// @Tags 监控
// @Produce json
// @Param data body apiModel.SchedulerJobIDRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/scheduler/job [DELETE]
func JobRemove(ctx context.Context, c *app.RequestContext) {
	var r apiModel.SchedulerJobIDRequest
	if err := c.BindAndValidate(&r); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := svc.SystemSchedulerSVC.Delete(ctx, r.ID); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// JobChangeStatus 启停定时任务
// @Summary 启停定时任务
// @Tags 监控
// @Produce json
// @Param data body apiModel.SchedulerJobIDRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/scheduler/job/status [PUT]
func JobChangeStatus(ctx context.Context, c *app.RequestContext) {
	var r apiModel.SchedulerJobIDRequest
	if err := c.BindAndValidate(&r); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	x, err := svc.SystemSchedulerSVC.Get(ctx, r.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	next := int16(0)
	if x.Status == 0 {
		next = 1
	}
	if err = svc.SystemSchedulerSVC.SetStatus(ctx, r.ID, next); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// JobRun 手动运行定时任务
// @Summary 手动运行定时任务
// @Tags 监控
// @Produce json
// @Param data body apiModel.SchedulerJobIDRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/scheduler/job/run [POST]
func JobRun(ctx context.Context, c *app.RequestContext) {
	var r apiModel.SchedulerJobIDRequest
	if err := c.BindAndValidate(&r); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := svc.SystemSchedulerSVC.Run(ctx, r.ID); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// RunRetry 使用运行快照重试定时任务
// @Summary 重试定时任务运行
// @Tags 监控
// @Produce json
// @Param data body apiModel.SchedulerJobIDRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/scheduler/run/retry [POST]
func RunRetry(ctx context.Context, c *app.RequestContext) {
	var r apiModel.SchedulerJobIDRequest
	if err := c.BindAndValidate(&r); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := svc.SystemSchedulerSVC.Retry(ctx, r.ID); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// RunStop 协作停止定时任务运行
// @Summary 停止定时任务运行
// @Tags 监控
// @Produce json
// @Param data body apiModel.SchedulerJobIDRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/scheduler/run/stop [POST]
func RunStop(ctx context.Context, c *app.RequestContext) {
	var r apiModel.SchedulerJobIDRequest
	if err := c.BindAndValidate(&r); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := svc.SystemSchedulerSVC.Stop(ctx, r.ID); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// RunList 定时任务运行记录
// @Summary 定时任务运行记录
// @Tags 监控
// @Produce json
// @Param data query apiModel.SchedulerJobRunListRequest true "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/scheduler/run/list [GET]
func RunList(ctx context.Context, c *app.RequestContext) {
	var r apiModel.SchedulerJobRunListRequest
	if err := c.BindAndValidate(&r); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	xs, n, err := svc.SystemSchedulerSVC.Runs(ctx, r.JobId, int(r.Page), int(r.PageSize))
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	out := make([]*apiModel.SchedulerJobRunDTO, 0, len(xs))
	for _, x := range xs {
		out = append(out, runDTO(x))
	}
	R.SuccessWithData(ctx, c, &apiModel.SchedulerJobRunListResponse{List: out, Total: n})
}

// RunLogList 定时任务执行日志
// @Summary 定时任务执行日志
// @Tags 监控
// @Produce json
// @Param data query apiModel.SchedulerJobRunLogListRequest true "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/scheduler/run/log/list [GET]
func RunLogList(ctx context.Context, c *app.RequestContext) {
	var r apiModel.SchedulerJobRunLogListRequest
	if err := c.BindAndValidate(&r); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	var afterID int64
	if r.AfterId != nil {
		afterID = *r.AfterId
	}
	xs, n, err := svc.SystemSchedulerSVC.RunLogs(ctx, r.RunId, afterID, int(r.Page), int(r.PageSize))
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	out := make([]*apiModel.SchedulerJobRunLogDTO, 0, len(xs))
	for _, x := range xs {
		out = append(out, runLogDTO(x))
	}
	R.SuccessWithData(ctx, c, &apiModel.SchedulerJobRunLogListResponse{List: out, Total: n})
}

// RunLogStream 定时任务执行日志实时订阅
// @Summary 定时任务执行日志实时订阅
// @Tags 监控
// @Produce text/event-stream
// @Param data query apiModel.SchedulerJobRunLogListRequest true "查询参数"
// @Success 200 {string} string
// @Router /system/scheduler/run/log/stream [GET]
func RunLogStream(ctx context.Context, c *app.RequestContext) {
	var r apiModel.SchedulerJobRunLogListRequest
	if err := c.BindAndValidate(&r); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	var afterID int64
	if r.AfterId != nil {
		afterID = *r.AfterId
	}
	if lastID, err := strconv.ParseInt(string(c.Request.Header.Peek("Last-Event-ID")), 10, 64); err == nil && lastID > afterID {
		afterID = lastID
	}
	live, unsubscribe := svc.SystemSchedulerSVC.SubscribeRunLogs(r.RunId)
	logs := make([]*dalModel.SystemJobRunLog, 0)
	for page := 1; ; page++ {
		batch, total, err := svc.SystemSchedulerSVC.RunLogs(ctx, r.RunId, afterID, page, 200)
		if err != nil {
			unsubscribe()
			R.Error(ctx, c, err)
			return
		}
		logs = append(logs, batch...)
		if int64(len(logs)) >= total || len(batch) == 0 {
			break
		}
	}
	defer unsubscribe()
	writer := sse.NewWriter(c)
	c.Response.Header.Set("Connection", "keep-alive")
	_ = streamRunLogs(ctx, writer, logs, live)
}

func streamRunLogs(ctx context.Context, writer *sse.Writer, replay []*dalModel.SystemJobRunLog, live <-chan *core.LogEntry) error {
	write := func(log *apiModel.SchedulerJobRunLogDTO) error {
		data, err := json.Marshal(log)
		if err != nil {
			return err
		}
		return writer.WriteEvent(strconv.FormatInt(log.ID, 10), "log", data)
	}
	cursor := int64(0)
	for _, log := range replay {
		if log == nil || log.ID <= cursor {
			continue
		}
		if err := write(runLogDTO(log)); err != nil {
			return err
		}
		cursor = log.ID
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case log, ok := <-live:
			if !ok {
				return nil
			}
			if log == nil || log.ID <= cursor {
				continue
			}
			if err := write(&apiModel.SchedulerJobRunLogDTO{ID: log.ID, RunId: log.RunID, JobKey: log.JobKey, Level: log.Level, Stage: log.Stage, Message: log.Message, CreatedAt: log.CreatedAt.Format(time.RFC3339)}); err != nil {
				return err
			}
			cursor = log.ID
		}
	}
}

// JobKeys 可用任务键
// @Summary 可用任务键
// @Tags 监控
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/scheduler/job/keys [GET]
func JobKeys(ctx context.Context, c *app.RequestContext) {
	R.SuccessWithData(ctx, c, &apiModel.SchedulerJobKeysResponse{List: svc.SystemSchedulerSVC.Keys()})
}
