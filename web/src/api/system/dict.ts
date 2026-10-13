import { request } from '@/utils/request';

import type {
  DictDataDTO,
  DictDataSaveParams,
  DictGroupDTO,
  DictGroupSaveParams,
  DictOption,
  DictTypeDTO,
  DictTypeSaveParams,
} from './dictModel';

export function listDictTypes(params: { page?: number; pageSize?: number; keyword?: string; groupId?: string }) {
  return request.get<{ list: DictTypeDTO[]; total: number }>({ url: '/system/dict/type/list', params });
}

export function createDictType(data: DictTypeSaveParams) {
  return request.post({ url: '/system/dict/type', data });
}

export function updateDictType(data: DictTypeSaveParams) {
  return request.put({ url: '/system/dict/type', data });
}

export function removeDictType(id: string) {
  return request.delete({ url: '/system/dict/type', data: { id }, params: { id } });
}

export function listDictGroups() {
  return request.get<{ list: DictGroupDTO[] }>({ url: '/system/dict/group/list' });
}

export function createDictGroup(data: DictGroupSaveParams) {
  return request.post({ url: '/system/dict/group', data });
}

export function updateDictGroup(data: DictGroupSaveParams) {
  return request.put({ url: '/system/dict/group', data });
}

export function removeDictGroup(id: string) {
  return request.delete({ url: '/system/dict/group', data: { id }, params: { id } });
}

export function listDictData(params: { page?: number; pageSize?: number; dictType?: string }) {
  return request.get<{ list: DictDataDTO[]; total: number }>({ url: '/system/dict/data/list', params });
}

export function createDictData(data: DictDataSaveParams) {
  return request.post({ url: '/system/dict/data', data });
}

export function updateDictData(data: DictDataSaveParams) {
  return request.put({ url: '/system/dict/data', data });
}

export function removeDictData(id: string) {
  return request.delete({ url: '/system/dict/data', data: { id }, params: { id } });
}

export function dictOptions(dictType: string) {
  return request.get<{ list: DictOption[] }>({ url: '/system/dict/data/options', params: { dictType } });
}
