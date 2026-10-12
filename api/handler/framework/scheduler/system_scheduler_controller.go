package scheduler

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	apiModel "gosilite/api/model/framework/scheduler"
	dalModel "gosilite/application/framework/dal/model"
	svc "gosilite/application/framework/service/scheduler"
	R "gosilite/pkg/utils/response"
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
	z := &apiModel.SchedulerJobRunDTO{ID: x.ID, JobId: x.JobID, JobKey: x.JobKey, Trigger: x.Trigger, RetryCount: x.Attempt - 1, Status: x.Status, TriggeredAt: x.TriggeredAt.Format(time.RFC3339), Output: x.Output, Error: x.Error}
	if x.FinishedAt != nil {
		z.FinishedAt = x.FinishedAt.Format(time.RFC3339)
	}
	return z
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

// JobKeys 可用任务键
// @Summary 可用任务键
// @Tags 监控
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/scheduler/job/keys [GET]
func JobKeys(ctx context.Context, c *app.RequestContext) {
	R.SuccessWithData(ctx, c, &apiModel.SchedulerJobKeysResponse{List: svc.SystemSchedulerSVC.Keys()})
}
