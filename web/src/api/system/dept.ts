import { request } from '@/utils/request';

// 部门 DTO（对应 idl/framework/system.thrift DeptDTO）
export type DeptDTO = {
  id: number;
  parentId: number;
  ancestors: string;
  deptName: string;
  orderNum: number;
  status: number;
};

// 部门树节点（DeptNode）
export type DeptNode = {
  id: number;
  parentId: number;
  label: string;
  orderNum: number;
  status: number;
  children?: DeptNode[];
};

export type DeptSaveParams = {
  id?: number;
  parentId: number;
  deptName: string;
  orderNum: number;
  status: number;
};

// 左侧树 / 父级选择器统一结构
export type DeptTreeItem = {
  value: number;
  label: string;
  children?: DeptTreeItem[];
};

const Api = {
  Tree: '/system/dept/tree',
  List: '/system/dept/list',
  Base: '/system/dept',
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

export function removeDept(id: number) {
  // DELETE 同时带 query 与 body，兼容 Hertz 的绑定方式
  return request.delete<number>({ url: Api.Base, data: { id }, params: { id } });
}
