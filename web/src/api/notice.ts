import { request } from '@/utils/request';

import type { NoticeDetail, NoticeInboxResult, UnreadCountResult } from './noticeModel';

/** 当前用户通知收件箱（login-only，列表不含 content） */
export function listNoticeInbox(params: { page?: number; pageSize?: number; keyword?: string }) {
  return request.get<NoticeInboxResult>({ url: '/notice/inbox', params });
}

/** 通知详情（login-only；服务端打开即自动已读） */
export function getNoticeDetail(id: string) {
  return request.get<NoticeDetail>({ url: '/notice/detail', params: { id } });
}

/** 未读数（Header 角标） */
export function getNoticeUnreadCount() {
  return request.get<UnreadCountResult>({ url: '/notice/unread-count' });
}

/** 标记已读；id 缺省 = 全部已读。返回影响行数 */
export function markNoticeRead(id?: string) {
  return request.post<number>({ url: '/notice/read', data: id ? { id } : {} });
}
