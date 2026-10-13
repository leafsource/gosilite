package errno

import "gosilite/pkg/errorx/code"

// Storage: System 段顺延（701000032 起）；上传语义复用 Upload 段（104xxxxxx）
const (
	ErrStorageListFailed   = 701000032
	ErrStorageDeleteFailed = 701000033
)

func init() {
	code.Register(ErrStorageListFailed, "storage list failed: {msg}", code.WithAffectStability(false))
	code.Register(ErrStorageDeleteFailed, "storage delete failed: {msg}", code.WithAffectStability(false))
}
