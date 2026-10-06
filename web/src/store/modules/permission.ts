import cloneDeep from 'lodash/cloneDeep';
import { defineStore } from 'pinia';
import type { RouteRecordRaw } from 'vue-router';

import type { RouteItem } from '@/api/model/permissionModel';
import { getMenuList } from '@/api/permission';
import router, { allRoutes, homepageRouterList } from '@/router';
import { store } from '@/store';
import { transformObjectToRoute } from '@/utils/route';
import { PAGE_NOT_FOUND_ROUTE } from '@/utils/route/constant';

const normalize = (path: string) => (path.startsWith('/') ? path : `/${path}`);

export const usePermissionStore = defineStore('permission', {
  state: () => ({
    whiteListRouters: ['/login'],
    routers: [] as Array<RouteRecordRaw>,
    removeRoutes: [] as Array<RouteRecordRaw>,
    // 待注册的动态路由（后端菜单中静态模块未覆盖的 path + 404）
    asyncRoutes: [] as Array<RouteRecordRaw>,
    // 侧栏业务菜单：以后端菜单（已按角色过滤）为唯一来源
    menuRoutes: [] as Array<RouteRecordRaw>,
  }),
  actions: {
    async initRoutes() {
      // 侧栏 = 首页（静态，权威） + 后端菜单（按角色过滤）；
      // fixedRouterList（system/user/result 等静态业务模块）不再进侧栏，
      // 但路由仍注册、URL 可直达，越权由 API 403 与 v-permission 兜底。
      this.routers = cloneDeep([...homepageRouterList, ...this.menuRoutes]);
    },
    async buildAsyncRoutes() {
      try {
        const rawList: Array<RouteItem> = (await getMenuList()).list;

        // 注册：与静态路由同 path 的交由静态模块提供组件，避免覆盖
        const staticPaths = new Set(allRoutes.map((route) => normalize(route.path)));
        const toRegister = rawList.filter((route) => !staticPaths.has(normalize(route.path)));
        this.asyncRoutes = transformObjectToRoute(cloneDeep(toRegister));

        // 侧栏：后端菜单全量展示（信任后端过滤结果），仅去掉与静态首页重复的根
        const homePaths = new Set(homepageRouterList.map((route) => normalize(route.path)));
        const toDisplay = rawList.filter((route) => !homePaths.has(normalize(route.path)));
        this.menuRoutes = transformObjectToRoute(cloneDeep(toDisplay)).filter(
          (route) => route.name !== PAGE_NOT_FOUND_ROUTE.name,
        );

        await this.initRoutes();
        return this.asyncRoutes;
      } catch (error) {
        throw new Error("Can't build routes", error as ErrorOptions);
      }
    },
    async restoreRoutes() {
      // 不需要在此额外调用initRoutes更新侧边导肮内容，在登录后asyncRoutes为空会调用
      this.asyncRoutes.forEach((item: RouteRecordRaw) => {
        if (item.name) {
          router.removeRoute(item.name);
        }
      });
      this.asyncRoutes = [];
      this.menuRoutes = [];
      this.routers = [];
    },
  },
});

export function getPermissionStore() {
  return usePermissionStore(store);
}
