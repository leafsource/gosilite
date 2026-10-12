namespace go framework.oss

// ===== 文件管理（对象存储 · 系统管理） =====

// ---- 台账 ----

struct OssListRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    // 对象 key 模糊（文件名）
    3: string fileName
    4: string originalName
    // 小写后缀，逗号分隔
    5: string fileSuffix
    // 上传人（create_by）
    6: string createBy
    // 服务商（system_oss_config.config_key）
    7: string service
    // yyyy-MM-dd HH:mm:ss 起止，缺省=不限
    8: string beginTime
    9: string endTime
}

struct OssIdRequest {
    1: i64 id (go.tag = "json:\"id,string\"")
}

struct OssIdsRequest {
    1: list<string> ids
}

struct OssDTO {
    1: i64 id (go.tag = "json:\"id,string\"")
    // 对象 key
    2: string fileName
    3: string originalName
    // 小写后缀（不带点）
    4: string fileSuffix
    5: string url
    6: i64 size
    7: string contentType
    8: bool isLock
    // 上传时使用的 config_key
    9: string service
    10: string createBy
    11: string createdAt
    12: string updatedAt
    // 处理规则命中的变体 url（key=ruleName，如 80x80）；overwrite 命中时 url 字段已被改写
    13: map<string, string> ruleUrls
}

struct OssListResponse {
    1: list<OssDTO> list
    2: i64 total
}

// 编辑：仅 originalName / isLock 可改
struct OssSaveRequest {
    1: i64 id (go.tag = "json:\"id,string\"")
    2: string originalName
    3: bool isLock
}

// ---- 桶视图（吸收自旧存储管理，权限 system:oss:bucket） ----

service SystemOssController {
    OssListResponse List(1: OssListRequest req)(api.get = "/system/oss/list")
    OssDTO Get(1: OssIdRequest req)(api.get = "/system/oss")
    OssDTO Update(1: OssSaveRequest req)(api.put = "/system/oss")
    i32 Remove(1: OssIdsRequest req)(api.delete = "/system/oss")
    // 上传/下载已收口至 /common/file/*（spec S2.11）
}
