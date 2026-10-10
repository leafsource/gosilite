package storage

import (
	"context"
	"io"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	storageModel "gosilite/api/model/framework/storage"
	storageService "gosilite/application/framework/service/storage"
	R "gosilite/pkg/utils/response"
)

// GetInfo 存储概览（配置 + 探活 + 用量）。
// @Summary 存储概览
// @Description 存储概览接口
// @Tags 系统管理
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/storage/info [GET]
func GetInfo(ctx context.Context, c *app.RequestContext) {
	vo, err := storageService.SystemStorageSVC.Info(ctx)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toStorageInfoResponse(vo))
}

// GetObjects 对象列表（分页 + 前缀 + key 模糊）。
// @Summary 存储对象列表
// @Description 存储对象列表接口
// @Tags 系统管理
// @Produce json
// @Param data query storageModel.StorageObjectListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/storage/list [GET]
func GetObjects(ctx context.Context, c *app.RequestContext) {
	var req storageModel.StorageObjectListRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	vos, total, err := storageService.SystemStorageSVC.List(ctx, int(req.Page), int(req.PageSize), req.Prefix, req.Keyword)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toStorageObjectListResponse(vos, total))
}

// DeleteObject 删除存储对象。
// @Summary 删除存储对象
// @Description 删除存储对象接口
// @Tags 系统管理
// @Produce json
// @Param data query storageModel.StorageObjectDeleteRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/storage/object [DELETE]
func DeleteObject(ctx context.Context, c *app.RequestContext) {
	var req storageModel.StorageObjectDeleteRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}

	if err := storageService.SystemStorageSVC.Delete(ctx, req.Key); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// Upload 上传测试对象（multipart 字段 file，≤10MB）。
// @Summary 上传存储对象
// @Description 上传存储对象接口
// @Tags 系统管理
// @Produce json
// @Param file formData file true "上传文件"
// @Success 200 {object} map[string]interface{}
// @Router /system/storage/upload [POST]
func Upload(ctx context.Context, c *app.RequestContext) {
	file, err := c.FormFile("file")
	if err != nil {
		hlog.CtxErrorf(ctx, "get upload file failed, err=%v", err)
		R.InvalidParamRequestResponse(c, "missing file field")
		return
	}
	src, err := file.Open()
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	defer src.Close()

	// 多读 1 字节：超限交给 Service 校验，内存有界
	content, err := io.ReadAll(io.LimitReader(src, int64(storageService.MaxUploadBytes)+1))
	if err != nil {
		R.Error(ctx, c, err)
		return
	}

	vo, err := storageService.SystemStorageSVC.Upload(ctx, file.Filename, file.Header.Get("Content-Type"), content)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toStorageUploadResponse(vo))
}
