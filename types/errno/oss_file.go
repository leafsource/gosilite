package errno

import "gosilite/pkg/errorx/code"

// OssFile: 对象存储文件台账（701000050 段）
const (
	ErrOssFileNotFound       = 701000050
	ErrOssFileLocked         = 701000051
	ErrOssFileInvalidParam   = 701000052
	ErrOssFileUploadFailed   = 701000053
	ErrOssFileDownloadFailed = 701000054
	ErrOssFileServiceMissing = 701000055
	ErrOssFileDeleteFailed   = 701000056
)

func init() {
	code.Register(ErrOssFileNotFound, "oss file not found: {id}", code.WithAffectStability(false))
	code.Register(ErrOssFileLocked, "locked file must be unlocked before delete", code.WithAffectStability(false))
	code.Register(ErrOssFileInvalidParam, "oss file invalid param: {msg}", code.WithAffectStability(false))
	code.Register(ErrOssFileUploadFailed, "oss file upload failed: {msg}", code.WithAffectStability(false))
	code.Register(ErrOssFileDownloadFailed, "oss file download failed: {msg}", code.WithAffectStability(false))
	code.Register(ErrOssFileServiceMissing, "storage config not found: {key}", code.WithAffectStability(false))
	code.Register(ErrOssFileDeleteFailed, "oss file delete failed: {msg}", code.WithAffectStability(false))
}
