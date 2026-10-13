package errno

import "gosilite/pkg/errorx/code"

// OssRule: 对象存储处理规则（701000040 段，避开 ossconfig 的 034+ 与 ossfile 的 050+）
const (
	ErrOssRuleNotFound      = 701000040
	ErrOssRuleExists        = 701000041
	ErrOssRuleInvalidParam  = 701000042
	ErrOssRuleResolveFailed = 701000043
)

func init() {
	code.Register(ErrOssRuleNotFound, "oss rule not found: {id}", code.WithAffectStability(false))
	code.Register(ErrOssRuleExists, "oss rule already exists in domain: {domain}", code.WithAffectStability(false))
	code.Register(ErrOssRuleInvalidParam, "oss rule invalid param: {msg}", code.WithAffectStability(false))
	code.Register(ErrOssRuleResolveFailed, "oss rule resolve failed: {msg}", code.WithAffectStability(false))
}
