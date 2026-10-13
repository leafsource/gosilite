# Feature Spec: 系统监控 · 数据监控

- status: delivered
- created: 2026-02-10

## Problem

需要在「系统监控」下增加**数据监控**页（单页多 Tab），覆盖数据源 / SQL 监控 / SQL 防火墙 / Web 应用 / URI / Hertz。约束：**一切以不影响业务性能为主**；并评估「是否可用 Prometheus 自实现数据监控页面」。

## 澄清拍板（Gate）

| 决策点 | 拍板 |
|--------|------|
| 页面数据源 | **可，且推荐**：进程内 Prometheus Registry 作唯一聚合层；Tab API 仅在管理端访问时 `Gather()` 出 JSON。**不依赖**部署 Prometheus Server |
| 业务热路径 | 只允许 O(1) 打点（counter/histogram `Inc`/`Observe`）；禁止业务路径查监控库/全表 count |
| 指标出口 | 可选：同一 `/metrics` 暴露给未来 Prometheus Server 抓取（conf 开关，不影响页面） |
| GORM 插件 | 用 `gorm.io/plugin/prometheus`：DBStats + `Postgres{}`，**`recordCount` 默认不启用**（全表 count 拖库）；`StartServer:false` |
| Hertz 插件 | `hertz-contrib/monitor-prometheus`：`WithTracer` 挂请求量/时延 |
| Session | 继续用「在线用户」独立菜单，不进本页 Tab |
| SQL 防火墙 | 简化规则（多语句/危险片段），可 conf 开关；命中只计数+内存环，不落库 |
| SQL 监控明细 | 有界内存聚合（模板粒度上限 + 慢 SQL 环），查询后异步语义（同 goroutine 廉价更新），不写表 |

## Design

### 双轨

```text
业务热路径 ──Inc/Observe──► 进程内 Prometheus Registry
                              ├─ Tab API：Gather() → JSON（管理端拉取）
                              └─ /metrics：可选暴露（运维 scrape）
```

### 契约（IDL `system_monitor.thrift` 追加，rootMw=permission.Default）

| 路由 | 权限 | 说明 |
|------|------|------|
| `GET /system/monitor/datasource` | `system:monitor:list` | PG 池 + Redis 探活 + gorm_dbstats 汇总 |
| `GET /system/monitor/server` | `system:monitor:list` | runtime/go/process（Web 应用） |
| `GET /system/monitor/sql` | `system:monitor:list` | 模板聚合 + 慢 SQL 环 |
| `GET /system/monitor/firewall` | `system:monitor:list` | 开关/规则/命中统计 |
| `GET /system/monitor/uri` | `system:monitor:list` | method+path 聚合 |
| `GET /system/monitor/http` | `system:monitor:list` | Hertz method×status / 时延摘要 |

### 实现落点

- `conf`: `monitor.enable_metrics` / `metrics_path` / `sql_slow_ms` / `sql_max_templates` / `firewall_enable`
- `application/base/metrics`: 自定义 Collector（URI、Hertz 汇总便于页面）、Gather 工具
- `application/infra/orm`: `db.Use(prometheus.New(...))`
- `main.go`: `server.WithTracer(...)`（enable 时）
- Service：`application/framework/service/monitor`（读 registry + db.Stats + redis Ping）
- 菜单迁移 `0016_monitor_data.sql`：`4702` C 数据监控 `monitor/data/index` perms `system:monitor:list`
- 前端：`web/src/pages/monitor/data/index.vue`，`t-tabs` 六页；`web/src/api/system/monitor.ts`

### 性能红线

1. 页面未打开时：零额外分配（仅已有指标自增）。
2. GORM Postgres Collector：禁 `recordCount`；`RefreshInterval≥15s`。
3. URI label 用 path 模板或固定路由，禁止 query 进 label。
4. SQL 模板 map / 慢 SQL 环 / 防火墙环均有上限。
5. 防火墙关闭时零检查成本。

## Out of Scope

- Prometheus Server / Grafana 部署与告警规则（仅保证 `/metrics` 可抓）
- cookie Session、多数据源、Druid 完整 WallFilter、SQL 落库、Client Tracer

## Tasks

- [x] T1 spec + conf + metrics 底座 + 两插件接入
- [x] T2 IDL + hz + Service/Handler
- [x] T3 菜单迁移 + 前端 Tab 页
- [x] T4 check.sh + 性能红线核对

## Report

**改版（用户反馈）**

- 去掉「Hertz 监控」Tab，信息并入 URI（后端 `/system/monitor/http` API 保留但页面不展示）。
- 「Web 应用」+「数据源」卡片合并为**首页**总览。
- 页面重设计：仪表盘式 stat 卡（色条/图标/大数字）+ 连接池进度条；Tab：首页 / SQL 监控 / SQL 防火墙 / URI 监控；子面板拆 `components/*Panel.vue`，首载只请求当前 Tab。

**验收（2026-02-10）**

- `WEB=1 ./script/check.sh` 全绿。
- 临时端口冒烟：六个 `/system/monitor/*` 均 `code=200` 且有真实数据；`http://127.0.0.1:<metrics>/metrics` 含 `hertz_server_throughput`、`gorm_dbstats_*`、`go_goroutines`。
- 迁移 `0016_monitor_data.sql` 已应用（菜单 4702 + system/admin 角色）。
- 单测 `application/base/metrics` 通过（模板归一化 / 慢 SQL / 防火墙观察）。

**性能红线（实现核对）**

1. 页面未打开：仅 tracer `Inc/Observe` + GORM logger Trace 内存更新；无额外 SQL。
2. GORM Postgres `recordCount` 采集器默认关闭（`gorm_postgres_metrics: false`）。
3. URI label = `FullPath()` 路由模板，无 query。
4. SQL 模板上限 200、慢 SQL/防火墙环 50。
5. 防火墙关闭时 `matchFirewall` 不执行。
