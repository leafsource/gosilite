---
feature: plugin-module
status: delivered
updated: 2026-06-27
branch: develop
commits: 7a4f6d04..working-tree
---

# 插件化 Module 与 Service 风格拉齐

## Report

**What was built** — 引入 `application.Module`（`Name` / `Init` / `RegisterRoutes`），`Bootstrap` 初始化 infra 与内置模块，`MountRoutes` 按注册表挂载 IDL/hz 生成路由；`main` 不再直接调用 `GeneratedRegister`。四个域 service 拉齐为具体 `Service` + `NewService`/`InitService` 包级门面，删除全部 `I*Service`，修正 `role`/`menu`/`single_agent_draft` 包名。codegen 模板改为 `gosilite` 路径与单文件 Service 样板，新增 `script/check.sh` 一键验收，并同步 AGENTS/README/skill/design 文档。

**Verification** — `./script/check.sh` → PASS（`go test` / `go vet` / `go build`）；独立评审 T1–T8 全部满足，无 CRITICAL。

**Journey log**
- 路由归属按用户选择：插件接管挂载，但路由代码仍由 IDL/hz 生成，模块 `RegisterRoutes` 调用生成的 `Register`。
- `main` 里 Bootstrap 返回值不可命名为 `app`（与 Hertz `app.RequestContext` 包名冲突），改用 `boot`。
- codegen 必须去掉 `service_impl` 生成路径，否则会再生成已删除的模板文件。

## [S1] Problem

当前 service 层每个域都是 `I*Service + *Impl + var *SVC + InitService` 样板，包名混乱（role/menu/single_agent_draft 为 `package dal`），初始化散落在 `application.Init` 手写列表中；路由靠 `router.GeneratedRegister` 整包挂载，Agent 新增业务域时无法按固定插件契约装配。codegen 模板仍引用 `skyx-cloud`，生成即错。需要一次重构拉齐风格并引入 Module 注册插件，同时保留 IDL → hz 生成 `api/model`、`api/router` 的能力。

## [S2] Design

### S2.1 Module 契约

- 新增 `application.Module`：
  - `Name() string`
  - `Init(ctx context.Context, deps *infra.AppDependencies) error` — 构造本域 Service，写入包级门面
  - `RegisterRoutes(r *server.Hertz)` — 调用本域 hz 生成的 `Register`（或等价路由）
- `application.Bootstrap(ctx)`：`infra.Init` → 依次 `Init` 内置模块；返回可挂载句柄。
- `application.MountRoutes(h *server.Hertz)`：依次 `RegisterRoutes`。
- `main.go`：`Bootstrap` 后不再直接 `router.GeneratedRegister`，改为 `MountRoutes`。
- 内置模块注册表集中在 `application`（或 `application/builtin.go`），新增域只加一个模块项。

### S2.2 域模块划分

| 模块 | 包路径 | Init 服务 | RegisterRoutes |
|------|--------|-----------|----------------|
| framework | `application/framework` | user / role / menu | `api/router/framework/user`、`api/router/framework/menu` |
| agent | `application/agent` | single_agent_draft | `api/router/agent` |

### S2.3 Service 拉齐

- 每个域只保留 `type Service struct`（或沿用对外名如 `SystemUserServiceImpl` 更名统一为 `Service`）。
- 提供 `NewService(...) *Service`；`InitService` 仅负责创建并写入包级门面变量（启动时一次）。
- 删除 `I*Service` 接口与 `*Impl` 双文件中的接口部分；方法实现合并进 `service.go` 或保留单文件。
- Repository **继续保留接口**。
- Handler / 中间件继续通过包级门面访问（`SystemUserSVC` 等），类型改为 `*Service`；禁止再出现 `I*Service`。
- 包名修正：`role` / `menu` / `single_agent_draft`（不再是 `dal`）。

### S2.4 codegen 与验收

- `script/codegen/templates` 全部改为 `gosilite/...` 模块路径，生成「具体 Service + NewService」样板，不生成 `I*Service`。
- 新增 `script/check.sh`：`go test` + `go vet` + `go build`（可选 web build 参数）。

### S2.5 文档

- 更新 `AGENTS.md`、`README`、`.agents/skills/gosilite-dev/SKILL.md`、`docs/patterns/design.md`：Module 契约、Service 形态、插件装配顺序。

## [S3] Out of Scope

- Go 动态插件（`plugin` / `.so`）或热加载。
- 前端插件化、TDesign Starter 结构调整。
- 改变 HTTP 响应信封或 OAuth2 会话语义。
- 业务功能行为变更（除装配方式外）。
- 迁移系统重做（仍用现有 `script/postgres.sql` 路径）。

## Tasks

- [x] T1: 定义 `application.Module` + `Bootstrap`/`MountRoutes` + 内置模块注册表 — acceptance: `application` 可编译，无对 `api/router.GeneratedRegister` 的直接依赖 (covers: S2.1)
- [x] T2: 实现 `framework` 与 `agent` 两个 Module 并挂入注册表 — acceptance: `MountRoutes` 能挂 user/menu/agent 路由，行为与原 `GeneratedRegister` 一致 (covers: S2.2; depends: T1)
- [x] T3: 拉齐 user service 为具体 `Service`，删除 `ISystemUserService` — acceptance: `go build` 通过，登录/会话调用点编译通过 (covers: S2.3; depends: T1)
- [x] T4: 拉齐 role / menu / single_agent_draft（含包名修正）— acceptance: 无 `I*Service`、无 `package dal` service 文件，handler 调用编译通过 (covers: S2.3; depends: T1)
- [x] T5: `main.go` 改为 `Bootstrap` + `MountRoutes` — acceptance: `go run .` 路由与 `/ping`、`/user/login` 可达 (covers: S2.1; depends: T2)
- [x] T6: 修正 codegen 模板为 `gosilite` + 具体 Service 样板 — acceptance: 模板无 `skyx-cloud`、无 `I*Service` (covers: S2.4)
- [x] T7: 新增 `script/check.sh` 验收脚本 — acceptance: 根目录执行脚本，test/vet/build 全绿 (covers: S2.4)
- [x] T8: 更新 AGENTS/README/skill/patterns 文档 — acceptance: 文档描述与实现一致，无 `I*Service` 作为推荐形态 (covers: S2.5; depends: T2 T3 T4 T5)
