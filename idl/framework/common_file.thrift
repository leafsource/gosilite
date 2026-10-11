namespace go framework.commonfile

// ===== 统一文件服务（spec system-oss S2.11：所有文件传输收口，业务不再自建 multipart handler）=====
// 登录即可用；上传自动写 system_oss 台账（create_by=登录用户）。

struct CommonFileDTO {
    1: i64 id
    2: string url
    // 对象 key
    3: string key
    4: string originalName
    5: i64 size
    6: string contentType
}

struct CommonFileDownloadRequest {
    1: i64 id
}

service CommonFileController {
    // 上传：multipart 字段 file 唯一参数（≤20MB），key 自动生成 uploads/<yyyyMMdd>/<uuid>.<ext>
    CommonFileDTO Upload()(api.post = "/common/file/upload")
    // 下载：handler 直接回流附件（非 JSON 信封）
    i32 Download(1: CommonFileDownloadRequest req)(api.get = "/common/file/download")
}
