package errno

import "gosilite/pkg/errorx/code"

// System: 701 000 000 ~ 701 999 999
const (
	ErrSystemInvalidParam   = 701000001
	ErrSystemDeptNotFound   = 701000002
	ErrSystemDeptHasChildren = 701000003
	ErrSystemDeptHasUsers   = 701000004
	ErrSystemPostNotFound   = 701000005
	ErrSystemPostCodeExists = 701000006
	ErrSystemRoleNotFound   = 701000007
	ErrSystemRoleCodeExists = 701000008
	ErrSystemMenuNotFound   = 701000009
	ErrSystemUserNotFound   = 701000010
	ErrSystemUserAccountExists = 701000011
	ErrSystemForbidden      = 701000012
)

func init() {
	code.Register(ErrSystemInvalidParam, "invalid parameter : {msg}", code.WithAffectStability(false))
	code.Register(ErrSystemDeptNotFound, "department not found: {id}", code.WithAffectStability(false))
	code.Register(ErrSystemDeptHasChildren, "cannot delete department with children: {id}", code.WithAffectStability(false))
	code.Register(ErrSystemDeptHasUsers, "cannot delete department with users: {id}", code.WithAffectStability(false))
	code.Register(ErrSystemPostNotFound, "post not found: {id}", code.WithAffectStability(false))
	code.Register(ErrSystemPostCodeExists, "post code already exists: {code}", code.WithAffectStability(false))
	code.Register(ErrSystemRoleNotFound, "role not found: {id}", code.WithAffectStability(false))
	code.Register(ErrSystemRoleCodeExists, "role code already exists: {code}", code.WithAffectStability(false))
	code.Register(ErrSystemMenuNotFound, "menu not found: {id}", code.WithAffectStability(false))
	code.Register(ErrSystemUserNotFound, "user not found: {id}", code.WithAffectStability(false))
	code.Register(ErrSystemUserAccountExists, "account already exists: {account}", code.WithAffectStability(false))
	code.Register(ErrSystemForbidden, "forbidden: {msg}", code.WithAffectStability(false))
}

// 字典/参数
const (
	ErrSystemDictTypeExists = 701000020
	ErrSystemDictNotFound   = 701000021
	ErrSystemDictHasData    = 701000022
	ErrSystemConfigNotFound = 701000023
	ErrSystemConfigKeyExists = 701000024
	ErrSystemDictGroupExists = 701000025
	ErrSystemDictGroupNotFound = 701000026
	ErrSystemDictGroupHasTypes = 701000027
	ErrSystemNoticeNotFound = 701000028
	ErrOperLogNotFound      = 701000029
	ErrErrorLogNotFound     = 701000030
	ErrOnlineSessionNotFound = 701000031
)

func init() {
	code.Register(ErrSystemDictTypeExists, "dict type already exists: {type}", code.WithAffectStability(false))
	code.Register(ErrSystemDictNotFound, "dict not found: {id}", code.WithAffectStability(false))
	code.Register(ErrSystemDictHasData, "cannot delete dict type with data: {type}", code.WithAffectStability(false))
	code.Register(ErrSystemConfigNotFound, "config not found: {id}", code.WithAffectStability(false))
	code.Register(ErrSystemConfigKeyExists, "config key already exists: {key}", code.WithAffectStability(false))
	code.Register(ErrSystemDictGroupExists, "dict group already exists: {name}", code.WithAffectStability(false))
	code.Register(ErrSystemDictGroupNotFound, "dict group not found: {id}", code.WithAffectStability(false))
	code.Register(ErrSystemDictGroupHasTypes, "cannot delete dict group with types: {id}", code.WithAffectStability(false))
	code.Register(ErrSystemNoticeNotFound, "notice not found: {id}", code.WithAffectStability(false))
	code.Register(ErrOperLogNotFound, "oper log not found: {id}", code.WithAffectStability(false))
	code.Register(ErrErrorLogNotFound, "error log not found: {id}", code.WithAffectStability(false))
	code.Register(ErrOnlineSessionNotFound, "online session not found: {id}", code.WithAffectStability(false))
}
