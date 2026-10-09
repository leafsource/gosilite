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

// ===== 数据监控（Tab） =====

// 数据源：PG 连接池 + Redis 探活 + gorm_dbstats 摘要
struct DatasourceDTO {
    1: i64 pgMaxOpen
    2: i64 pgOpen
    3: i64 pgInUse
    4: i64 pgIdle
    5: i64 pgWaitCount
    6: i64 pgWaitDurationMs
    7: i64 pgMaxIdleClosed
    8: i64 pgMaxLifetimeClosed
    9: bool redisUp
    10: i64 redisLatencyMs
    11: string checkedAt
}

// Web 应用：Go runtime / 进程
struct ServerDTO {
    1: string goVersion
    2: i64 goMaxProcs
    3: i64 goroutine
    4: i64 uptimeSec
    5: i64 heapAlloc
    6: i64 heapSys
    7: i64 heapIdle
    8: i64 heapInuse
    9: i64 sysBytes
    10: i64 gcNum
    11: i64 gcPauseTotalMs
    12: i64 numCPU
    13: string pid
}

struct SqlAggDTO {
    1: string sql
    2: i64 count
    3: i64 avgMs
    4: i64 maxMs
    5: i64 totalMs
    6: i64 errorCount
    7: i64 lastMs
}

struct SlowSqlDTO {
    1: string sql
    2: i64 costMs
    3: i64 at
    4: bool err
}

struct SqlMonitorResponse {
    1: list<SqlAggDTO> list
    2: list<SlowSqlDTO> slowList
    3: i64 slowThresholdMs
    4: bool limited
}

struct FwHitDTO {
    1: string rule
    2: string sql
    3: i64 at
}

struct FirewallResponse {
    1: bool enabled
    2: i64 hitCount
    3: list<FwHitDTO> recent
    4: list<string> rules
}

struct UriStatDTO {
    1: string path
    2: string method
    3: i64 count
    4: i64 errorCount
    5: i64 avgMs
    6: i64 maxMs
}

struct UriListResponse {
    1: list<UriStatDTO> list
    2: i64 total
}

struct HttpStatusDTO {
    1: string method
    2: string statusCode
    3: i64 count
}

struct HttpLatencyDTO {
    1: string method
    2: string path
    3: i64 count
    4: i64 avgUs
}

struct HttpMonitorResponse {
    1: list<HttpStatusDTO> statusList
    2: list<HttpLatencyDTO> topLatency
    3: i64 totalCount
}

service SystemMonitorController {
    OnlineListResponse OnlineList(1: OnlineListRequest req)(api.get = "/system/online/list")
    i32 OnlineForce(1: OnlineForceRequest req)(api.post = "/system/online/force")
    DatasourceDTO Datasource()(api.get = "/system/monitor/datasource")
    ServerDTO Server()(api.get = "/system/monitor/server")
    SqlMonitorResponse SqlMonitor()(api.get = "/system/monitor/sql")
    FirewallResponse Firewall()(api.get = "/system/monitor/firewall")
    UriListResponse UriList()(api.get = "/system/monitor/uri")
    HttpMonitorResponse HttpMonitor()(api.get = "/system/monitor/http")
}
