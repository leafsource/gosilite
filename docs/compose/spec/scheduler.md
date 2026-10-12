---
feature: scheduler
status: in-progress
updated: 2026-10-12
branch: feat/scheduler
commits:
---

# 定时调度

## Report

- Delivered migration 0024, IDL-generated management APIs, Redis lease execution, cron runtime, framework module wiring, and the built-in `system.scheduler.test` job.
- Executor returns promptly on timeout or lost lease after cancelling the job context, but keeps the durable active-execution fence until the worker goroutine exits; the terminal audit record is finalized only then, so a non-cooperative job cannot overlap a local retry or later execution even if its Redis lease is lost.
- Added paginated per-job run history to the scheduler management page, including status and output/error inspection.
- Verification: `WEB=1 ./script/check.sh` passed; the real local database migration and `system.scheduler.test` runtime smoke test are pending before delivery.

## [S1] Problem

集群部署下需要由管理员配置已编译进服务的定时任务，并确保同一任务在同一触发时刻只由一个应用副本执行。系统还需审计每次尝试、失败和重试，并提供一个可观测的内置测试任务作为验收入口。

## [S2] Design

任务能力由 Go 代码中的稳定 `job_key` 常量及注册中心提供。`job_key` 使用小写点分业务语义，禁止从函数名、类型名或数据库 ID 推导；同一 Key 只能注册一个 Job，实现重构时 Key 不变。管理员仅能从注册表选择 Job Key，并配置 cron、时区、超时、任务参数与额外重试次数；不支持脚本或任意 HTTP 回调。

`system_job` 保存任务定义和 `active_run_token`，`system_job_run` 保存每次尝试的不可变运行记录。执行器在启动 worker 前以条件更新原子认领空的 `active_run_token`；仅持有相同 token 的 worker 在实际退出后才能清除它。随后使用 Redis `SET key token NX PX` 获取任务级租约；成功持锁后以 token 校验的 Lua 续租与释放锁。抢锁失败、数据库 fence 被占用写入 `skipped_locked`，Redis 或 fence 不可用时拒绝执行。若 Redis 续租失败，已启动 worker 的 fence 继续保留到其退出，因此后续执行即使重新获得 Redis 也必须跳过。为安全起见，系统绝不自动回收遗留 token；进程崩溃或无法完成清理后，恢复必须由未来显式的运维操作完成。任务运行带超时 context；失败后的总尝试次数为 `retry_count + 1`，每次重试重新获取锁。停机、禁用与抢锁期间遗漏的 cron 时刻不补跑。

任务定义和运行记录通过 IDL 暴露为系统管理接口，所有 i64 ID 使用字符串 JSON 契约。任务变更热更新本进程 cron 表；每个副本均装载任务，Redis 锁选择唯一执行者。内置 `system.scheduler.test` Job 使用安全的计数型测试行为，并能在运行记录和日志中验证执行结果。

## [S3] Out of Scope

不支持自定义脚本、任意 HTTP 回调、DAG 编排、补偿历史遗漏时刻、跨 Redis 仲裁、强制中断已开始的业务操作或编辑运行记录。

## Tasks

- [x] T1: 建立调度器注册、Redis 租约锁与执行器测试 — acceptance: 重复 Key 被拒绝，抢锁、失败重试、超时与运行记录状态可由单测验证 (covers: S2)
- [x] T2: 增加任务定义与运行记录持久化、IDL 和权限路由 — acceptance: 管理接口可校验配置、加载已启用任务并查询不可变执行记录 (covers: S2; depends: T1)
- [x] T3: 接入 framework Module、内置测试任务与管理端页面 — acceptance: 管理员能创建、启停、手动运行测试任务并在页面看到记录 (covers: S2; depends: T1, T2)
- [x] T4: 运行后端与前端验收并进行独立评审 — acceptance: `WEB=1 ./script/check.sh` 通过，评审确认规格、正确性与项目一致性 (covers: S2; depends: T1, T2, T3)
