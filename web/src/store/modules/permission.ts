import cloneDeep from 'lodash/cloneDeep';
import { defineStore } from 'pinia';
import type { RouteRecordRaw } from 'vue-router';

import type { RouteItem } from '@/api/model/permissionModel';
import { getMenuList } from '@/api/permission';
import router, { homepageRouterList } from '@/router';
import { store } from '@/store';
import { transformObjectToRoute } from '@/utils/route';
import { PAGE_NOT_FOUND_ROUTE } from '@/utils/route/constant';

const normalize = (path: string) => (path.startsWith('/') ? path : `/${path}`);

/** 静态独占 path：动态注册时跳过（登录 + 首页整树） */
function reservedPaths(): Set<string> {
  const out = new Set<string>(['/login']);
  const walk = (list: RouteRecordRaw[], prefix = '') => {
    list.forEach((r) => {
      const p = r.path.startsWith('/') ? normalize(r.path) : normalize(`${prefix}/${r.path}`);
      out.add(p);
      if (r.children) walk(r.children as RouteRecordRaw[], p);
    });
  };
  walk(homepageRouterList);
  return out;
}

export const usePermissionStore = defineStore('permission', {
  state: () => ({
    whiteListRouters: ['/login'],
    routers: [] as Array<RouteRecordRaw>,
    removeRoutes: [] as Array<RouteRecordRaw>,
    /** 动态注册（含 404） */
    asyncRoutes: [] as Array<RouteRecordRaw>,
    /** 侧栏 = 后端菜单（角色过滤）；首页根另行拼接 */
    menuRoutes: [] as Array<RouteRecordRaw>,
  }),
  actions: {
    async initRoutes() {
      this.routers = cloneDeep([...homepageRouterList, ...this.menuRoutes]);
    },
    async buildAsyncRoutes() {
      try {
        const rawList: Array<RouteItem> = (await getMenuList()).list;
        const reserved = reservedPaths();
        const homeRoots = new Set(homepageRouterList.map((route) => normalize(route.path)));

        // 业务路由唯一注册源 = 后端菜单；跳过登录/首页静态独占
        const toRegister = rawList.filter((route) => !reserved.has(normalize(route.path)));
        this.asyncRoutes = transformObjectToRoute(cloneDeep(toRegister));

        // 侧栏：后端全量（已按角色过滤），去掉与静态首页重复的根
        const toDisplay = rawList.filter((route) => !homeRoots.has(normalize(route.path)));
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
