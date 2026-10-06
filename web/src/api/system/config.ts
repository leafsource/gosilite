import { request } from '@/utils/request';
import type { ConfigDTO, ConfigSaveParams } from './configModel';

export function listConfigs(params: { page?: number; pageSize?: number; keyword?: string }) {
  return request.get<{ list: ConfigDTO[]; total: number }>({ url: '/system/config/list', params });
}

export function createConfig(data: ConfigSaveParams) {
  return request.post({ url: '/system/config', data });
}

export function updateConfig(data: ConfigSaveParams) {
  return request.put({ url: '/system/config', data });
}

export function removeConfig(id: number) {
  return request.delete({ url: '/system/config', data: { id }, params: { id } });
}

export function getConfigByKey(key: string) {
  return request.get<ConfigDTO>({ url: '/system/config/key', params: { key } });
}
