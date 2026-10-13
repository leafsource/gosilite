package menu

import (
	"context"
	menuService "gosilite/application/framework/service/menu"
	R "gosilite/pkg/utils/response"

	menu "gosilite/api/model/framework/menu"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// GetRoutes .
// @Summary 获取前端路由
// @Description 获取前端路由接口
// @Tags 菜单
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /menu/routes [GET]
func GetRoutes(ctx context.Context, c *app.RequestContext) {
	var err error

	resp := &menu.RouteResponse{}
	// 按当前登录用户角色过滤（admin 全量），剔除按钮行
	menus, err := menuService.SystemMenuSVC.VisibleMenus(ctx)
	if err != nil {
		R.Failed(ctx, c, consts.StatusOK, err)
		return
	}
	resp.Routes = toRoutes(menus)

	R.SuccessWithData(ctx, c, resp)
}
