import { request } from '@/utils/request';

import type {
  OssConfigDTO,
  OssConfigListParams,
  OssConfigSaveParams,
  OssConfigTestParams,
  OssConfigTestResult,
} from './ossConfigModel';

// 查询对象存储配置列表
export function listOssConfigs(params: OssConfigListParams) {
  return request.get<{ list: OssConfigDTO[]; total: number }>({ url: '/system/oss/config/list', params });
}

// 查询对象存储配置详情
export function getOssConfig(id: string) {
  return request.get<OssConfigDTO>({ url: '/system/oss/config', params: { id } });
}

// 新增对象存储配置
export function createOssConfig(data: OssConfigSaveParams) {
  return request.post<OssConfigDTO>({ url: '/system/oss/config', data });
}

// 修改对象存储配置
export function updateOssConfig(data: OssConfigSaveParams) {
  return request.put<OssConfigDTO>({ url: '/system/oss/config', data });
}

// 删除对象存储配置（默认配置拒绝）
export function removeOssConfig(ids: string[]) {
  return request.delete<number>({ url: '/system/oss/config', data: { ids } });
}

// 设为默认（切换即时生效）
export function changeOssConfigDefault(id: string) {
  return request.put<OssConfigDTO>({ url: '/system/oss/config/status', data: { id } });
}

// 连通测试（不写库）
export function testOssConfig(data: OssConfigTestParams) {
  return request.post<OssConfigTestResult>({ url: '/system/oss/config/test', data });
}
