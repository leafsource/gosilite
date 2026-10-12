import { request } from '@/utils/request';

import type { MessageLogDTO, MessageLogListParams } from './messageLogModel';

// 查询消息发送记录列表
export function listMessageLogs(params: MessageLogListParams) {
  return request.get<{ list: MessageLogDTO[]; total: number }>({ url: '/system/message/log/list', params });
}

// 查询消息发送记录详细
export function getMessageLog(id: string) {
  return request.get<MessageLogDTO>({ url: '/system/message/log', params: { id } });
}

// 删除消息发送记录
export function removeMessageLog(id: string) {
  return request.delete({ url: '/system/message/log', data: { id }, params: { id } });
}

// 清空消息发送记录
export function clearMessageLogs() {
  return request.delete<number>({ url: '/system/message/log/clear' });
}
