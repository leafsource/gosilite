package syscfgglobal

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	syscfgglobal "gosilite/api/model/framework/syscfgglobal"
	"gosilite/application/framework/service/config"
	R "gosilite/pkg/utils/response"
)

// GlobalList 全局参数（config_type=global，免登录）。
// 本包由独立 IDL 文件 idl/framework/system_config_global.thrift 生成，
// rootMw 不挂登录（middleware.go 中保持 return nil）——spec: config-consumption S2.2。
// @Summary 全局参数
// @Description 全局参数接口
// @Tags 系统配置
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/config/global [GET]
func GlobalList(ctx context.Context, c *app.RequestContext) {
	list, err := config.SystemConfigSVC.ListGlobal(ctx)
	if err != nil {
		// 错误必须走 R.Error（非 200 信封），否则前端按成功处理、loadGlobal 不再重试
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, &syscfgglobal.ConfigGlobalListResponse{List: toConfigGlobalDTOs(list), Total: int64(len(list))})
}
