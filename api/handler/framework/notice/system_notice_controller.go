package notice

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	notice "gosilite/api/model/framework/notice"
	noticeSvc "gosilite/application/framework/service/notice"
	R "gosilite/pkg/utils/response"
)

// 路由与请求/响应 DTO 均由 hz 从 idl/framework/system_notice.thrift 生成
// （api/model/framework/notice + api/router/framework/notice），
// 本文件只保留 Handler 实现（绑定 → 调 Service → 回包）。
// spec: docs/compose/spec/notice.md S2.2

// List 分页列表
// @Summary 通知列表
// @Description 通知列表接口
// @Tags 通知
// @Produce json
// @Param data query notice.NoticeListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/notice/list [GET]
func List(ctx context.Context, c *app.RequestContext) {
	var req notice.NoticeListRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	var noticeType, status *int16
	if req.NoticeType != nil {
		v := int16(*req.NoticeType)
		noticeType = &v
	}
	if req.Status != nil {
		v := int16(*req.Status)
		status = &v
	}
	list, total, err := noticeSvc.SystemNoticeSVC.Page(ctx, int(req.Page), int(req.PageSize), req.Keyword, noticeType, status)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, &notice.NoticeListResponse{List: toNoticeDTOs(list), Total: total})
}

// Get 详情
// @Summary 通知
// @Description 通知接口
// @Tags 通知
// @Produce json
// @Param data query notice.NoticeIdRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/notice [GET]
func Get(ctx context.Context, c *app.RequestContext) {
	var req notice.NoticeIdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m, err := noticeSvc.SystemNoticeSVC.Get(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toNoticeDTO(m))
}

// Create 新增
// @Summary 通知
// @Description 通知接口
// @Tags 通知
// @Produce json
// @Param data body notice.NoticeSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/notice [POST]
func Create(ctx context.Context, c *app.RequestContext) {
	var req notice.NoticeSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m := fromNoticeSaveRequest(&req)
	if err := noticeSvc.SystemNoticeSVC.Create(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toNoticeDTO(m))
}

// Update 更新
// @Summary 通知
// @Description 通知接口
// @Tags 通知
// @Produce json
// @Param data body notice.NoticeSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/notice [PUT]
func Update(ctx context.Context, c *app.RequestContext) {
	var req notice.NoticeSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if req.ID == nil {
		R.InvalidParamRequestResponse(c, "id required")
		return
	}
	m := fromNoticeSaveRequest(&req)
	m.ID = *req.ID
	if err := noticeSvc.SystemNoticeSVC.Update(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	// 回包落库后的完整行（Save 为选择性更新，请求体不含 publisher/时间列）
	saved, err := noticeSvc.SystemNoticeSVC.Get(ctx, m.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toNoticeDTO(saved))
}

// Remove 删除
// @Summary 通知
// @Description 通知接口
// @Tags 通知
// @Produce json
// @Param data body notice.NoticeIdRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/notice [DELETE]
func Remove(ctx context.Context, c *app.RequestContext) {
	var req notice.NoticeIdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := noticeSvc.SystemNoticeSVC.Delete(ctx, req.ID); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// Inbox 当前用户通知收件箱（login-only；列表不含 content）
// @Summary 通知收件箱
// @Description 通知收件箱接口
// @Tags 通知
// @Produce json
// @Param data query notice.NoticeInboxRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /notice/inbox [GET]
func Inbox(ctx context.Context, c *app.RequestContext) {
	var req notice.NoticeInboxRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	items, total, unread, err := noticeSvc.SystemNoticeSVC.Inbox(ctx, int(req.Page), int(req.PageSize), req.Keyword)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, &notice.NoticeInboxListResponse{
		List: toInboxItemDTOs(items), Total: total, Unread: unread,
	})
}

// Detail 通知详情（login-only；打开即自动已读）
// @Summary 详情
// @Description 详情接口
// @Tags 通知
// @Produce json
// @Param data query notice.NoticeDetailRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /notice/detail [GET]
func Detail(ctx context.Context, c *app.RequestContext) {
	var req notice.NoticeDetailRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m, err := noticeSvc.SystemNoticeSVC.Detail(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toNoticeDetailDTO(m))
}

// UnreadCount 未读数（login-only；Header 角标）
// @Summary 通知未读数
// @Description 通知未读数接口
// @Tags 通知
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /notice/unread-count [GET]
func UnreadCount(ctx context.Context, c *app.RequestContext) {
	count, err := noticeSvc.SystemNoticeSVC.UnreadCount(ctx)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, &notice.UnreadCountResponse{Count: count})
}

// Read 标记已读（login-only；id 缺省=全部已读）
// @Summary 标记已读
// @Description 标记已读接口
// @Tags 通知
// @Produce json
// @Param data body notice.NoticeReadRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /notice/read [POST]
func Read(ctx context.Context, c *app.RequestContext) {
	var req notice.NoticeReadRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	n, err := noticeSvc.SystemNoticeSVC.Read(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(n))
}
