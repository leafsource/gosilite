import { request } from '@/utils/request';

import type { CacheKeyDTO, CachePrefixDTO, CacheValueResponse } from './cacheMonitorModel';

export function listCachePrefixes() {
  return request.get<{ list: CachePrefixDTO[]; total: number }>({
    url: '/system/cache/prefix/list',
  });
}

export function listCacheKeys(prefix: string) {
  return request.get<{ list: CacheKeyDTO[]; total: number }>({
    url: '/system/cache/key/list',
    params: { prefix },
  });
}

export function getCacheValue(key: string, field?: string) {
  return request.get<CacheValueResponse>({
    url: '/system/cache/value',
    params: field ? { key, field } : { key },
  });
}

export function deleteCacheKey(key: string, field?: string) {
  return request.post<number>({ url: '/system/cache/key/delete', data: { key, field: field ?? '' } });
}

export function clearCachePrefix(prefix: string) {
  return request.post<number>({ url: '/system/cache/prefix/clear', data: { prefix } });
}
