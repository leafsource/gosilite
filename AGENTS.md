# AGENTS.md

轻量 Go 框架 + Vue3 前端 + Agent 工程闭环，用于 agent 快速交付 Web 功能。约定细节见 `.agents/skills/gosilite-dev/SKILL.md` 与 `docs/patterns/`、`docs/process/requirements.md`。

## 命令

```bash
# 后端验收（顺序：test → vet → build）
go test ./...
go test ./httpx ./dbx ./mcpx ./examples/tasks   # 单包
go vet ./... && go build ./...

# 前端
cd web && npm install && npm run build   # vue-tsc + vite build；验收必跑
cd web && npm run dev                    # http://127.0.0.1:5173
```

本地联调（本机 8080 常被占用，用 18080）：

```bash
GOSILITE_HTTP_ADDR=:18080 GOSILITE_DB_DSN='sqlite://file:/tmp/gosilite.db' go run ./cmd/tasks
# MCP 业务模式（stdio JSON-RPC，不是系统工具）
GOSILITE_MCP_STDIO=1 go run ./cmd/tasks
```

`web` 默认把 `/api` 代理到 `http://127.0.0.1:18080`（`GOSILITE_API_URL` / `GOSILITE_WEB_PORT` 可覆盖）。

## 结构

| 路径 | 职责 |
|------|------|
| `httpx` `dbx` `mcpx` `app` | 框架包（错误/envelope、连库/方言/schema、业务 MCP、env 配置） |
| `examples/tasks` | 唯一端到端参考：Domain → dto → repo → service → http/mcp + `migrations/` |
| `cmd/tasks` | 进程入口（HTTP 或 MCP stdio） |
| `web/` | Vue3 + Vite + TS + TDesign |
| `docs/compose/spec/` | Compose feature 文档 |
| `.agents/skills/gosilite-dev/` | 开发 skill（不是 `.mimocode`） |

新业务功能照 `examples/tasks/` 文件夹复制分层，不要在 handler 里写业务。

## 硬约定（易错）

1. **分层单向**：Handler/MCP → Service → Repository。Service 只实现一次；HTTP 与 MCP 都是适配器。
2. **错误**只用 `*httpx.Error`（`invalid`/`not_found`/`conflict`/`internal`…）；HTTP 与 MCP 共用 envelope：`{"ok":true,"data"}` / `{"ok":false,"error":{code,message,detail}}`。
3. **数据库 = 启动时连一种库**（`GOSILITE_DB_DSN`：`sqlite://` / `mysql://` / `postgres://`），换库换 DSN，不是迁库工具。`examples/*/migrations/*.sql` 是**应用 schema 版本**（启动幂等建表），不是 MySQL↔PG 数据迁移。
4. **SQL 占位符**必须经 `dbx.Dialect.Placeholder`（`?` vs `$n`）；schema 记账插入同样如此。DDL 用可移植类型（`VARCHAR`+默认值，勿 `TEXT DEFAULT`）。迁移文件可多语句（`;` 分隔）。
5. **`sqlite://:memory:` 会强制 `MaxOpenConns(1)`**——内存库按连接隔离，池化会“表不见了”。事务内只用 `tx.*`，不要另开 `db.Query`。
6. **MCP 只暴露业务工具**（`tasks_*`），禁止脚手架/运维类 tool。
7. **前端**依赖名是 `tdesign-vue-next` / `tdesign-icons-vue-next`（不是 `@tdesign/*`）；`web/src/types` 与后端 JSON 字段同名。`web/node_modules`、`web/dist` 已 gitignore。
8. 验收全绿才算完成：`go test ./... && go vet ./... && go build ./...` 且 `cd web && npm run build`。修 bug 先补回归测试。

## 环境

- Go 1.27+（`go.mod`）；Node 24+ 实测可用。国内网络 npm 可用 `--registry=https://registry.npmmirror.com`。
- 默认 DSN 是内存 SQLite；要持久化用 `sqlite://file:/path.db`。
- `DELETE` 成功返回 `204`；其余 2xx 走成功 envelope。
