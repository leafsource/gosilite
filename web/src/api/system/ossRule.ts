import { request } from '@/utils/request';

import type { OssRuleDTO, OssRuleListParams, OssRuleSaveParams } from './ossRuleModel';

// 查询OSS处理规则列表
export function listOssRules(params: OssRuleListParams) {
  return request.get<{ list: OssRuleDTO[]; total: number }>({ url: '/system/oss/rule/list', params });
}

// 查询OSS处理规则详情
export function getOssRule(id: string) {
  return request.get<OssRuleDTO>({ url: '/system/oss/rule', params: { id } });
}

// 新增OSS处理规则
export function createOssRule(data: OssRuleSaveParams) {
  return request.post<OssRuleDTO>({ url: '/system/oss/rule', data });
}

// 修改OSS处理规则
export function updateOssRule(data: OssRuleSaveParams) {
  return request.put<OssRuleDTO>({ url: '/system/oss/rule', data });
}

// 删除OSS处理规则
export function removeOssRule(ids: string[]) {
  return request.delete<number>({ url: '/system/oss/rule', data: { ids } });
}

// 规则覆盖字段值开关
export function changeOssRuleOverwrite(id: string, isOverwrite: 'Y' | 'N') {
  return request.put<OssRuleDTO>({ url: '/system/oss/rule/overwrite', data: { id, isOverwrite } });
}

// 刷新规则缓存
export function refreshOssRuleCache() {
  return request.delete<number>({ url: '/system/oss/rule/cache' });
}
