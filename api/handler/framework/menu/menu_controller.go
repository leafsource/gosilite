package menu

import (
	"context"
	menuService "gosilite/application/framework/service/menu"
	R "gosilite/pkg/utils/response"

	menu "gosilite/api/model/framework/menu"
	"gosilite/api/utils"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// GetRoutes .
// @router /menu/routes [GET]
func GetRoutes(ctx context.Context, c *app.RequestContext) {
	var err error

	resp := &menu.RouteResponse{}
	// 按当前登录用户角色过滤（admin 全量），剔除按钮行
	menus, err := menuService.SystemMenuSVC.VisibleMenus(ctx)
	if err != nil {
		R.Failed(ctx, c, consts.StatusOK, err)
		return
	}
	resp.Routes = utils.ConvertToRoutes(menus)

	R.SuccessWithData(ctx, c, resp)
}
