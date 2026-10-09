namespace go framework.monitor

struct OnlineListRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    // 登录名称模糊
    3: string userName
}

struct OnlineDTO {
    // 行标识 = access token 摘要（同用户多端各自成行）
    1: string id
    2: i64 userId
    // 登录名称（账号）
    3: string userName
    4: string deptName
    5: string clientIp
    6: string browser
    7: string os
    8: string loginAt
}

struct OnlineListResponse {
    1: list<OnlineDTO> list
    2: i64 total
}

struct OnlineForceRequest {
    1: string id
}

service SystemMonitorController {
    OnlineListResponse OnlineList(1: OnlineListRequest req)(api.get = "/system/online/list")
    i32 OnlineForce(1: OnlineForceRequest req)(api.post = "/system/online/force")
}
