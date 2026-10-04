# gosilite

轻量级 Go 框架 + Agent 工程体系闭环，用于**由 agent 驱动的 Web 功能快捷开发**。

业务能力一次实现，双通道暴露：

- **HTTP / JSON API** — 给用户与常规客户端
- **MCP business tools** — 给 agent 客户端（只暴露业务接口，不是系统运维工具）

## 快速开始

```bash
go test ./...
go build ./...
```

运行示例 HTTP 服务（内存 SQLite）：

```bash
go run ./cmd/tasks
# POST http://localhost:8080/api/tasks  {"title":"hello"}
```

以 MCP 业务服务器模式运行（stdio JSON-RPC）：

```bash
GOSILITE_MCP_STDIO=1 go run ./cmd/tasks
# initialize / tools/list / tools/call  → tasks_create 等
```

环境变量：

| 变量 | 默认 | 说明 |
|------|------|------|
| `GOSILITE_HTTP_ADDR` | `:8080` | HTTP 监听地址 |
| `GOSILITE_DB_DSN` | `sqlite://:memory:` | 启动时连接**一种**库：`sqlite://` / `mysql://user:pass@host/db` / `postgres://...` |
| `GOSILITE_MCP_STDIO` | 关 | 置 1 时以 MCP stdio 模式启动 |

## 架构

```text
HTTP Handler / MCP Tool → Service → Repository → database/sql (sqlite/mysql/postgres)
         ↓                     ↓
      DTO / JSON          Domain + httpx.Error
```

| 包 | 职责 |
|----|------|
| `httpx` | 统一错误模型、JSON envelope、中间件 |
| `dbx` | 按 DSN 启动连一种库、方言占位符、schema 版本文件 |
| `mcpx` | 业务能力 → MCP tools（stdio JSON-RPC） |
| `app` | 环境变量配置 |
| `examples/tasks` | 端到端参考（CRUD 双通道） |

## Agent 工程闭环

1. 需求写入 `docs/compose/spec/<feature>.md`
2. 按 `docs/process/requirements.md` 澄清契约
3. 按 `docs/patterns/{design,coding,testing}.md` 分层实现
4. 用 skill `.mimocode/skills/gosilite-dev` 引导生成与自检
5. `go test ./... && go vet ./... && go build ./...` 验收

## 文档

- `docs/patterns/design.md` — 设计模式
- `docs/patterns/coding.md` — 编码模式
- `docs/patterns/testing.md` — 测试模式
- `docs/process/requirements.md` — 需求开发流程
- `docs/compose/spec/agent-dev-loop-mvp.md` — 本 MVP 规格

## License

MIT
