package monitor

// VO 层：Service → Handler 的数据监控视图（camelCase JSON）。

type DatasourceVO struct {
	PgMaxOpen           int64  `json:"pgMaxOpen"`
	PgOpen              int64  `json:"pgOpen"`
	PgInUse             int64  `json:"pgInUse"`
	PgIdle              int64  `json:"pgIdle"`
	PgWaitCount         int64  `json:"pgWaitCount"`
	PgWaitDurationMs    int64  `json:"pgWaitDurationMs"`
	PgMaxIdleClosed     int64  `json:"pgMaxIdleClosed"`
	PgMaxLifetimeClosed int64  `json:"pgMaxLifetimeClosed"`
	RedisUp             bool   `json:"redisUp"`
	RedisLatencyMs      int64  `json:"redisLatencyMs"`
	CheckedAt           string `json:"checkedAt"`
}

type ServerVO struct {
	GoVersion      string `json:"goVersion"`
	GoMaxProcs     int64  `json:"goMaxProcs"`
	Goroutine      int64  `json:"goroutine"`
	UptimeSec      int64  `json:"uptimeSec"`
	HeapAlloc      int64  `json:"heapAlloc"`
	HeapSys        int64  `json:"heapSys"`
	HeapIdle       int64  `json:"heapIdle"`
	HeapInuse      int64  `json:"heapInuse"`
	SysBytes       int64  `json:"sysBytes"`
	GcNum          int64  `json:"gcNum"`
	GcPauseTotalMs int64  `json:"gcPauseTotalMs"`
	NumCPU         int64  `json:"numCPU"`
	Pid            string `json:"pid"`
}

type SqlAggVO struct {
	SQL        string `json:"sql"`
	Count      int64  `json:"count"`
	AvgMs      int64  `json:"avgMs"`
	MaxMs      int64  `json:"maxMs"`
	TotalMs    int64  `json:"totalMs"`
	ErrorCount int64  `json:"errorCount"`
	LastMs     int64  `json:"lastMs"`
}

type SlowSqlVO struct {
	SQL    string `json:"sql"`
	CostMs int64  `json:"costMs"`
	At     int64  `json:"at"`
	Err    bool   `json:"err"`
}

type SqlMonitorVO struct {
	List            []*SqlAggVO   `json:"list"`
	SlowList        []*SlowSqlVO  `json:"slowList"`
	SlowThresholdMs int64         `json:"slowThresholdMs"`
	Limited         bool          `json:"limited"`
}

type FwHitVO struct {
	Rule string `json:"rule"`
	SQL  string `json:"sql"`
	At   int64  `json:"at"`
}

type FirewallVO struct {
	Enabled  bool         `json:"enabled"`
	HitCount int64        `json:"hitCount"`
	Recent   []*FwHitVO   `json:"recent"`
	Rules    []string     `json:"rules"`
}

type UriStatVO struct {
	Path       string `json:"path"`
	Method     string `json:"method"`
	Count      int64  `json:"count"`
	ErrorCount int64  `json:"errorCount"`
	AvgMs      int64  `json:"avgMs"`
	MaxMs      int64  `json:"maxMs"`
}

type UriListVO struct {
	List  []*UriStatVO `json:"list"`
	Total int64        `json:"total"`
}

type HttpStatusVO struct {
	Method     string `json:"method"`
	StatusCode string `json:"statusCode"`
	Count      int64  `json:"count"`
}

type HttpLatencyVO struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Count  int64  `json:"count"`
	AvgUs  int64  `json:"avgUs"`
}

type HttpMonitorVO struct {
	StatusList []*HttpStatusVO `json:"statusList"`
	TopLatency []*HttpLatencyVO `json:"topLatency"`
	TotalCount int64            `json:"totalCount"`
}
