package menu

import (
	menu "gosilite/api/model/framework/menu"
	"gosilite/application/framework/dal/model"
	"sort"
	"strings"
)

// toRoutes 将菜单列表转换为路由树（根：parent_id 为 0 或 -1）
func toRoutes(menus []*model.SystemMenu) []*menu.Route {
	menuMap := make(map[int64]*model.SystemMenu)
	for _, m := range menus {
		menuMap[m.ID] = m
	}
	return buildRouteTree(menus, menuMap, 0)
}

func isRootParent(parentID int64) bool {
	return parentID == 0 || parentID == -1
}

func buildRouteTree(menus []*model.SystemMenu, menuMap map[int64]*model.SystemMenu, parentID int64) []*menu.Route {
	var routes []*menu.Route

	for _, menuData := range menus {
		if parentID == 0 {
			if !isRootParent(menuData.ParentID) {
				continue
			}
		} else if menuData.ParentID != parentID {
			continue
		}

		// 检查是否为外部链接
		isExternal := false
		if len(menuData.Path) >= 7 && menuData.Path[:7] == "http://" {
			isExternal = true
		} else if len(menuData.Path) >= 8 && menuData.Path[:8] == "https://" {
			isExternal = true
		}

		// 构建完整路径
		fullPath := menuData.Path
		if !isExternal {
			fullPath = buildFullPath(menuData, menuMap)
		}

		// 构建路由
		component := menuData.Component
		if menuData.MenuType == "parent" {
			component = menuData.ViewMode
		}

		// 创建Meta结构体
		meta := &menu.Meta{
			Title:            menuData.Title,
			Icon:             menuData.Icon,
			Hidden:           menuData.Visible == "hide",
			Single:           menuData.MenuType == "parent" && len(getChildMenus(menus, menuData.ID)) == 1,
			OrderNo:          int32(menuData.Sort),
			KeepAlive:        menuData.IsCache != "1", // 0-开启，1-关闭
			HiddenBreadcrumb: menuData.HiddenBreadcrumb,
			Expanded:         menuData.Expanded,
		}

		// 处理iframe相关属性
		if menuData.ViewMode == "iframe" || menuData.ViewMode == "iframe_blank" {
			meta.FrameSrc = menuData.Component // 使用原始component作为frameSrc
			if menuData.ViewMode == "iframe_blank" {
				meta.FrameBlank = true
			}
		}

		// 构建路由
		// Name 不从数据库取：路由 name 由前端根据 path 自动生成（与页面 SFC name 约定对齐）
		route := &menu.Route{
			Path:      fullPath,
			Component: component,
			Meta:      meta,
		}

		// 递归构建子路由
		children := buildRouteTree(menus, menuMap, menuData.ID)
		if len(children) > 0 {
			route.Children = children

			// 显式 redirect 优先；否则目录默认指向排序最前的子路由（避免点目录白屏）
			if menuData.Redirect != "" {
				route.Redirect = menuData.Redirect
			} else if len(children) == 1 || menuData.MenuType == "M" {
				route.Redirect = children[0].Path
			}
		}

		routes = append(routes, route)
	}

	// 按排序字段排序
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Meta.OrderNo < routes[j].Meta.OrderNo
	})

	return routes
}

func buildFullPath(menu *model.SystemMenu, menuMap map[int64]*model.SystemMenu) string {
	// 外部链接直接返回
	if len(menu.Path) >= 7 && menu.Path[:7] == "http://" {
		return menu.Path
	} else if len(menu.Path) >= 8 && menu.Path[:8] == "https://" {
		return menu.Path
	}

	// 根菜单直接返回
	if menu.ParentID == -1 || menu.ParentID == 0 {
		return normalizeMenuPath(menu.Path)
	}

	parent, exists := menuMap[menu.ParentID]
	if !exists {
		return normalizeMenuPath(menu.Path)
	}

	// 递归构建完整路径
	parentPath := buildFullPath(parent, menuMap)

	// 子级 path 允许写成 `log` 或 `/log`：拼接前去掉前导 /，避免 /system//log
	child := normalizeMenuPath(menu.Path)
	child = strings.TrimPrefix(child, "/")

	// 子级若已是含父前缀的完整路径，直接返回
	if child != "" && parentPath != "/" && strings.HasPrefix(child, strings.TrimPrefix(parentPath, "/")+"/") {
		return "/" + child
	}
	if parentPath == "/" || parentPath == "" {
		return "/" + child
	}
	return parentPath + "/" + child
}

// normalizeMenuPath 去空白；根返回 "/"。
func normalizeMenuPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "/"
	}
	return p
}

func getChildMenus(menus []*model.SystemMenu, parentID int64) []*model.SystemMenu { // 返回值切片
	var children []*model.SystemMenu
	for _, menu := range menus {
		if menu.ParentID == parentID {
			children = append(children, menu)
		}
	}
	return children
}
