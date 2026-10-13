package monitor

import (
	monitor "gosilite/api/model/framework/monitor"
	monitorService "gosilite/application/framework/service/monitor"
	userService "gosilite/application/framework/service/user"
)

const timeLayout = "2006-01-02 15:04:05"

func toOnlineDTO(r userService.OnlineRow) *monitor.OnlineDTO {
	return &monitor.OnlineDTO{
		ID:       r.AccessHash,
		UserId:   r.UserID,
		UserName: r.Account,
		DeptName: r.DeptName,
		ClientIp: r.IP,
		Browser:  r.Browser,
		Os:       r.OS,
		LoginAt:  r.LoginAt.Format(timeLayout),
	}
}

func toOnlineDTOs(rows []userService.OnlineRow) []*monitor.OnlineDTO {
	out := make([]*monitor.OnlineDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toOnlineDTO(r))
	}
	return out
}

func toDatasourceDTO(v *monitorService.DatasourceVO) *monitor.DatasourceDTO {
	if v == nil {
		return &monitor.DatasourceDTO{}
	}
	return &monitor.DatasourceDTO{
		PgMaxOpen: v.PgMaxOpen, PgOpen: v.PgOpen, PgInUse: v.PgInUse, PgIdle: v.PgIdle,
		PgWaitCount: v.PgWaitCount, PgWaitDurationMs: v.PgWaitDurationMs,
		PgMaxIdleClosed: v.PgMaxIdleClosed, PgMaxLifetimeClosed: v.PgMaxLifetimeClosed,
		RedisUp: v.RedisUp, RedisLatencyMs: v.RedisLatencyMs, CheckedAt: v.CheckedAt,
	}
}

func toServerDTO(v *monitorService.ServerVO) *monitor.ServerDTO {
	if v == nil {
		return &monitor.ServerDTO{}
	}
	return &monitor.ServerDTO{
		GoVersion: v.GoVersion, GoMaxProcs: v.GoMaxProcs, Goroutine: v.Goroutine,
		UptimeSec: v.UptimeSec, HeapAlloc: v.HeapAlloc, HeapSys: v.HeapSys,
		HeapIdle: v.HeapIdle, HeapInuse: v.HeapInuse, SysBytes: v.SysBytes,
		GcNum: v.GcNum, GcPauseTotalMs: v.GcPauseTotalMs, NumCPU: v.NumCPU, Pid: v.Pid,
	}
}

func toSqlMonitorDTO(v *monitorService.SqlMonitorVO) *monitor.SqlMonitorResponse {
	out := &monitor.SqlMonitorResponse{}
	if v == nil {
		return out
	}
	out.SlowThresholdMs = v.SlowThresholdMs
	out.Limited = v.Limited
	for _, r := range v.List {
		out.List = append(out.List, &monitor.SqlAggDTO{
			SQL: r.SQL, Count: r.Count, AvgMs: r.AvgMs, MaxMs: r.MaxMs,
			TotalMs: r.TotalMs, ErrorCount: r.ErrorCount, LastMs: r.LastMs,
		})
	}
	for _, x := range v.SlowList {
		out.SlowList = append(out.SlowList, &monitor.SlowSqlDTO{
			SQL: x.SQL, CostMs: x.CostMs, At: x.At, Err: x.Err,
		})
	}
	return out
}

func toFirewallDTO(v *monitorService.FirewallVO) *monitor.FirewallResponse {
	out := &monitor.FirewallResponse{}
	if v == nil {
		return out
	}
	out.Enabled = v.Enabled
	out.HitCount = v.HitCount
	out.Rules = v.Rules
	for _, h := range v.Recent {
		out.Recent = append(out.Recent, &monitor.FwHitDTO{Rule: h.Rule, SQL: h.SQL, At: h.At})
	}
	return out
}

func toUriListDTO(v *monitorService.UriListVO) *monitor.UriListResponse {
	out := &monitor.UriListResponse{}
	if v == nil {
		return out
	}
	out.Total = v.Total
	for _, r := range v.List {
		out.List = append(out.List, &monitor.UriStatDTO{
			Path: r.Path, Method: r.Method, Count: r.Count,
			ErrorCount: r.ErrorCount, AvgMs: r.AvgMs, MaxMs: r.MaxMs,
		})
	}
	return out
}

func toHttpMonitorDTO(v *monitorService.HttpMonitorVO) *monitor.HttpMonitorResponse {
	out := &monitor.HttpMonitorResponse{}
	if v == nil {
		return out
	}
	out.TotalCount = v.TotalCount
	for _, r := range v.StatusList {
		out.StatusList = append(out.StatusList, &monitor.HttpStatusDTO{
			Method: r.Method, StatusCode: r.StatusCode, Count: r.Count,
		})
	}
	for _, r := range v.TopLatency {
		out.TopLatency = append(out.TopLatency, &monitor.HttpLatencyDTO{
			Method: r.Method, Path: r.Path, Count: r.Count, AvgUs: r.AvgUs,
		})
	}
	return out
}

func toCachePrefixList(rows []monitorService.CachePrefixVO) *monitor.CachePrefixListResponse {
	out := &monitor.CachePrefixListResponse{List: []*monitor.CachePrefixDTO{}, Total: int64(len(rows))}
	for _, r := range rows {
		out.List = append(out.List, &monitor.CachePrefixDTO{Prefix: r.Prefix, Count: r.Count, Remark: r.Remark})
	}
	return out
}

func toCacheKeyList(rows []monitorService.CacheKeyVO) *monitor.CacheKeyListResponse {
	out := &monitor.CacheKeyListResponse{List: []*monitor.CacheKeyDTO{}, Total: int64(len(rows))}
	for _, r := range rows {
		out.List = append(out.List, &monitor.CacheKeyDTO{Key: r.Key, Type: r.Type, Field: r.Field})
	}
	return out
}

func toCacheValue(v *monitorService.CacheValueVO) *monitor.CacheValueResponse {
	if v == nil {
		return &monitor.CacheValueResponse{}
	}
	return &monitor.CacheValueResponse{
		Key: v.Key, Type: v.Type, Value: v.Value, DisplayKey: v.DisplayKey,
	}
}
