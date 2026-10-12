/** 消息模板 DTO（MessageTemplateDTO） */
export interface MessageTemplateDTO {
  id: string;
  templateName: string;
  messageConfigId: string;
  messageKeyId: string;
  messageKey: string;
  /** MAIL / SMS */
  messageType: string;
  /** TEMPLATE_ID / TEMPLATE_CONTENT */
  templateMode: string;
  title: string;
  signature: string;
  templateId: string;
  content: string;
  /** 输入变量定义 JSON */
  varsJson: string;
  /** 1 正常 0 停用 */
  status: number;
  remark: string;
  createdAt: string;
  updatedAt: string;
}

export interface MessageTemplateListParams {
  page?: number;
  pageSize?: number;
  keyword?: string;
  messageType?: string;
  messageKeyId?: string;
  status?: number;
}

export interface MessageTemplateSaveParams {
  id?: string;
  templateName: string;
  /** 未选择时省略该字段（Go 侧置 0 走业务校验），禁止传空串（会绑成 int64 失败） */
  messageConfigId?: string;
  messageKeyId?: string;
  messageKey: string;
  messageType: string;
  templateMode: string;
  title: string;
  signature: string;
  templateId: string;
  content: string;
  varsJson: string;
  status: number;
  remark?: string;
}

/** 模板变量（前端表单态，随 varsJson 持久化） */
export interface MessageTemplateVar {
  key: string;
  value?: string;
}

/** 测试发送请求（messageTemplateId 以字符串传递） */
export interface MessageSendTestParams {
  messageTemplateId: string;
  /** 收件账号（邮箱地址或手机号） */
  account: string;
  vars: Record<string, string>;
}

/** 测试发送结果 */
export interface MessageSendTestResult {
  isSuccess: boolean;
  responseBody: string;
  /** 毫秒 */
  costTime: number;
  logId: string;
}
