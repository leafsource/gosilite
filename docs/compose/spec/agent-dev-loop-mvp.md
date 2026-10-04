---
feature: agent-dev-loop-mvp
status: delivered
updated: 2026-10-04
branch: feat/agent-dev-loop-mvp
commits: 4a58e75..6ac32b1
---

# Agent-Driven Web Framework MVP

## Report

**What was built** — 落地了 gosilite 闭环 MVP：`httpx` 统一错误/JSON envelope/中间件；`dbx` **启动时按 DSN 连接一种数据库**（sqlite/mysql/postgres，换库只换连接配置）+ 方言占位符 + 在所选库上幂等应用 schema 版本文件；`mcpx` 最小 MCP stdio server（仅业务 tools）；`app` 环境配置。`examples/tasks` 以 Handler→Service→Repository 分层实现 CRUD，同一 Service 同时暴露 HTTP REST 与 `tasks_*` MCP 工具；配套 design/coding/testing 模式文档、Compose 对齐的需求流程、`gosilite-dev` skill 与 README。

**Verification** — `go test ./...` PASS；`go vet ./...` PASS；`go build ./...` PASS；HTTP 冒烟（POST/GET `/api/tasks`）PASS；MCP 冒烟（initialize/tools/list/tools/call `tasks_create`）PASS。复审确认 3 个 critical（schema 记账占位符、内存 SQLite 连接池、MySQL DDL 可移植性）与 MCP tools/call 测试缺口已修复，无新增 critical。

**Journey log** —
1. MCP 定位纠偏：不是脚手架/系统工具，而是业务接口能力（与 HTTP 同一 Service）。
2. 支持多数据库 = **启动连一种库**；SQL 占位符与 schema 版本 DDL 都要按当前驱动方言生成，不能只改 Repository。
3. `sqlite://:memory:` + 连接池会各见各的库；内存 DSN 必须单连接。
4. schema 版本文件可多语句（`;` 分隔）；这是应用建表，不是 MySQL↔Postgres 迁库。
5. JSON 内嵌 envelope 的断言要先解一层，避免转义引号误判。

## [S1] Problem

需要一套由 agent 高频驱动、可快速交付 Web 功能需求的 Go 开发闭环：既要有一致且可被 agent 稳定生成的轻量框架与编码/测试模式，又要把业务能力同时暴露为 HTTP API 与 MCP 业务工具（非系统运维工具），并配套需求流程与 skill，使「需求 → 设计 → 编码 → 验收」可重复、可校验。

## [S2] Design

### Product shape

gosilite = 轻量 Go 业务框架 + Agent 工程体系闭环。

- 框架层：HTTP/JSON 服务标准积木（路由、错误模型、中间件、DB、迁移）。
- 能力层：业务 Service 一次实现，双通道暴露：HTTP handler 与 MCP business tool。
- 工程层：设计/编码/测试模式文档、Compose Next 对齐的需求流程、单一主 skill。

### Package layout

```text
gosilite/
  httpx/          # JSON 响应、错误模型、中间件
  dbx/            # database/sql 打开、方言、迁移
  mcpx/           # 业务能力 → MCP tools（stdio JSON-RPC）
  app/            # 环境变量 Config
  examples/tasks/ # 示例业务（分层 + 组装）+ HTTP/MCP 双暴露
  cmd/tasks/      # 示例进程入口（HTTP，可选 MCP stdio）
  docs/patterns/  # design / coding / testing
  docs/process/   # requirements 闭环
  .agents/skills/gosilite-dev/
```

### [S2.1] Layering & coding model

业务代码严格分层，依赖单向：

```text
HTTP Handler → Service → Repository (interface) → SQL/dialect
     ↓             ↓
   DTO/JSON    Domain + Error
MCP Tool   ↗ 同一 Service
```

- Handler：只做 bind/validate 边界、调用 Service、映射错误到 JSON。
- Service：业务规则与事务边界；构造器注入 Repository 接口。
- Repository：接口由 Service 消费方定义；实现基于 `database/sql` + 方言。
- DTO：请求/响应结构体与 Domain 分离，JSON tag 稳定。
- `context.Context` 永远是第一个参数；错误用 `%w` 包装为 `*httpx.Error`。

### [S2.2] Error model（`httpx`）

统一错误结构，HTTP/MCP 共用：

```json
{"ok": false, "error": {"code": "not_found", "message": "...", "detail": {}}}
```

成功：

```json
{"ok": true, "data": <payload>}
```

| code | HTTP |
|------|------|
| invalid | 400 |
| unauthorized | 401 |
| forbidden | 403 |
| not_found | 404 |
| conflict | 409 |
| internal | 500 |

提供 `httpx.Invalid/NotFound/Conflict/Internal` 构造器与 `httpx.HTTPStatus(err)`。

### [S2.3] HTTP contracts（`httpx`）

- 基于 `net/http`（Go 1.22+ pattern routing：`GET /api/tasks/{id}`）。
- 中间件链：Recover → RequestID → Logging → (业务 handler)。
- `RespondOK` / `RespondErr` / `DecodeJSON`。
- 2xx 成功 envelope；`DELETE` 成功返回 `204 No Content`；非 2xx 一律错误 envelope。

### [S2.4] Data layer（`dbx`）

- **启动时按 DSN 连接一种数据库**（不是数据库之间的迁移）：
  - `sqlite://` / `file:` / `:memory:` → `modernc.org/sqlite`（纯 Go，无 CGO）
  - `mysql://` → `go-sql-driver/mysql`
  - `postgres://` / `postgresql://` → `jackc/pgx/v5/stdlib`
- 进程启动读 `GOSILITE_DB_DSN`（或等价配置）选定连接；换库 = 换 DSN，不改业务代码。
- `Dialect`：按当前驱动生成 `Placeholder(n)`（`?` vs `$n`）、`QuoteIdent`、`Name`。
- **Schema 版本文件**（应用建表，不是迁库）：有序 `.sql` 文件，启动时在**所选库**上幂等应用，记录于 `schema_migrations(version)`。
- Repository 测试默认 SQLite 内存库；同一代码可在 mysql/postgres 上运行，SQL 占位符由方言生成。

### [S2.5] Business MCP（`mcpx`）— 非系统能力

MCP 只暴露**业务接口能力**（与 HTTP 同一 Service），不是脚手架/测试等系统工具。

- 最小 MCP stdio server（JSON-RPC 2.0）：`initialize` / `tools/list` / `tools/call`。
- `mcpx.Tool{Name, Description, InputSchema(JSON Schema), Handler}`。
- 调用失败映射为 `isError` + 上述 error envelope 文本/结构，与 HTTP code 语义对齐。
- 示例工具：`tasks_create` / `tasks_list` / `tasks_get` / `tasks_update` / `tasks_delete`。

### [S2.6] Example feature: tasks

HTTP：

- `POST /api/tasks` body `{title, description?, status?}`
- `GET /api/tasks?status=`
- `GET /api/tasks/{id}`
- `PATCH /api/tasks/{id}` body 可更新字段
- `DELETE /api/tasks/{id}`

Domain 字段：`id, title, description, status(todo|doing|done), created_at, updated_at`。

### [S2.7] Engineering loop

1. **需求**：在 `docs/compose/spec/<feature>.md` 按 Compose Next 模板记录 Problem/Design/Tasks。
2. **设计**：先定 DTO/错误码/路由/表结构/测试边界，再写代码。
3. **编码**：repo → service → handler/mcp；禁止跨层调用。
4. **测试**：表驱动；Service 用真 SQLite 仓储（少 mock）；HTTP 用 `httptest`；MCP 用 JSON-RPC 报文测 `tools/call`。
5. **验收**：`go test ./...`、`go vet ./...`、`go build ./...`。
6. **skill**：`.agents/skills/gosilite-dev/SKILL.md` 引用上述文档与检查清单。

### [S2.8] Patterns catalog（文档沉淀）

`docs/patterns/design.md`：分层、DTO 边界、错误模型、Repository+Dialect、双通道能力、中间件链。

`docs/patterns/coding.md`：ctx 首参、构造器注入、接口由消费方定义、错误包装、禁全局、命名。

`docs/patterns/testing.md`：表驱动、真依赖优先、httptest、内存 SQLite、MCP 报文测试、验收命令。

## [S3] Out of Scope

- SPA/SSR 复杂前端工程（路由框架、状态管理、组件库自研）；`web/` 仅为 Vite + Vue3 + TDesign 单页参考。
- ORM、代码生成 CLI、脚手架 MCP、鉴权/多租户、生产级连接池调优。
- 完整 MCP 资源/订阅/采样等能力；仅 tools 三方法。
- OpenAPI 生成、分布式追踪、消息队列。

## Amendment 2026-10-04

- 工作区由 `.worktrees` 改为仓库主目录；功能已合入 `develop`。
- 明确多数据库 = **启动时按 DSN 连接一种库**，不是库间迁数；`migrations/` 仅应用 schema 版本。
- **新增 `web/`**：TypeScript + Vite + Vue3 + TDesign（`tdesign-vue-next`），入口 `web/index.html`，经 `/api` 代理调用 tasks HTTP API。

## Tasks

- [x] T1: 搭建 httpx（错误模型/响应/中间件） — acceptance: 包测试通过，错误码映射与 JSON envelope 正确 (covers: S2.2, S2.3)
- [x] T2: 搭建 dbx（Open/方言/迁移） — acceptance: SQLite 内存迁移成功，DSN scheme 分发正确，方言占位符测试通过 (covers: S2.4)
- [x] T3: 搭建 mcpx（stdio MCP tools） — acceptance: initialize/tools/list/tools/call 报文测试通过 (covers: S2.5)
- [x] T4: 搭建 app 配置与组装 — acceptance: 可从 env 组装 HTTP handler 与 MCP tools (covers: S2.1, S2.6)
- [x] T5: 实现 examples/tasks 分层与 HTTP 路由 — acceptance: CRUD  httptest 断言通过，错误 envelope 符合约定 (covers: S2.1, S2.2, S2.6)
- [x] T6: 实现 tasks 的 MCP 工具双暴露 — acceptance: tools/call 可创建/查询/更新/删除并与 HTTP 同数据 (covers: S2.5, S2.6)
- [x] T7: 编写 design/coding/testing/process 文档 — acceptance: 文档覆盖 [S2.7]/[S2.8] 且无 TBD (covers: S2.7, S2.8)
- [x] T8: 编写 gosilite-dev skill — acceptance: skill 可引导需求到验收，引用文档路径有效 (covers: S2.7)
- [x] T9: 全量验证与 README — acceptance: go test/vet/build 全绿，README 说明闭环用法 (covers: S2.7)
