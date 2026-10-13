import { request } from '@/utils/request';

// 管理端用户（AdminUserDTO）
export type AdminUserDTO = {
  id: string;
  account: string;
  username: string;
  mobilePhone: string;
  email: string;
  state: string;
  deptId: string;
  deptName: string;
  roleIds: string[];
  postIds: string[];
  avatarUrl: string;
  loginFailureCount: number;
  loginPermanentlyLocked: boolean;
};

export type AdminUserListParams = {
  page?: number;
  pageSize?: number;
  keyword?: string;
  deptId?: string;
  state?: string;
};

export type AdminUserSaveParams = {
  id?: string;
  account: string;
  username: string;
  mobilePhone: string;
  email: string;
  password: string;
  deptId: string;
  state: string;
  roleIds: string[];
  postIds: string[];
};

export type UserStatusParams = {
  id: string;
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

export function removeUser(id: string) {
  return request.delete<number>({ url: Api.Base, data: { id }, params: { id } });
}

export function updateUserStatus(data: UserStatusParams) {
  return request.put<number>({ url: Api.Status, data });
}

export function fetchUserRoles(id: string) {
  return request.get<{ ids: string[] }>({ url: Api.Roles, params: { id } });
}
