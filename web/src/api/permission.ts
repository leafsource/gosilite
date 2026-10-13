import type { MenuListResult, RouteItem } from '@/api/model/permissionModel';
import { allRoutes } from '@/router';
import { request } from '@/utils/request';

const Api = {
  Routes: '/menu/routes',
};

interface MenuMeta {
  title?: string;
  icon?: string;
  hidden?: boolean;
  single?: boolean;
  orderNo?: number;
  keepAlive?: boolean;
  hiddenBreadcrumb?: boolean;
  expanded?: boolean;
  frameSrc?: string;
  frameBlank?: boolean;
}

interface MenuRouteDto {
  path: string;
  component?: string;
  redirect?: string;
  name?: string;
  meta?: MenuMeta;
  children?: MenuRouteDto[];
}

// 由完整 path 生成路由 name（PascalCase），与 starter 静态路由命名约定一致（/dashboard/base -> DashboardBase）
function genRouteName(fullPath: string): string {
  const cleaned = fullPath.replace(/^https?:\/\//, '');
  const name = cleaned
    .split('/')
    .filter(Boolean)
    .map((seg) => seg.replace(/[^a-z0-9]+(.)?/gi, (_, c: string | undefined) => (c ? c.toUpperCase() : '')))
    .map((seg) => seg.charAt(0).toUpperCase() + seg.slice(1))
    .join('');
  return name || 'Route';
}

function dedupeNames(list: RouteItem[], seen: Set<string>) {
  list.forEach((item) => {
    let name = item.name;
    let seq = 2;
    while (seen.has(name)) {
      name = `${item.name}_${seq}`;
      seq += 1;
    }
    item.name = name;
    seen.add(name);
    if (item.children) {
      dedupeNames(item.children, seen);
    }
  });
}

// 后端 meta.title 为字符串，统一转成本地化结构供菜单/面包屑渲染
function normalizeRoute(dto: MenuRouteDto): RouteItem {
  const meta = { ...dto.meta } as RouteItem['meta'];
  if (typeof meta.title === 'string') {
    meta.title = { zh_CN: meta.title, en_US: meta.title };
  }
  const item: RouteItem = {
    path: dto.path,
    // 后端不下发路由 name，由前端根据 path 自动生成（keep-alive 需要时在页面 SFC 中 defineOptions 保持同名）
    name: genRouteName(dto.path),
    component: dto.component,
    redirect: dto.redirect,
    meta,
  };
  if (dto.children && dto.children.length > 0) {
    item.children = dto.children.map(normalizeRoute);
  }
  return item;
}

export async function getMenuList(): Promise<MenuListResult> {
  const res = await request.get<{ routes: MenuRouteDto[] }>({ url: Api.Routes });
  const list = (res.routes || []).map(normalizeRoute);
  dedupeNames(list, new Set(allRoutes.map((route) => route.name as string).filter(Boolean)));
  return { list };
}
