package errno

import "gosilite/pkg/errorx/code"

// 消息管理（spec system-message）
const (
	ErrMessageTemplateNotFound = 701000060
	ErrMessageConfigNotFound   = 701000061
	ErrMessageKeyNotFound      = 701000062
	ErrMessageKeyExists        = 701000063
	ErrMessageLogNotFound      = 701000064
	// ErrMessageSupplierNotSupported 未接入的短信供应商（发送失败但照常落发送记录）
	ErrMessageSupplierNotSupported = 701000065
	ErrMessageSendFailed           = 701000066
	ErrMessageInvalidParam         = 701000067
)

func init() {
	code.Register(ErrMessageTemplateNotFound, "message template not found: {id}", code.WithAffectStability(false))
	code.Register(ErrMessageConfigNotFound, "message config not found: {id}", code.WithAffectStability(false))
	code.Register(ErrMessageKeyNotFound, "message key not found: {id}", code.WithAffectStability(false))
	code.Register(ErrMessageKeyExists, "message key already exists: {key}", code.WithAffectStability(false))
	code.Register(ErrMessageLogNotFound, "message log not found: {id}", code.WithAffectStability(false))
	code.Register(ErrMessageSupplierNotSupported, "短信供应商未接入: {supplier}", code.WithAffectStability(false))
	code.Register(ErrMessageSendFailed, "message send failed: {reason}", code.WithAffectStability(false))
	code.Register(ErrMessageInvalidParam, "invalid parameter: {msg}", code.WithAffectStability(false))
}
