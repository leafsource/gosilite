namespace go framework.storage

// ===== 存储管理（系统管理） =====

// 存储概览：配置 + 探活 + 用量聚合
struct StorageInfoResponse {
    1: string endpoint
    2: string bucket
    3: string domain
    // 探活结果：list 调用是否成功
    4: bool connected
    5: i64 objectCount
    6: i64 totalSize
    // 探活失败时的错误信息
    7: string message
}

struct StorageObjectListRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    // key 前缀过滤
    3: string prefix
    // key 包含模糊
    4: string keyword
}

struct StorageObjectDTO {
    1: string key
    2: i64 size
    3: string etag
    // yyyy-MM-dd HH:mm:ss
    4: string lastModified
    // 扩展名（小写，无点；目录为 dir，无扩展名为 file）
    5: string type
}

struct StorageObjectListResponse {
    1: list<StorageObjectDTO> list
    2: i64 total
}

struct StorageObjectDeleteRequest {
    1: string key
}

// 上传测试：随机 key 落 S3 后返回可访问地址
struct StorageUploadResponse {
    1: string webURI
    2: string key
}

service SystemStorageController {
    StorageInfoResponse GetInfo()(api.get = "/system/storage/info")
    StorageObjectListResponse GetObjects(1: StorageObjectListRequest req)(api.get = "/system/storage/list")
    i32 DeleteObject(1: StorageObjectDeleteRequest req)(api.delete = "/system/storage/object")
    StorageUploadResponse Upload()(api.post = "/system/storage/upload")
}
