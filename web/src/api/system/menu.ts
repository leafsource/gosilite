import { request } from '@/utils/request';

// 菜单 DTO（MenuDTO）
export interface MenuDTO {
  id: string;
  parentId: string;
  menuName: string;
  menuType: string;
  path: string;
  component: string;
  perms: string;
  icon: string;
  orderNum: number;
  visible: boolean;
  status: number;
  /** 按钮操作类型（sys_action_type），仅 menuType=F */
  actionType: string;
}

// 菜单树节点（MenuNode）
export interface MenuNode {
  id: string;
  parentId: string;
  label: string;
  menuType: string;
  perms: string;
  orderNum: number;
  visible: boolean;
  status: number;
  path: string;
  component: string;
  icon: string;
  /** 按钮操作类型（sys_action_type），仅 menuType=F */
  actionType: string;
  children?: MenuNode[];
}

export interface MenuSaveParams {
  id?: string;
  parentId: string;
  menuName: string;
  menuType: string;
  path?: string;
  component?: string;
  perms?: string;
  icon?: string;
  orderNum: number;
  visible: boolean;
  status: number;
  /** 按钮操作类型（sys_action_type），仅 menuType=F 必填 */
  actionType?: string;
}

// 树选择 / 树勾选统一结构
export interface MenuTreeItem {
  value: string;
  label: string;
  menuType?: string;
  children?: MenuTreeItem[];
}

const Api = {
  Tree: '/system/menu/tree',
  List: '/system/menu/list',
  Base: '/system/menu',
};

export function fetchMenuTree() {
  return request.get<{ list: MenuNode[] }>({ url: Api.Tree });
}

export function fetchMenuList() {
  return request.get<{ list: MenuDTO[]; total: number }>({ url: Api.List });
}

export function createMenu(data: MenuSaveParams) {
  return request.post<MenuDTO>({ url: Api.Base, data });
}

export function updateMenu(data: MenuSaveParams) {
  return request.put<MenuDTO>({ url: Api.Base, data });
}

export function removeMenu(id: string) {
  return request.delete<number>({ url: Api.Base, data: { id }, params: { id } });
}
