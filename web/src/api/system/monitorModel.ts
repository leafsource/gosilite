/** 数据监控（系统监控 · Tab） */

export interface DatasourceDTO {
  pgMaxOpen: number;
  pgOpen: number;
  pgInUse: number;
  pgIdle: number;
  pgWaitCount: number;
  pgWaitDurationMs: number;
  pgMaxIdleClosed: number;
  pgMaxLifetimeClosed: number;
  redisUp: boolean;
  redisLatencyMs: number;
  checkedAt: string;
}

export interface ServerDTO {
  goVersion: string;
  goMaxProcs: number;
  goroutine: number;
  uptimeSec: number;
  heapAlloc: number;
  heapSys: number;
  heapIdle: number;
  heapInuse: number;
  sysBytes: number;
  gcNum: number;
  gcPauseTotalMs: number;
  numCPU: number;
  pid: string;
}

export interface SqlAggDTO {
  sql: string;
  count: number;
  avgMs: number;
  maxMs: number;
  totalMs: number;
  errorCount: number;
  lastMs: number;
}

export interface SlowSqlDTO {
  sql: string;
  costMs: number;
  at: number;
  err: boolean;
}

export interface SqlMonitorResponse {
  list: SqlAggDTO[];
  slowList: SlowSqlDTO[];
  slowThresholdMs: number;
  limited: boolean;
}

export interface FwHitDTO {
  rule: string;
  sql: string;
  at: number;
}

export interface FirewallResponse {
  enabled: boolean;
  hitCount: number;
  recent: FwHitDTO[];
  rules: string[];
}

export interface UriStatDTO {
  path: string;
  method: string;
  count: number;
  errorCount: number;
  avgMs: number;
  maxMs: number;
}

export interface UriListResponse {
  list: UriStatDTO[];
  total: number;
}

export interface HttpStatusDTO {
  method: string;
  statusCode: string;
  count: number;
}

export interface HttpLatencyDTO {
  method: string;
  path: string;
  count: number;
  avgUs: number;
}

export interface HttpMonitorResponse {
  statusList: HttpStatusDTO[];
  topLatency: HttpLatencyDTO[];
  totalCount: number;
}
