import { request } from '@/utils/request';

/** 统一文件服务（spec system-oss S2.11）：所有文件传输收口，上传自动记台账 */
export interface CommonFileDTO {
  id: string;
  url: string;
  /** 对象 key */
  key: string;
  originalName: string;
  size: number;
  contentType: string;
}

/** 统一上传（multipart 字段 file，≤20MB，登录即可；t-upload 走原生 XHR 需自行拼前缀与 token） */
export function uploadCommonFile(file: File) {
  return request.upload<CommonFileDTO>('file', file, { url: '/common/file/upload' });
}

/**
 * 统一下载（带 Authorization 的 blob 下载）。
 * 不能用 window.open 直开链接：浏览器导航请求不带 Authorization 头，/common/file/download 会 401。
 * 走 axios 拿 blob（享受 401→refresh→重试拦截器），再触发本地保存。
 */
export async function downloadCommonFile(id: string, fallbackName = 'download') {
  const res = await request.get<{ data: Blob; headers: Record<string, string> }>(
    {
      url: '/common/file/download',
      params: { id },
      responseType: 'blob',
    },
    { isTransformResponse: false, isReturnNativeResponse: true },
  );

  // 业务错误（如 not found）响应体是 JSON 信封：读出错误抛出，避免把错误页存成文件
  const ct = String(res.headers?.['content-type'] || '');
  if (ct.includes('application/json')) {
    const text = await (res.data as Blob).text();
    let msg = '下载失败';
    try {
      msg = (JSON.parse(text) as { msg?: string }).msg || msg;
    } catch {
      /* 保留默认文案 */
    }
    throw new Error(msg);
  }

  const filename = parseContentDisposition(res.headers?.['content-disposition']) || fallbackName;
  const blob = res.data;
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

/** 解析 Content-Disposition 里的文件名（兼容 filename*=UTF-8'' 与 filename="..."） */
function parseContentDisposition(cd?: string): string | undefined {
  if (!cd) return undefined;
  const star = /filename\*=(?:UTF-8'')?([^;]+)/i.exec(cd);
  if (star) {
    try {
      return decodeURIComponent(star[1].trim().replace(/^"|"$/g, ''));
    } catch {
      /* 落回 filename= */
    }
  }
  const plain = /filename="?([^";]+)"?/i.exec(cd);
  return plain ? plain[1] : undefined;
}
