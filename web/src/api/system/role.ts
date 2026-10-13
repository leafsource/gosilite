import { request } from '@/utils/request';

// 角色 DTO（RoleDTO）
export interface RoleDTO {
  id: string;
  name: string;
  code: string;
  description: string;
  state: string;
  dataScope: number;
  orderNum: number;
}

export interface RoleQuery {
  page?: number;
  pageSize?: number;
  keyword?: string;
}

export interface RoleSaveParams {
  id?: string;
  name: string;
  code: string;
  description?: string;
  state: string;
  dataScope: number;
  orderNum: number;
}

// ===== 授权快照（GET /system/role/authorize） =====

export type RoleId = string;

export interface RolePermItem {
  id: RoleId;
  menuName: string;
  perms: string;
  checked: boolean;
}

export interface RoleAuthMenuNode {
  id: RoleId;
  parentId: RoleId;
  menuName: string;
  menuType: string;
  icon: string;
  orderNum: number;
  checked: boolean;
  perms: RolePermItem[];
  children?: RoleAuthMenuNode[];
}

export interface RoleAuthorizeSnapshot {
  dataScope: number;
  menus: RoleAuthMenuNode[];
  deptIds: RoleId[];
}

export interface RoleAuthorizeSaveParams {
  roleId: string;
  menuIds: RoleId[];
  deptIds: RoleId[];
  dataScope: number;
}

const Api = {
  List: '/system/role/list',
  Base: '/system/role',
  Authorize: '/system/role/authorize',
};

export function fetchRoleList(params?: RoleQuery) {
  return request.get<{ list: RoleDTO[]; total: number }>({ url: Api.List, params });
}

export function createRole(data: RoleSaveParams) {
  return request.post<RoleDTO>({ url: Api.Base, data });
}

export function updateRole(data: RoleSaveParams) {
  return request.put<RoleDTO>({ url: Api.Base, data });
}

export function removeRole(id: string) {
  return request.delete<number>({ url: Api.Base, data: { id }, params: { id } });
}

export function fetchRoleAuthorize(id: string) {
  return request.get<RoleAuthorizeSnapshot>({ url: Api.Authorize, params: { id } });
}

export function saveRoleAuthorize(data: RoleAuthorizeSaveParams) {
  return request.put<number>({ url: Api.Authorize, data });
}
