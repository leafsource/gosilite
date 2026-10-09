package monitor

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	monitor "gosilite/api/model/framework/monitor"
	cacheMonitor "gosilite/application/framework/service/monitor"
	R "gosilite/pkg/utils/response"
)

// PrefixList 缓存前缀列表。
// @Summary 缓存前缀列表
// @Description 缓存前缀列表接口
// @Tags 监控
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/cache/prefix/list [GET]
func PrefixList(ctx context.Context, c *app.RequestContext) {
	list, err := cacheMonitor.SystemCacheSVC.PrefixList(ctx)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toCachePrefixList(list))
}

// KeyList 前缀下 key 列表。
// @Summary 缓存键列表
// @Description 缓存键列表接口
// @Tags 监控
// @Produce json
// @Param data query monitor.CacheKeyListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/cache/key/list [GET]
func KeyList(ctx context.Context, c *app.RequestContext) {
	var req monitor.CacheKeyListRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	list, err := cacheMonitor.SystemCacheSVC.KeyList(ctx, req.Prefix)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toCacheKeyList(list))
}

// Value 单 key 内容。
// @Summary 读取缓存值
// @Description 读取缓存值接口
// @Tags 监控
// @Produce json
// @Param data query monitor.CacheValueRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/cache/value [GET]
func Value(ctx context.Context, c *app.RequestContext) {
	var req monitor.CacheValueRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	v, err := cacheMonitor.SystemCacheSVC.Value(ctx, req.Key, req.Field)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toCacheValue(v))
}

// DeleteKey 删除单 key 或 hash field。
// @Summary 删除缓存键
// @Description 删除缓存键接口
// @Tags 监控
// @Produce json
// @Param data body monitor.CacheKeyDeleteRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/cache/key/delete [POST]
func DeleteKey(ctx context.Context, c *app.RequestContext) {
	var req monitor.CacheKeyDeleteRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := cacheMonitor.SystemCacheSVC.DeleteKey(ctx, req.Key, req.Field); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(1))
}

// ClearPrefix 清空前缀。
// @Summary 清空前缀缓存
// @Description 清空前缀缓存接口
// @Tags 监控
// @Produce json
// @Param data body monitor.CachePrefixClearRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/cache/prefix/clear [POST]
func ClearPrefix(ctx context.Context, c *app.RequestContext) {
	var req monitor.CachePrefixClearRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	n, err := cacheMonitor.SystemCacheSVC.ClearPrefix(ctx, req.Prefix)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(n))
}
