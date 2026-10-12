import { request } from '@/utils/request';

import type { MessageKeyDTO, MessageKeyListParams, MessageKeySaveParams } from './messageKeyModel';

// 查询消息常量列表
export function listMessageKeys(params: MessageKeyListParams) {
  return request.get<{ list: MessageKeyDTO[]; total: number }>({ url: '/system/message/key/list', params });
}

// 查询消息常量详细
export function getMessageKey(id: string) {
  return request.get<MessageKeyDTO>({ url: '/system/message/key', params: { id } });
}

// 新增消息常量
export function createMessageKey(data: MessageKeySaveParams) {
  return request.post({ url: '/system/message/key', data });
}

// 修改消息常量
export function updateMessageKey(data: MessageKeySaveParams) {
  return request.put({ url: '/system/message/key', data });
}

// 删除消息常量
export function removeMessageKey(id: string) {
  return request.delete({ url: '/system/message/key', data: { id }, params: { id } });
}
