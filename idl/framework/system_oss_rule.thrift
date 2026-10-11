namespace go framework.ossrule

// ===== OSS处理规则（对象存储 · 系统管理） =====

struct OssRuleListRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    3: string ruleName
    4: string domain
    // 缺省=全部；1=启用 0=停用
    5: optional string status
    // 缺省=全部；Y=默认 N=非默认
    6: optional string isDefault
}

struct OssRuleIdRequest {
    1: i64 id
}

struct OssRuleIdsRequest {
    1: list<i64> ids
}

struct OssRuleDTO {
    1: i64 id
    2: string ruleName
    3: string domain
    4: string mimeType
    5: string rule
    6: string isOverwrite
    7: string isDefault
    8: string status
    9: i32 ruleSort
    10: string remark
    11: string createdAt
    12: string updatedAt
}

struct OssRuleListResponse {
    1: list<OssRuleDTO> list
    2: i64 total
}

struct OssRuleSaveRequest {
    1: optional i64 id
    2: string ruleName
    3: string domain
    4: string mimeType
    5: string rule
    6: string isDefault
    7: string status
    8: i32 ruleSort
    9: string remark
}

// 覆盖字段值开关（目标为 Y 且该规则为默认时，同域名其他默认规则覆盖自动关闭）
struct OssRuleOverwriteRequest {
    1: i64 id
    2: string isOverwrite
}

service SystemOssRuleController {
    OssRuleListResponse List(1: OssRuleListRequest req)(api.get = "/system/oss/rule/list")
    OssRuleDTO Get(1: OssRuleIdRequest req)(api.get = "/system/oss/rule")
    OssRuleDTO Create(1: OssRuleSaveRequest req)(api.post = "/system/oss/rule")
    OssRuleDTO Update(1: OssRuleSaveRequest req)(api.put = "/system/oss/rule")
    i32 Remove(1: OssRuleIdsRequest req)(api.delete = "/system/oss/rule")
    OssRuleDTO ChangeOverwrite(1: OssRuleOverwriteRequest req)(api.put = "/system/oss/rule/overwrite")
    i32 RefreshCache()(api.delete = "/system/oss/rule/cache")
}
