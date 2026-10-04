---
name: gosilite-dev
description: Use when developing, reviewing, or extending web API features with the gosilite Go framework — layered Handler/Service/Repository code, dual HTTP + MCP business exposure, startup DB connection via DSN (sqlite/mysql/postgres), application schema version files, and the Compose-aligned requirements loop. Trigger on gosilite, tasks example, business MCP tool, or "用 gosilite 开发接口".
---

# gosilite 开发技能

把 Web 功能需求变成可验收的 Go 代码：**分层实现 + HTTP/MCP 双通道 + 统一错误模型**。

## 何时使用

- 在本仓库新增/修改 HTTP API 业务功能
- 为业务能力增加 MCP tool（给 agent 调用，不是系统工具）
- 修 bug、补迁移、补测试
- 评审 gosilite 风格的代码

## 硬性约定

1. 分层单向：Handler/MCP → Service → Repository，禁止跨层。
2. 错误只用 `httpx.Error`（`invalid` / `not_found` / `conflict` / `internal`…）。
3. 同一业务只实现一次 Service；HTTP 与 MCP 都是适配器。
4. **启动连一种库**：`GOSILITE_DB_DSN` 选定 sqlite/mysql/postgres；SQL 占位符经 `dbx.Dialect`。这是换连接，不是迁库。
5. 验收三连：`go test ./...`、`go vet ./...`、`go build ./...`。

## 标准开发流程

1. **定契约**：字段、路由、错误码、表、MCP tool 名。
2. **写 spec**（若功能值得成文）：`docs/compose/spec/<feature>.md`，含 Tasks + acceptance。
3. **Schema 版本文件**：`migrations/NNNN_name.sql`（应用建表，启动时在所选库上幂等应用）。
4. **Repo → Service → HTTP → MCP**，每层带测试。
5. **双通道断言**：MCP 创建的数据 HTTP 能读到，反之亦然。
6. **跑验收三连**，修到全绿。

## 代码骨架

新增 `examples/<feature>/`：

```text
task.go       # Domain
dto.go        # Request/Query DTO
repo.go       # Repository 接口 + SQLRepository
service.go    # 业务规则
http.go       # Routes + handlers
mcp.go        # MCPTools(svc)
app.go        # 组装 + embed migrations
migrations/   # 应用 schema 版本文件
```

MCP tool 命名：`<feature>_<verb>`（例：`tasks_create`）。禁止注册系统运维类工具。

## 测试要求

- Service：真 SQLite + 固定时钟（`WithClock`）
- HTTP：`httptest` 断言 status + envelope
- MCP：JSON-RPC 报文测 `tools/list` / `tools/call`（含 get/update/delete）
- 修 bug 必有回归测试
- 默认测 sqlite；若宣称支持 mysql/postgres，用同一 SQL 流程在对应驱动上跑 schema 版本与 CRUD

## 必读文档

- `docs/patterns/design.md` — 架构与设计模式
- `docs/patterns/coding.md` — 编码规范
- `docs/patterns/testing.md` — 测试规范
- `docs/process/requirements.md` — 需求开发闭环
- `examples/tasks/` — 端到端参考实现

## 检查清单

- [ ] 无 Handler 业务逻辑 / 无 Repo 外泄漏 SQL
- [ ] JSON 契约与错误码符合统一 envelope
- [ ] 新增能力同时暴露 HTTP 与 MCP（或在 spec 中写明只做其一的原因）
- [ ] 启动只连 DSN 指定的一种库；schema 版本幂等可重复执行
- [ ] `go test/vet/build` 全绿
