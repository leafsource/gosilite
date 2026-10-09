import cloneDeep from 'lodash/cloneDeep';

import type { RouteItem } from '@/api/model/permissionModel';
import {
  BLANK_LAYOUT,
  IFRAME,
  LAYOUT,
  NOT_FOUND_COMPONENT,
  PAGE_NOT_FOUND_ROUTE,
  PARENT_LAYOUT,
} from '@/utils/route/constant';

const LayoutMap = new Map<string, () => Promise<typeof import('*.vue')>>();

LayoutMap.set('LAYOUT', LAYOUT);
LayoutMap.set('BLANK', BLANK_LAYOUT);
LayoutMap.set('IFRAME', IFRAME);

let dynamicViewsModules: Record<string, () => Promise<Recordable>>;

/**
 * 动态引入路由组件。
 * @param nested true = 已在某个 LAYOUT 子树内；此时再遇到 layout 只透传 router-view，避免双层侧栏
 */
function asyncImportRoute(routes: RouteItem[] | undefined, nested = false) {
  dynamicViewsModules = dynamicViewsModules || import.meta.glob('../../pages/**/*.vue');
  if (!routes) return;

  routes.forEach((item) => {
    const { component, name } = item;
    const { children } = item;

    if (component) {
      const upper = String(component).toUpperCase();
      if (upper === 'LAYOUT') {
        // 根级目录用完整 Layout；嵌套目录只透传，防止内嵌第二个 layout
        item.component = nested ? ROUTER_PASS : LAYOUT;
      } else if (upper === 'BLANK') {
        item.component = BLANK_LAYOUT;
      } else if (upper === 'IFRAME') {
        item.component = IFRAME;
      } else {
        item.component = dynamicImport(dynamicViewsModules, String(component));
      }
    } else if (name) {
      item.component = PARENT_LAYOUT();
    }

    children && asyncImportRoute(children, true);
  });
}

/** 嵌套 M 目录（如 /system/log）：仅渲染子路由，不套完整 Layout */
const ROUTER_PASS = () => import('@/layouts/router-pass.vue');

function dynamicImport(dynamicViewsModules: Record<string, () => Promise<Recordable>>, component: string) {
  const keys = Object.keys(dynamicViewsModules);
  const matchKeys = keys.filter((key) => {
    const k = key.replace('../../pages', '');
    const startFlag = component.startsWith('/');
    const endFlag = component.endsWith('.vue') || component.endsWith('.tsx');
    const startIndex = startFlag ? 0 : 1;
    const lastIndex = endFlag ? k.length : k.lastIndexOf('.');
    return k.substring(startIndex, lastIndex) === component;
  });
  if (matchKeys?.length === 1) {
    const matchKey = matchKeys[0];
    return dynamicViewsModules[matchKey];
  }
  if (matchKeys?.length > 1) {
    throw new Error(
      'Please do not create `.vue` and `.TSX` files with the same file name in the same hierarchical directory under the views folder. This will cause dynamic introduction failure',
    );
  }
  console.warn(`[router] component not found in pages, fallback 404: ${component}`);
  // 缺页 → 404（不是 500）
  return NOT_FOUND_COMPONENT;
}

// 将菜单对象变成路由对象
export function transformObjectToRoute<T = RouteItem>(routeList: RouteItem[]): T[] {
  routeList.forEach((route) => {
    const component = route.component as string;

    if (component) {
      if (String(component).toUpperCase() === 'LAYOUT') {
        route.component = LAYOUT;
      } else {
        // 无 LAYOUT 目录的业务节点：包一层空 path 的 LAYOUT，子节点承载真实 path
        route.children = [cloneDeep(route)];
        route.component = LAYOUT;
        route.name = `${route.name}Parent`;
        route.path = '';
        route.meta = route.meta || {};
      }
    } else {
      throw new Error('component is undefined');
    }

    // 子节点均在外层已套 Layout 的壳内：再遇 layout 一律透传
    route.children && asyncImportRoute(route.children, true);
  });

  return [PAGE_NOT_FOUND_ROUTE, ...routeList] as unknown as T[];
}

// 统一 path 编解码
export function normalizePath(path: string): string {
  if (!path || !path.includes('%')) return path;
  try {
    return decodeURIComponent(path);
  } catch {
    return path;
  }
}