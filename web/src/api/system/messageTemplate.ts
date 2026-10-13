import { request } from '@/utils/request';

import type {
  MessageSendTestParams,
  MessageSendTestResult,
  MessageTemplateDTO,
  MessageTemplateListParams,
  MessageTemplateSaveParams,
} from './messageTemplateModel';

// 查询消息模板列表
export function listMessageTemplates(params: MessageTemplateListParams) {
  return request.get<{ list: MessageTemplateDTO[]; total: number }>({ url: '/system/message/template/list', params });
}

// 查询消息模板详细
export function getMessageTemplate(id: string) {
  return request.get<MessageTemplateDTO>({ url: '/system/message/template', params: { id } });
}

// 新增消息模板
export function createMessageTemplate(data: MessageTemplateSaveParams) {
  return request.post({ url: '/system/message/template', data });
}

// 修改消息模板
export function updateMessageTemplate(data: MessageTemplateSaveParams) {
  return request.put({ url: '/system/message/template', data });
}

// 删除消息模板
export function removeMessageTemplate(id: string) {
  return request.delete({ url: '/system/message/template', data: { id }, params: { id } });
}

// 发送测试消息（无论成败均落发送记录）
export function sendMessageTest(data: MessageSendTestParams) {
  return request.post<MessageSendTestResult>({ url: '/system/message/template/send-test', data });
}
