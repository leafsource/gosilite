namespace go framework.ossconfig

// ===== OSS配置（对象存储 · 系统管理） =====

struct OssConfigListRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    // configKey / endpoint 模糊
    3: string keyword
}

struct OssConfigIdRequest {
    1: i64 id (go.tag = "json:\"id,string\"")
}

struct OssConfigIdsRequest {
    1: list<string> ids
}

struct OssConfigDTO {
    1: i64 id (go.tag = "json:\"id,string\"")
    2: string configKey
    3: string accessKey
    4: string secretKey
    5: string bucketName
    6: string prefix
    7: string endpoint
    8: string domain
    // Y/N
    9: string isHttps
    10: string region
    // 0=private 1=public 2=custom
    11: string accessPolicy
    // 是否默认：1=默认 0=否
    12: string status
    13: string ext1
    14: string remark
    15: string createdAt
    16: string updatedAt
}

struct OssConfigListResponse {
    1: list<OssConfigDTO> list
    2: i64 total
}

// 新增/编辑共用；编辑时必带 id；config_key 建库后不可改
struct OssConfigSaveRequest {
    1: optional i64 id (go.tag = "json:\"id,string\"")
    2: string configKey
    3: string accessKey
    4: string secretKey
    5: string bucketName
    6: string prefix
    7: string endpoint
    8: string domain
    9: string isHttps
    10: string region
    11: string accessPolicy
    12: string remark
}

// 设为默认（设默认即时生效：事务内全部置 0 → 目标置 1）
struct OssConfigStatusRequest {
    1: i64 id (go.tag = "json:\"id,string\"")
}

// 连通测试：用请求体配置临时构建客户端探活，不写库
struct OssConfigTestRequest {
    1: string accessKey
    2: string secretKey
    3: string bucketName
    4: string endpoint
    5: string region
}

struct OssConfigTestResponse {
    1: bool connected
    2: string message
}

service SystemOssConfigController {
    OssConfigListResponse List(1: OssConfigListRequest req)(api.get = "/system/oss/config/list")
    OssConfigDTO Get(1: OssConfigIdRequest req)(api.get = "/system/oss/config")
    OssConfigDTO Create(1: OssConfigSaveRequest req)(api.post = "/system/oss/config")
    OssConfigDTO Update(1: OssConfigSaveRequest req)(api.put = "/system/oss/config")
    i32 Remove(1: OssConfigIdsRequest req)(api.delete = "/system/oss/config")
    OssConfigDTO ChangeStatus(1: OssConfigStatusRequest req)(api.put = "/system/oss/config/status")
    OssConfigTestResponse Test(1: OssConfigTestRequest req)(api.post = "/system/oss/config/test")
}
