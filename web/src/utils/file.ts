/** HTTP URL / 文件辅助（移植自 ruoyi-tdesign utils/ruoyi.ts + utils/file.ts） */

/** 提取 http 链接文件后缀名（去 query） */
export function getHttpFileSuffix(http: string) {
  const index = http.indexOf('?');
  if (index !== -1) {
    http = http.substring(0, index);
  }
  return http.substring(http.lastIndexOf('.') + 1);
}

/** 提取 http 链接文件名（去 query） */
export function getHttpFileName(http: string) {
  const index = http.indexOf('?');
  if (index !== -1) {
    http = http.substring(0, index);
  }
  return http.substring(http.lastIndexOf('/') + 1);
}

/** Blob → File（图片压缩后还原文件名/类型，移植自 ruoyi utils/file.ts） */
export function blobToFile(blob: Blob, fileName: string, fileMimeType: string) {
  const newBlob = blob.slice(0, blob.size, fileMimeType);
  return new File([newBlob], fileName, {
    type: fileMimeType,
    lastModified: Date.now(),
  });
}
