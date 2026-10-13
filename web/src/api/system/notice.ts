import { request } from '@/utils/request';

import type { NoticeDTO, NoticeListParams, NoticeSaveParams } from './noticeModel';

export function listNotices(params: NoticeListParams) {
  return request.get<{ list: NoticeDTO[]; total: number }>({ url: '/system/notice/list', params });
}

export function getNotice(id: string) {
  return request.get<NoticeDTO>({ url: '/system/notice', params: { id } });
}

export function createNotice(data: NoticeSaveParams) {
  return request.post({ url: '/system/notice', data });
}

export function updateNotice(data: NoticeSaveParams) {
  return request.put({ url: '/system/notice', data });
}

export function removeNotice(id: string) {
  return request.delete({ url: '/system/notice', data: { id }, params: { id } });
}
