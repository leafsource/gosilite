import { request } from '@/utils/request';

import type { StorageInfo, StorageListParams, StorageObject, StorageUploadResult } from './storageModel';

export function getStorageInfo() {
  return request.get<StorageInfo>({ url: '/system/storage/info' });
}

export function listStorageObjects(params: StorageListParams) {
  return request.get<{ list: StorageObject[]; total: number }>({ url: '/system/storage/list', params });
}

export function removeStorageObject(key: string) {
  return request.delete({ url: '/system/storage/object', data: { key }, params: { key } });
}

/** 上传测试（multipart 字段 file），成功返回 { webURI, key } */
export function uploadStorageObject(file: File) {
  return request.upload<StorageUploadResult>('file', file, { url: '/system/storage/upload' });
}
