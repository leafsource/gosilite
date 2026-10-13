// 文件管理 DTO / 参数（对应 idl/framework/system_oss.thrift）
export interface OssFileDTO {
  id: string;
  /** 对象 key */
  fileName: string;
  originalName: string;
  /** 小写后缀（不带点） */
  fileSuffix: string;
  url: string;
  size: number;
  contentType: string;
  isLock: boolean;
  /** 上传时使用的 config_key */
  service: string;
  createBy: string;
  createdAt: string;
  updatedAt: string;
  /** 规则命中的变体 url（key=ruleName，如 80x80） */
  ruleUrls?: Record<string, string>;
}

export interface OssFileListParams {
  page?: number;
  pageSize?: number;
  /** 对象 key 模糊（文件名） */
  fileName?: string;
  originalName?: string;
  /** 服务商 config_key 精确 */
  service?: string;
  /** 小写后缀，逗号分隔 */
  fileSuffix?: string;
  createBy?: string;
  beginTime?: string;
  endTime?: string;
}

export interface OssFileSaveParams {
  id: string;
  originalName: string;
  isLock: boolean;
}
