package storage

import (
	"path"
	"strings"

	infraStorage "gosilite/application/infra/storage"
)

// toObjectVO infra FileInfo → 列表 VO（etag 去 S3 引号；时间统一展示格式）
func toObjectVO(f *infraStorage.FileInfo) *ObjectVO {
	return &ObjectVO{
		Key:          f.Key,
		Size:         f.Size,
		Etag:         strings.Trim(f.ETag, `"`),
		LastModified: f.LastModified.Format("2006-01-02 15:04:05"),
		Type:         objectType(f.Key),
	}
}

// objectType key 派生展示类型：目录 dir、扩展名小写、无扩展名 file
func objectType(key string) string {
	if strings.HasSuffix(key, "/") {
		return "dir"
	}
	ext := strings.ToLower(strings.TrimPrefix(path.Ext(key), "."))
	if ext == "" {
		return "file"
	}
	return ext
}
