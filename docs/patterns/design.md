# 设计模式（Design Patterns）

gosilite 面向 agent 高频生成与修改 Web 功能，因此模式数量刻意收敛到可重复、可校验的一组。

## 1. 分层 + 单向依赖

```text
HTTP Handler / MCP Tool  →  Service  →  Repository 接口  →  SQL 实现
         ↓                     ↓
     DTO / JSON            Domain + httpx.Error
```

- Handler/MCP 只做边界校验、DTO 绑定、错误映射，不写业务规则。
- Service 唯一持有业务规则与事务边界。
- Repository 接口由 Service 所在包定义；实现可替换（SQL / 内存）。
- 禁止 Handler 直接碰 `database/sql`。Repository 失败时返回 `*httpx.Error`（业务错误码），**不要**返回裸 `http.StatusXxx` 数字。

## 2. DTO 边界

- `CreateXRequest` / `UpdateXRequest` / `ListXQuery` 与 Domain 结构分离。
- JSON 字段名一经交付保持稳定；内部字段名可重构。
- 更新用指针字段表达「未传 / 显式置空」。

## 3. 统一错误模型

- 全系统只用 `*httpx.Error`，code 固定为：`invalid` / `unauthorized` / `forbidden` / `not_found` / `conflict` / `internal`。
- HTTP 与 MCP 共用同一 code 词表，agent 客户端只学一套语义。
- 成功 envelope：`{"ok":true,"data":...}`；失败：`{"ok":false,"error":{code,message,detail}}`。

## 4. Repository + Dialect（启动连库，非迁库）

- 项目启动时通过 `GOSILITE_DB_DSN`（或等价配置）**连接一种数据库**；换库 = 换 DSN。
- Repository 只写可移植 SQL；占位符由 `dbx.Dialect` 按**当前驱动**生成（`?` / `$n`）。
- DSN scheme 决定驱动：`sqlite://` / `mysql://` / `postgres://`。
- `migrations/NNNN_name.sql` 是**应用 schema 版本**（建表/索引），在所选库上启动时幂等应用；不是 MySQL↔PostgreSQL 之间的数据迁移工具。

## 5. 双通道业务能力（HTTP + MCP）

- 同一 Service 方法被两种适配器复用：
  - HTTP：REST/JSON，给用户与常规客户端。
  - MCP Tool：`tasks_*` 命名，给 agent 客户端。
- MCP **只暴露业务接口能力**，不暴露脚手架/测试/运维等系统工具。
- 新增能力时先写 Service 测试，再挂 HTTP 与 MCP，禁止两套业务实现。

## 6. 中间件链

顺序固定：`Recover → RequestID → Logging → mux`。

- Recover：panic → 500 envelope。
- RequestID：传播/生成 `X-Request-ID`。
- Logging：一行结构化访问日志。

## 7. 前端（Vue3 + TDesign）

- 目录：`web/`，栈固定为 **TypeScript + Vite + Vue3 + TDesign（`tdesign-vue-next`）**。
- 入口：`web/index.html` → `src/main.ts`；页面用 TDesign 组件，不引入额外 UI 体系。
- `src/types` 与后端 DTO JSON 字段同名（snake_case 保持一致）。
- `src/api` 只做 fetch + envelope 解包（`ok/data/error`）；业务规则仍在 Go Service。
- 开发代理：`/api` → 后端；生产可将 `web/dist` 交给任意静态服务或反代。

## 8. 构造器注入

- `NewService(repo)`、`NewHandler(svc)`、`NewSQLRepository(db)`。
- 时钟等环境因素用 `WithClock` 在测试中注入，生产不传。
- 不使用全局可变状态。
