# gosilite

轻量级 Go 框架 + Vue3 前端 + Agent 工程体系闭环，用于**由 agent 驱动的 Web 功能快捷开发**。

业务能力一次实现，三端共用契约：

- **HTTP / JSON API** — 给前端与常规客户端
- **MCP business tools** — 给 agent 客户端（只暴露业务接口，不是系统运维工具）
- **Vue3 前端**（`web/`）— TypeScript + Vite + TDesign

## 快速开始

### 后端

```bash
go test ./...
go build ./...
```

运行示例 HTTP 服务（默认内存 SQLite；本机 8080 被占用时换端口）：

```bash
GOSILITE_HTTP_ADDR=:18080 go run ./cmd/tasks
# POST http://127.0.0.1:18080/api/tasks  {"title":"hello"}
```

以 MCP 业务服务器模式运行（stdio JSON-RPC）：

```bash
GOSILITE_MCP_STDIO=1 go run ./cmd/tasks
# initialize / tools/list / tools/call  → tasks_create 等
```

### 前端（web/）

```bash
cd web
npm install
npm run dev
# http://127.0.0.1:5173/  （入口 web/index.html）
```

默认把 `/api` 代理到 `http://127.0.0.1:18080`，可用环境变量覆盖：

```bash
GOSILITE_API_URL=http://127.0.0.1:8080 npm run dev
```

生产构建：`npm run build`（输出 `web/dist`）。

环境变量（后端）：

| 变量 | 默认 | 说明 |
|------|------|------|
| `GOSILITE_HTTP_ADDR` | `:8080` | HTTP 监听地址 |
| `GOSILITE_DB_DSN` | `sqlite://:memory:` | 启动时连接**一种**库：`sqlite://` / `mysql://user:pass@host/db` / `postgres://...` |
| `GOSILITE_MCP_STDIO` | 关 | 置 1 时以 MCP stdio 模式启动 |

## 架构

```text
Vue3/TDesign web/  ──HTTP JSON──►  HTTP Handler ──┐
                                                ├─► Service ─► Repository ─► 一种数据库
Agent (MCP client) ──MCP tools──►  mcpx tools ───┘
         ↓                     ↓
      DTO / JSON          Domain + httpx.Error
```

| 路径 | 职责 |
|------|------|
| `httpx` | 统一错误模型、JSON envelope、中间件 |
| `dbx` | 按 DSN 启动连一种库、方言占位符、schema 版本文件 |
| `mcpx` | 业务能力 → MCP tools（stdio JSON-RPC） |
| `app` | 环境变量配置 |
| `examples/tasks` | 端到端参考（CRUD 双通道） |
| `web/` | Vue3 + Vite + TypeScript + TDesign 前端 |

## Agent 工程闭环

1. 需求写入 `docs/compose/spec/<feature>.md`
2. 按 `docs/process/requirements.md` 澄清契约（含 API JSON 与页面字段）
3. 按 `docs/patterns/{design,coding,testing}.md` 分层实现
4. 后端 Service → HTTP/MCP；前端 `web/` 用 TDesign 对齐同一 DTO
5. 用 skill `.mimocode/skills/gosilite-dev` 引导生成与自检
6. `go test ./... && go vet ./... && go build ./...` 与 `cd web && npm run build` 验收

## 文档

- `docs/patterns/design.md` — 设计模式
- `docs/patterns/coding.md` — 编码模式
- `docs/patterns/testing.md` — 测试模式
- `docs/process/requirements.md` — 需求开发流程
- `docs/compose/spec/agent-dev-loop-mvp.md` — 本 MVP 规格

## License

MIT
