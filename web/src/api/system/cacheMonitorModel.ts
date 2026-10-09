/** 缓存监控 */

export interface CachePrefixDTO {
  prefix: string;
  count: number;
  remark: string;
}

export interface CacheKeyDTO {
  key: string;
  type: string;
  /** hash 展开后的 field；string 为空 */
  field?: string;
}

export interface CacheValueResponse {
  key: string;
  type: string;
  value: string;
  /** 展示键名：hash 用 field，否则 key */
  displayKey: string;
}
