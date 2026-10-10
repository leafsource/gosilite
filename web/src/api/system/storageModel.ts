/** 存储管理相关类型（对齐 idl/framework/system_storage.thrift 生成契约） */

export interface StorageInfo {
  endpoint: string;
  bucket: string;
  domain: string;
  connected: boolean;
  objectCount: number;
  totalSize: number;
  message?: string;
}

export interface StorageObject {
  key: string;
  size: number;
  etag?: string;
  lastModified?: string;
  type?: string;
}

export interface StorageListParams {
  page: number;
  pageSize: number;
  prefix?: string;
  keyword?: string;
}

export interface StorageUploadResult {
  webURI: string;
  key: string;
}
