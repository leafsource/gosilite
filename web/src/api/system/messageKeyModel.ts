/** 消息常量 DTO（MessageKeyDTO） */
export interface MessageKeyDTO {
  id: string;
  name: string;
  messageKey: string;
  remark: string;
  createdAt: string;
  updatedAt: string;
}

export interface MessageKeyListParams {
  page?: number;
  pageSize?: number;
  keyword?: string;
}

export interface MessageKeySaveParams {
  id?: string;
  name: string;
  messageKey: string;
  remark?: string;
}
