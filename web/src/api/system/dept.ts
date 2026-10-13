import { request } from '@/utils/request';

// 部门 DTO（对应 idl/framework/system.thrift DeptDTO）
export interface DeptDTO {
  id: string;
  parentId: string;
  ancestors: string;
  deptName: string;
  orderNum: number;
  status: number;
}

// 部门树节点（DeptNode）
export interface DeptNode {
  id: string;
  parentId: string;
  label: string;
  orderNum: number;
  status: number;
  leaderId: string;
  leaderName: string;
  children?: DeptNode[];
}

export interface DeptSaveParams {
  id?: string;
  parentId: string;
  deptName: string;
  orderNum: number;
  status: number;
}

// 左侧树 / 父级选择器统一结构
export interface DeptTreeItem {
  value: string;
  label: string;
  children?: DeptTreeItem[];
}

const Api = {
  Tree: '/system/dept/tree',
  List: '/system/dept/list',
  Base: '/system/dept',
  Leader: '/system/dept/leader',
};

export function fetchDeptTree() {
  return request.get<{ list: DeptNode[] }>({ url: Api.Tree });
}

export function fetchDeptList() {
  return request.get<{ list: DeptDTO[]; total: number }>({ url: Api.List });
}

export function createDept(data: DeptSaveParams) {
  return request.post<DeptDTO>({ url: Api.Base, data });
}

export function updateDept(data: DeptSaveParams) {
  return request.put<DeptDTO>({ url: Api.Base, data });
}

export function removeDept(id: string) {
  // DELETE 同时带 query 与 body，兼容 Hertz 的绑定方式
  return request.delete<number>({ url: Api.Base, data: { id }, params: { id } });
}

// 设置/清除部门负责人（userId=0 清除）
export function setDeptLeader(deptId: string, userId: string) {
  return request.put<number>({ url: Api.Leader, data: { deptId, userId } });
}
