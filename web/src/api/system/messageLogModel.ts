/** 消息发送记录 DTO（MessageLogDTO） */
export interface MessageLogDTO {
  id: string;
  messageTemplateId: string;
  messageKey: string;
  messageTemplateName: string;
  messageType: string;
  templateMode: string;
  account: string;
  title: string;
  templateId: string;
  content: string;
  messageConfigTitle: string;
  supplierType: string;
  /** 1 成功 0 失败 */
  isSuccess: number;
  responseBody: string;
  logTime: string;
  /** 毫秒 */
  costTime: number;
}

export interface MessageLogListParams {
  page?: number;
  pageSize?: number;
  keyword?: string;
  messageType?: string;
  messageKey?: string;
  isSuccess?: number;
}
