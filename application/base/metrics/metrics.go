// Package metrics 进程内 Prometheus 指标底座：
// 业务热路径仅 Inc/Observe；管理端 Tab 在访问时 Gather 出 JSON；
// 可选经 metrics_addr 对外暴露 /metrics（与页面解耦）。
package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
)

var (
	// Registry 供 Hertz tracer 与 runtime/process 指标使用（与 DefaultGatherer 分离）。
	Registry = prometheus.NewRegistry()

	initOnce sync.Once
)

// Init 注册 Go/runtime 与进程指标（幂等）。
func Init() {
	initOnce.Do(func() {
		Registry.MustRegister(
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
	})
}

// multiGatherer 合并自定义 Registry 与默认 Registry（GORM dbstats 在默认侧）。
type multiGatherer struct{}

func (multiGatherer) Gather() ([]*dto.MetricFamily, error) {
	a, err := Registry.Gather()
	if err != nil {
		return nil, err
	}
	b, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return a, nil
	}
	return append(a, b...), nil
}

// Gatherer 暴露与页面共用的合并视图。
var Gatherer prometheus.Gatherer = multiGatherer{}

// Handler /metrics 文本暴露（监控端口，与业务端口隔离）。
func Handler() http.Handler {
	return promhttp.HandlerFor(Gatherer, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
		MaxRequestsInFlight: 4,
		Timeout:              5 * time.Second,
	})
}

// Family 返回合并视图中指定名称的指标族。
func Family(name string) *dto.MetricFamily {
	mfs, err := Gatherer.Gather()
	if err != nil {
		return nil
	}
	for _, mf := range mfs {
		if mf.GetName() == name {
			return mf
		}
	}
	return nil
}
