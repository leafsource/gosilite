import { hasPerm } from '@/directives/permission';
import { request } from '@/utils/request';

import type { OssFileDTO, OssFileListParams, OssFileSaveParams } from './ossModel';

// ===== 台账 =====

// 文件台账分页列表
export function listOssFiles(params: OssFileListParams) {
  return request.get<{ list: OssFileDTO[]; total: number }>({ url: '/system/oss/list', params });
}

// 文件详情
export function getOssFile(id: string) {
  return request.get<OssFileDTO>({ url: '/system/oss', params: { id } });
}

// 修改文件（原名 / 锁定）
export function updateOssFile(data: OssFileSaveParams) {
  return request.put<OssFileDTO>({ url: '/system/oss', data });
}

// 批量删除台账（锁定文件拒绝）
export function removeOssFile(ids: string[]) {
  return request.delete<number>({ url: '/system/oss', data: { ids } });
}

// ===== 上传组件回显辅助（s-upload-image / s-upload-file 移植自 ruoyi-tdesign） =====

/**
 * 按 id 批量查台账 → Map<id, OssFileDTO>（ruoyi listByIds 对应）。
 * gosilite 无批量接口，逐个查询；无 system:oss:list 权限直接返回空（组件回退展示），单条失败跳过不抛出。
 */
export async function listByIds(ids: string): Promise<Map<string, OssFileDTO>> {
  const map = new Map<string, OssFileDTO>();
  if (!hasPerm('system:oss:list')) return map;
  const parts = ids.split(',').filter(Boolean);
  const rows = await Promise.all(parts.map((id) => getOssFile(id).catch(() => null)));
  rows.forEach((row) => {
    if (row) map.set(String(row.id), row);
  });
  return map;
}

/** url → 对象 key（uploads/xxx/yyy.ext），台账按 key 精确匹配回显 */
function urlToKey(url: string): string {
  try {
    return decodeURIComponent(new URL(url, window.location.origin).pathname.replace(/^\//, ''));
  } catch {
    return '';
  }
}

/**
 * 按 url 批量查台账 → Map<url, OssFileDTO>（ruoyi listByUrls 对应）。
 * gosilite 无按 url 查询接口：从 url 提取对象 key 走 fileName 精确匹配。
 * 无 system:oss:list 权限（如普通用户头像回显）直接返回空 Map，组件回退文件名展示，避免无谓 403。
 */
export async function listByUrls(urls: string[]): Promise<Map<string, OssFileDTO>> {
  const map = new Map<string, OssFileDTO>();
  if (!hasPerm('system:oss:list')) return map;
  for (const url of urls) {
    try {
      const key = urlToKey(url);
      if (!key) continue;
      const res = await listOssFiles({ fileName: key, pageSize: 1 });
      const hit = res.list.find((row) => row.fileName === key);
      if (hit) map.set(url, hit);
    } catch {
      // 单条查询失败不影响其余回显
    }
  }
  return map;
}

/**
 * 删除单条台账（上传组件移除直传文件时清理，ruoyi delOss 对应）。
 * 需 system:oss:remove 权限；组件内对失败静默（普通用户移除仅清本地）。
 */
export function delOss(id: string) {
  return removeOssFile([id]);
}
