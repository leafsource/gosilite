import { request } from '@/utils/request';

import type { OnlineDTO, OnlineListParams } from './onlineModel';

export function listOnline(params: OnlineListParams) {
  return request.get<{ list: OnlineDTO[]; total: number }>({ url: '/system/online/list', params });
}

/** 会话级强退 */
export function forceOffline(id: string) {
  return request.post<number>({ url: '/system/online/force', data: { id } });
}
