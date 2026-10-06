import { request } from '@/utils/request';

// 角色 DTO（RoleDTO）
export type RoleDTO = {
  id: number;
  name: string;
  code: string;
  description: string;
  state: string;
  dataScope: number;
  orderNum: number;
};

export type RoleQuery = {
  page?: number;
  pageSize?: number;
  keyword?: string;
};

export type RoleSaveParams = {
  id?: number;
  name: string;
  code: string;
  description?: string;
  state: string;
  dataScope: number;
  orderNum: number;
};

export type RoleMenuParams = {
  roleId: number;
  menuIds: number[];
};

export type RoleDeptParams = {
  roleId: number;
  deptIds: number[];
};

const Api = {
  List: '/system/role/list',
  Base: '/system/role',
  Menu: '/system/role/menu',
  Dept: '/system/role/dept',
  Menus: '/system/role/menus',
  Depts: '/system/role/depts',
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

export function removeRole(id: number) {
  return request.delete<number>({ url: Api.Base, data: { id }, params: { id } });
}

export function assignRoleMenu(data: RoleMenuParams) {
  return request.put<number>({ url: Api.Menu, data });
}

export function assignRoleDept(data: RoleDeptParams) {
  return request.put<number>({ url: Api.Dept, data });
}

export function fetchRoleMenus(roleId: number) {
  return request.get<{ ids: number[] }>({ url: Api.Menus, params: { id: roleId } });
}

export function fetchRoleDepts(roleId: number) {
  return request.get<{ ids: number[] }>({ url: Api.Depts, params: { id: roleId } });
}
