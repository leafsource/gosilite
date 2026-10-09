namespace go framework.log

// ===== 操作日志 =====

struct OperLogListRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    3: string traceId
    4: string userName
    // sys_oper_type，缺省=全部
    5: optional i32 operType
    // sys_oper_status，缺省=全部
    6: optional i32 execStatus
    // 执行方法（request_uri 模糊）
    7: string method
    8: optional string httpMethod
    9: optional string startTime
    10: optional string endTime
}

struct OperLogDTO {
    1: i64 id
    2: string traceId
    3: string userName
    // sys_oper_type
    4: i32 operType
    5: string requestUri
    6: string httpMethod
    7: string module
    // sys_oper_status: 0 成功 1 失败
    8: i32 operStatus
    9: string remark
    10: string errorMsg
    11: i32 costMs
    12: string clientIp
    13: string createdAt
}

struct OperLogDetailDTO {
    1: OperLogDTO base
    2: string requestHeaders
    3: string requestBody
    4: string responseBody
}

struct LogListResponse {
    1: list<OperLogDTO> list
    2: i64 total
}

struct LogIdRequest {
    1: i64 id
}

// ids 与 all 二选一：ids 按选中删除，all=true 清空全部
struct LogCleanRequest {
    1: optional list<i64> ids
    2: bool all
}

struct ErrorLogListRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    3: string traceId
    4: string userName
    // 异常信息模糊
    5: string keyword
    // sys_handle_status，缺省=全部
    6: optional i32 handleStatus
    // sys_error_level，缺省=全部
    7: optional i32 errLevel
    8: optional string httpMethod
    9: optional string startTime
    10: optional string endTime
}

struct ErrorLogDTO {
    1: i64 id
    2: string traceId
    3: string userName
    // 1 业务异常 2 panic
    4: i32 errType
    // sys_error_level: 1 普通 2 严重
    5: i32 errLevel
    // sys_handle_status: 0 待处理 1 已处理
    6: i32 handleStatus
    7: i32 errCode
    8: string errMsg
    9: string requestUri
    10: string httpMethod
    11: string module
    12: i32 costMs
    13: string clientIp
    14: string createdAt
    15: string handledAt
}

struct ErrorLogDetailDTO {
    1: ErrorLogDTO base
    2: string stack
    3: string requestHeaders
    4: string requestBody
    5: string handleRemark
}

struct ErrorLogListResponse {
    1: list<ErrorLogDTO> list
    2: i64 total
}

struct ErrorHandleRequest {
    1: i64 id
    2: string handleRemark
}

// ===== 登录日志 =====

struct LoginLogListRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    3: string userName
    // sys_login_status，缺省=全部
    4: optional i32 status
    5: string clientIp
    6: optional string startTime
    7: optional string endTime
}

struct LoginLogDTO {
    1: i64 id
    2: string traceId
    3: string userName
    4: i64 userId
    // sys_login_status: 0 成功 1 失败
    5: i32 status
    6: string failReason
    7: string clientIp
    8: string userAgent
    9: string browser
    10: string os
    11: string createdAt
}

struct LoginLogListResponse {
    1: list<LoginLogDTO> list
    2: i64 total
}

service SystemLogController {
    LogListResponse OperLogList(1: OperLogListRequest req)(api.get = "/system/operlog/list")
    OperLogDetailDTO OperLogDetail(1: LogIdRequest req)(api.get = "/system/operlog")
    i32 OperLogClean(1: LogCleanRequest req)(api.post = "/system/operlog/clean")

    ErrorLogListResponse ErrorLogList(1: ErrorLogListRequest req)(api.get = "/system/errorlog/list")
    ErrorLogDetailDTO ErrorLogDetail(1: LogIdRequest req)(api.get = "/system/errorlog")
    i32 ErrorLogHandle(1: ErrorHandleRequest req)(api.post = "/system/errorlog/handle")
    i32 ErrorLogClean(1: LogCleanRequest req)(api.post = "/system/errorlog/clean")

    LoginLogListResponse LoginLogList(1: LoginLogListRequest req)(api.get = "/system/loginlog/list")
}
