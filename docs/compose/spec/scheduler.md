---
feature: scheduler
status: delivered
updated: 2026-10-12
branch: feat/scheduler
commits: e5382ca..958d876
---

# 定时调度

## Report

**What was built** — Added immutable, ordered `system_job_run_log` records for scheduler lifecycle and Job business logs. The executor, Redis/fence stages, run-bound robfig cron panic wrapper, and Job context logging write both Hertz `hlog` and an asynchronous persist-before-broadcast sink. The monitoring UI replays history and follows live logs through permission-protected SSE.

**What was verified** — Applied migrations 0027/0028/0029 to the local PostgreSQL development database. A real manual `system.scheduler.test` run produced persisted trigger, lock, execute, Job-step, and finish logs. The authenticated SSE endpoint replayed the same `id / event: log / data` frames immediately. Snapshot retry was verified through the real API: the new run retained the original task key, name, parameters, timeout and configuration version.

**Run control console** — The monitoring page now separates task management and run records. A task row opens the run-record Tab with its task filter selected. Runs retain snapshots, show details/logs/retry/stop only, and provide copy/download logs. Stop is cooperative and durable across replicas: a local request cancels immediately when possible, while the owner polls the persisted stopping state every second and preserves its lease/fence until the worker exits as `stopped`. Stable keys are centralized in `types/jobkeys`, framework jobs are in `application/framework/jobs`, and SSE is centralized in `web/src/utils/request/sse.ts`.

**Verification** — `WEB=1 ./script/check.sh` passed: Go tests, vet, build, Vue type checking, Vite production build, and Hugo documentation build. Focused cron-panic, cross-replica stop, stopped-run snapshot retry, and SSE ordering regressions passed. Final independent review confirmed S4/S5 specification compliance, correctness, and project consistency.

**Journey log** — ① Hertz `SetBodyStream` with a deferred pipe did not flush SSE replay reliably; switching to Hertz’s native `protocol/sse.Writer` made replay immediate and preserves live streaming. ② Browser `EventSource` cannot send the existing Bearer header; the UI uses authenticated fetch streaming behind one shared SSE utility. ③ Cron-wide events lack a valid run ID and remain in hlog; wrapper panics are bound to the run created by the executor and persist into that run’s log. ④ Retention is YAML-driven (`scheduler.log_retention_days`, default 30) and the cleanup Job never removes run summaries. ⑤ Cross-replica stop polling is intentionally independent from lease renewal, preventing a one-minute lease from delaying cancellation by 30 seconds.

## [S1] Problem

集群部署下需要由管理员配置已编译进服务的定时任务，并确保同一任务在同一触发时刻只由一个应用副本执行。系统还需审计每次尝试、失败和重试，并提供一个可观测的内置测试任务作为验收入口。

## [S2] Design

任务能力由 Go 代码中的稳定 `job_key` 常量及注册中心提供。`job_key` 使用小写点分业务语义，禁止从函数名、类型名或数据库 ID 推导；同一 Key 只能注册一个 Job，实现重构时 Key 不变。管理员仅能从注册表选择 Job Key，并配置 cron、超时、任务参数与额外重试次数；cron 按服务进程系统本地时间运行；不支持脚本或任意 HTTP 回调。

`system_job` 保存任务定义和 `active_run_token`，`system_job_run` 保存每次尝试的不可变运行记录。执行器在启动 worker 前以条件更新原子认领空的 `active_run_token`；仅持有相同 token 的 worker 在实际退出后才能清除它。随后使用 Redis `SET key token NX PX` 获取任务级租约；成功持锁后以 token 校验的 Lua 续租与释放锁。抢锁失败、数据库 fence 被占用写入 `skipped_locked`，Redis 或 fence 不可用时拒绝执行。若 Redis 续租失败，已启动 worker 的 fence 继续保留到其退出，因此后续执行即使重新获得 Redis 也必须跳过。为安全起见，系统绝不自动回收遗留 token；进程崩溃或无法完成清理后，恢复必须由未来显式的运维操作完成。任务运行带超时 context；失败后的总尝试次数为 `retry_count + 1`，每次重试重新获取锁。停机、禁用与抢锁期间遗漏的 cron 时刻不补跑。

任务定义和运行记录通过 IDL 暴露为系统管理接口，所有 i64 ID 使用字符串 JSON 契约。任务变更热更新本进程 cron 表；每个副本均装载任务，Redis 锁选择唯一执行者。内置 `system.scheduler.test` Job 使用安全的计数型测试行为，并能在运行记录和日志中验证执行结果。

## [S3] Out of Scope

不支持自定义脚本、任意 HTTP 回调、DAG 编排、补偿历史遗漏时刻、跨 Redis 仲裁、强制中断已开始的业务操作或编辑运行记录。

## [S4] Log Streaming and Retention

每一次 `system_job_run` 关联有序、不可变的执行日志。调度器适配 robfig/cron `Logger`，将调度器事件、包装器 panic、锁与执行器生命周期统一输出到 Hertz `hlog`，同时通过异步持久化队列写入任务日志表。Job 可通过执行上下文中的 logger 写入业务步骤；日志行包含 run ID、job key、级别、阶段、消息与时间。

管理端通过受权限保护的 SSE 端点订阅一个运行记录的新增日志。服务先回放该运行记录已持久化日志，再推送连接期间的新日志；事件携带单调递增日志 ID，浏览器断线可用 `Last-Event-ID` 恢复。SSE 订阅仅用于实时显示，数据库是历史查询的唯一来源。

日志保留天数取自 `scheduler.log_retention_days`，默认 30 天，并允许环境变量覆盖。内置清理 Job 每日执行，删除过期执行日志；不得删除 `system_job_run` 摘要记录。持久化或广播失败不得中断业务 Job，但要通过 `hlog` 报错。

## [S5] Run Control and Operations Console

调度页面分为“任务管理”和“运行记录”两个 Tab。任务管理的运行记录入口切换到运行记录 Tab，并将任务 ID 写入查询条件。运行记录表展示运行 ID、任务名称、任务键、触发方式、重试次数、状态、触发时间与结束时间；它是独立页面内容，不再使用抽屉承载列表。

每条运行记录持久化创建时的任务快照：任务键、任务名称、参数 JSON、超时秒数与调度配置版本。重试使用此快照创建新的手动运行，不读取后来编辑的任务定义。运行记录只提供详情、日志、重试、停止四种操作。详情展示运行基本信息和快照；日志在运行中订阅 SSE、已结束时查询持久化日志，并支持复制和下载纯文本日志。

停止是协作式取消：可停止状态仅为运行中或停止中之前的记录。停止请求将运行状态改为 `stopping`、取消该运行上下文并阻止后续重试；Job/Service 通过 `ctx.Done()` 尽快收敛，最终记为 `stopped`。不强杀 goroutine；未响应取消时保留活动执行栅栏、继续显示停止中，直至实际退出。

稳定 Job Key 常量位于 `types/jobkeys/`，按业务域拆分。框架自身 Job 实现在 `application/framework/jobs/`；业务 Job 实现在所属 `application/<domain>/jobs/` 并调用同域 Service。SSE 前端客户端归入 `web/src/utils/request/sse.ts`，统一处理 Bearer 鉴权、帧解析、断线重连、Last-Event-ID、关闭和错误回调；领域 API 仅暴露已定义的 SSE 连接函数。

## Tasks

- [x] T1: 建立调度器注册、Redis 租约锁与执行器测试 — acceptance: 重复 Key 被拒绝，抢锁、失败重试、超时与运行记录状态可由单测验证 (covers: S2)
- [x] T2: 增加任务定义与运行记录持久化、IDL 和权限路由 — acceptance: 管理接口可校验配置、加载已启用任务并查询不可变执行记录 (covers: S2; depends: T1)
- [x] T3: 接入 framework Module、内置测试任务与管理端页面 — acceptance: 管理员能创建、启停、手动运行测试任务并在页面看到记录 (covers: S2; depends: T1, T2)
- [x] T4: 运行后端与前端验收并进行独立评审 — acceptance: `WEB=1 ./script/check.sh` 通过，评审确认规格、正确性与项目一致性 (covers: S2; depends: T1, T2, T3)
- [x] T5: 持久化任务执行日志并适配 cron/hlog — acceptance: 每个运行记录可按序查询触发、锁、执行、重试与结束日志，业务任务失败不因日志写入失败中断 (covers: S4)
- [x] T6: 提供任务日志 SSE 实时订阅与管理端日志面板 — acceptance: 页面打开运行记录可回放历史并实时追加日志，断线后按日志 ID 恢复 (covers: S4; depends: T5)
- [x] T7: 增加 YAML 保留期配置与自动清理任务并完成验收 — acceptance: 默认 30 天配置生效，过期日志被清理且运行摘要保留，`WEB=1 ./script/check.sh` 与独立评审通过 (covers: S4; depends: T5, T6)
- [x] T8: 持久化运行快照并实现停止、重试契约 — acceptance: 停止协作取消且阻止后续重试；重试使用原运行快照 (covers: S5)
- [x] T9: 集中稳定 Job Key 与 Job 实现落位 — acceptance: Key 位于 `types/jobkeys`，框架与业务 Job 按域组织且现有任务可执行 (covers: S5)
- [x] T10: 建立统一 SSE 请求客户端并实现双 Tab 运行控制台 — acceptance: 任务管理可跳转筛选运行记录；日志可实时/历史显示、复制、下载，组件不直接 fetch (covers: S5; depends: T8, T9)
- [x] T11: 验收运行控制、双 Tab 与独立评审 — acceptance: `WEB=1 ./script/check.sh` 通过，真实任务可停止/重试并满足快照语义 (covers: S5; depends: T8, T9, T10)
