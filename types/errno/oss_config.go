package errno

import "gosilite/pkg/errorx/code"

// OssConfig: 对象存储配置（701000034 段，避开 storage 032/033）
const (
	ErrOssConfigNotFound   = 701000034
	ErrOssConfigKeyExists  = 701000035
	ErrOssConfigInvalid    = 701000036
	ErrOssConfigIsDefault  = 701000037
	ErrOssConfigSaveFailed = 701000038
)

func init() {
	code.Register(ErrOssConfigNotFound, "oss config not found: {id}", code.WithAffectStability(false))
	code.Register(ErrOssConfigKeyExists, "oss config key already exists: {key}", code.WithAffectStability(false))
	code.Register(ErrOssConfigInvalid, "oss config invalid param: {msg}", code.WithAffectStability(false))
	code.Register(ErrOssConfigIsDefault, "cannot delete the default storage config", code.WithAffectStability(false))
	code.Register(ErrOssConfigSaveFailed, "oss config save failed: {msg}", code.WithAffectStability(false))
}
