package monitor

import (
	"context"
	"os"
	"runtime"
	"sort"
	"strconv"
	"time"

	"gosilite/application/base/metrics"
	"gosilite/application/infra/cache"
	"gosilite/conf"
	dto "github.com/prometheus/client_model/go"
	"gorm.io/gorm"
)

// Service 数据监控查询：读进程内指标与连接池，不写业务表。
type Service struct {
	db    *gorm.DB
	cache cache.CacheClient
	start time.Time
}

// SystemMonitorSVC 包级门面。
var SystemMonitorSVC *Service

// InitService 安装门面。
func InitService(db *gorm.DB, cacheClient cache.CacheClient) *Service {
	SystemMonitorSVC = &Service{db: db, cache: cacheClient, start: time.Now()}
	return SystemMonitorSVC
}

// Datasource 数据源：PG 池实时 Stats + Redis 探活（仅管理端访问时执行）。
func (s *Service) Datasource(ctx context.Context) (*DatasourceVO, error) {
	out := &DatasourceVO{CheckedAt: time.Now().Format("2006-01-02 15:04:05")}
	if s.db != nil {
		if pool, err := s.db.DB(); err == nil {
			st := pool.Stats()
			out.PgMaxOpen = int64(st.MaxOpenConnections)
			out.PgOpen = int64(st.OpenConnections)
			out.PgInUse = int64(st.InUse)
			out.PgIdle = int64(st.Idle)
			out.PgWaitCount = st.WaitCount
			out.PgWaitDurationMs = st.WaitDuration.Milliseconds()
			out.PgMaxIdleClosed = st.MaxIdleClosed
			out.PgMaxLifetimeClosed = st.MaxLifetimeClosed
		}
	}
	if s.cache != nil {
		begin := time.Now()
		// Exists 轻量探活，不写数据
		_, err := s.cache.Exists(ctx, "gosilite:health:ping").Result()
		out.RedisLatencyMs = time.Since(begin).Milliseconds()
		out.RedisUp = err == nil
	}
	return out, nil
}

// Server Web 应用：Go runtime 快照。
func (s *Service) Server() *ServerVO {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return &ServerVO{
		GoVersion:     runtime.Version(),
		GoMaxProcs:    int64(runtime.GOMAXPROCS(0)),
		Goroutine:     int64(runtime.NumGoroutine()),
		UptimeSec:     int64(time.Since(s.start).Seconds()),
		HeapAlloc:     int64(ms.HeapAlloc),
		HeapSys:       int64(ms.HeapSys),
		HeapIdle:      int64(ms.HeapIdle),
		HeapInuse:     int64(ms.HeapInuse),
		SysBytes:      int64(ms.Sys),
		GcNum:         int64(ms.NumGC),
		GcPauseTotalMs: int64(ms.PauseTotalNs) / 1e6,
		NumCPU:        int64(runtime.NumCPU()),
		Pid:           strconv.Itoa(os.Getpid()),
	}
}

// SqlMonitor SQL 模板聚合 + 慢 SQL。
func (s *Service) SqlMonitor() *SqlMonitorVO {
	rows, slow, _, _, threshold, limited := metrics.SQL().Snapshot()
	// 模板列表已按 count 降序；限前 100
	if len(rows) > 100 {
		rows = rows[:100]
	}
	// 慢 SQL 环为时间倒序近似（追加序），反转为新→旧
	for i, j := 0, len(slow)-1; i < j; i, j = i+1, j-1 {
		slow[i], slow[j] = slow[j], slow[i]
	}
	out := &SqlMonitorVO{SlowThresholdMs: threshold, Limited: limited}
	for _, r := range rows {
		out.List = append(out.List, &SqlAggVO{
			SQL: r.SQL, Count: r.Count, AvgMs: r.AvgMs, MaxMs: r.MaxMs,
			TotalMs: r.TotalMs, ErrorCount: r.ErrorCount, LastMs: r.LastMs,
		})
	}
	for _, x := range slow {
		out.SlowList = append(out.SlowList, &SlowSqlVO{SQL: x.SQL, CostMs: x.CostMs, At: x.At, Err: x.Err})
	}
	return out
}

// Firewall 防火墙观察统计。
func (s *Service) Firewall() *FirewallVO {
	_, _, fw, hitCount, _, _ := metrics.SQL().Snapshot()
	// 近似总数：环内条数在环未溢出前准确；溢出后为下限
	out := &FirewallVO{
		Enabled:  conf.GetConf().Monitor.FirewallEnable,
		HitCount: hitCount,
		Rules:    firewallRules(),
	}
	for _, h := range fw {
		out.Recent = append(out.Recent, &FwHitVO{Rule: h.Rule, SQL: h.SQL, At: h.At})
	}
	// recent 环为旧→新，反转
	for i, j := 0, len(out.Recent)-1; i < j; i, j = i+1, j-1 {
		out.Recent[i], out.Recent[j] = out.Recent[j], out.Recent[i]
	}
	return out
}

func firewallRules() []string {
	return []string{"multi_statement", "dangerous_pattern", "union_injection"}
}

// UriList 从 Hertz 指标聚合 URI 维度（path=路由模板）。
func (s *Service) UriList() *UriListVO {
	type key struct{ method, path string }
	type agg struct {
		count      int64
		errorCount int64
	}
	acc := map[key]*agg{}
	if mf := metrics.Family("hertz_server_throughput"); mf != nil {
		for _, m := range mf.GetMetric() {
			k := key{method: labelValue(m, "method"), path: labelValue(m, "path")}
			if k.path == "" || k.path == "unknown" {
				continue
			}
			a := acc[k]
			if a == nil {
				a = &agg{}
				acc[k] = a
			}
			n := int64(m.GetCounter().GetValue())
			a.count += n
			if st := labelValue(m, "statusCode"); st != "" && st != "200" {
				if c, err := strconv.Atoi(st); err == nil && c >= 400 {
					a.errorCount += n
				}
			}
		}
	}
	lat := map[key]*latAgg{}
	if hmf := metrics.Family("hertz_server_latency_us"); hmf != nil {
		for _, m := range hmf.GetMetric() {
			k := key{method: labelValue(m, "method"), path: labelValue(m, "path")}
			if k.path == "" || k.path == "unknown" {
				continue
			}
			h := m.GetHistogram()
			if h == nil || h.GetSampleCount() == 0 {
				continue
			}
			la := lat[k]
			if la == nil {
				la = &latAgg{}
				lat[k] = la
			}
			la.count += int64(h.GetSampleCount())
			la.sumUs += h.GetSampleSum()
		}
	}

	out := &UriListVO{}
	for k, a := range acc {
		row := &UriStatVO{Path: k.path, Method: k.method, Count: a.count, ErrorCount: a.errorCount}
		if la := lat[k]; la != nil && la.count > 0 {
			row.AvgMs = int64(la.sumUs / float64(la.count) / 1000)
		}
		out.List = append(out.List, row)
	}
	sort.Slice(out.List, func(i, j int) bool {
		return out.List[i].Count > out.List[j].Count
	})
	out.Total = int64(len(out.List))
	return out
}

type latAgg struct {
	count int64
	sumUs float64
}

// HttpMonitor method×status 与高时延路由。
func (s *Service) HttpMonitor() *HttpMonitorVO {
	out := &HttpMonitorVO{}
	if mf := metrics.Family("hertz_server_throughput"); mf != nil {
		for _, m := range mf.GetMetric() {
			n := int64(m.GetCounter().GetValue())
			out.TotalCount += n
			out.StatusList = append(out.StatusList, &HttpStatusVO{
				Method:     labelValue(m, "method"),
				StatusCode: labelValue(m, "statusCode"),
				Count:      n,
			})
		}
		sort.Slice(out.StatusList, func(i, j int) bool {
			return out.StatusList[i].Count > out.StatusList[j].Count
		})
		if len(out.StatusList) > 50 {
			out.StatusList = out.StatusList[:50]
		}
	}
	type row struct {
		method, path string
		count        int64
		avgUs        float64
	}
	var rows []row
	if hmf := metrics.Family("hertz_server_latency_us"); hmf != nil {
		for _, m := range hmf.GetMetric() {
			h := m.GetHistogram()
			if h == nil || h.GetSampleCount() == 0 {
				continue
			}
			rows = append(rows, row{
				method: labelValue(m, "method"),
				path:   labelValue(m, "path"),
				count:  int64(h.GetSampleCount()),
				avgUs:  h.GetSampleSum() / float64(h.GetSampleCount()),
			})
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].avgUs > rows[j].avgUs })
	if len(rows) > 20 {
		rows = rows[:20]
	}
	for _, r := range rows {
		out.TopLatency = append(out.TopLatency, &HttpLatencyVO{
			Method: r.method, Path: r.path, Count: r.count, AvgUs: int64(r.avgUs),
		})
	}
	return out
}

func labelValue(m *dto.Metric, name string) string {
	for _, lp := range m.GetLabel() {
		if lp.GetName() == name {
			return lp.GetValue()
		}
	}
	return ""
}
