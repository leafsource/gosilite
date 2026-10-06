package syscfg

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	syscfg "gosilite/api/model/framework/syscfg"
	"gosilite/application/framework/service/config"
	R "gosilite/pkg/utils/response"
)

// GlobalList 全局参数（config_type=global，免登录）
// @router /system/config/global [GET]
func GlobalList(ctx context.Context, c *app.RequestContext) {
	list, err := config.SystemConfigSVC.ListGlobal(ctx)
	if err != nil {
		// 错误必须走 R.Error（非 200 信封），否则前端按成功处理、loadGlobal 不再重试
		R.Error(ctx, c, err)
		return
	}
	resp := &syscfg.ConfigListResponse{List: make([]*syscfg.ConfigDTO, 0, len(list)), Total: int64(len(list))}
	for _, m := range list {
		resp.List = append(resp.List, configDTO(m))
	}
	R.SuccessWithData(ctx, c, resp)
}
