# 测试模式（Testing Patterns）

目标：agent 改完业务后能快速确认行为，且测试不复制生产逻辑。

## 分层策略

| 层 | 测什么 | 怎么测 |
|----|--------|--------|
| Service | 业务规则、校验、not_found | 真 SQLite 仓储 + 固定时钟 |
| Repository | SQL、迁移、方言占位符 | `sqlite://:memory:` + 迁移 |
| HTTP | 状态码、envelope、路由 | `httptest` |
| MCP | tools/list、tools/call、错误 envelope | stdio JSON-RPC 报文 |
| 组装 | HTTP 与 MCP 共享同一 Service/数据 | 双通道交叉断言 |

## 原则

1. **表驱动**：同构输入输出用 `tests := []struct{...}`。
2. **少 mock**：优先真实 `database/sql` + SQLite；接口 mock 只在无法落地时使用。
3. **测公共行为**：断言状态码/错误 code/持久化结果，不重复实现生产算法。
4. **回归必补**：修 bug 先写能复现的失败测试，再改代码。
5. **禁止测试专用生产 API**；时钟等注入点用 `WithXxx` 且生产默认合理。

## 约定用例命名

- `TestServiceCRUD`、`TestHTTPCRUD`、`TestMCPBusinessToolsEndToEnd`
- 失败分支显式命名意图：`TestServiceInvalidStatus`

## 启动连库与 schema 版本注意

- 启动时只连**一种**库（读 DSN）；测试用 `sqlite://:memory:` 即可。
- schema 版本文件可多语句（`;` 分隔）；`dbx.Migrate` 切开后在所选库上逐条执行。
- 业务 SQL 与 schema 记账表的占位符都必须经当前 `Dialect.Placeholder`。
- `sqlite://:memory:` 强制单连接池，避免多连接各见各的内存库；并发场景用文件型 DSN。
- DDL 尽量可移植（`VARCHAR` + 默认值），避免 MySQL 禁止的 `TEXT DEFAULT`。
- 换数据库 = 换 DSN 重启；不提供、也不需要“把 MySQL 数据迁到 PostgreSQL”的工具。

## 验收命令

在仓库根目录：

```bash
go test ./...
go vet ./...
go build ./...
```

三条全绿才算可交付。失败先修代码或补测试，不跳过。

## MCP 测试要点

- `initialize` / `tools/list` / `tools/call` 各测一次。
- 业务错误必须映射为 `isError=true` 且 payload 含 `not_found` 等 code。
- 断言工具命名只包含业务前缀（如 `tasks_`），避免误暴露系统能力。
