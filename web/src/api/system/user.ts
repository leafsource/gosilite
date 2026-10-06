import { request } from '@/utils/request';

// 管理端用户（AdminUserDTO）
export type AdminUserDTO = {
  id: number;
  account: string;
  username: string;
  mobilePhone: string;
  email: string;
  state: string;
  deptId: number;
  deptName: string;
  roleIds: number[];
  postIds: number[];
  avatarUrl: string;
};

export type AdminUserListParams = {
  page?: number;
  pageSize?: number;
  keyword?: string;
  deptId?: number;
  state?: string;
};

export type AdminUserSaveParams = {
  id?: number;
  account: string;
  username: string;
  mobilePhone: string;
  email: string;
  password: string;
  deptId: number;
  state: string;
  roleIds: number[];
  postIds: number[];
};

export type UserStatusParams = {
  id: number;
  state: string;
};

const Api = {
  List: '/system/user/list',
  Base: '/system/user',
  Status: '/system/user/status',
  Roles: '/system/user/roles',
};

export function fetchUserList(params?: AdminUserListParams) {
  return request.get<{ list: AdminUserDTO[]; total: number }>({ url: Api.List, params });
}

export function createUser(data: AdminUserSaveParams) {
  return request.post<AdminUserDTO>({ url: Api.Base, data });
}

export function updateUser(data: AdminUserSaveParams) {
  return request.put<AdminUserDTO>({ url: Api.Base, data });
}

export function removeUser(id: number) {
  return request.delete<number>({ url: Api.Base, data: { id }, params: { id } });
}

export function updateUserStatus(data: UserStatusParams) {
  return request.put<number>({ url: Api.Status, data });
}

export function fetchUserRoles(id: number) {
  return request.get<{ ids: number[] }>({ url: Api.Roles, params: { id } });
}
